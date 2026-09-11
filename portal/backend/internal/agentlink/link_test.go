package agentlink

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

const testKey = "test-agent-key"

// fakeAgent ist ein minimaler Agent: er meldet sich an und beantwortet
// Kommandos über die vom Test gesetzte Antwortfunktion.
type fakeAgent struct {
	t     *testing.T
	ws    *websocket.Conn
	reply func(Frame) Frame
}

// startHub bringt Handler und Hub hinter einem httptest-Server hoch.
func startHub(t *testing.T, primaryID string) (*Hub, *httptest.Server) {
	t.Helper()

	hub := NewHub(primaryID)
	srv := httptest.NewServer(NewHandler(hub, testKey))
	t.Cleanup(srv.Close)

	return hub, srv
}

// dialAgent verbindet einen fakeAgent mit dem Portal.
func dialAgent(t *testing.T, srv *httptest.Server, agentID string, hello Hello, reply func(Frame) Frame) *fakeAgent {
	t.Helper()

	header := http.Header{}
	header.Set("X-Agent-Key", testKey)
	header.Set("X-Agent-ID", agentID)

	ws, resp, err := websocket.DefaultDialer.Dial(wsURL(srv.URL), header)
	if err != nil {
		if resp != nil {
			t.Fatalf("dial failed: %v (HTTP %d)", err, resp.StatusCode)
		}
		t.Fatalf("dial failed: %v", err)
	}
	t.Cleanup(func() { ws.Close() })

	a := &fakeAgent{t: t, ws: ws, reply: reply}

	data, err := json.Marshal(hello)
	if err != nil {
		t.Fatalf("failed to encode hello: %v", err)
	}
	if err := ws.WriteJSON(Frame{Type: FrameTypeHello, Data: data}); err != nil {
		t.Fatalf("failed to send hello: %v", err)
	}

	go a.serve()

	return a
}

func (a *fakeAgent) serve() {
	for {
		var frame Frame
		if err := a.ws.ReadJSON(&frame); err != nil {
			return
		}

		if frame.Type != FrameTypeCommand {
			continue
		}

		resp := a.reply(frame)
		resp.ID = frame.ID
		resp.Type = FrameTypeResponse

		if err := a.ws.WriteJSON(resp); err != nil {
			return
		}
	}
}

// sendEvent pusht ein Event ans Portal.
func (a *fakeAgent) sendEvent(event string, payload interface{}) {
	a.t.Helper()

	data, err := json.Marshal(payload)
	if err != nil {
		a.t.Fatalf("failed to encode event: %v", err)
	}

	if err := a.ws.WriteJSON(Frame{Type: FrameTypeEvent, Event: event, Data: data}); err != nil {
		a.t.Fatalf("failed to send event: %v", err)
	}
}

func wsURL(httpURL string) string {
	return "ws" + strings.TrimPrefix(httpURL, "http")
}

func waitFor(t *testing.T, cond func() bool) {
	t.Helper()

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}

	t.Fatal("timed out waiting for condition")
}

func TestAgentConnectsAndRegisters(t *testing.T) {
	hub, srv := startHub(t, "")
	dialAgent(t, srv, "streaming-pc", Hello{
		AgentID:  "streaming-pc",
		Hostname: "studio",
		Version:  "1.2.3",
		OBSLocal: true,
	}, func(Frame) Frame { return Frame{OK: true} })

	waitFor(t, func() bool { return hub.Count() == 1 })

	agents := hub.List()
	if len(agents) != 1 {
		t.Fatalf("expected 1 agent, got %d", len(agents))
	}

	if agents[0].AgentID != "streaming-pc" {
		t.Errorf("agent id = %q, want %q", agents[0].AgentID, "streaming-pc")
	}

	// Hello wird asynchron verarbeitet.
	waitFor(t, func() bool { return hub.List()[0].Hostname == "studio" })

	if got := hub.List()[0].Version; got != "1.2.3" {
		t.Errorf("version = %q, want %q", got, "1.2.3")
	}
}

func TestRejectsWrongKey(t *testing.T) {
	_, srv := startHub(t, "")

	header := http.Header{}
	header.Set("X-Agent-Key", "wrong-key")
	header.Set("X-Agent-ID", "intruder")

	_, resp, err := websocket.DefaultDialer.Dial(wsURL(srv.URL), header)
	if err == nil {
		t.Fatal("expected dial to fail with a wrong key")
	}

	if resp == nil || resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected HTTP 401, got %v", resp)
	}
}

func TestRejectsMissingAgentID(t *testing.T) {
	_, srv := startHub(t, "")

	header := http.Header{}
	header.Set("X-Agent-Key", testKey)

	_, resp, err := websocket.DefaultDialer.Dial(wsURL(srv.URL), header)
	if err == nil {
		t.Fatal("expected dial to fail without an agent id")
	}

	if resp == nil || resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected HTTP 400, got %v", resp)
	}
}

func TestDisabledHandlerRefusesConnections(t *testing.T) {
	hub := NewHub("")
	srv := httptest.NewServer(NewHandler(hub, ""))
	t.Cleanup(srv.Close)

	header := http.Header{}
	header.Set("X-Agent-Key", "")
	header.Set("X-Agent-ID", "agent")

	_, resp, err := websocket.DefaultDialer.Dial(wsURL(srv.URL), header)
	if err == nil {
		t.Fatal("expected dial to fail when the link is unconfigured")
	}

	if resp == nil || resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("expected HTTP 503, got %v", resp)
	}
}

func TestCallReturnsAgentResult(t *testing.T) {
	hub, srv := startHub(t, "")
	var gotMethod string
	var gotParams json.RawMessage

	dialAgent(t, srv, "pc", Hello{AgentID: "pc", OBSLocal: true}, func(f Frame) Frame {
		gotMethod = f.Method
		gotParams = f.Params
		return Frame{OK: true, Result: json.RawMessage(`{"current_scene":"Intro"}`)}
	})

	waitFor(t, func() bool { return hub.Count() == 1 })

	conn, ok := hub.Get("pc")
	if !ok {
		t.Fatal("agent not registered")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	result, err := conn.Call(ctx, MethodOBSSwitchScene, SwitchSceneParams{SceneName: "Intro"})
	if err != nil {
		t.Fatalf("call failed: %v", err)
	}

	if gotMethod != MethodOBSSwitchScene {
		t.Errorf("method = %q, want %q", gotMethod, MethodOBSSwitchScene)
	}

	var params SwitchSceneParams
	if err := json.Unmarshal(gotParams, &params); err != nil {
		t.Fatalf("agent received invalid params: %v", err)
	}
	if params.SceneName != "Intro" {
		t.Errorf("scene_name = %q, want %q", params.SceneName, "Intro")
	}

	var payload struct {
		CurrentScene string `json:"current_scene"`
	}
	if err := json.Unmarshal(result, &payload); err != nil {
		t.Fatalf("invalid result: %v", err)
	}
	if payload.CurrentScene != "Intro" {
		t.Errorf("current_scene = %q, want %q", payload.CurrentScene, "Intro")
	}
}

func TestCallPropagatesAgentError(t *testing.T) {
	hub, srv := startHub(t, "")
	dialAgent(t, srv, "pc", Hello{AgentID: "pc", OBSLocal: true}, func(Frame) Frame {
		return Frame{OK: false, Error: "not connected to OBS"}
	})

	waitFor(t, func() bool { return hub.Count() == 1 })

	conn, _ := hub.Get("pc")

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := conn.Call(ctx, MethodOBSStatus, nil)
	if err == nil {
		t.Fatal("expected an error from the agent")
	}

	if err.Error() != "not connected to OBS" {
		t.Errorf("error = %q, want %q", err.Error(), "not connected to OBS")
	}
}

// Ein Kommando an einen stummen Agent darf nicht ewig hängen.
func TestCallRespectsContextDeadline(t *testing.T) {
	hub, srv := startHub(t, "")
	// Antwortet nie: Frames werden geschluckt.
	silent := make(chan struct{})
	t.Cleanup(func() { close(silent) })

	header := http.Header{}
	header.Set("X-Agent-Key", testKey)
	header.Set("X-Agent-ID", "mute")

	ws, _, err := websocket.DefaultDialer.Dial(wsURL(srv.URL), header)
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}
	t.Cleanup(func() { ws.Close() })

	go func() {
		for {
			var f Frame
			if err := ws.ReadJSON(&f); err != nil {
				return
			}
			select {
			case <-silent:
				return
			default:
			}
		}
	}()

	waitFor(t, func() bool { return hub.Count() == 1 })

	conn, _ := hub.Get("mute")

	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()

	start := time.Now()
	if _, err := conn.Call(ctx, MethodOBSStatus, nil); err == nil {
		t.Fatal("expected a timeout error")
	}

	if elapsed := time.Since(start); elapsed > time.Second {
		t.Errorf("call took %s, expected it to abort near the deadline", elapsed)
	}
}

func TestEventsReachHandler(t *testing.T) {
	hub, srv := startHub(t, "")
	events := make(chan string, 4)
	hub.SetEventHandler(func(agentID, event string, data json.RawMessage) {
		var payload struct {
			Type      string `json:"type"`
			SceneName string `json:"scene_name"`
		}
		_ = json.Unmarshal(data, &payload)
		events <- agentID + "/" + event + "/" + payload.SceneName
	})

	agent := dialAgent(t, srv, "pc", Hello{AgentID: "pc", OBSLocal: true},
		func(Frame) Frame { return Frame{OK: true} })

	waitFor(t, func() bool { return hub.Count() == 1 })

	agent.sendEvent(EventOBS, map[string]interface{}{
		"type":       "scene_changed",
		"scene_name": "Outro",
	})

	select {
	case got := <-events:
		if got != "pc/obs_event/Outro" {
			t.Errorf("event = %q, want %q", got, "pc/obs_event/Outro")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for the event")
	}
}

// Meldet der Agent einen OBS-Verbindungswechsel, muss die Agentenliste das zeigen.
func TestOBSConnectionStateIsTracked(t *testing.T) {
	hub, srv := startHub(t, "")
	agent := dialAgent(t, srv, "pc", Hello{AgentID: "pc", OBSLocal: true, OBSActive: false},
		func(Frame) Frame { return Frame{OK: true} })

	waitFor(t, func() bool { return hub.Count() == 1 })

	agent.sendEvent(EventOBS, map[string]interface{}{
		"type":      "connection_changed",
		"connected": true,
	})

	waitFor(t, func() bool {
		conn, ok := hub.Get("pc")
		return ok && conn.Hello().OBSActive
	})

	agent.sendEvent(EventOBS, map[string]interface{}{
		"type":      "connection_changed",
		"connected": false,
	})

	waitFor(t, func() bool {
		conn, ok := hub.Get("pc")
		return ok && !conn.Hello().OBSActive
	})
}

// Ohne konfigurierten primären Agent wird der erste Agent mit OBS gewählt.
func TestOBSAgentPrefersOBSCapableAgent(t *testing.T) {
	hub, srv := startHub(t, "")
	dialAgent(t, srv, "monitor-only", Hello{AgentID: "monitor-only", OBSLocal: false},
		func(Frame) Frame { return Frame{OK: true} })
	dialAgent(t, srv, "streaming-pc", Hello{AgentID: "streaming-pc", OBSLocal: true},
		func(Frame) Frame { return Frame{OK: true} })

	waitFor(t, func() bool { return hub.Count() == 2 })

	// Hello wird asynchron verarbeitet — auf das OBS-Flag warten.
	waitFor(t, func() bool {
		conn, ok := hub.OBSAgent()
		return ok && conn.AgentID() == "streaming-pc"
	})
}

// Ein konfigurierter primärer Agent hat Vorrang.
func TestOBSAgentHonoursPrimaryID(t *testing.T) {
	hub, srv := startHub(t, "second")
	dialAgent(t, srv, "first", Hello{AgentID: "first", OBSLocal: true},
		func(Frame) Frame { return Frame{OK: true} })
	dialAgent(t, srv, "second", Hello{AgentID: "second", OBSLocal: true},
		func(Frame) Frame { return Frame{OK: true} })

	waitFor(t, func() bool { return hub.Count() == 2 })

	conn, ok := hub.OBSAgent()
	if !ok {
		t.Fatal("expected the primary agent to be selected")
	}
	if conn.AgentID() != "second" {
		t.Errorf("selected %q, want %q", conn.AgentID(), "second")
	}
}

// Verbindet sich derselbe Agent erneut, ersetzt die neue Verbindung die alte,
// statt sie zu verdoppeln.
func TestReconnectReplacesStaleConnection(t *testing.T) {
	hub, srv := startHub(t, "")
	dialAgent(t, srv, "pc", Hello{AgentID: "pc", OBSLocal: true},
		func(Frame) Frame { return Frame{OK: true} })
	waitFor(t, func() bool { return hub.Count() == 1 })

	first, _ := hub.Get("pc")

	dialAgent(t, srv, "pc", Hello{AgentID: "pc", OBSLocal: true},
		func(Frame) Frame { return Frame{OK: true} })

	waitFor(t, func() bool {
		current, ok := hub.Get("pc")
		return ok && current != first
	})

	if hub.Count() != 1 {
		t.Errorf("agent count = %d, want 1", hub.Count())
	}
}

// Bricht die Verbindung während eines Calls ab, muss der Aufrufer aufwachen.
func TestCallFailsWhenAgentDisappears(t *testing.T) {
	hub, srv := startHub(t, "")
	header := http.Header{}
	header.Set("X-Agent-Key", testKey)
	header.Set("X-Agent-ID", "flaky")

	ws, _, err := websocket.DefaultDialer.Dial(wsURL(srv.URL), header)
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}

	waitFor(t, func() bool { return hub.Count() == 1 })

	conn, _ := hub.Get("flaky")

	errCh := make(chan error, 1)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, err := conn.Call(ctx, MethodOBSStatus, nil)
		errCh <- err
	}()

	// Verbindung hart schließen, während der Call läuft.
	time.Sleep(50 * time.Millisecond)
	ws.Close()

	select {
	case err := <-errCh:
		if err == nil {
			t.Fatal("expected the call to fail after the agent disappeared")
		}
	case <-time.After(3 * time.Second):
		t.Fatal("call did not return after the agent disconnected")
	}
}
