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

// LocalAgentID kennzeichnet Ereignisse aus der direkten OBS-Verbindung
// (obs.mode: direct). Dort gibt es keinen Agent, die Herkunft soll aber
// trotzdem benannt sein — sonst müssten Empfänger zwei Fälle unterscheiden.
const LocalAgentID = "local"

// WithAgentID ergänzt ein Ereignis um seine Herkunft.
//
// Die Ereignisdaten stammen vom Agent und werden an mehrere Empfänger
// verteilt; deshalb wird kopiert statt die ursprüngliche Map zu verändern.
func WithAgentID(agentID string, event map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(event)+1)
	for k, v := range event {
		out[k] = v
	}
	out["agent_id"] = agentID

	return out
}
