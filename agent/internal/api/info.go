package api

import "kit.workmate/gaming-agent/internal/system/specs"

type Info struct {
	Name      string      `json:"name"`
	Version   string      `json:"version"`
	Commit    string      `json:"commit"`
	BuildTime string      `json:"build_time"`
	Specs     specs.Specs `json:"specs"`
}
