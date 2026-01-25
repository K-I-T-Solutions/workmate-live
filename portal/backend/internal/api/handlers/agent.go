package handlers

import (
	"encoding/json"
	"net/http"

	"kit.workmate/live-portal/internal/services/agent"
)

type AgentHandler struct {
	client *agent.Client
}

func NewAgentHandler(client *agent.Client) *AgentHandler {
	return &AgentHandler{
		client: client,
	}
}

func (h *AgentHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	status, err := h.client.GetStatus()
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

func (h *AgentHandler) GetCapabilities(w http.ResponseWriter, r *http.Request) {
	caps, err := h.client.GetCapabilities()
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(caps)
}

func (h *AgentHandler) GetInfo(w http.ResponseWriter, r *http.Request) {
	info, err := h.client.GetInfo()
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(info)
}
