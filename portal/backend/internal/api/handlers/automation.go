package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"kit.workmate/live-portal/internal/automation"
)

// AutomationHandler handles automation rule CRUD and test execution
type AutomationHandler struct {
	store  *automation.Store
	engine *automation.Engine
}

// NewAutomationHandler creates a new AutomationHandler
func NewAutomationHandler(store *automation.Store, engine *automation.Engine) *AutomationHandler {
	return &AutomationHandler{store: store, engine: engine}
}

// ListRules returns all automation rules
func (h *AutomationHandler) ListRules(w http.ResponseWriter, r *http.Request) {
	rules := h.store.List()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rules)
}

// CreateRule creates a new automation rule
func (h *AutomationHandler) CreateRule(w http.ResponseWriter, r *http.Request) {
	var rule automation.Rule
	if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if rule.Name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}
	if rule.Trigger.Type == "" {
		http.Error(w, "Trigger type is required", http.StatusBadRequest)
		return
	}
	if err := h.store.Add(rule); err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// UpdateRule replaces an existing automation rule
func (h *AutomationHandler) UpdateRule(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	var rule automation.Rule
	if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if err := h.store.Update(name, rule); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// DeleteRule removes an automation rule
func (h *AutomationHandler) DeleteRule(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	if err := h.store.Delete(name); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// TestRule manually fires a rule by name with empty vars
func (h *AutomationHandler) TestRule(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	if err := h.engine.Test(name); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
