package youtube

import "time"

// YouTubeStatus represents the connection status to YouTube
type YouTubeStatus struct {
	Connected     bool   `json:"connected"`
	LiveChatReady bool   `json:"live_chat_ready"`
	ChannelID     string `json:"channel_id,omitempty"`
	BroadcastID   string `json:"broadcast_id,omitempty"`
}

// StreamStats represents YouTube live stream statistics
type StreamStats struct {
	IsLive            bool      `json:"is_live"`
	ViewerCount       int       `json:"viewer_count"`
	SubscriberCount   int       `json:"subscriber_count"`
	VideoCount        int       `json:"video_count"`
	Title             string    `json:"title,omitempty"`
	Description       string    `json:"description,omitempty"`
	ScheduledStartTime *time.Time `json:"scheduled_start_time,omitempty"`
	ActualStartTime    *time.Time `json:"actual_start_time,omitempty"`
}

// ChatMessage represents a YouTube live chat message
type ChatMessage struct {
	ID              string    `json:"id"`
	AuthorName      string    `json:"author_name"`
	AuthorChannelID string    `json:"author_channel_id"`
	Message         string    `json:"message"`
	Timestamp       time.Time `json:"timestamp"`
	IsModerator     bool      `json:"is_moderator"`
	IsSponsor       bool      `json:"is_sponsor"`
	IsOwner         bool      `json:"is_owner"`
}

// UpdateStreamRequest represents a request to update stream metadata
type UpdateStreamRequest struct {
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	CategoryID  string `json:"category_id,omitempty"`
}

// Internal YouTube API Response types

// channelResponse represents the YouTube Data API channel response
type channelResponse struct {
	Items []struct {
		ID      string `json:"id"`
		Snippet struct {
			Title string `json:"title"`
		} `json:"snippet"`
		Statistics struct {
			SubscriberCount string `json:"subscriberCount"`
			VideoCount      string `json:"videoCount"`
		} `json:"statistics"`
	} `json:"items"`
}

// liveBroadcastResponse represents the YouTube Data API liveBroadcasts response
type liveBroadcastResponse struct {
	Items []struct {
		ID      string `json:"id"`
		Snippet struct {
			Title              string    `json:"title"`
			Description        string    `json:"description"`
			ScheduledStartTime time.Time `json:"scheduledStartTime"`
			ActualStartTime    time.Time `json:"actualStartTime,omitempty"`
			LiveChatID         string    `json:"liveChatId"`
		} `json:"snippet"`
		Status struct {
			LifeCycleStatus string `json:"lifeCycleStatus"` // "live", "upcoming", "complete"
		} `json:"status"`
		Statistics struct {
			ConcurrentViewers string `json:"concurrentViewers,omitempty"`
		} `json:"statistics,omitempty"`
	} `json:"items"`
}

// liveChatMessagesResponse represents the YouTube Data API liveChatMessages response
type liveChatMessagesResponse struct {
	NextPageToken  string `json:"nextPageToken"`
	PollingIntervalMillis int `json:"pollingIntervalMillis"`
	Items          []struct {
		ID      string `json:"id"`
		Snippet struct {
			Type               string    `json:"type"`
			LiveChatID         string    `json:"liveChatId"`
			AuthorChannelID    string    `json:"authorChannelId"`
			PublishedAt        time.Time `json:"publishedAt"`
			HasDisplayContent  bool      `json:"hasDisplayContent"`
			DisplayMessage     string    `json:"displayMessage"`
			TextMessageDetails *struct {
				MessageText string `json:"messageText"`
			} `json:"textMessageDetails,omitempty"`
		} `json:"snippet"`
		AuthorDetails struct {
			ChannelID       string `json:"channelId"`
			DisplayName     string `json:"displayName"`
			IsChatOwner     bool   `json:"isChatOwner"`
			IsChatModerator bool   `json:"isChatModerator"`
			IsChatSponsor   bool   `json:"isChatSponsor"`
		} `json:"authorDetails"`
	} `json:"items"`
}
