package obs

// Scene represents an OBS scene
type Scene struct {
	Name      string `json:"name"`
	Active    bool   `json:"active"`
	Index     int    `json:"index"`
	SceneUUID string `json:"scene_uuid,omitempty"`
}

// Source represents an OBS source
type Source struct {
	Name         string `json:"name"`
	Type         string `json:"type"`
	Visible      bool   `json:"visible"`
	Muted        bool   `json:"muted,omitempty"`
	Volume       float64 `json:"volume,omitempty"`
	InputUUID    string `json:"input_uuid,omitempty"`
}

// StreamStatus represents OBS streaming status
type StreamStatus struct {
	Active        bool   `json:"active"`
	Reconnecting  bool   `json:"reconnecting"`
	Duration      int64  `json:"duration"` // seconds
	Bytes         int64  `json:"bytes"`
	Frames        int64  `json:"frames"`
	DroppedFrames int64  `json:"dropped_frames"`
}

// RecordingStatus represents OBS recording status
type RecordingStatus struct {
	Active   bool   `json:"active"`
	Paused   bool   `json:"paused"`
	Duration int64  `json:"duration"` // seconds
	Bytes    int64  `json:"bytes"`
	Path     string `json:"path,omitempty"`
}

// OBSStatus represents the overall OBS status
type OBSStatus struct {
	Connected       bool             `json:"connected"`
	Version         string           `json:"version,omitempty"`
	CurrentScene    string           `json:"current_scene,omitempty"`
	Streaming       *StreamStatus    `json:"streaming,omitempty"`
	Recording       *RecordingStatus `json:"recording,omitempty"`
}
