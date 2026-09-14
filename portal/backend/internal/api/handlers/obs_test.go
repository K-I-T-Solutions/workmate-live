package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	"kit.workmate/live-portal/internal/services/obs"
)

// stubController merkt sich, welcher Agent bedient wurde.
type stubController struct {
	agentID  string
	scenes   []obs.Scene
	switched string
}

func (s *stubController) IsConnected() bool { return true }
func (s *stubController) GetStatus() (*obs.OBSStatus, error) {
	return &obs.OBSStatus{Connected: true, CurrentScene: s.agentID}, nil
}
func (s *stubController) GetScenes() ([]obs.Scene, error)                   { return s.scenes, nil }
func (s *stubController) GetCurrentScene() (string, error)                  { return "", nil }
func (s *stubController) SwitchScene(name string) error                     { s.switched = name; return nil }
func (s *stubController) GetSources(string) ([]obs.Source, error)           { return nil, nil }
func (s *stubController) ToggleSourceVisibility(string, string, bool) error { return nil }
func (s *stubController) StartStreaming() error                             { return nil }
func (s *stubController) StopStreaming() error                              { return nil }
func (s *stubController) StartRecording() error                             { return nil }
func (s *stubController) StopRecording() error                              { return nil }
func (s *stubController) PauseRecording() error                             { return nil }
func (s *stubController) ResumeRecording() error                            { return nil }

// stubResolver bildet Kennungen auf vorbereitete Controller ab.
type stubResolver struct {
	byID    map[string]*stubController
	primary string
}

func (r *stubResolver) ControllerFor(agentID string) (obs.Controller, error) {
	if agentID == "" {
		agentID = r.primary
	}
	if c, ok := r.byID[agentID]; ok {
		return c, nil
	}
	return nil, obs.ErrAgentNotFound
}

// router baut die Routen so auf, wie routes.go es tut.
func router(h *OBSHandler) chi.Router {
	r := chi.NewRouter()
	mount := func(r chi.Router) {
		r.Get("/status", h.GetStatus)
		r.Get("/scenes", h.GetScenes)
		r.Post("/scenes/switch", h.SwitchScene)
	}
	r.Route("/agents/{agent}/obs", mount)
	r.Route("/obs", mount)
	return r
}

func newTestHandler() (*OBSHandler, *stubResolver) {
	res := &stubResolver{
		primary: "cisco",
		byID: map[string]*stubController{
			"cisco": {agentID: "cisco", scenes: []obs.Scene{{Name: "Studio"}}},
			"barry": {agentID: "barry", scenes: []obs.Scene{{Name: "Kamera 2"}}},
		},
	}
	return NewOBSHandler(res), res
}

func do(t *testing.T, r chi.Router, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()

	var req *http.Request
	if body == "" {
		req = httptest.NewRequest(method, path, nil)
	} else {
		req = httptest.NewRequest(method, path, strings.NewReader(body))
	}

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

// Die Agent-Route spricht den genannten Rechner an, nicht den primären.
func TestAgentRouteAddressesNamedAgent(t *testing.T) {
	h, res := newTestHandler()
	r := router(h)

	rec := do(t, r, http.MethodGet, "/agents/barry/obs/scenes", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", rec.Code, rec.Body)
	}

	var got struct {
		Scenes []obs.Scene `json:"scenes"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("invalid response: %v", err)
	}
	if len(got.Scenes) != 1 || got.Scenes[0].Name != "Kamera 2" {
		t.Errorf("scenes = %+v, want barry's", got.Scenes)
	}

	// Ein Schaltbefehl muss beim selben Agent landen.
	if rec := do(t, r, http.MethodPost, "/agents/barry/obs/scenes/switch", `{"scene_name":"Kamera 2"}`); rec.Code != http.StatusOK {
		t.Fatalf("switch status = %d: %s", rec.Code, rec.Body)
	}
	if res.byID["barry"].switched != "Kamera 2" {
		t.Errorf("barry got %q, want %q", res.byID["barry"].switched, "Kamera 2")
	}
	if res.byID["cisco"].switched != "" {
		t.Errorf("cisco was switched to %q but should not have been touched", res.byID["cisco"].switched)
	}
}

// Die bisherige Route ohne Kennung spricht weiterhin den primären Agent an.
func TestLegacyRouteUsesPrimary(t *testing.T) {
	h, res := newTestHandler()
	r := router(h)

	if rec := do(t, r, http.MethodPost, "/obs/scenes/switch", `{"scene_name":"Studio"}`); rec.Code != http.StatusOK {
		t.Fatalf("status = %d: %s", rec.Code, rec.Body)
	}

	if res.byID["cisco"].switched != "Studio" {
		t.Errorf("primary agent got %q, want %q", res.byID["cisco"].switched, "Studio")
	}
	if res.byID["barry"].switched != "" {
		t.Error("the legacy route touched the wrong agent")
	}
}

// Ein unbekannter Agent ist ein Fehler des Aufrufers, kein Serverfehler —
// sonst ist "Rechner weg" nicht von "OBS kaputt" zu unterscheiden.
func TestUnknownAgentYields404(t *testing.T) {
	h, _ := newTestHandler()
	r := router(h)

	for _, path := range []string{
		"/agents/laptop/obs/status",
		"/agents/laptop/obs/scenes",
	} {
		if rec := do(t, r, http.MethodGet, path, ""); rec.Code != http.StatusNotFound {
			t.Errorf("GET %s = %d, want 404", path, rec.Code)
		}
	}

	rec := do(t, r, http.MethodPost, "/agents/laptop/obs/scenes/switch", `{"scene_name":"X"}`)
	if rec.Code != http.StatusNotFound {
		t.Errorf("switch = %d, want 404", rec.Code)
	}
}
