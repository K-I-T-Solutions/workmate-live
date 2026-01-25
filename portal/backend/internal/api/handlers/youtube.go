package handlers

import (
	"encoding/json"
	"net/http"

	"kit.workmate/live-portal/internal/services/youtube"
)

type YouTubeHandler struct {
	client *youtube.Client
}

func NewYouTubeHandler(client *youtube.Client) *YouTubeHandler {
	return &YouTubeHandler{
		client: client,
	}
}

// GetStatus returns YouTube connection status
func (h *YouTubeHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	if h.client == nil {
		http.Error(w, "YouTube not configured", http.StatusServiceUnavailable)
		return
	}

	status := h.client.GetStatus()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// GetStats returns YouTube stream statistics
func (h *YouTubeHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	if h.client == nil {
		http.Error(w, "YouTube not configured", http.StatusServiceUnavailable)
		return
	}

	stats, err := h.client.GetStats()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// UpdateStream updates stream metadata (title, description)
func (h *YouTubeHandler) UpdateStream(w http.ResponseWriter, r *http.Request) {
	if h.client == nil {
		http.Error(w, "YouTube not configured", http.StatusServiceUnavailable)
		return
	}

	var req youtube.UpdateStreamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.client.UpdateStream(&req); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "success",
	})
}
