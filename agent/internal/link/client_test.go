package link

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"

	"kit.workmate/live-agent/internal/config"
	"kit.workmate/live-agent/internal/health"
	"kit.workmate/live-agent/internal/services/obs"
)

const testKey = "test-agent-key"

// fakeOBS steht für die lokale OBS-Instanz und zeichnet auf, was der Agent
// aufgrund der Portal-Kommandos tut.
type fakeOBS struct {
	mu sync.Mutex

	connected     bool
	currentScene  string
	switched      []string
	toggled       []string
	streamStarted bool
	recordStarted bool
	failWith      error

	eventCallback func(map[string]interface{})
}

func newFakeOBS() *fakeOBS {
	return &fakeOBS{connected: true, currentScene: "Intro"}
}

func (f *fakeOBS) SetEventCallback(cb func(map[string]interface{})) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.eventCallback = cb
}

// emit simuliert ein Event aus OBS.
func (f *fakeOBS) emit(event map[string]interface{}) {
	f.mu.Lock()
	cb := f.eventCallback
	f.mu.Unlock()

	if cb != nil {
		cb(event)
	}
}

func (f *fakeOBS) IsConnected() bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.connected
}

func (f *fakeOBS) GetStatus() (*obs.OBSStatus, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.failWith != nil {
		return nil, f.failWith
	}

	return &obs.OBSStatus{
		Connected:    f.connected,
		Version:      "30.0.0",
		CurrentScene: f.currentScene,
	}, nil
}

func (f *fakeOBS) GetScenes() ([]obs.Scene, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.failWith != nil {
		return nil, f.failWith
	}

	return []obs.Scene{
		{Name: "Intro", Active: true, Index: 0},
		{Name: "Outro", Active: false, Index: 1},
	}, nil
}

func (f *fakeOBS) GetCurrentScene() (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.currentScene, f.failWith
}

func (f *fakeOBS) SwitchScene(sceneName string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.failWith != nil {
		return f.failWith
	}

	f.switched = append(f.switched, sceneName)
	f.currentScene = sceneName
	return nil
}

func (f *fakeOBS) GetSources(sceneName string) ([]obs.Source, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.failWith != nil {
		return nil, f.failWith
	}

	return []obs.Source{{Name: "Cam", Type: "input", Visible: true}}, nil
}

func (f *fakeOBS) ToggleSourceVisibility(sceneName, sourceName string, visible bool) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.failWith != nil {
		return f.failWith
	}

	state := "off"
	if visible {
		state = "on"
	}
	f.toggled = append(f.toggled, sceneName+"/"+sourceName+"="+state)
	return nil
}

func (f *fakeOBS) StartStreaming() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.streamStarted = true
	return f.failWith
}

func (f *fakeOBS) StopStreaming() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.streamStarted = false
	return f.failWith
}

func (f *fakeOBS) StartRecording() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.recordStarted = true
	return f.failWith
}

func (f *fakeOBS) StopRecording() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.recordStarted = false
	return f.failWith
}

func (f *fakeOBS) PauseRecording() error  { return f.failWith }
func (f *fakeOBS) ResumeRecording() error { return f.failWith }

func (f *fakeOBS) snapshot(fn func(*fakeOBS)) {
	f.mu.Lock()
	defer f.mu.Unlock()
	fn(f)
}

// fakePortal nimmt die Agent-Verbindung entgegen und stellt die empfangenen
// Frames dem Test zur Verfügung.
type fakePortal struct {
	srv *httptest.Server

	mu    sync.Mutex
	conns []*websocket.Conn

	hellos  chan Hello
	events  chan Frame
	dialled chan struct{}

	gotKey     string
	gotAgentID string
}

func newFakePortal(t *testing.T) *fakePortal {
	t.Helper()

	p := &fakePortal{
		hellos:  make(chan Hello, 8),
		events:  make(chan Frame, 64),
		dialled: make(chan struct{}, 8),
	}

	upgrader := websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}

	p.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/ws/agent") {
			http.Error(w, "unexpected path "+r.URL.Path, http.StatusNotFound)
			return
		}

		p.mu.Lock()
		p.gotKey = r.Header.Get("X-Agent-Key")
		p.gotAgentID = r.Header.Get("X-Agent-ID")
		p.mu.Unlock()

		if r.Header.Get("X-Agent-Key") != testKey {
			http.Error(w, "bad key", http.StatusUnauthorized)
			return
		}

		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}

		p.mu.Lock()
		p.conns = append(p.conns, ws)
		p.mu.Unlock()

		select {
		case p.dialled <- struct{}{}:
		default:
		}

		go p.read(ws)
	}))

	t.Cleanup(p.srv.Close)
	return p
}

func (p *fakePortal) read(ws *websocket.Conn) {
	for {
		var frame Frame
		if err := ws.ReadJSON(&frame); err != nil {
			return
		}

		switch frame.Type {
		case FrameTypeHello:
			var hello Hello
			if err := json.Unmarshal(frame.Data, &hello); err == nil {
				select {
				case p.hellos <- hello:
				default:
				}
			}
		default:
			select {
			case p.events <- frame:
			default:
			}
		}
	}
}

// latest liefert die zuletzt angenommene Verbindung.
func (p *fakePortal) latest(t *testing.T) *websocket.Conn {
	t.Helper()

	p.mu.Lock()
	defer p.mu.Unlock()

	if len(p.conns) == 0 {
		t.Fatal("no agent connected yet")
	}

	return p.conns[len(p.conns)-1]
}

// command schickt ein Kommando an den Agent und wartet auf die Antwort.
func (p *fakePortal) command(t *testing.T, method string, params interface{}) Frame {
	t.Helper()

	frame := Frame{ID: "req-1", Type: FrameTypeCommand, Method: method}
	if params != nil {
		raw, err := json.Marshal(params)
		if err != nil {
			t.Fatalf("failed to encode params: %v", err)
		}
		frame.Params = raw
	}

	if err := p.latest(t).WriteJSON(frame); err != nil {
		t.Fatalf("failed to send command: %v", err)
	}

	deadline := time.After(3 * time.Second)
	for {
		select {
		case got := <-p.events:
			if got.Type == FrameTypeResponse && got.ID == frame.ID {
				return got
			}
			// Events dürfen dazwischenkommen — weiterlesen.
		case <-deadline:
			t.Fatalf("timed out waiting for a response to %s", method)
		}
	}
}

// waitForEvent wartet auf ein Event des angegebenen Namens.
func (p *fakePortal) waitForEvent(t *testing.T, name string) Frame {
	t.Helper()

	deadline := time.After(3 * time.Second)
	for {
		select {
		case got := <-p.events:
			if got.Type == FrameTypeEvent && got.Event == name {
				return got
			}
		case <-deadline:
			t.Fatalf("timed out waiting for event %q", name)
		}
	}
}

// startAgent bringt einen echten Link-Client gegen das Fake-Portal hoch.
func startAgent(t *testing.T, portal *fakePortal, fake *fakeOBS) (*Client, *health.Cache) {
	t.Helper()

	cache := health.NewCache()
	cache.Set(&health.Status{Hostname: "studio", Timestamp: time.Now()})

	cfg := config.PortalConfig{
		Enabled:       true,
		URL:           portal.srv.URL,
		APIKey:        testKey,
		AgentID:       "streaming-pc",
		Timeout:       2 * time.Second,
		RetryAttempts: 1,
		RetryDelay:    50 * time.Millisecond,
	}

	client := NewClient(cfg, "streaming-pc", fake, cache)

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	go client.Run(ctx)

	select {
	case <-portal.dialled:
	case <-time.After(3 * time.Second):
		t.Fatal("agent did not connect")
	}

	return client, cache
}

func TestAgentSendsHelloOnConnect(t *testing.T) {
	portal := newFakePortal(t)
	fake := newFakeOBS()
	startAgent(t, portal, fake)

	// Der Hostname kommt direkt vom System, damit er auch beim allerersten
	// Verbinden gesetzt ist — bevor der Health-Poller geschrieben hat.
	wantHost, err := os.Hostname()
	if err != nil {
		t.Fatalf("failed to read hostname: %v", err)
	}

	select {
	case hello := <-portal.hellos:
		if hello.AgentID != "streaming-pc" {
			t.Errorf("agent_id = %q, want %q", hello.AgentID, "streaming-pc")
		}
		if hello.Hostname != wantHost {
			t.Errorf("hostname = %q, want %q", hello.Hostname, wantHost)
		}
		if !hello.OBSLocal {
			t.Error("obs_local = false, want true")
		}
		if !hello.OBSActive {
			t.Error("obs_active = false, want true")
		}
	case <-time.After(3 * time.Second):
		t.Fatal("no hello received")
	}

	portal.mu.Lock()
	defer portal.mu.Unlock()
	if portal.gotKey != testKey {
		t.Errorf("key header = %q, want %q", portal.gotKey, testKey)
	}
	if portal.gotAgentID != "streaming-pc" {
		t.Errorf("agent id header = %q, want %q", portal.gotAgentID, "streaming-pc")
	}
}

func TestSwitchSceneCommandReachesOBS(t *testing.T) {
	portal := newFakePortal(t)
	fake := newFakeOBS()
	startAgent(t, portal, fake)

	resp := portal.command(t, MethodOBSSwitchScene, SwitchSceneParams{SceneName: "Outro"})
	if !resp.OK {
		t.Fatalf("command failed: %s", resp.Error)
	}

	fake.snapshot(func(f *fakeOBS) {
		if len(f.switched) != 1 || f.switched[0] != "Outro" {
			t.Errorf("switched = %v, want [Outro]", f.switched)
		}
	})
}

func TestToggleSourceCommandReachesOBS(t *testing.T) {
	portal := newFakePortal(t)
	fake := newFakeOBS()
	startAgent(t, portal, fake)

	resp := portal.command(t, MethodOBSToggleSource, ToggleSourceParams{
		SceneName:  "Intro",
		SourceName: "Cam",
		Visible:    false,
	})
	if !resp.OK {
		t.Fatalf("command failed: %s", resp.Error)
	}

	fake.snapshot(func(f *fakeOBS) {
		if len(f.toggled) != 1 || f.toggled[0] != "Intro/Cam=off" {
			t.Errorf("toggled = %v, want [Intro/Cam=off]", f.toggled)
		}
	})
}

func TestScenesCommandReturnsPayload(t *testing.T) {
	portal := newFakePortal(t)
	fake := newFakeOBS()
	startAgent(t, portal, fake)

	resp := portal.command(t, MethodOBSScenes, nil)
	if !resp.OK {
		t.Fatalf("command failed: %s", resp.Error)
	}

	var scenes []obs.Scene
	if err := json.Unmarshal(resp.Result, &scenes); err != nil {
		t.Fatalf("invalid result: %v", err)
	}

	if len(scenes) != 2 || scenes[0].Name != "Intro" || !scenes[0].Active {
		t.Errorf("scenes = %+v, want Intro (active) and Outro", scenes)
	}
}

func TestStreamAndRecordCommands(t *testing.T) {
	portal := newFakePortal(t)
	fake := newFakeOBS()
	startAgent(t, portal, fake)

	if resp := portal.command(t, MethodOBSStreamStart, nil); !resp.OK {
		t.Fatalf("stream start failed: %s", resp.Error)
	}
	if resp := portal.command(t, MethodOBSRecordStart, nil); !resp.OK {
		t.Fatalf("record start failed: %s", resp.Error)
	}

	fake.snapshot(func(f *fakeOBS) {
		if !f.streamStarted {
			t.Error("stream not started")
		}
		if !f.recordStarted {
			t.Error("recording not started")
		}
	})
}

func TestOBSErrorIsReportedToPortal(t *testing.T) {
	portal := newFakePortal(t)
	fake := newFakeOBS()
	fake.snapshot(func(f *fakeOBS) { f.failWith = obs.ErrNotConnected })
	startAgent(t, portal, fake)

	resp := portal.command(t, MethodOBSSwitchScene, SwitchSceneParams{SceneName: "Outro"})
	if resp.OK {
		t.Fatal("expected the command to fail")
	}

	if resp.Error != obs.ErrNotConnected.Error() {
		t.Errorf("error = %q, want %q", resp.Error, obs.ErrNotConnected.Error())
	}
}

func TestUnknownMethodIsRejected(t *testing.T) {
	portal := newFakePortal(t)
	fake := newFakeOBS()
	startAgent(t, portal, fake)

	resp := portal.command(t, "obs.explode", nil)
	if resp.OK {
		t.Fatal("expected an unknown method to fail")
	}

	if !strings.Contains(resp.Error, "unknown method") {
		t.Errorf("error = %q, want it to mention an unknown method", resp.Error)
	}
}

func TestOBSEventsArePushedToPortal(t *testing.T) {
	portal := newFakePortal(t)
	fake := newFakeOBS()
	startAgent(t, portal, fake)

	// Warten, bis der Link den Event-Callback gesetzt hat.
	deadline := time.Now().Add(2 * time.Second)
	for {
		var registered bool
		fake.mu.Lock()
		registered = fake.eventCallback != nil
		fake.mu.Unlock()
		if registered {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("link never registered an OBS event callback")
		}
		time.Sleep(5 * time.Millisecond)
	}

	fake.emit(map[string]interface{}{"type": "scene_changed", "scene_name": "Outro"})

	frame := portal.waitForEvent(t, EventOBS)

	var payload struct {
		Type      string `json:"type"`
		SceneName string `json:"scene_name"`
	}
	if err := json.Unmarshal(frame.Data, &payload); err != nil {
		t.Fatalf("invalid event payload: %v", err)
	}

	if payload.Type != "scene_changed" || payload.SceneName != "Outro" {
		t.Errorf("event = %+v, want scene_changed/Outro", payload)
	}
}

func TestAgentStatusIsPushedPeriodically(t *testing.T) {
	portal := newFakePortal(t)
	fake := newFakeOBS()
	startAgent(t, portal, fake)

	frame := portal.waitForEvent(t, EventAgentStatus)

	var status health.Status
	if err := json.Unmarshal(frame.Data, &status); err != nil {
		t.Fatalf("invalid status payload: %v", err)
	}

	if status.Hostname != "studio" {
		t.Errorf("hostname = %q, want %q", status.Hostname, "studio")
	}
}

// Reißt die Verbindung ab, muss der Agent von sich aus neu verbinden.
func TestAgentReconnectsAfterDrop(t *testing.T) {
	portal := newFakePortal(t)
	fake := newFakeOBS()
	startAgent(t, portal, fake)

	// Erstes Hello abwarten, dann die Verbindung hart schließen.
	select {
	case <-portal.hellos:
	case <-time.After(3 * time.Second):
		t.Fatal("no initial hello")
	}

	portal.latest(t).Close()

	select {
	case <-portal.dialled:
	case <-time.After(5 * time.Second):
		t.Fatal("agent did not reconnect")
	}

	select {
	case hello := <-portal.hellos:
		if hello.AgentID != "streaming-pc" {
			t.Errorf("agent_id after reconnect = %q, want %q", hello.AgentID, "streaming-pc")
		}
	case <-time.After(3 * time.Second):
		t.Fatal("no hello after reconnect")
	}
}

// Ein Agent ohne OBS meldet das und weist OBS-Kommandos ab.
func TestAgentWithoutOBSRejectsCommands(t *testing.T) {
	portal := newFakePortal(t)

	cache := health.NewCache()
	cache.Set(&health.Status{Hostname: "monitor-only"})

	cfg := config.PortalConfig{
		Enabled:    true,
		URL:        portal.srv.URL,
		APIKey:     testKey,
		Timeout:    2 * time.Second,
		RetryDelay: 50 * time.Millisecond,
	}

	client := NewClient(cfg, "monitor-only", nil, cache)

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	go client.Run(ctx)

	select {
	case <-portal.dialled:
	case <-time.After(3 * time.Second):
		t.Fatal("agent did not connect")
	}

	select {
	case hello := <-portal.hellos:
		if hello.OBSLocal {
			t.Error("obs_local = true, want false for an agent without OBS")
		}
	case <-time.After(3 * time.Second):
		t.Fatal("no hello received")
	}

	resp := portal.command(t, MethodOBSStatus, nil)
	if resp.OK {
		t.Fatal("expected the command to be rejected")
	}
	if !strings.Contains(resp.Error, "does not control an OBS instance") {
		t.Errorf("error = %q, want it to mention the missing OBS instance", resp.Error)
	}
}

func TestEndpointDerivation(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		want    string
		wantErr bool
	}{
		{name: "http", url: "http://portal.local:8080", want: "ws://portal.local:8080/ws/agent"},
		{name: "https", url: "https://portal.example.com", want: "wss://portal.example.com/ws/agent"},
		{name: "trailing slash", url: "http://portal.local:8080/", want: "ws://portal.local:8080/ws/agent"},
		{name: "sub path", url: "https://example.com/workmate", want: "wss://example.com/workmate/ws/agent"},
		{name: "ws passthrough", url: "ws://portal.local:8080", want: "ws://portal.local:8080/ws/agent"},
		{name: "whitespace", url: "  http://portal.local:8080  ", want: "ws://portal.local:8080/ws/agent"},
		{name: "bad scheme", url: "ftp://portal.local", wantErr: true},
		{name: "no host", url: "http://", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{cfg: config.PortalConfig{URL: tt.url}}

			got, err := c.endpoint()
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected an error for %q, got %q", tt.url, got)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error for %q: %v", tt.url, err)
			}
			if got != tt.want {
				t.Errorf("endpoint = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestBackoffGrowsAndIsCapped(t *testing.T) {
	c := &Client{cfg: config.PortalConfig{
		RetryDelay:    time.Second,
		RetryAttempts: 3,
	}}

	tests := []struct {
		attempt int
		want    time.Duration
	}{
		{attempt: 1, want: time.Second},
		{attempt: 2, want: 2 * time.Second},
		{attempt: 3, want: 4 * time.Second},
		{attempt: 4, want: 8 * time.Second},
		// Ab retry_attempts Verdopplungen bleibt die Wartezeit konstant.
		{attempt: 5, want: 8 * time.Second},
		{attempt: 20, want: 8 * time.Second},
	}

	for _, tt := range tests {
		if got := c.backoff(tt.attempt); got != tt.want {
			t.Errorf("backoff(%d) = %s, want %s", tt.attempt, got, tt.want)
		}
	}
}

func TestBackoffRespectsMaximum(t *testing.T) {
	c := &Client{cfg: config.PortalConfig{
		RetryDelay:    30 * time.Second,
		RetryAttempts: 10,
	}}

	if got := c.backoff(9); got != maxBackoff {
		t.Errorf("backoff(9) = %s, want the %s cap", got, maxBackoff)
	}
}
