package websocket

import "testing"

func TestWithAgentIDAddsOrigin(t *testing.T) {
	event := map[string]interface{}{"type": "scene_changed", "scene_name": "Intro"}

	got := WithAgentID("barry", event)

	if got["agent_id"] != "barry" {
		t.Errorf("agent_id = %v, want %q", got["agent_id"], "barry")
	}
	if got["type"] != "scene_changed" || got["scene_name"] != "Intro" {
		t.Errorf("original fields lost: %v", got)
	}
}

// Das Ereignis geht an mehrere Empfänger — die Vorlage darf dabei nicht
// verändert werden.
func TestWithAgentIDDoesNotMutateInput(t *testing.T) {
	event := map[string]interface{}{"type": "scene_changed"}

	WithAgentID("cisco", event)

	if _, exists := event["agent_id"]; exists {
		t.Error("the source event was modified")
	}
	if len(event) != 1 {
		t.Errorf("source event has %d keys, want 1", len(event))
	}
}

// Zwei Agents dürfen sich nicht gegenseitig überschreiben.
func TestWithAgentIDIsIndependentPerCall(t *testing.T) {
	event := map[string]interface{}{"type": "scene_changed"}

	a := WithAgentID("cisco", event)
	b := WithAgentID("barry", event)

	if a["agent_id"] != "cisco" || b["agent_id"] != "barry" {
		t.Errorf("results interfere: %v / %v", a["agent_id"], b["agent_id"])
	}
}
