package twitch

// TwitchStatus represents the connection status of the Twitch client
type TwitchStatus struct {
	Connected       bool   `json:"connected"`
	ChatConnected   bool   `json:"chat_connected"`
	EventsConnected bool   `json:"events_connected"`
	Channel         string `json:"channel"`
	UserID          string `json:"user_id,omitempty"`
}

// StreamStats represents aggregated stream statistics
type StreamStats struct {
	IsLive        bool   `json:"is_live"`
	ViewerCount   int    `json:"viewer_count"`
	FollowerCount int    `json:"follower_count"`
	Uptime        int64  `json:"uptime"` // seconds
	StartedAt     string `json:"started_at,omitempty"`
	Title         string `json:"title"`
	GameName      string `json:"game_name"`
	Language      string `json:"language,omitempty"`
	ThumbnailURL  string `json:"thumbnail_url,omitempty"`
	StreamID      string `json:"stream_id,omitempty"`
}

// ChatMessage represents a Twitch chat message
type ChatMessage struct {
	Username     string   `json:"username"`
	DisplayName  string   `json:"display_name"`
	Message      string   `json:"message"`
	Color        string   `json:"color"`
	Timestamp    string   `json:"timestamp"`
	IsModerator  bool     `json:"is_moderator"`
	IsSubscriber bool     `json:"is_subscriber"`
	Badges       []string `json:"badges"`
}

// EventSubEvent represents a generic EventSub event
type EventSubEvent struct {
	Type      string      `json:"type"` // "follow", "subscribe", "raid"
	Data      interface{} `json:"data"`
	Timestamp string      `json:"timestamp"`
}

// FollowEvent represents a channel.follow event
type FollowEvent struct {
	UserID     string `json:"user_id"`
	UserLogin  string `json:"user_login"`
	UserName   string `json:"user_name"`
	FollowedAt string `json:"followed_at"`
}

// SubscribeEvent represents a channel.subscribe event
type SubscribeEvent struct {
	UserID    string `json:"user_id"`
	UserLogin string `json:"user_login"`
	UserName  string `json:"user_name"`
	Tier      string `json:"tier"`
	IsGift    bool   `json:"is_gift"`
}

// RaidEvent represents a channel.raid event
type RaidEvent struct {
	FromUserID    string `json:"from_user_id"`
	FromUserLogin string `json:"from_user_login"`
	FromUserName  string `json:"from_user_name"`
	Viewers       int    `json:"viewers"`
}

// UpdateStreamRequest represents a request to update stream metadata
type UpdateStreamRequest struct {
	Title    string `json:"title,omitempty"`
	GameID   string `json:"game_id,omitempty"`
	GameName string `json:"game_name,omitempty"`
}

// Internal API response types

// helixUser represents a user from the Twitch Helix API
type helixUser struct {
	ID          string `json:"id"`
	Login       string `json:"login"`
	DisplayName string `json:"display_name"`
}

// helixStream represents a stream from the Twitch Helix API
type helixStream struct {
	ID           string `json:"id"`
	UserID       string `json:"user_id"`
	UserLogin    string `json:"user_login"`
	UserName     string `json:"user_name"`
	GameID       string `json:"game_id"`
	GameName     string `json:"game_name"`
	Type         string `json:"type"`
	Title        string `json:"title"`
	ViewerCount  int    `json:"viewer_count"`
	StartedAt    string `json:"started_at"`
	Language     string `json:"language"`
	ThumbnailURL string `json:"thumbnail_url"`
}

// helixFollowers represents the followers response from the Twitch Helix API
type helixFollowers struct {
	Total int `json:"total"`
}

// helixGame represents a game from the Twitch Helix API
type helixGame struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// helixResponse represents a generic Helix API response wrapper
type helixResponse struct {
	Data []interface{} `json:"data"`
}
