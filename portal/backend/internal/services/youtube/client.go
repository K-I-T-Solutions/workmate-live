package youtube

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// Client handles YouTube Data API interactions
type Client struct {
	apiKey       string
	channelID    string
	clientID     string
	clientSecret string

	httpClient *http.Client

	// State
	mu            sync.RWMutex
	connected     bool
	liveChatReady bool
	broadcastID   string
	liveChatID    string

	// Event callback for chat messages
	eventCallback func(interface{})

	// Chat polling
	chatPoller     *chatPoller
	stopChatPoller chan struct{}
}

// NewClient creates a new YouTube client
func NewClient(apiKey, channelID, clientID, clientSecret string) *Client {
	return &Client{
		apiKey:       apiKey,
		channelID:    channelID,
		clientID:     clientID,
		clientSecret: clientSecret,
		httpClient:   &http.Client{Timeout: 10 * time.Second},
	}
}

// Connect initializes the YouTube client and fetches channel info
func (c *Client) Connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Verify channel exists
	_, err := c.getChannelInfo()
	if err != nil {
		return fmt.Errorf("failed to get channel info: %w", err)
	}

	c.connected = true

	// Try to find active live broadcast
	if err := c.findActiveBroadcast(); err != nil {
		// Not an error if no active broadcast
		c.liveChatReady = false
	} else {
		c.liveChatReady = true
		// Start chat polling if we have an active broadcast
		c.startChatPolling()
	}

	return nil
}

// Disconnect stops all YouTube operations
func (c *Client) Disconnect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.stopChatPolling()

	c.connected = false
	c.liveChatReady = false
	c.broadcastID = ""
	c.liveChatID = ""

	return nil
}

// GetStatus returns the current YouTube connection status
func (c *Client) GetStatus() *YouTubeStatus {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return &YouTubeStatus{
		Connected:     c.connected,
		LiveChatReady: c.liveChatReady,
		ChannelID:     c.channelID,
		BroadcastID:   c.broadcastID,
	}
}

// GetStats fetches current stream statistics
func (c *Client) GetStats() (*StreamStats, error) {
	c.mu.RLock()
	channelID := c.channelID
	c.mu.RUnlock()

	if channelID == "" {
		return nil, fmt.Errorf("channel ID not set")
	}

	// Get channel info (subscribers, video count)
	channelInfo, err := c.getChannelInfo()
	if err != nil {
		return nil, err
	}

	stats := &StreamStats{
		SubscriberCount: channelInfo.SubscriberCount,
		VideoCount:      channelInfo.VideoCount,
	}

	// Get live broadcast info if available
	c.mu.RLock()
	broadcastID := c.broadcastID
	c.mu.RUnlock()

	if broadcastID != "" {
		broadcast, err := c.getBroadcastInfo(broadcastID)
		if err == nil {
			stats.IsLive = broadcast.IsLive
			stats.ViewerCount = broadcast.ViewerCount
			stats.Title = broadcast.Title
			stats.Description = broadcast.Description
			stats.ScheduledStartTime = broadcast.ScheduledStartTime
			stats.ActualStartTime = broadcast.ActualStartTime
		}
	}

	return stats, nil
}

// UpdateStream updates the live stream metadata
func (c *Client) UpdateStream(req *UpdateStreamRequest) error {
	c.mu.RLock()
	broadcastID := c.broadcastID
	c.mu.RUnlock()

	if broadcastID == "" {
		return fmt.Errorf("no active broadcast")
	}

	// YouTube Data API v3: liveBroadcasts.update would be called here
	// URL: https://www.googleapis.com/youtube/v3/liveBroadcasts?part=snippet&id=%s&key=%s
	// This would require OAuth token, simplified here
	// In production, use OAuth2 token instead of API key
	_ = broadcastID // Used for future OAuth implementation
	return fmt.Errorf("update stream requires OAuth token (not implemented with API key only)")
}

// SetEventCallback sets the callback for YouTube events (chat messages)
func (c *Client) SetEventCallback(callback func(interface{})) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.eventCallback = callback
}

// Internal helper methods

func (c *Client) getChannelInfo() (*struct {
	SubscriberCount int
	VideoCount      int
}, error) {
	url := fmt.Sprintf("https://www.googleapis.com/youtube/v3/channels?part=statistics&id=%s&key=%s",
		c.channelID, c.apiKey)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("YouTube API error: %d", resp.StatusCode)
	}

	var data channelResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	if len(data.Items) == 0 {
		return nil, fmt.Errorf("channel not found")
	}

	subCount, _ := strconv.Atoi(data.Items[0].Statistics.SubscriberCount)
	vidCount, _ := strconv.Atoi(data.Items[0].Statistics.VideoCount)

	return &struct {
		SubscriberCount int
		VideoCount      int
	}{
		SubscriberCount: subCount,
		VideoCount:      vidCount,
	}, nil
}

func (c *Client) findActiveBroadcast() error {
	url := fmt.Sprintf("https://www.googleapis.com/youtube/v3/liveBroadcasts?part=snippet,status,statistics&broadcastStatus=active&key=%s",
		c.apiKey)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var data liveBroadcastResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return err
	}

	if len(data.Items) == 0 {
		return fmt.Errorf("no active broadcast")
	}

	broadcast := data.Items[0]
	c.broadcastID = broadcast.ID
	c.liveChatID = broadcast.Snippet.LiveChatID

	return nil
}

func (c *Client) getBroadcastInfo(broadcastID string) (*struct {
	IsLive             bool
	ViewerCount        int
	Title              string
	Description        string
	ScheduledStartTime *time.Time
	ActualStartTime    *time.Time
}, error) {
	url := fmt.Sprintf("https://www.googleapis.com/youtube/v3/liveBroadcasts?part=snippet,status,statistics&id=%s&key=%s",
		broadcastID, c.apiKey)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data liveBroadcastResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	if len(data.Items) == 0 {
		return nil, fmt.Errorf("broadcast not found")
	}

	broadcast := data.Items[0]
	viewerCount, _ := strconv.Atoi(broadcast.Statistics.ConcurrentViewers)

	result := &struct {
		IsLive             bool
		ViewerCount        int
		Title              string
		Description        string
		ScheduledStartTime *time.Time
		ActualStartTime    *time.Time
	}{
		IsLive:             broadcast.Status.LifeCycleStatus == "live",
		ViewerCount:        viewerCount,
		Title:              broadcast.Snippet.Title,
		Description:        broadcast.Snippet.Description,
		ScheduledStartTime: &broadcast.Snippet.ScheduledStartTime,
	}

	if !broadcast.Snippet.ActualStartTime.IsZero() {
		result.ActualStartTime = &broadcast.Snippet.ActualStartTime
	}

	return result, nil
}

func (c *Client) startChatPolling() {
	if c.liveChatID == "" {
		return
	}

	c.stopChatPoller = make(chan struct{})
	c.chatPoller = newChatPoller(c, c.liveChatID, c.apiKey)
	go c.chatPoller.start(c.stopChatPoller)
}

func (c *Client) stopChatPolling() {
	if c.stopChatPoller != nil {
		close(c.stopChatPoller)
		c.stopChatPoller = nil
	}
	c.chatPoller = nil
}
