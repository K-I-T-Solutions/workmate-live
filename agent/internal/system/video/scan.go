package video

import (
	"os"
	"path/filepath"
	"sort"
)

// ScanDevices findet alle /dev/video* Devices,
// die wirklich existieren und zugreifbar sind.
func ScanDevices() ([]string, error) {
	matches, err := filepath.Glob("/dev/video*")
	if err != nil {
		return nil, err
	}

	devices := make([]string, 0, len(matches))

	for _, path := range matches {
		info, err := os.Stat(path)
		if err != nil {
			// Device verschwunden oder keine Rechte → ignorieren
			continue
		}

		// Nur Character Devices akzeptieren
		if info.Mode()&os.ModeCharDevice == 0 {
			continue
		}

		devices = append(devices, path)
	}

	sort.Strings(devices)
	return devices, nil
}
