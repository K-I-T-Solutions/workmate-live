package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// writeConfig legt eine temporäre Konfigurationsdatei an.
func writeConfig(t *testing.T, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "portal.yaml")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	return path
}

func TestDefaultModeIsDirect(t *testing.T) {
	cfg := Default()

	if cfg.OBS.Mode != OBSModeDirect {
		t.Errorf("default mode = %q, want %q", cfg.OBS.Mode, OBSModeDirect)
	}
}

// Eine Konfiguration ohne obs.mode muss weiterhin wie bisher laufen.
func TestMissingModeNormalizesToDirect(t *testing.T) {
	path := writeConfig(t, `
obs:
    host: 192.168.1.50
    port: 4455
    password: secret
`)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.OBS.Mode != OBSModeDirect {
		t.Errorf("mode = %q, want %q", cfg.OBS.Mode, OBSModeDirect)
	}
	if cfg.OBS.Host != "192.168.1.50" {
		t.Errorf("host = %q, want %q", cfg.OBS.Host, "192.168.1.50")
	}
}

func TestAgentModeLoads(t *testing.T) {
	path := writeConfig(t, `
agent:
    url: ""
    api_key: shared-secret
    primary_id: streaming-pc
    command_timeout: 15s
obs:
    mode: agent
`)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.OBS.Mode != OBSModeAgent {
		t.Errorf("mode = %q, want %q", cfg.OBS.Mode, OBSModeAgent)
	}
	if !cfg.AgentLinkEnabled() {
		t.Error("AgentLinkEnabled = false, want true when an api_key is set")
	}
	if cfg.Agent.PrimaryID != "streaming-pc" {
		t.Errorf("primary_id = %q, want %q", cfg.Agent.PrimaryID, "streaming-pc")
	}
	if cfg.Agent.CommandTimeout != 15*time.Second {
		t.Errorf("command_timeout = %s, want 15s", cfg.Agent.CommandTimeout)
	}
}

// Der Agent-Modus ohne API-Key wäre still wirkungslos — das muss auffallen.
func TestAgentModeRequiresAPIKey(t *testing.T) {
	path := writeConfig(t, `
agent:
    api_key: ""
obs:
    mode: agent
`)

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected agent mode without an api_key to be rejected")
	}

	if !strings.Contains(err.Error(), "api_key") {
		t.Errorf("error = %q, want it to mention api_key", err)
	}
}

func TestUnknownModeIsRejected(t *testing.T) {
	path := writeConfig(t, `
obs:
    mode: telepathy
`)

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected an unknown mode to be rejected")
	}

	if !strings.Contains(err.Error(), "telepathy") {
		t.Errorf("error = %q, want it to name the invalid mode", err)
	}
}

// Im Agent-Modus verbindet das Portal nicht selbst zu OBS — Host und Port
// aus dem OBS-Block dürfen die Validierung dann nicht blockieren.
func TestAgentModeIgnoresOBSAddress(t *testing.T) {
	path := writeConfig(t, `
agent:
    api_key: shared-secret
obs:
    mode: agent
    port: 0
`)

	if _, err := Load(path); err != nil {
		t.Fatalf("agent mode should not validate the OBS address: %v", err)
	}
}

func TestDirectModeStillValidatesPort(t *testing.T) {
	path := writeConfig(t, `
obs:
    mode: direct
    port: 0
`)

	if _, err := Load(path); err == nil {
		t.Fatal("expected an invalid port to be rejected in direct mode")
	}
}

func TestCommandTimeoutFallsBackToDefault(t *testing.T) {
	path := writeConfig(t, `
agent:
    command_timeout: 0s
`)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.Agent.CommandTimeout != 10*time.Second {
		t.Errorf("command_timeout = %s, want the 10s default", cfg.Agent.CommandTimeout)
	}
}

// Der geteilte Agent-Schlüssel darf nicht über die Config-API nach außen gehen.
func TestAgentKeyIsMasked(t *testing.T) {
	cfg := Default()
	cfg.Agent.APIKey = "super-secret-agent-key"

	safe := cfg.ToSafe()

	if safe.Agent.APIKey == cfg.Agent.APIKey {
		t.Error("agent api_key was returned unmasked")
	}
	if strings.Contains(safe.Agent.APIKey, "secret") {
		t.Errorf("masked key %q still leaks the original value", safe.Agent.APIKey)
	}
}
