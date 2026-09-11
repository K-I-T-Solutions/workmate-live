package agentlink

import (
	"encoding/json"
	"log"
	"sort"
	"sync"
	"time"
)

// AgentInfo beschreibt einen verbundenen Agent für die API.
type AgentInfo struct {
	AgentID   string    `json:"agent_id"`
	Hostname  string    `json:"hostname,omitempty"`
	Version   string    `json:"version,omitempty"`
	Commit    string    `json:"commit,omitempty"`
	OBSLocal  bool      `json:"obs_local"`
	OBSActive bool      `json:"obs_active"`
	Connected bool      `json:"connected"`
	Since     time.Time `json:"since"`
}

// EventHandler wird für jedes vom Agent gepushte Event aufgerufen.
type EventHandler func(agentID string, event string, data json.RawMessage)

// Hub verwaltet die aktiven Agent-Verbindungen.
type Hub struct {
	mu     sync.RWMutex
	agents map[string]*Conn

	// primaryID benennt den Agent, der OBS steuert. Leer bedeutet:
	// den erstbesten verbundenen Agent mit OBS verwenden.
	primaryID string

	onEvent EventHandler
}

// NewHub creates a new agent hub. primaryID darf leer sein.
func NewHub(primaryID string) *Hub {
	return &Hub{
		agents:    make(map[string]*Conn),
		primaryID: primaryID,
	}
}

// SetEventHandler registriert den Callback für Agent-Events.
// Vor dem Start aufrufen.
func (h *Hub) SetEventHandler(fn EventHandler) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.onEvent = fn
}

func (h *Hub) eventHandler() EventHandler {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.onEvent
}

// register nimmt eine Verbindung auf und verdrängt eine bestehende Verbindung
// desselben Agents (etwa nach einem Reconnect, dessen Abriss noch nicht
// erkannt wurde).
func (h *Hub) register(c *Conn) {
	h.mu.Lock()
	old, exists := h.agents[c.agentID]
	h.agents[c.agentID] = c
	h.mu.Unlock()

	if exists && old != c {
		log.Printf("agentlink: replacing stale connection for agent %s", c.agentID)
		old.close()
	}

	log.Printf("agentlink: agent %s connected", c.agentID)
}

// unregister entfernt die Verbindung, sofern sie noch die aktuelle ist.
func (h *Hub) unregister(c *Conn) {
	h.mu.Lock()
	if current, ok := h.agents[c.agentID]; ok && current == c {
		delete(h.agents, c.agentID)
	}
	h.mu.Unlock()

	log.Printf("agentlink: agent %s disconnected", c.agentID)
}

// Get liefert die Verbindung zu einem bestimmten Agent.
func (h *Hub) Get(agentID string) (*Conn, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	c, ok := h.agents[agentID]
	return c, ok
}

// OBSAgent liefert den Agent, der OBS steuern soll: entweder den konfigurierten
// primären Agent oder — wenn keiner konfiguriert ist — den ersten verbundenen
// Agent mit lokaler OBS-Instanz. Die Auswahl ist bei mehreren Kandidaten
// deterministisch (alphabetisch nach Agent-ID).
func (h *Hub) OBSAgent() (*Conn, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if h.primaryID != "" {
		c, ok := h.agents[h.primaryID]
		return c, ok
	}

	ids := make([]string, 0, len(h.agents))
	for id, c := range h.agents {
		if c.Hello().OBSLocal {
			ids = append(ids, id)
		}
	}
	if len(ids) == 0 {
		return nil, false
	}

	sort.Strings(ids)
	return h.agents[ids[0]], true
}

// List liefert alle verbundenen Agents, sortiert nach Agent-ID.
func (h *Hub) List() []AgentInfo {
	h.mu.RLock()
	conns := make([]*Conn, 0, len(h.agents))
	for _, c := range h.agents {
		conns = append(conns, c)
	}
	h.mu.RUnlock()

	infos := make([]AgentInfo, 0, len(conns))
	for _, c := range conns {
		infos = append(infos, c.Info())
	}

	sort.Slice(infos, func(i, j int) bool { return infos[i].AgentID < infos[j].AgentID })
	return infos
}

// Count liefert die Anzahl verbundener Agents.
func (h *Hub) Count() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.agents)
}
