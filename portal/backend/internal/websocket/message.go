package websocket

// Message represents a WebSocket message
type Message struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// Message types
const (
	MessageTypeAgentStatus     = "agent_status"
	MessageTypeOBSEvent        = "obs_event"
	MessageTypeTwitchChat      = "twitch_chat"
	MessageTypeTwitchEvent     = "twitch_event"
	MessageTypeYouTubeChat     = "youtube_chat"
	MessageTypeTwitchCommand   = "twitch_command"
	MessageTypeAutomationFired = "automation_fired"
	MessageTypePing            = "ping"
	MessageTypePong            = "pong"
)
