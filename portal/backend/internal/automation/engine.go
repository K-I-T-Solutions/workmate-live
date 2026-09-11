package automation

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"kit.workmate/live-portal/internal/websocket"
)

// Engine processes incoming events and executes matching automation rules
type Engine struct {
	store   *Store
	eventCh chan Event
	hub     *websocket.Hub
	twitch  TwitchExecutor
	obs     OBSExecutor
	done    chan struct{}
	mu      sync.RWMutex
}

// NewEngine creates a new Engine. Call Run() in a goroutine to start processing
func NewEngine(store *Store, hub *websocket.Hub) *Engine {
	return &Engine{
		store:   store,
		eventCh: make(chan Event, 256),
		hub:     hub,
		done:    make(chan struct{}),
	}
}

// SetTwitchExecutor wires up the Twitch client for chat actions
func (e *Engine) SetTwitchExecutor(exec TwitchExecutor) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.twitch = exec
}

// SetOBSExecutor wires up the OBS client for OBS actions
func (e *Engine) SetOBSExecutor(exec OBSExecutor) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.obs = exec
}

// Send enqueues an event for processing (non-blocking)
func (e *Engine) Send(event Event) {
	select {
	case e.eventCh <- event:
	default:
		log.Printf("Automation: event channel full, dropping: %s", event.Type)
	}
}

// Run starts the event loop. Call this in a goroutine
func (e *Engine) Run() {
	for {
		select {
		case event := <-e.eventCh:
			e.processEvent(event)
		case <-e.done:
			return
		}
	}
}

// Stop shuts down the engine
func (e *Engine) Stop() {
	close(e.done)
}

// Test manually fires a rule by name with empty vars
func (e *Engine) Test(ruleName string) error {
	rule, ok := e.store.Get(ruleName)
	if !ok {
		return fmt.Errorf("rule '%s' not found", ruleName)
	}
	go e.runActions(*rule, Event{Type: rule.Trigger.Type, Vars: map[string]string{}})
	return nil
}

func (e *Engine) processEvent(event Event) {
	for _, rule := range e.store.List() {
		if !rule.Enabled || rule.Trigger.Type != event.Type {
			continue
		}
		r := rule
		go e.runActions(r, event)
	}
}

func (e *Engine) runActions(rule Rule, event Event) {
	log.Printf("Automation: rule '%s' fired by '%s'", rule.Name, event.Type)

	e.mu.RLock()
	twitch := e.twitch
	obs := e.obs
	e.mu.RUnlock()

	for _, action := range rule.Actions {
		if err := e.executeAction(action, event.Vars, twitch, obs); err != nil {
			log.Printf("Automation: action '%s' failed in rule '%s': %v", action.Type, rule.Name, err)
		}
	}

	e.hub.Broadcast(websocket.Message{
		Type: websocket.MessageTypeAutomationFired,
		Data: FiredEvent{
			RuleName:    rule.Name,
			TriggerType: event.Type,
			Timestamp:   time.Now(),
		},
	})
}

func (e *Engine) executeAction(action Action, vars map[string]string, twitch TwitchExecutor, obs OBSExecutor) error {
	switch action.Type {
	case "twitch:chat_message":
		return executeChatMessage(twitch, action.Params, vars)
	case "obs:scene_switch":
		return executeSceneSwitch(obs, action.Params)
	case "obs:source_toggle":
		return executeSourceToggle(obs, action.Params)
	case "delay":
		return executeDelay(action.Params)
	default:
		log.Printf("Automation: unknown action type: %s", action.Type)
		return nil
	}
}

// applyVars replaces {key} placeholders in s with values from vars
func applyVars(s string, vars map[string]string) string {
	if len(vars) == 0 {
		return s
	}
	pairs := make([]string, 0, len(vars)*2)
	for k, v := range vars {
		pairs = append(pairs, "{"+k+"}", v)
	}
	return strings.NewReplacer(pairs...).Replace(s)
}
