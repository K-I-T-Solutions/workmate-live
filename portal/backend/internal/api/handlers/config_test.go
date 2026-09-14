package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"kit.workmate/live-portal/internal/config"
	"kit.workmate/live-portal/internal/storage"
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

// fakeUsers ersetzt die Benutzerdatenbank im Test.
type fakeUsers struct {
	user      *storage.User
	setFor    int
	setTo     string
	failFind  bool
	failWrite bool
}

func (f *fakeUsers) GetByUsername(username string) (*storage.User, error) {
	if f.failFind || f.user == nil || f.user.Username != username {
		return nil, fmt.Errorf("no such user: %s", username)
	}
	return f.user, nil
}

func (f *fakeUsers) UpdatePassword(userID int, newPassword string) error {
	if f.failWrite {
		return fmt.Errorf("database is read-only")
	}
	f.setFor, f.setTo = userID, newPassword
	return nil
}

func newUsers() *fakeUsers {
	return &fakeUsers{user: &storage.User{ID: 7, Username: "admin"}}
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
	h := NewConfigHandler(cfg, newUsers())

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
	h := NewConfigHandler(cfg, newUsers())

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
	h := NewConfigHandler(cfg, newUsers())

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
	h := NewConfigHandler(cfg, newUsers())

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
	h := NewConfigHandler(cfg, newUsers())

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
	h := NewConfigHandler(cfg, newUsers())

	updated := cfg.OBS
	updated.Mode = "telepathy"

	if rec := patch(t, h, "obs", updated); rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

// Das Anmeldepasswort steht in der Datenbank. Wird es in den Einstellungen
// geändert, muss es dort ankommen — sonst gilt beim Login weiter das alte.
func TestPasswordChangeReachesTheDatabase(t *testing.T) {
	cfg := newTestConfig(t)
	cfg.Auth.DefaultUser.Username = "admin"
	cfg.Auth.DefaultUser.Password = "altes-passwort"

	users := newUsers()
	h := NewConfigHandler(cfg, users)

	body := map[string]interface{}{
		"default_user": map[string]string{"username": "admin", "password": "neues-passwort"},
	}
	if rec := patch(t, h, "auth", body); rec.Code != http.StatusOK {
		t.Fatalf("update failed: %d %s", rec.Code, rec.Body)
	}

	if users.setTo != "neues-passwort" {
		t.Errorf("database got %q, want %q", users.setTo, "neues-passwort")
	}
	if users.setFor != 7 {
		t.Errorf("updated user id %d, want 7", users.setFor)
	}
	if cfg.Auth.DefaultUser.Password != "neues-passwort" {
		t.Errorf("config password = %q, want it updated too", cfg.Auth.DefaultUser.Password)
	}
}

// Speichert die UI die Sektion unverändert, kommt der Platzhalter zurück —
// das ist keine Passwortänderung und darf die Datenbank nicht anfassen.
func TestMaskedPasswordDoesNotTouchTheDatabase(t *testing.T) {
	cfg := newTestConfig(t)
	cfg.Auth.DefaultUser.Username = "admin"
	cfg.Auth.DefaultUser.Password = "echtes-passwort"

	users := newUsers()
	h := NewConfigHandler(cfg, users)

	safe := getConfig(t, h)
	if safe.Auth.DefaultUser.Password != "***" {
		t.Fatalf("GetConfig returned %q, want it masked", safe.Auth.DefaultUser.Password)
	}

	body := map[string]interface{}{
		"default_user": map[string]string{
			"username": safe.Auth.DefaultUser.Username,
			"password": safe.Auth.DefaultUser.Password,
		},
	}
	if rec := patch(t, h, "auth", body); rec.Code != http.StatusOK {
		t.Fatalf("update failed: %d %s", rec.Code, rec.Body)
	}

	if users.setTo != "" {
		t.Errorf("database was written with %q, want no write at all", users.setTo)
	}
	if cfg.Auth.DefaultUser.Password != "echtes-passwort" {
		t.Errorf("config password = %q, want the original preserved", cfg.Auth.DefaultUser.Password)
	}
}

// Schlägt der Datenbankschreibvorgang fehl, darf die Konfiguration nicht
// still auf ein Passwort umgestellt werden, mit dem niemand sich anmelden kann.
func TestFailedPasswordWriteLeavesConfigUntouched(t *testing.T) {
	cfg := newTestConfig(t)
	cfg.Auth.DefaultUser.Username = "admin"
	cfg.Auth.DefaultUser.Password = "altes-passwort"

	users := newUsers()
	users.failWrite = true
	h := NewConfigHandler(cfg, users)

	body := map[string]interface{}{
		"default_user": map[string]string{"username": "admin", "password": "neues-passwort"},
	}
	rec := patch(t, h, "auth", body)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec.Code)
	}

	if cfg.Auth.DefaultUser.Password != "altes-passwort" {
		t.Errorf("config password = %q, want the old one to survive", cfg.Auth.DefaultUser.Password)
	}
}

// Passt der Benutzername zu keinem Konto, muss das auffallen.
func TestPasswordChangeForUnknownUserFails(t *testing.T) {
	cfg := newTestConfig(t)
	cfg.Auth.DefaultUser.Username = "admin"
	cfg.Auth.DefaultUser.Password = "altes-passwort"

	users := newUsers()
	users.failFind = true
	h := NewConfigHandler(cfg, users)

	body := map[string]interface{}{
		"default_user": map[string]string{"username": "admin", "password": "neues-passwort"},
	}
	if rec := patch(t, h, "auth", body); rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec.Code)
	}
}
