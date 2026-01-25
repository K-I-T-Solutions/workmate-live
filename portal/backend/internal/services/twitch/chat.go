package twitch

import (
	"bufio"
	"fmt"
	"log"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	ircWebSocketURL = "wss://irc-ws.chat.twitch.tv:443"
	pingInterval    = 5 * time.Minute
)

// ChatClient handles Twitch IRC chat via WebSocket
type ChatClient struct {
	channel   string
	oauthToken string
	conn      *websocket.Conn

	connected       bool
	messageCallback func(*ChatMessage)

	mu       sync.RWMutex
	stopChan chan struct{}
	done     chan struct{}
}

// NewChatClient creates a new Twitch chat client
func NewChatClient(channel, oauthToken string) *ChatClient {
	return &ChatClient{
		channel:    channel,
		oauthToken: oauthToken,
		stopChan:   make(chan struct{}),
		done:       make(chan struct{}),
	}
}

// Connect establishes connection to Twitch IRC
func (c *ChatClient) Connect() error {
	conn, _, err := websocket.DefaultDialer.Dial(ircWebSocketURL, nil)
	if err != nil {
		return fmt.Errorf("failed to dial IRC WebSocket: %w", err)
	}

	c.conn = conn

	// Authenticate with Twitch IRC
	if err := c.authenticate(); err != nil {
		c.conn.Close()
		return fmt.Errorf("authentication failed: %w", err)
	}

	c.mu.Lock()
	c.connected = true
	c.mu.Unlock()

	// Start reading messages in background
	go c.readMessages()
	go c.pingHandler()

	return nil
}

// Disconnect closes the IRC connection
func (c *ChatClient) Disconnect() error {
	c.mu.Lock()
	c.connected = false
	c.mu.Unlock()

	close(c.stopChan)

	if c.conn != nil {
		// Send PART command to leave the channel
		_ = c.sendRaw("PART #" + c.channel)
		_ = c.conn.Close()
	}

	// Wait for goroutines to finish
	<-c.done

	return nil
}

// IsConnected returns whether the client is connected
func (c *ChatClient) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connected
}

// SetMessageCallback sets the callback for incoming messages
func (c *ChatClient) SetMessageCallback(callback func(*ChatMessage)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.messageCallback = callback
}

// authenticate sends the IRC authentication sequence
func (c *ChatClient) authenticate() error {
	// Request IRCv3 capabilities for rich message metadata
	if err := c.sendRaw("CAP REQ :twitch.tv/tags twitch.tv/commands"); err != nil {
		return err
	}

	// Authenticate with OAuth token
	if err := c.sendRaw("PASS oauth:" + c.oauthToken); err != nil {
		return err
	}

	// Set nickname to channel name
	if err := c.sendRaw("NICK " + c.channel); err != nil {
		return err
	}

	// Join the channel
	if err := c.sendRaw("JOIN #" + c.channel); err != nil {
		return err
	}

	log.Printf("Twitch IRC: Authenticated and joined #%s", c.channel)
	return nil
}

// readMessages reads messages from the IRC WebSocket
func (c *ChatClient) readMessages() {
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
				log.Printf("Twitch IRC: Read error: %v", err)
			}
			return
		}

		// Handle each line in the message
		scanner := bufio.NewScanner(strings.NewReader(string(message)))
		for scanner.Scan() {
			line := scanner.Text()
			c.handleMessage(line)
		}
	}
}

// handleMessage processes a single IRC message
func (c *ChatClient) handleMessage(line string) {
	// Respond to PING with PONG
	if strings.HasPrefix(line, "PING") {
		_ = c.sendRaw("PONG :tmi.twitch.tv")
		return
	}

	// Parse PRIVMSG (chat messages)
	if strings.Contains(line, "PRIVMSG") {
		msg := c.parsePrivMsg(line)
		if msg != nil {
			c.mu.RLock()
			callback := c.messageCallback
			c.mu.RUnlock()

			if callback != nil {
				callback(msg)
			}
		}
	}
}

// parsePrivMsg parses an IRC PRIVMSG with IRCv3 tags
func (c *ChatClient) parsePrivMsg(line string) *ChatMessage {
	// Example format:
	// @badge-info=;badges=moderator/1;color=#FF0000;display-name=Username;mod=1;subscriber=0 :username!username@username.tmi.twitch.tv PRIVMSG #channel :Hello world

	// Split tags from the rest
	var tags string
	var rest string
	if strings.HasPrefix(line, "@") {
		parts := strings.SplitN(line, " ", 2)
		if len(parts) == 2 {
			tags = parts[0]
			rest = parts[1]
		}
	} else {
		rest = line
	}

	// Parse tags
	tagMap := make(map[string]string)
	if tags != "" {
		tags = strings.TrimPrefix(tags, "@")
		for _, tag := range strings.Split(tags, ";") {
			kv := strings.SplitN(tag, "=", 2)
			if len(kv) == 2 {
				tagMap[kv[0]] = kv[1]
			}
		}
	}

	// Extract username from IRC prefix
	usernameRegex := regexp.MustCompile(`:([^!]+)!`)
	usernameMatches := usernameRegex.FindStringSubmatch(rest)
	if len(usernameMatches) < 2 {
		return nil
	}
	username := usernameMatches[1]

	// Extract message content
	messageRegex := regexp.MustCompile(`PRIVMSG #[^ ]+ :(.+)$`)
	messageMatches := messageRegex.FindStringSubmatch(rest)
	if len(messageMatches) < 2 {
		return nil
	}
	message := messageMatches[1]

	// Build ChatMessage
	chatMsg := &ChatMessage{
		Username:     username,
		DisplayName:  tagMap["display-name"],
		Message:      message,
		Color:        tagMap["color"],
		Timestamp:    time.Now().Format(time.RFC3339),
		IsModerator:  tagMap["mod"] == "1",
		IsSubscriber: tagMap["subscriber"] == "1",
		Badges:       []string{},
	}

	// Parse badges
	if badgesStr := tagMap["badges"]; badgesStr != "" {
		for _, badge := range strings.Split(badgesStr, ",") {
			badgeName := strings.Split(badge, "/")[0]
			chatMsg.Badges = append(chatMsg.Badges, badgeName)
		}
	}

	// Use username if display name is not set
	if chatMsg.DisplayName == "" {
		chatMsg.DisplayName = username
	}

	return chatMsg
}

// pingHandler sends periodic PING messages to keep the connection alive
func (c *ChatClient) pingHandler() {
	ticker := time.NewTicker(pingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-c.stopChan:
			return
		case <-ticker.C:
			if c.IsConnected() {
				_ = c.sendRaw("PING :tmi.twitch.tv")
			}
		}
	}
}

// sendRaw sends a raw IRC command
func (c *ChatClient) sendRaw(command string) error {
	if c.conn == nil {
		return fmt.Errorf("not connected")
	}

	return c.conn.WriteMessage(websocket.TextMessage, []byte(command+"\r\n"))
}
