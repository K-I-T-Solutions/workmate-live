package twitch

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

const (
	helixBaseURL = "https://api.twitch.tv/helix"
)

// Client wraps the Twitch Helix API and WebSocket clients
type Client struct {
	clientID     string
	clientSecret string
	channel      string
	oauthToken   string
	httpClient   *http.Client

	userID   string
	userName string

	connected       bool
	chatConnected   bool
	eventsConnected bool

	chatWS     *ChatClient
	eventSubWS *EventSubClient

	eventCallback func(interface{})
	mu            sync.RWMutex
}

// NewClient creates a new Twitch client
func NewClient(clientID, clientSecret, channel, oauthToken string) *Client {
	return &Client{
		clientID:     clientID,
		clientSecret: clientSecret,
		channel:      channel,
		oauthToken:   oauthToken,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Connect initializes all Twitch connections (HTTP API, Chat, EventSub)
func (c *Client) Connect() error {
	// First, fetch user info to get UserID
	if err := c.fetchUserInfo(); err != nil {
		return fmt.Errorf("failed to fetch user info: %w", err)
	}

	log.Printf("Twitch: Resolved channel '%s' to user ID '%s'", c.channel, c.userID)

	c.mu.Lock()
	c.connected = true
	c.mu.Unlock()

	// Initialize chat client
	c.chatWS = NewChatClient(c.channel, c.oauthToken)
	c.chatWS.SetMessageCallback(c.handleChatMessage)

	if err := c.chatWS.Connect(); err != nil {
		log.Printf("Warning: Failed to connect to Twitch chat: %v", err)
	} else {
		c.mu.Lock()
		c.chatConnected = true
		c.mu.Unlock()
		log.Println("Twitch: Chat connected")
	}

	// Initialize EventSub client
	c.eventSubWS = NewEventSubClient(c.clientID, c.oauthToken, c.userID, c)
	c.eventSubWS.SetEventCallback(c.handleEventSubEvent)

	if err := c.eventSubWS.Connect(); err != nil {
		log.Printf("Warning: Failed to connect to Twitch EventSub: %v", err)
	} else {
		c.mu.Lock()
		c.eventsConnected = true
		c.mu.Unlock()
		log.Println("Twitch: EventSub connected")
	}

	return nil
}

// Disconnect closes all Twitch connections
func (c *Client) Disconnect() error {
	c.mu.Lock()
	c.connected = false
	c.chatConnected = false
	c.eventsConnected = false
	c.mu.Unlock()

	var errors []error

	if c.chatWS != nil {
		if err := c.chatWS.Disconnect(); err != nil {
			errors = append(errors, fmt.Errorf("chat disconnect: %w", err))
		}
	}

	if c.eventSubWS != nil {
		if err := c.eventSubWS.Disconnect(); err != nil {
			errors = append(errors, fmt.Errorf("eventsub disconnect: %w", err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("disconnect errors: %v", errors)
	}

	return nil
}

// IsConnected returns whether the client is connected
func (c *Client) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connected
}

// SetEventCallback sets the callback function for events
func (c *Client) SetEventCallback(callback func(interface{})) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.eventCallback = callback
}

// GetStatus returns the current Twitch connection status
func (c *Client) GetStatus() (*TwitchStatus, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return &TwitchStatus{
		Connected:       c.connected,
		ChatConnected:   c.chatConnected,
		EventsConnected: c.eventsConnected,
		Channel:         c.channel,
		UserID:          c.userID,
	}, nil
}

// GetStreamStats returns aggregated stream statistics
func (c *Client) GetStreamStats() (*StreamStats, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("not connected to Twitch")
	}

	stats := &StreamStats{
		IsLive:        false,
		ViewerCount:   0,
		FollowerCount: 0,
		Uptime:        0,
		Title:         "",
		GameName:      "",
	}

	// Get stream info
	stream, err := c.getStream()
	if err != nil {
		log.Printf("Warning: Failed to get stream info: %v", err)
	} else if stream != nil {
		stats.IsLive = true
		stats.ViewerCount = stream.ViewerCount
		stats.Title = stream.Title
		stats.GameName = stream.GameName
		stats.StartedAt = stream.StartedAt
		stats.Language = stream.Language
		stats.ThumbnailURL = stream.ThumbnailURL
		stats.StreamID = stream.ID

		// Calculate uptime
		startedAt, err := time.Parse(time.RFC3339, stream.StartedAt)
		if err == nil {
			stats.Uptime = int64(time.Since(startedAt).Seconds())
		}
	}

	// Get follower count
	followerCount, err := c.getFollowerCount()
	if err != nil {
		log.Printf("Warning: Failed to get follower count: %v", err)
	} else {
		stats.FollowerCount = followerCount
	}

	return stats, nil
}

// UpdateStreamMetadata updates the stream title and/or game
func (c *Client) UpdateStreamMetadata(req *UpdateStreamRequest) error {
	if !c.IsConnected() {
		return fmt.Errorf("not connected to Twitch")
	}

	if req.Title == "" && req.GameID == "" && req.GameName == "" {
		return fmt.Errorf("at least one field required")
	}

	gameID := req.GameID

	// If GameName is provided but not GameID, search for the game
	if req.GameName != "" && gameID == "" {
		game, err := c.searchGame(req.GameName)
		if err != nil {
			return fmt.Errorf("failed to search game: %w", err)
		}
		gameID = game.ID
	}

	return c.updateChannelInfo(req.Title, gameID)
}

// CreateEventSubSubscription creates an EventSub subscription
func (c *Client) CreateEventSubSubscription(eventType, version, sessionID string) error {
	if !c.IsConnected() {
		return fmt.Errorf("not connected to Twitch")
	}

	// Build condition based on event type
	condition := make(map[string]string)
	switch eventType {
	case "channel.raid":
		// For raids, we want to know when someone raids US (incoming raids)
		condition["to_broadcaster_user_id"] = c.userID
	case "channel.follow":
		// channel.follow v2 requires both broadcaster_user_id and moderator_user_id
		condition["broadcaster_user_id"] = c.userID
		condition["moderator_user_id"] = c.userID
	default:
		// Default: just broadcaster_user_id
		condition["broadcaster_user_id"] = c.userID
	}

	body := map[string]interface{}{
		"type":      eventType,
		"version":   version,
		"condition": condition,
		"transport": map[string]string{
			"method":     "websocket",
			"session_id": sessionID,
		},
	}

	resp, err := c.makeRequest("POST", "/eventsub/subscriptions", body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusOK {
		bodyData, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to create subscription (status %d): %s", resp.StatusCode, string(bodyData))
	}

	log.Printf("Twitch: Created EventSub subscription: %s (v%s)", eventType, version)
	return nil
}

// Private methods

// fetchUserInfo fetches user information from the Twitch API
func (c *Client) fetchUserInfo() error {
	resp, err := c.makeRequest("GET", "/users?login="+c.channel, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result struct {
		Data []helixUser `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if len(result.Data) == 0 {
		return fmt.Errorf("channel not found: %s", c.channel)
	}

	c.userID = result.Data[0].ID
	c.userName = result.Data[0].DisplayName
	return nil
}

// getStream fetches current stream information
func (c *Client) getStream() (*helixStream, error) {
	resp, err := c.makeRequest("GET", "/streams?user_id="+c.userID, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result struct {
		Data []helixStream `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(result.Data) == 0 {
		return nil, nil // Stream is offline
	}

	return &result.Data[0], nil
}

// getFollowerCount fetches the total follower count
func (c *Client) getFollowerCount() (int, error) {
	resp, err := c.makeRequest("GET", "/channels/followers?broadcaster_id="+c.userID, nil)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result helixFollowers

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Total, nil
}

// searchGame searches for a game by name
func (c *Client) searchGame(name string) (*helixGame, error) {
	resp, err := c.makeRequest("GET", "/games?name="+name, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result struct {
		Data []helixGame `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(result.Data) == 0 {
		return nil, fmt.Errorf("game not found: %s", name)
	}

	return &result.Data[0], nil
}

// updateChannelInfo updates the channel's title and/or game
func (c *Client) updateChannelInfo(title, gameID string) error {
	body := make(map[string]interface{})

	if title != "" {
		body["title"] = title
	}
	if gameID != "" {
		body["game_id"] = gameID
	}

	resp, err := c.makeRequest("PATCH", "/channels?broadcaster_id="+c.userID, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		bodyData, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to update channel (status %d): %s", resp.StatusCode, string(bodyData))
	}

	log.Printf("Twitch: Updated stream metadata (title: %s, game_id: %s)", title, gameID)
	return nil
}

// makeRequest makes an HTTP request to the Twitch Helix API
func (c *Client) makeRequest(method, endpoint string, body interface{}) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, helixBaseURL+endpoint, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.oauthToken)
	req.Header.Set("Client-ID", c.clientID)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	// Check for authentication errors
	if resp.StatusCode == http.StatusUnauthorized {
		return resp, fmt.Errorf("unauthorized: OAuth token may be expired or invalid")
	}

	return resp, nil
}

// Event handlers

// handleChatMessage handles chat messages from the IRC client
func (c *Client) handleChatMessage(msg *ChatMessage) {
	c.mu.RLock()
	callback := c.eventCallback
	c.mu.RUnlock()

	if callback != nil {
		callback(map[string]interface{}{
			"type": "chat_message",
			"data": msg,
		})
	}
}

// handleEventSubEvent handles events from the EventSub client
func (c *Client) handleEventSubEvent(event *EventSubEvent) {
	c.mu.RLock()
	callback := c.eventCallback
	c.mu.RUnlock()

	if callback != nil {
		callback(map[string]interface{}{
			"type": "eventsub_event",
			"data": event,
		})
	}
}
