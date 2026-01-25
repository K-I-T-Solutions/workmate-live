package audio

import (
	"os"
	"path/filepath"
)

// probePipeWire prüft, ob PipeWire aktiv ist.
// Wir prüfen das PipeWire-Socket im XDG_RUNTIME_DIR.
func probePipeWire() Status {
	runtimeDir := os.Getenv("XDG_RUNTIME_DIR")
	if runtimeDir == "" {
		return Status{
			Backend: "pipewire",
			Ready:   false,
		}
	}

	socket := filepath.Join(runtimeDir, "pipewire-0")
	if _, err := os.Stat(socket); err == nil {
		return Status{
			Backend: "pipewire",
			Ready:   true,
		}
	}

	return Status{
		Backend: "pipewire",
		Ready:   false,
	}
}
