package agentlink

import "encoding/json"

// Dieses Protokoll wird von agent/internal/link/protocol.go gespiegelt.
// Agent und Portal sind getrennte Go-Module — beide Dateien müssen synchron bleiben.

// Frame-Typen auf der Agent<->Portal-Verbindung.
const (
	FrameTypeHello    = "hello"    // Agent -> Portal, erster Frame nach Verbindungsaufbau
	FrameTypeCommand  = "command"  // Portal -> Agent
	FrameTypeResponse = "response" // Agent -> Portal, Antwort auf ein Command
	FrameTypeEvent    = "event"    // Agent -> Portal, unaufgefordert
)

// Command-Methoden, die ein Agent ausführen kann.
const (
	MethodOBSStatus       = "obs.status"
	MethodOBSScenes       = "obs.scenes"
	MethodOBSCurrentScene = "obs.current_scene"
	MethodOBSSwitchScene  = "obs.switch_scene"
	MethodOBSSources      = "obs.sources"
	MethodOBSToggleSource = "obs.toggle_source"
	MethodOBSStreamStart  = "obs.stream_start"
	MethodOBSStreamStop   = "obs.stream_stop"
	MethodOBSRecordStart  = "obs.record_start"
	MethodOBSRecordStop   = "obs.record_stop"
	MethodOBSRecordPause  = "obs.record_pause"
	MethodOBSRecordResume = "obs.record_resume"
)

// Event-Namen, die der Agent pusht.
const (
	EventOBS         = "obs_event"
	EventAgentStatus = "agent_status"
)

// Frame ist der Umschlag für alle Nachrichten auf der Link-Verbindung.
type Frame struct {
	// ID korreliert Command und Response. Leer bei Events und Hello.
	ID   string `json:"id,omitempty"`
	Type string `json:"type"`

	// Command
	Method string          `json:"method,omitempty"`
	Params json.RawMessage `json:"params,omitempty"`

	// Response
	OK     bool            `json:"ok,omitempty"`
	Result json.RawMessage `json:"result,omitempty"`
	Error  string          `json:"error,omitempty"`

	// Event und Hello
	Event string          `json:"event,omitempty"`
	Data  json.RawMessage `json:"data,omitempty"`
}

// Hello beschreibt den Agent beim Verbindungsaufbau.
type Hello struct {
	AgentID   string `json:"agent_id"`
	Hostname  string `json:"hostname"`
	Version   string `json:"version"`
	Commit    string `json:"commit,omitempty"`
	OBSLocal  bool   `json:"obs_local"`  // Agent steuert eine lokale OBS-Instanz
	OBSActive bool   `json:"obs_active"` // OBS-Verbindung steht gerade
}

// SwitchSceneParams für MethodOBSSwitchScene.
type SwitchSceneParams struct {
	SceneName string `json:"scene_name"`
}

// SourcesParams für MethodOBSSources.
type SourcesParams struct {
	SceneName string `json:"scene_name"`
}

// ToggleSourceParams für MethodOBSToggleSource.
type ToggleSourceParams struct {
	SceneName  string `json:"scene_name"`
	SourceName string `json:"source_name"`
	Visible    bool   `json:"visible"`
}
