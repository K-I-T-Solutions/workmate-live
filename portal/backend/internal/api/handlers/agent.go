package handlers

import (
	"encoding/json"
	"net/http"

	"kit.workmate/live-portal/internal/agentlink"
	"kit.workmate/live-portal/internal/services/agent"
)

type AgentHandler struct {
	// client pollt einen Agent per HTTP. Nil, wenn keine Agent-URL
	// konfiguriert ist (typisch im Link-Betrieb hinter NAT).
	client *agent.Client
	// cache hält die Status, die verbundene Agents über den Link pushen.
	cache *agent.StatusCache
	// hub kennt die aktiven Agent-Verbindungen. Nil, wenn der Link aus ist.
	hub *agentlink.Hub
}

func NewAgentHandler(client *agent.Client, cache *agent.StatusCache, hub *agentlink.Hub) *AgentHandler {
	return &AgentHandler{
		client: client,
		cache:  cache,
		hub:    hub,
	}
}

// GetStatus liefert den Agent-Status. Gepushte Daten aus dem Link haben Vorrang,
// weil ein Agent hinter NAT per HTTP gar nicht erreichbar ist.
func (h *AgentHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	if agentID := r.URL.Query().Get("agent_id"); agentID != "" {
		status, ok := h.cache.Get(agentID)
		if !ok {
			http.Error(w, "unknown or disconnected agent: "+agentID, http.StatusNotFound)
			return
		}
		writeJSON(w, status)
		return
	}

	if status, ok := h.cache.Latest(); ok {
		writeJSON(w, status)
		return
	}

	if h.client == nil {
		http.Error(w, "no agent connected", http.StatusServiceUnavailable)
		return
	}

	status, err := h.client.GetStatus()
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	writeJSON(w, status)
}

func (h *AgentHandler) GetCapabilities(w http.ResponseWriter, r *http.Request) {
	if h.client == nil {
		http.Error(w, "capabilities require a directly reachable agent", http.StatusServiceUnavailable)
		return
	}

	caps, err := h.client.GetCapabilities()
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	writeJSON(w, caps)
}

func (h *AgentHandler) GetInfo(w http.ResponseWriter, r *http.Request) {
	if h.client == nil {
		http.Error(w, "info requires a directly reachable agent", http.StatusServiceUnavailable)
		return
	}

	info, err := h.client.GetInfo()
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	writeJSON(w, info)
}

// ListAgents liefert alle Agents, die aktuell über den Link verbunden sind.
func (h *AgentHandler) ListAgents(w http.ResponseWriter, r *http.Request) {
	agents := []agentlink.AgentInfo{}
	if h.hub != nil {
		agents = h.hub.List()
	}

	writeJSON(w, map[string]interface{}{
		"agents":       agents,
		"link_enabled": h.hub != nil,
	})
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}
