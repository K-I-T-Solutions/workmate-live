package obs

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"

	"kit.workmate/live-portal/internal/agentlink"
)

const testKey = "test-agent-key"

// stubAgent beantwortet Kommandos mit vorbereiteten Ergebnissen und merkt sich,
// was er empfangen hat.
type stubAgent struct {
	ws      *websocket.Conn
	replies map[string]agentlink.Frame

	received chan agentlink.Frame
}

// startRemote verbindet einen stubAgent mit einem Hub und liefert den fertig
// verdrahteten RemoteController.
func startRemote(t *testing.T, hello agentlink.Hello, replies map[string]agentlink.Frame) (*RemoteController, *stubAgent) {
	t.Helper()

	hub := agentlink.NewHub("")
	srv := httptest.NewServer(agentlink.NewHandler(hub, testKey))
	t.Cleanup(srv.Close)

	header := http.Header{}
	header.Set("X-Agent-Key", testKey)
	header.Set("X-Agent-ID", hello.AgentID)

	ws, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(srv.URL, "http"), header)
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}
	t.Cleanup(func() { ws.Close() })

	agent := &stubAgent{
		ws:       ws,
		replies:  replies,
		received: make(chan agentlink.Frame, 16),
	}

	data, err := json.Marshal(hello)
	if err != nil {
		t.Fatalf("failed to encode hello: %v", err)
	}
	if err := ws.WriteJSON(agentlink.Frame{Type: agentlink.FrameTypeHello, Data: data}); err != nil {
		t.Fatalf("failed to send hello: %v", err)
	}

	go agent.serve()

	// Warten, bis Registrierung und Hello verarbeitet sind.
	deadline := time.Now().Add(2 * time.Second)
	for {
		if conn, ok := hub.Get(hello.AgentID); ok && conn.Hello().AgentID == hello.AgentID {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("agent did not register in time")
		}
		time.Sleep(5 * time.Millisecond)
	}

	return NewRemoteController(hub, 2*time.Second), agent
}

func (a *stubAgent) serve() {
	for {
		var frame agentlink.Frame
		if err := a.ws.ReadJSON(&frame); err != nil {
			return
		}

		if frame.Type != agentlink.FrameTypeCommand {
			continue
		}

		select {
		case a.received <- frame:
		default:
		}

		resp, ok := a.replies[frame.Method]
		if !ok {
			resp = agentlink.Frame{OK: true}
		}
		resp.ID = frame.ID
		resp.Type = agentlink.FrameTypeResponse

		if err := a.ws.WriteJSON(resp); err != nil {
			return
		}
	}
}

func (a *stubAgent) lastCommand(t *testing.T) agentlink.Frame {
	t.Helper()

	select {
	case frame := <-a.received:
		return frame
	case <-time.After(2 * time.Second):
		t.Fatal("agent received no command")
		return agentlink.Frame{}
	}
}

func okWith(t *testing.T, v interface{}) agentlink.Frame {
	t.Helper()

	data, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("failed to encode reply: %v", err)
	}

	return agentlink.Frame{OK: true, Result: data}
}

func TestRemoteGetStatus(t *testing.T) {
	want := &OBSStatus{
		Connected:    true,
		Version:      "30.1.2",
		CurrentScene: "Intro",
		Streaming:    &StreamStatus{Active: true, Duration: 42},
		Recording:    &RecordingStatus{Active: false},
	}

	remote, _ := startRemote(t,
		agentlink.Hello{AgentID: "pc", OBSLocal: true, OBSActive: true},
		map[string]agentlink.Frame{agentlink.MethodOBSStatus: okWith(t, want)},
	)

	got, err := remote.GetStatus()
	if err != nil {
		t.Fatalf("GetStatus failed: %v", err)
	}

	if !got.Connected || got.Version != "30.1.2" || got.CurrentScene != "Intro" {
		t.Errorf("status = %+v, want %+v", got, want)
	}
	if got.Streaming == nil || !got.Streaming.Active || got.Streaming.Duration != 42 {
		t.Errorf("streaming = %+v, want active with duration 42", got.Streaming)
	}
}

func TestRemoteGetScenes(t *testing.T) {
	want := []Scene{
		{Name: "Intro", Active: true, Index: 0},
		{Name: "Outro", Active: false, Index: 1},
	}

	remote, _ := startRemote(t,
		agentlink.Hello{AgentID: "pc", OBSLocal: true, OBSActive: true},
		map[string]agentlink.Frame{agentlink.MethodOBSScenes: okWith(t, want)},
	)

	got, err := remote.GetScenes()
	if err != nil {
		t.Fatalf("GetScenes failed: %v", err)
	}

	if len(got) != 2 || got[0].Name != "Intro" || !got[0].Active || got[1].Name != "Outro" {
		t.Errorf("scenes = %+v, want %+v", got, want)
	}
}

func TestRemoteSwitchSceneSendsParams(t *testing.T) {
	remote, agent := startRemote(t,
		agentlink.Hello{AgentID: "pc", OBSLocal: true, OBSActive: true}, nil)

	if err := remote.SwitchScene("Outro"); err != nil {
		t.Fatalf("SwitchScene failed: %v", err)
	}

	cmd := agent.lastCommand(t)
	if cmd.Method != agentlink.MethodOBSSwitchScene {
		t.Errorf("method = %q, want %q", cmd.Method, agentlink.MethodOBSSwitchScene)
	}

	var params agentlink.SwitchSceneParams
	if err := json.Unmarshal(cmd.Params, &params); err != nil {
		t.Fatalf("invalid params: %v", err)
	}
	if params.SceneName != "Outro" {
		t.Errorf("scene_name = %q, want %q", params.SceneName, "Outro")
	}
}

func TestRemoteToggleSourceSendsParams(t *testing.T) {
	remote, agent := startRemote(t,
		agentlink.Hello{AgentID: "pc", OBSLocal: true, OBSActive: true}, nil)

	if err := remote.ToggleSourceVisibility("Intro", "Cam", false); err != nil {
		t.Fatalf("ToggleSourceVisibility failed: %v", err)
	}

	cmd := agent.lastCommand(t)

	var params agentlink.ToggleSourceParams
	if err := json.Unmarshal(cmd.Params, &params); err != nil {
		t.Fatalf("invalid params: %v", err)
	}

	if params.SceneName != "Intro" || params.SourceName != "Cam" || params.Visible {
		t.Errorf("params = %+v, want Intro/Cam invisible", params)
	}
}

func TestRemoteGetSourcesSendsScene(t *testing.T) {
	want := []Source{{Name: "Cam", Type: "input", Visible: true}}

	remote, agent := startRemote(t,
		agentlink.Hello{AgentID: "pc", OBSLocal: true, OBSActive: true},
		map[string]agentlink.Frame{agentlink.MethodOBSSources: okWith(t, want)},
	)

	got, err := remote.GetSources("Intro")
	if err != nil {
		t.Fatalf("GetSources failed: %v", err)
	}

	if len(got) != 1 || got[0].Name != "Cam" {
		t.Errorf("sources = %+v, want %+v", got, want)
	}

	var params agentlink.SourcesParams
	if err := json.Unmarshal(agent.lastCommand(t).Params, &params); err != nil {
		t.Fatalf("invalid params: %v", err)
	}
	if params.SceneName != "Intro" {
		t.Errorf("scene_name = %q, want %q", params.SceneName, "Intro")
	}
}

func TestRemoteStreamAndRecordMethods(t *testing.T) {
	remote, agent := startRemote(t,
		agentlink.Hello{AgentID: "pc", OBSLocal: true, OBSActive: true}, nil)

	tests := []struct {
		name string
		call func() error
		want string
	}{
		{"start streaming", remote.StartStreaming, agentlink.MethodOBSStreamStart},
		{"stop streaming", remote.StopStreaming, agentlink.MethodOBSStreamStop},
		{"start recording", remote.StartRecording, agentlink.MethodOBSRecordStart},
		{"stop recording", remote.StopRecording, agentlink.MethodOBSRecordStop},
		{"pause recording", remote.PauseRecording, agentlink.MethodOBSRecordPause},
		{"resume recording", remote.ResumeRecording, agentlink.MethodOBSRecordResume},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.call(); err != nil {
				t.Fatalf("%s failed: %v", tt.name, err)
			}

			if got := agent.lastCommand(t).Method; got != tt.want {
				t.Errorf("method = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRemotePropagatesAgentError(t *testing.T) {
	remote, _ := startRemote(t,
		agentlink.Hello{AgentID: "pc", OBSLocal: true, OBSActive: true},
		map[string]agentlink.Frame{
			agentlink.MethodOBSSwitchScene: {OK: false, Error: "source not found: Cam"},
		},
	)

	err := remote.SwitchScene("Outro")
	if err == nil {
		t.Fatal("expected an error from the agent")
	}

	if err.Error() != "source not found: Cam" {
		t.Errorf("error = %q, want %q", err.Error(), "source not found: Cam")
	}
}

// Ohne verbundenen Agent liefert GetStatus "nicht verbunden" statt eines
// Fehlers — die UI soll diesen Zustand anzeigen können.
func TestRemoteStatusWithoutAgent(t *testing.T) {
	remote := NewRemoteController(agentlink.NewHub(""), time.Second)

	status, err := remote.GetStatus()
	if err != nil {
		t.Fatalf("GetStatus should not fail without an agent: %v", err)
	}

	if status.Connected {
		t.Error("connected = true, want false without an agent")
	}

	if remote.IsConnected() {
		t.Error("IsConnected = true, want false without an agent")
	}
}

// Steuerbefehle ohne Agent müssen dagegen klar scheitern.
func TestRemoteCommandsWithoutAgentFail(t *testing.T) {
	remote := NewRemoteController(agentlink.NewHub(""), time.Second)

	if err := remote.SwitchScene("Intro"); err != ErrNoAgent {
		t.Errorf("error = %v, want %v", err, ErrNoAgent)
	}

	if _, err := remote.GetScenes(); err != ErrNoAgent {
		t.Errorf("error = %v, want %v", err, ErrNoAgent)
	}
}

// IsConnected spiegelt den vom Agent gemeldeten OBS-Zustand, nicht nur die
// Tatsache, dass ein Agent verbunden ist.
func TestRemoteIsConnectedFollowsOBSState(t *testing.T) {
	remote, _ := startRemote(t,
		agentlink.Hello{AgentID: "pc", OBSLocal: true, OBSActive: false}, nil)

	if remote.IsConnected() {
		t.Error("IsConnected = true, want false while the agent reports OBS down")
	}

	if got := remote.AgentID(); got != "pc" {
		t.Errorf("AgentID = %q, want %q", got, "pc")
	}
}
