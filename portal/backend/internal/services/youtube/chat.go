package youtube

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// chatPoller polls YouTube Live Chat API for new messages
type chatPoller struct {
	client      *Client
	liveChatID  string
	apiKey      string
	httpClient  *http.Client
	nextPageToken string
	pollingInterval time.Duration
}

func newChatPoller(client *Client, liveChatID, apiKey string) *chatPoller {
	return &chatPoller{
		client:          client,
		liveChatID:      liveChatID,
		apiKey:          apiKey,
		httpClient:      &http.Client{Timeout: 10 * time.Second},
		pollingInterval: 5 * time.Second, // Default, will be updated from API
	}
}

func (p *chatPoller) start(stop chan struct{}) {
	ticker := time.NewTicker(p.pollingInterval)
	defer ticker.Stop()

	// Initial fetch
	p.fetchMessages()

	for {
		select {
		case <-ticker.C:
			p.fetchMessages()
			ticker.Reset(p.pollingInterval)
		case <-stop:
			return
		}
	}
}

func (p *chatPoller) fetchMessages() {
	url := fmt.Sprintf("https://www.googleapis.com/youtube/v3/liveChat/messages?liveChatId=%s&part=snippet,authorDetails&key=%s",
		p.liveChatID, p.apiKey)

	if p.nextPageToken != "" {
		url += fmt.Sprintf("&pageToken=%s", p.nextPageToken)
	}

	resp, err := p.httpClient.Get(url)
	if err != nil {
		fmt.Printf("Error fetching chat messages: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("YouTube Chat API error: %d\n", resp.StatusCode)
		return
	}

	var data liveChatMessagesResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		fmt.Printf("Error decoding chat messages: %v\n", err)
		return
	}

	// Update pagination and polling interval
	p.nextPageToken = data.NextPageToken
	if data.PollingIntervalMillis > 0 {
		p.pollingInterval = time.Duration(data.PollingIntervalMillis) * time.Millisecond
	}

	// Process new messages
	for _, item := range data.Items {
		// Only process text messages
		if item.Snippet.Type != "textMessageEvent" {
			continue
		}

		// Extract message text
		messageText := item.Snippet.DisplayMessage
		if item.Snippet.TextMessageDetails != nil {
			messageText = item.Snippet.TextMessageDetails.MessageText
		}

		chatMsg := &ChatMessage{
			ID:              item.ID,
			AuthorName:      item.AuthorDetails.DisplayName,
			AuthorChannelID: item.AuthorDetails.ChannelID,
			Message:         messageText,
			Timestamp:       item.Snippet.PublishedAt,
			IsModerator:     item.AuthorDetails.IsChatModerator,
			IsSponsor:       item.AuthorDetails.IsChatSponsor,
			IsOwner:         item.AuthorDetails.IsChatOwner,
		}

		// Send to callback
		p.client.mu.RLock()
		callback := p.client.eventCallback
		p.client.mu.RUnlock()

		if callback != nil {
			callback(map[string]interface{}{
				"type": "chat_message",
				"data": chatMsg,
			})
		}
	}
}
