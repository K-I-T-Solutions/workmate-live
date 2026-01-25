package audio

// Status beschreibt den Audio-Zustand des Systems.
type Status struct {
	Backend string
	Ready   bool
}

// Probe ermittelt den aktuellen Audio-Status.
// Aktuell: PipeWire (Stub, aber realistisch).
func Probe() Status {
	return probePipeWire()
}
