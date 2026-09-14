package video

import (
	"context"
	"encoding/json"
	"os/exec"
	"sort"
	"strings"
	"sync"
	"time"
)

var (
	once  sync.Once
	value []string
)

// cameraReport bildet den benötigten Teil der system_profiler-Ausgabe ab.
type cameraReport struct {
	Cameras []struct {
		Name string `json:"_name"`
	} `json:"SPCameraDataType"`
}

// ScanDevices listet die Kameras des Systems über system_profiler.
//
// Der Aufruf dauert rund eine Sekunde und wird deshalb nur einmal gemacht.
// Eine später angesteckte Kamera erscheint damit erst nach einem Neustart des
// Agents — der Kompromiss ist bewusst: sonst würde jeder Poll-Durchlauf einen
// sekundenlangen Prozess starten.
func ScanDevices() ([]string, error) {
	once.Do(func() { value = probeCameras() })
	return value, nil
}

func probeCameras() []string {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	out, err := exec.CommandContext(ctx, "system_profiler", "-json", "SPCameraDataType").Output()
	if err != nil {
		return []string{}
	}

	var report cameraReport
	if err := json.Unmarshal(out, &report); err != nil {
		return []string{}
	}

	devices := make([]string, 0, len(report.Cameras))
	for _, c := range report.Cameras {
		if name := strings.TrimSpace(c.Name); name != "" {
			devices = append(devices, name)
		}
	}

	sort.Strings(devices)
	return devices
}
