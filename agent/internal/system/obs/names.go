package obs

import "strings"

// processNames sind die Namen, unter denen OBS je nach Plattform in der
// Prozessliste auftaucht. Linux und macOS liefern den reinen Namen, Windows
// den Dateinamen mit Endung.
var processNames = map[string]bool{
	"obs":       true, // Linux, macOS
	"obs64":     true, // Linux (Eigenbau), macOS
	"obs.exe":   true, // Windows
	"obs64.exe": true, // Windows, Standardfall
	"obs32.exe": true, // Windows, 32-Bit-Installation
}

// IsOBSProcess prüft, ob ein Prozessname zu OBS gehört.
// Die Prüfung ist bewusst exakt: "obs-ffmpeg-mux" und ähnliche Hilfsprozesse
// laufen auch ohne geöffnetes OBS und dürfen nicht mitzählen.
func IsOBSProcess(name string) bool {
	return processNames[strings.ToLower(strings.TrimSpace(name))]
}
