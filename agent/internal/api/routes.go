package api

import (
	"encoding/json"
	"net/http"

	"kit.workmate/live-agent/internal/buildinfo"
	"kit.workmate/live-agent/internal/health"
	"kit.workmate/live-agent/internal/system/specs"
)

func Routes(cache *health.Cache) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		status := cache.Get()
		if status == nil {
			http.Error(w, "status not ready", http.StatusServiceUnavailable)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(status)
	})

	mux.HandleFunc("/info", func(w http.ResponseWriter, r *http.Request) {
		info := Info{
			Name:      buildinfo.Name,
			Version:   buildinfo.Version,
			Commit:    buildinfo.Commit,
			BuildTime: buildinfo.BuildTime,
			Specs:     specs.Probe(),
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(info)
	})
	mux.HandleFunc("/capabilities", func(w http.ResponseWriter, r *http.Request) {
		caps := cache.Capabilities()

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(caps)
	})

	return mux
}
