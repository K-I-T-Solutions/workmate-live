package automation

// FromOBSEvent übersetzt ein OBS-Ereignis in ein Automation-Event.
//
// agentID benennt den Rechner, von dem das Ereignis stammt, und steht Regeln
// als {agent} zur Verfügung. Bei direkter OBS-Verbindung ist das
// websocket.LocalAgentID.
//
// Der zweite Rückgabewert ist false, wenn das Ereignis keinen Trigger hat —
// OBS meldet mehr, als die Engine auswertet.
func FromOBSEvent(agentID string, event map[string]interface{}) (Event, bool) {
	vars := map[string]string{"agent": agentID}

	switch event["type"] {
	case "scene_changed":
		vars["scene"], _ = event["scene_name"].(string)
		return Event{Type: "obs:scene_changed", Vars: vars}, true

	case "stream_state_changed":
		if active, _ := event["active"].(bool); active {
			return Event{Type: "obs:stream_started", Vars: vars}, true
		}
		return Event{Type: "obs:stream_stopped", Vars: vars}, true
	}

	return Event{}, false
}
