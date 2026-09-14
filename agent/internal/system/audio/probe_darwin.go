package audio

import (
	"os/exec"
)

// Probe prüft, ob CoreAudio bereitsteht.
//
// coreaudiod ist der Systemdienst, über den sämtliche Audioein- und -ausgabe
// läuft. Läuft er, ist Audio benutzbar; fehlt er, hilft auch ein vorhandenes
// Ausgabegerät nicht weiter.
func Probe() Status {
	if err := exec.Command("pgrep", "-x", "coreaudiod").Run(); err == nil {
		return Status{Backend: "coreaudio", Ready: true}
	}

	return Status{Backend: "coreaudio", Ready: false}
}
