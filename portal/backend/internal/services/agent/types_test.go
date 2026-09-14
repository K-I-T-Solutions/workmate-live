package agent

import (
	"encoding/json"
	"testing"
)

// Die Kennung kommt hinzu, ohne die bisherige Struktur zu verändern:
// bestehende Empfänger lesen weiter hostname, video und so weiter.
func TestStatusMessageStaysFlat(t *testing.T) {
	msg := StatusMessage{
		Status:  &Status{Hostname: "barry"},
		AgentID: "barry-01",
	}

	raw, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var got map[string]interface{}
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if got["agent_id"] != "barry-01" {
		t.Errorf("agent_id = %v, want %q", got["agent_id"], "barry-01")
	}

	// hostname muss auf oberster Ebene stehen, nicht unter "status"
	if got["hostname"] != "barry" {
		t.Errorf("hostname = %v, want it flat alongside agent_id", got["hostname"])
	}
	if _, nested := got["Status"]; nested {
		t.Error("Status appears as a nested object, breaking existing readers")
	}
}

// Der Status lässt sich weiterhin ohne Kennung lesen — so kommt er vom Agent.
func TestStatusRoundTrip(t *testing.T) {
	raw := []byte(`{"hostname":"cisco","obs":{"running":true}}`)

	var status Status
	if err := json.Unmarshal(raw, &status); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if status.Hostname != "cisco" {
		t.Errorf("hostname = %q, want %q", status.Hostname, "cisco")
	}
}
