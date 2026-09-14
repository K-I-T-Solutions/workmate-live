package obs

import "testing"

func TestIsOBSProcess(t *testing.T) {
	running := []string{
		"obs",       // Linux, macOS
		"obs64",     // macOS, Eigenbau unter Linux
		"obs64.exe", // Windows, Standardfall
		"obs.exe",
		"obs32.exe",
		"OBS64.EXE", // Windows meldet Namen nicht einheitlich klein
		"  obs  ",   // ps-Ausgabe kann Leerzeichen mitbringen
	}

	for _, name := range running {
		if !IsOBSProcess(name) {
			t.Errorf("IsOBSProcess(%q) = false, want true", name)
		}
	}

	// Hilfsprozesse laufen auch ohne geöffnetes OBS und dürfen nicht zählen.
	notRunning := []string{
		"obs-ffmpeg-mux",
		"obs-browser-helper",
		"obsidian",
		"obs-studio-launcher",
		"",
		"chrome",
	}

	for _, name := range notRunning {
		if IsOBSProcess(name) {
			t.Errorf("IsOBSProcess(%q) = true, want false", name)
		}
	}
}
