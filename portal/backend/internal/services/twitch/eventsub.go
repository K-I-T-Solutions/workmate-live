package twitch

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	eventSubWebSocketURL = "wss://eventsub.wss.twitch.tv/ws"
)

// EventSubClient handles Twitch EventSub via WebSocket
type EventSubClient struct {
	clientID   string
	oauthToken string
	userID     string
	mainClient *Client // Reference to main client for subscription creation

	conn      *websocket.Conn
	sessionID string

	connected     bool
	eventCallback func(*EventSubEvent)

	mu       sync.RWMutex
	stopChan chan struct{}
	done     chan struct{}
}

// EventSub message types
type eventSubMessage struct {
	Metadata eventSubMetadata `json:"metadata"`
	Payload  json.RawMessage  `json:"payload"`
}

type eventSubMetadata struct {
	MessageID           string    `json:"message_id"`
	MessageType         string    `json:"message_type"`
	MessageTimestamp    time.Time `json:"message_timestamp"`
	SubscriptionType    string    `json:"subscription_type,omitempty"`
	SubscriptionVersion string    `json:"subscription_version,omitempty"`
}

type eventSubWelcome struct {
	Session eventSubSession `json:"session"`
}

type eventSubSession struct {
	ID                      string    `json:"id"`
	Status                  string    `json:"status"`
	KeepaliveTimeoutSeconds int       `json:"keepalive_timeout_seconds"`
	ReconnectURL            string    `json:"reconnect_url,omitempty"`
	ConnectedAt             time.Time `json:"connected_at"`
}

type eventSubNotification struct {
	Subscription eventSubSubscription `json:"subscription"`
	Event        json.RawMessage      `json:"event"`
}

type eventSubSubscription struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Version string `json:"version"`
}

type eventSubReconnect struct {
	Session eventSubSession `json:"session"`
}

// NewEventSubClient creates a new EventSub client
func NewEventSubClient(clientID, oauthToken, userID string, mainClient *Client) *EventSubClient {
	return &EventSubClient{
		clientID:   clientID,
		oauthToken: oauthToken,
		userID:     userID,
		mainClient: mainClient,
		stopChan:   make(chan struct{}),
		done:       make(chan struct{}),
	}
}

// Connect establishes connection to EventSub WebSocket
func (c *EventSubClient) Connect() error {
	conn, _, err := websocket.DefaultDialer.Dial(eventSubWebSocketURL, nil)
	if err != nil {
		return fmt.Errorf("failed to dial EventSub WebSocket: %w", err)
	}

	c.conn = conn

	c.mu.Lock()
	c.connected = true
	c.mu.Unlock()

	// Start reading messages in background
	go c.readMessages()

	return nil
}

// Disconnect closes the EventSub connection
func (c *EventSubClient) Disconnect() error {
	c.mu.Lock()
	c.connected = false
	c.mu.Unlock()

	close(c.stopChan)

	if c.conn != nil {
		_ = c.conn.Close()
	}

	// Wait for goroutines to finish
	<-c.done

	return nil
}

// IsConnected returns whether the client is connected
func (c *EventSubClient) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connected
}

// SetEventCallback sets the callback for incoming events
func (c *EventSubClient) SetEventCallback(callback func(*EventSubEvent)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.eventCallback = callback
}

// readMessages reads messages from the EventSub WebSocket
func (c *EventSubClient) readMessages() {
	defer close(c.done)

	for {
		select {
		case <-c.stopChan:
			return
		default:
		}

		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if c.IsConnected() {
				log.Printf("Twitch EventSub: Read error: %v", err)
			}
			return
		}

		c.handleMessage(message)
	}
}

// handleMessage processes an EventSub message
func (c *EventSubClient) handleMessage(data []byte) {
	var msg eventSubMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		log.Printf("Twitch EventSub: Failed to unmarshal message: %v", err)
		return
	}

	switch msg.Metadata.MessageType {
	case "session_welcome":
		c.handleWelcome(msg.Payload)
	case "session_keepalive":
		// No action needed, just acknowledging keepalive
	case "notification":
		c.handleNotification(msg.Payload)
	case "session_reconnect":
		c.handleReconnect(msg.Payload)
	case "revocation":
		log.Printf("Twitch EventSub: Subscription revoked")
	default:
		log.Printf("Twitch EventSub: Unknown message type: %s", msg.Metadata.MessageType)
	}
}

// handleWelcome handles the session_welcome message
func (c *EventSubClient) handleWelcome(payload []byte) {
	var welcome eventSubWelcome
	if err := json.Unmarshal(payload, &welcome); err != nil {
		log.Printf("Twitch EventSub: Failed to unmarshal welcome: %v", err)
		return
	}

	c.mu.Lock()
	c.sessionID = welcome.Session.ID
	c.mu.Unlock()

	log.Printf("Twitch EventSub: Session established (ID: %s)", welcome.Session.ID)

	// Subscribe to events
	go c.subscribeToEvents()
}

// handleNotification handles event notifications
func (c *EventSubClient) handleNotification(payload []byte) {
	var notification eventSubNotification
	if err := json.Unmarshal(payload, &notification); err != nil {
		log.Printf("Twitch EventSub: Failed to unmarshal notification: %v", err)
		return
	}

	// Route event based on type
	var event *EventSubEvent
	switch notification.Subscription.Type {
	case "channel.follow":
		var followData FollowEvent
		if err := json.Unmarshal(notification.Event, &followData); err != nil {
			log.Printf("Twitch EventSub: Failed to unmarshal follow event: %v", err)
			return
		}
		event = &EventSubEvent{
			Type:      "follow",
			Data:      followData,
			Timestamp: time.Now().Format(time.RFC3339),
		}

	case "channel.subscribe":
		var subData SubscribeEvent
		if err := json.Unmarshal(notification.Event, &subData); err != nil {
			log.Printf("Twitch EventSub: Failed to unmarshal subscribe event: %v", err)
			return
		}
		event = &EventSubEvent{
			Type:      "subscribe",
			Data:      subData,
			Timestamp: time.Now().Format(time.RFC3339),
		}

	case "channel.raid":
		var raidData RaidEvent
		if err := json.Unmarshal(notification.Event, &raidData); err != nil {
			log.Printf("Twitch EventSub: Failed to unmarshal raid event: %v", err)
			return
		}
		event = &EventSubEvent{
			Type:      "raid",
			Data:      raidData,
			Timestamp: time.Now().Format(time.RFC3339),
		}

	default:
		log.Printf("Twitch EventSub: Unhandled event type: %s", notification.Subscription.Type)
		return
	}

	// Call event callback
	c.mu.RLock()
	callback := c.eventCallback
	c.mu.RUnlock()

	if callback != nil {
		callback(event)
	}
}

// handleReconnect handles the session_reconnect message
func (c *EventSubClient) handleReconnect(payload []byte) {
	var reconnect eventSubReconnect
	if err := json.Unmarshal(payload, &reconnect); err != nil {
		log.Printf("Twitch EventSub: Failed to unmarshal reconnect: %v", err)
		return
	}

	log.Printf("Twitch EventSub: Server requested reconnect to: %s", reconnect.Session.ReconnectURL)
	// TODO: Implement seamless reconnection
}

// subscribeToEvents subscribes to all desired event types
func (c *EventSubClient) subscribeToEvents() {
	c.mu.RLock()
	sessionID := c.sessionID
	c.mu.RUnlock()

	if sessionID == "" {
		log.Printf("Twitch EventSub: Cannot subscribe without session ID")
		return
	}

	// Subscribe to channel.follow (v2)
	if err := c.mainClient.CreateEventSubSubscription("channel.follow", "2", sessionID); err != nil {
		log.Printf("Twitch EventSub: Failed to subscribe to channel.follow: %v", err)
	}

	// Subscribe to channel.subscribe (v1)
	if err := c.mainClient.CreateEventSubSubscription("channel.subscribe", "1", sessionID); err != nil {
		log.Printf("Twitch EventSub: Failed to subscribe to channel.subscribe: %v", err)
	}

	// Subscribe to channel.raid (v1)
	if err := c.mainClient.CreateEventSubSubscription("channel.raid", "1", sessionID); err != nil {
		log.Printf("Twitch EventSub: Failed to subscribe to channel.raid: %v", err)
	}

	log.Println("Twitch EventSub: Subscribed to all events")
}
