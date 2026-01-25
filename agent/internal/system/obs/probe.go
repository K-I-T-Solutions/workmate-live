package obs

import (
	"os"
	"strings"
)

func Probe() Status {
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return Status{Running: false}
	}

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}

		pid := e.Name()
		if !isNumeric(pid) {
			continue
		}

		comm, err := os.ReadFile("/proc/" + pid + "/comm")
		if err != nil {
			continue
		}

		name := strings.TrimSpace(string(comm))
		if name == "obs" || name == "obs64" {
			return Status{Running: true}
		}
	}

	return Status{Running: false}
}

func isNumeric(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}
