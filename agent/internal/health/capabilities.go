package health

type Capabilities struct {
	CanVideo  bool `json:"can_video"`
	CanAudio  bool `json:"can_audio"`
	CanStream bool `json:"can_stream"`
}
