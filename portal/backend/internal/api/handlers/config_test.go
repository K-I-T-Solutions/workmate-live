package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"kit.workmate/live-portal/internal/config"
)

// newTestConfig liefert eine gültige Konfiguration, die in eine temporäre
// Datei gespeichert werden kann.
func newTestConfig(t *testing.T) *config.Config {
	t.Helper()

	path := filepath.Join(t.TempDir(), "portal.yaml")
	if err := os.WriteFile(path, []byte("server:\n    port: 8080\n"), 0o600); err != nil {
		t.Fatalf("failed to seed config: %v", err)
	}

	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	return cfg
}

// patch schickt eine Sektion an UpdateConfig.
func patch(t *testing.T, h *ConfigHandler, section string, data interface{}) *httptest.ResponseRecorder {
	t.Helper()

	raw, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("failed to encode section: %v", err)
	}

	body, err := json.Marshal(map[string]interface{}{
		"section": section,
		"data":    json.RawMessage(raw),
	})
	if err != nil {
		t.Fatalf("failed to encode request: %v", err)
	}

	req := httptest.NewRequest(http.MethodPatch, "/api/config", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	h.UpdateConfig(rec, req)

	return rec
}

// getConfig liest die maskierte Konfiguration über den Handler.
func getConfig(t *testing.T, h *ConfigHandler) config.Config {
	t.Helper()

	rec := httptest.NewRecorder()
	h.GetConfig(rec, httptest.NewRequest(http.MethodGet, "/api/config", nil))

	var got config.Config
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("failed to decode config: %v", err)
	}

	return got
}

// Speichert die UI die Sektion unverändert zurück, enthält sie den maskierten
// Platzhalter — der darf das echte Geheimnis nicht überschreiben.
func TestMaskedAgentKeySurvivesRoundTrip(t *testing.T) {
	cfg := newTestConfig(t)
	cfg.Agent.APIKey = "real-agent-key"
	h := NewConfigHandler(cfg)

	safe := getConfig(t, h)
	if safe.Agent.APIKey != "***" {
		t.Fatalf("GetConfig returned api_key %q, want it masked", safe.Agent.APIKey)
	}

	// Genau das, was die UI zurückschickt.
	if rec := patch(t, h, "agent", safe.Agent); rec.Code != http.StatusOK {
		t.Fatalf("update failed: %d %s", rec.Code, rec.Body)
	}

	if cfg.Agent.APIKey != "real-agent-key" {
		t.Errorf("api_key = %q, want the original to be preserved", cfg.Agent.APIKey)
	}
}

func TestMaskedOBSPasswordSurvivesRoundTrip(t *testing.T) {
	cfg := newTestConfig(t)
	cfg.OBS.Password = "real-obs-password"
	h := NewConfigHandler(cfg)

	safe := getConfig(t, h)
	if safe.OBS.Password != "***" {
		t.Fatalf("GetConfig returned password %q, want it masked", safe.OBS.Password)
	}

	if rec := patch(t, h, "obs", safe.OBS); rec.Code != http.StatusOK {
		t.Fatalf("update failed: %d %s", rec.Code, rec.Body)
	}

	if cfg.OBS.Password != "real-obs-password" {
		t.Errorf("password = %q, want the original to be preserved", cfg.OBS.Password)
	}
}

func TestNewSecretsAreStored(t *testing.T) {
	cfg := newTestConfig(t)
	cfg.Agent.APIKey = "old-key"
	h := NewConfigHandler(cfg)

	updated := cfg.Agent
	updated.APIKey = "brand-new-key"

	if rec := patch(t, h, "agent", updated); rec.Code != http.StatusOK {
		t.Fatalf("update failed: %d %s", rec.Code, rec.Body)
	}

	if cfg.Agent.APIKey != "brand-new-key" {
		t.Errorf("api_key = %q, want the new value to be stored", cfg.Agent.APIKey)
	}
}

func TestSwitchingToAgentModeIsStored(t *testing.T) {
	cfg := newTestConfig(t)
	cfg.Agent.APIKey = "shared-secret"
	h := NewConfigHandler(cfg)

	updated := cfg.OBS
	updated.Mode = config.OBSModeAgent

	if rec := patch(t, h, "obs", updated); rec.Code != http.StatusOK {
		t.Fatalf("update failed: %d %s", rec.Code, rec.Body)
	}

	if cfg.OBS.Mode != config.OBSModeAgent {
		t.Errorf("mode = %q, want %q", cfg.OBS.Mode, config.OBSModeAgent)
	}
}

// Die UI darf keine Konfiguration schreiben können, mit der das Portal beim
// nächsten Start nicht mehr hochkommt.
func TestAgentModeWithoutKeyIsRejected(t *testing.T) {
	cfg := newTestConfig(t)
	cfg.Agent.APIKey = ""
	h := NewConfigHandler(cfg)

	updated := cfg.OBS
	updated.Mode = config.OBSModeAgent

	rec := patch(t, h, "obs", updated)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}

	if cfg.OBS.Mode == config.OBSModeAgent {
		t.Error("invalid mode was applied despite the rejection")
	}
}

func TestInvalidModeIsRejected(t *testing.T) {
	cfg := newTestConfig(t)
	h := NewConfigHandler(cfg)

	updated := cfg.OBS
	updated.Mode = "telepathy"

	if rec := patch(t, h, "obs", updated); rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}
