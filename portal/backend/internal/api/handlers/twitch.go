package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"kit.workmate/live-portal/internal/services/twitch"
	"kit.workmate/live-portal/internal/services/twitch/commands"
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

// SendMessage sends a message to the Twitch chat
func (h *TwitchHandler) SendMessage(w http.ResponseWriter, r *http.Request) {
	if h.client == nil {
		http.Error(w, "Twitch not enabled", http.StatusServiceUnavailable)
		return
	}

	var req struct {
		Message string `json:"message"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.Message == "" {
		http.Error(w, "Message is required", http.StatusBadRequest)
		return
	}

	if err := h.client.SendChatMessage(req.Message); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// ListCommands returns all registered chat commands
func (h *TwitchHandler) ListCommands(w http.ResponseWriter, r *http.Request) {
	if h.client == nil {
		http.Error(w, "Twitch not enabled", http.StatusServiceUnavailable)
		return
	}

	commands := h.client.ListCommands()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(commands)
}

// ExecuteCommand executes a chat command from the UI
func (h *TwitchHandler) ExecuteCommand(w http.ResponseWriter, r *http.Request) {
	if h.client == nil {
		http.Error(w, "Twitch not enabled", http.StatusServiceUnavailable)
		return
	}

	var req struct {
		Command string `json:"command"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.Command == "" {
		http.Error(w, "Command is required", http.StatusBadRequest)
		return
	}

	result := h.client.ExecuteCommand(req.Command, "portal")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
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

// createCommandRequest is the JSON body for creating/updating a command
type createCommandRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Response    string  `json:"response"`
	ModOnly     bool    `json:"mod_only"`
	Cooldown    float64 `json:"cooldown"` // seconds
	Enabled     bool    `json:"enabled"`
}

// CreateCommand creates a new custom command
func (h *TwitchHandler) CreateCommand(w http.ResponseWriter, r *http.Request) {
	if h.client == nil {
		http.Error(w, "Twitch not enabled", http.StatusServiceUnavailable)
		return
	}

	var req createCommandRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.Name == "" || req.Response == "" {
		http.Error(w, "Name and response are required", http.StatusBadRequest)
		return
	}

	cmd := commands.CustomCommand{
		Name:        req.Name,
		Description: req.Description,
		Response:    req.Response,
		ModOnly:     req.ModOnly,
		Cooldown:    time.Duration(req.Cooldown * float64(time.Second)),
		Enabled:     req.Enabled,
	}

	if err := h.client.AddCommand(cmd); err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// UpdateCommand updates an existing custom command
func (h *TwitchHandler) UpdateCommand(w http.ResponseWriter, r *http.Request) {
	if h.client == nil {
		http.Error(w, "Twitch not enabled", http.StatusServiceUnavailable)
		return
	}

	name := chi.URLParam(r, "name")
	if name == "" {
		http.Error(w, "Command name is required", http.StatusBadRequest)
		return
	}

	var req createCommandRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	cmd := commands.CustomCommand{
		Name:        name,
		Description: req.Description,
		Response:    req.Response,
		ModOnly:     req.ModOnly,
		Cooldown:    time.Duration(req.Cooldown * float64(time.Second)),
		Enabled:     req.Enabled,
	}

	if err := h.client.UpdateCommand(name, cmd); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// DeleteCommand deletes a custom command
func (h *TwitchHandler) DeleteCommand(w http.ResponseWriter, r *http.Request) {
	if h.client == nil {
		http.Error(w, "Twitch not enabled", http.StatusServiceUnavailable)
		return
	}

	name := chi.URLParam(r, "name")
	if name == "" {
		http.Error(w, "Command name is required", http.StatusBadRequest)
		return
	}

	if err := h.client.DeleteCommand(name); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
