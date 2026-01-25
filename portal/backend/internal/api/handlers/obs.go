package handlers

import (
	"encoding/json"
	"net/http"

	"kit.workmate/live-portal/internal/services/obs"
)

type OBSHandler struct {
	client *obs.Client
}

func NewOBSHandler(client *obs.Client) *OBSHandler {
	return &OBSHandler{
		client: client,
	}
}

// GetStatus returns the OBS status
func (h *OBSHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	status, err := h.client.GetStatus()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// GetScenes returns all OBS scenes
func (h *OBSHandler) GetScenes(w http.ResponseWriter, r *http.Request) {
	scenes, err := h.client.GetScenes()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"scenes": scenes,
	})
}

// SwitchScene switches to a different scene
func (h *OBSHandler) SwitchScene(w http.ResponseWriter, r *http.Request) {
	var req struct {
		SceneName string `json:"scene_name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.SceneName == "" {
		http.Error(w, "scene_name is required", http.StatusBadRequest)
		return
	}

	if err := h.client.SwitchScene(req.SceneName); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
		"scene":  req.SceneName,
	})
}

// GetSources returns all sources in a scene
func (h *OBSHandler) GetSources(w http.ResponseWriter, r *http.Request) {
	sceneName := r.URL.Query().Get("scene")
	if sceneName == "" {
		// Get current scene if not specified
		currentScene, err := h.client.GetCurrentScene()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		sceneName = currentScene
	}

	sources, err := h.client.GetSources(sceneName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"scene":   sceneName,
		"sources": sources,
	})
}

// ToggleSource toggles source visibility
func (h *OBSHandler) ToggleSource(w http.ResponseWriter, r *http.Request) {
	var req struct {
		SceneName  string `json:"scene_name"`
		SourceName string `json:"source_name"`
		Visible    bool   `json:"visible"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.SceneName == "" || req.SourceName == "" {
		http.Error(w, "scene_name and source_name are required", http.StatusBadRequest)
		return
	}

	if err := h.client.ToggleSourceVisibility(req.SceneName, req.SourceName, req.Visible); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "ok",
		"source":  req.SourceName,
		"visible": req.Visible,
	})
}

// StartStreaming starts OBS streaming
func (h *OBSHandler) StartStreaming(w http.ResponseWriter, r *http.Request) {
	if err := h.client.StartStreaming(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "streaming_started",
	})
}

// StopStreaming stops OBS streaming
func (h *OBSHandler) StopStreaming(w http.ResponseWriter, r *http.Request) {
	if err := h.client.StopStreaming(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "streaming_stopped",
	})
}

// StartRecording starts OBS recording
func (h *OBSHandler) StartRecording(w http.ResponseWriter, r *http.Request) {
	if err := h.client.StartRecording(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "recording_started",
	})
}

// StopRecording stops OBS recording
func (h *OBSHandler) StopRecording(w http.ResponseWriter, r *http.Request) {
	if err := h.client.StopRecording(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "recording_stopped",
	})
}

// PauseRecording pauses OBS recording
func (h *OBSHandler) PauseRecording(w http.ResponseWriter, r *http.Request) {
	if err := h.client.PauseRecording(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "recording_paused",
	})
}

// ResumeRecording resumes OBS recording
func (h *OBSHandler) ResumeRecording(w http.ResponseWriter, r *http.Request) {
	if err := h.client.ResumeRecording(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "recording_resumed",
	})
}
