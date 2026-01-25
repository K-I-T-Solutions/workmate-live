package handlers

import (
	"encoding/json"
	"net/http"

	"kit.workmate/gaming-portal/internal/services/twitch"
)

// TwitchHandler handles Twitch-related HTTP requests
type TwitchHandler struct {
	client *twitch.Client
}

// NewTwitchHandler creates a new Twitch handler
func NewTwitchHandler(client *twitch.Client) *TwitchHandler {
	return &TwitchHandler{
		client: client,
	}
}

// GetStatus returns the Twitch connection status
func (h *TwitchHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	if h.client == nil {
		http.Error(w, "Twitch not enabled", http.StatusServiceUnavailable)
		return
	}

	status, err := h.client.GetStatus()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// GetStats returns current stream statistics
func (h *TwitchHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	if h.client == nil {
		http.Error(w, "Twitch not enabled", http.StatusServiceUnavailable)
		return
	}

	stats, err := h.client.GetStreamStats()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// UpdateStream updates stream metadata (title, game)
func (h *TwitchHandler) UpdateStream(w http.ResponseWriter, r *http.Request) {
	if h.client == nil {
		http.Error(w, "Twitch not enabled", http.StatusServiceUnavailable)
		return
	}

	var req twitch.UpdateStreamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Title == "" && req.GameID == "" && req.GameName == "" {
		http.Error(w, "At least one field required (title, game_id, or game_name)", http.StatusBadRequest)
		return
	}

	if err := h.client.UpdateStreamMetadata(&req); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
