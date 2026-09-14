package automation

import "testing"

func TestFromOBSEventCarriesAgent(t *testing.T) {
	ev, ok := FromOBSEvent("barry", map[string]interface{}{
		"type":       "scene_changed",
		"scene_name": "Kamera 2",
	})
	if !ok {
		t.Fatal("scene_changed should produce an event")
	}

	if ev.Type != "obs:scene_changed" {
		t.Errorf("type = %q, want %q", ev.Type, "obs:scene_changed")
	}
	if ev.Vars["agent"] != "barry" {
		t.Errorf("agent = %q, want %q", ev.Vars["agent"], "barry")
	}
	if ev.Vars["scene"] != "Kamera 2" {
		t.Errorf("scene = %q, want %q", ev.Vars["scene"], "Kamera 2")
	}
}

// Auch Ereignisse ohne eigene Nutzdaten müssen den Absender tragen, sonst
// kann eine Regel nicht auf einen bestimmten Rechner eingeschränkt werden.
func TestStreamStateCarriesAgent(t *testing.T) {
	tests := []struct {
		active bool
		want   string
	}{
		{true, "obs:stream_started"},
		{false, "obs:stream_stopped"},
	}

	for _, tt := range tests {
		ev, ok := FromOBSEvent("cisco", map[string]interface{}{
			"type":   "stream_state_changed",
			"active": tt.active,
		})
		if !ok {
			t.Fatalf("stream_state_changed(active=%v) should produce an event", tt.active)
		}
		if ev.Type != tt.want {
			t.Errorf("type = %q, want %q", ev.Type, tt.want)
		}
		if ev.Vars["agent"] != "cisco" {
			t.Errorf("agent = %q, want %q", ev.Vars["agent"], "cisco")
		}
	}
}

// OBS meldet mehr, als die Engine auswertet — Unbekanntes darf keine Regel auslösen.
func TestUnhandledEventsProduceNothing(t *testing.T) {
	for _, typ := range []string{"record_state_changed", "source_visibility_changed", "connection_changed", ""} {
		if _, ok := FromOBSEvent("cisco", map[string]interface{}{"type": typ}); ok {
			t.Errorf("FromOBSEvent(%q) produced an event, want none", typ)
		}
	}
}

// Fehlt der Szenenname, darf das Ereignis trotzdem durchkommen — nur eben
// mit leerer Variable statt gar nicht.
func TestSceneChangedWithoutName(t *testing.T) {
	ev, ok := FromOBSEvent("cisco", map[string]interface{}{"type": "scene_changed"})
	if !ok {
		t.Fatal("scene_changed should produce an event even without a name")
	}
	if ev.Vars["scene"] != "" {
		t.Errorf("scene = %q, want empty", ev.Vars["scene"])
	}
	if ev.Vars["agent"] != "cisco" {
		t.Errorf("agent = %q, want %q", ev.Vars["agent"], "cisco")
	}
}
