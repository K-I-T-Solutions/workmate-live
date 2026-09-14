package audio

// Status beschreibt den Audio-Zustand des Systems.
// Backend benennt das jeweils verwendete Audiosystem: pipewire (Linux),
// wasapi (Windows) oder coreaudio (macOS).
type Status struct {
	Backend string
	Ready   bool
}
