package automation

import "time"

// Rule defines an automation rule: when trigger fires, run actions sequentially
type Rule struct {
	Name    string   `yaml:"name" json:"name"`
	Enabled bool     `yaml:"enabled" json:"enabled"`
	Trigger Trigger  `yaml:"trigger" json:"trigger"`
	Actions []Action `yaml:"actions" json:"actions"`
}

// Trigger defines what event activates a rule
type Trigger struct {
	Type string `yaml:"type" json:"type"`
}

// Action defines a single step to execute when a rule fires
type Action struct {
	Type   string                 `yaml:"type" json:"type"`
	Params map[string]interface{} `yaml:"params" json:"params"`
}

// Event is an internal event passed to the engine for trigger matching.
// Vars holds template variables extracted from the source event (e.g. {user}, {viewers}).
type Event struct {
	Type string
	Vars map[string]string
}

// FiredEvent is broadcast via WebSocket when a rule successfully executes
type FiredEvent struct {
	RuleName    string    `json:"rule_name"`
	TriggerType string    `json:"trigger_type"`
	Timestamp   time.Time `json:"timestamp"`
}

// TwitchExecutor is the minimal interface required for Twitch chat actions
type TwitchExecutor interface {
	SendChatMessage(message string) error
}

// OBSExecutor is the minimal interface required for OBS actions
type OBSExecutor interface {
	SwitchScene(sceneName string) error
	ToggleSourceVisibility(sceneName, sourceName string, visible bool) error
	GetCurrentScene() (string, error)
}
