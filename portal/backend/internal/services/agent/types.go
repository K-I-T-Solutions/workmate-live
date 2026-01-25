package agent

import "time"

// Status represents the agent's status response
type Status struct {
	Timestamp time.Time   `json:"timestamp"`
	Hostname  string      `json:"hostname"`
	Headless  bool        `json:"headless"`
	Video     VideoStatus `json:"video"`
	Audio     AudioStatus `json:"audio"`
	OBS       OBSStatus   `json:"obs"`
	GPU       GPUStatus   `json:"gpu"`
}

type VideoStatus struct {
	DeviceCount int      `json:"device_count"`
	Devices     []string `json:"devices"`
}

type AudioStatus struct {
	Backend string `json:"backend"`
	Ready   bool   `json:"ready"`
}

type OBSStatus struct {
	Running bool `json:"running"`
}

type GPUStatus struct {
	Present     bool     `json:"present"`
	Vendors     []string `json:"vendors,omitempty"`
	RenderNodes []string `json:"render_nodes,omitempty"`
}

// Capabilities represents agent capabilities
type Capabilities struct {
	CanVideo  bool `json:"can_video"`
	CanAudio  bool `json:"can_audio"`
	CanStream bool `json:"can_stream"`
}

// Info represents agent build info
type Info struct {
	Name      string    `json:"name"`
	Version   string    `json:"version"`
	Commit    string    `json:"commit"`
	BuildTime string    `json:"build_time"`
	Specs     SpecsInfo `json:"specs"`
}

type SpecsInfo struct {
	OS     string     `json:"os"`
	Arch   string     `json:"arch"`
	Kernel string     `json:"kernel"`
	CPU    CPUInfo    `json:"cpu"`
	Memory MemoryInfo `json:"memory"`
}

type CPUInfo struct {
	Model   string `json:"model"`
	Cores   int    `json:"cores"`
	Threads int    `json:"threads"`
}

type MemoryInfo struct {
	TotalMB int `json:"total_mb"`
}
