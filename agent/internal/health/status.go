package health

import (
	"time"

	"kit.workmate/gaming-agent/internal/system/gpu"
)

//Status ist das zentrale Objekt,
//das der Agent nach außen liefert.

type Status struct {
	Timestamp time.Time   `json:"timestamp"`
	Hostname  string      `json:"hostname"`
	Headless  bool        `json:"headless"`
	Video     VideoStatus `json:"video"`
	Audio     AudioStatus `json:"audio"`
	OBS       OBSStatus   `json:"obs"`
	GPU       gpu.Status  `json:"gpu"`
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
