package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"kit.workmate/live-portal/internal/services/obs"
)

type OBSHandler struct {
	// resolver liefert je Anfrage den zuständigen Controller: entweder die
	// direkte OBS-Verbindung oder einen Agent. Die Handler sehen keinen
	// Unterschied.
	resolver obs.Resolver
}

func NewOBSHandler(resolver obs.Resolver) *OBSHandler {
	return &OBSHandler{
		resolver: resolver,
	}
}

// controller löst den angesprochenen Agent auf. Die Routen gibt es zweimal:
// unter /api/agents/{agent}/obs/... mit Kennung und unter /api/obs/... ohne —
// dort greift dann der primäre Agent.
func (h *OBSHandler) controller(r *http.Request) (obs.Controller, error) {
	return h.resolver.ControllerFor(chi.URLParam(r, "agent"))
}

// writeResolveError übersetzt einen Auflösungsfehler in eine Antwort.
// Ein unbekannter Agent ist ein Fehler des Aufrufers, kein Serverfehler.
func writeResolveError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, obs.ErrAgentNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)
	case errors.Is(err, obs.ErrAgentNotAddressable):
		http.Error(w, err.Error(), http.StatusBadRequest)
	default:
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// GetStatus returns the OBS status
func (h *OBSHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	ctrl, err := h.controller(r)
	if err != nil {
		writeResolveError(w, err)
		return
	}

	status, err := ctrl.GetStatus()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// GetScenes returns all OBS scenes
func (h *OBSHandler) GetScenes(w http.ResponseWriter, r *http.Request) {
	ctrl, err := h.controller(r)
	if err != nil {
		writeResolveError(w, err)
		return
	}

	scenes, err := ctrl.GetScenes()
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
	ctrl, err := h.controller(r)
	if err != nil {
		writeResolveError(w, err)
		return
	}

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

	if err := ctrl.SwitchScene(req.SceneName); err != nil {
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
	ctrl, err := h.controller(r)
	if err != nil {
		writeResolveError(w, err)
		return
	}

	sceneName := r.URL.Query().Get("scene")
	if sceneName == "" {
		// Get current scene if not specified
		currentScene, err := ctrl.GetCurrentScene()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		sceneName = currentScene
	}

	sources, err := ctrl.GetSources(sceneName)
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
	ctrl, err := h.controller(r)
	if err != nil {
		writeResolveError(w, err)
		return
	}

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

	if err := ctrl.ToggleSourceVisibility(req.SceneName, req.SourceName, req.Visible); err != nil {
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
	ctrl, err := h.controller(r)
	if err != nil {
		writeResolveError(w, err)
		return
	}

	if err := ctrl.StartStreaming(); err != nil {
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
	ctrl, err := h.controller(r)
	if err != nil {
		writeResolveError(w, err)
		return
	}

	if err := ctrl.StopStreaming(); err != nil {
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
	ctrl, err := h.controller(r)
	if err != nil {
		writeResolveError(w, err)
		return
	}

	if err := ctrl.StartRecording(); err != nil {
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
	ctrl, err := h.controller(r)
	if err != nil {
		writeResolveError(w, err)
		return
	}

	if err := ctrl.StopRecording(); err != nil {
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
	ctrl, err := h.controller(r)
	if err != nil {
		writeResolveError(w, err)
		return
	}

	if err := ctrl.PauseRecording(); err != nil {
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
	ctrl, err := h.controller(r)
	if err != nil {
		writeResolveError(w, err)
		return
	}

	if err := ctrl.ResumeRecording(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "recording_resumed",
	})
}
