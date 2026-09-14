package obs

import (
	"os/exec"
	"strings"
)

// Probe prüft über pgrep, ob OBS läuft.
//
// Der direkte Weg wäre sysctl kern.proc.all, dessen Auswertung aber feste
// Offsets in struct kinfo_proc voraussetzt — die je nach macOS-Version
// abweichen und bei einem Fehlgriff still falsche Ergebnisse liefern.
// pgrep gehört zum Basissystem, kostet nur wenige Millisekunden und ist
// damit auch für das Polling im Sekundentakt unproblematisch.
func Probe() Status {
	for _, name := range []string{"OBS", "obs"} {
		// -x: exakter Name, damit "obs-ffmpeg-mux" o. Ä. nicht mitzählt.
		if err := exec.Command("pgrep", "-x", name).Run(); err == nil {
			return Status{Running: true}
		}
	}

	// Fallback für den Fall, dass pgrep fehlt: die Prozessliste durchsuchen.
	out, err := exec.Command("ps", "-axco", "command").Output()
	if err != nil {
		return Status{Running: false}
	}

	for _, line := range strings.Split(string(out), "\n") {
		if IsOBSProcess(line) {
			return Status{Running: true}
		}
	}

	return Status{Running: false}
}
