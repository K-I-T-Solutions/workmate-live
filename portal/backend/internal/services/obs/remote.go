package obs

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"kit.workmate/live-portal/internal/agentlink"
)

// ErrNoAgent wird zurückgegeben, wenn gerade kein Agent verbunden ist, der OBS
// steuern könnte.
var ErrNoAgent = fmt.Errorf("no OBS-capable agent connected")

// RemoteController steuert OBS über einen Agent. Jeder Aufruf wird als Kommando
// über die vom Agent aufgebaute WebSocket-Verbindung geschickt; der Agent führt
// ihn gegen seine lokale OBS-Instanz aus.
type RemoteController struct {
	hub *agentlink.Hub
	// agentID benennt den angesprochenen Agent. Leer bedeutet: der primäre,
	// also der aus agent.primary_id oder der erste verbundene mit OBS.
	agentID string
	timeout time.Duration
}

// NewRemoteController creates an OBS controller for the primary agent.
func NewRemoteController(hub *agentlink.Hub, timeout time.Duration) *RemoteController {
	return NewRemoteControllerFor(hub, "", timeout)
}

// NewRemoteControllerFor creates an OBS controller for one specific agent.
// Ein leerer agentID spricht den primären Agent an.
func NewRemoteControllerFor(hub *agentlink.Hub, agentID string, timeout time.Duration) *RemoteController {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	return &RemoteController{hub: hub, agentID: agentID, timeout: timeout}
}

// conn löst den zuständigen Agent auf.
//
// Die Auflösung passiert bei jedem Aufruf neu und nicht einmalig im
// Konstruktor: Verbindet sich ein Agent nach einem Abriss erneut, entsteht
// eine neue Verbindung — ein festgehaltener Zeiger zeigte dann ins Leere.
func (r *RemoteController) conn() (*agentlink.Conn, bool) {
	if r.agentID == "" {
		return r.hub.OBSAgent()
	}

	return r.hub.Get(r.agentID)
}

// IsConnected meldet, ob ein Agent verbunden ist, dessen OBS-Verbindung steht.
func (r *RemoteController) IsConnected() bool {
	conn, ok := r.conn()
	if !ok {
		return false
	}

	return conn.Hello().OBSActive
}

// AgentID liefert die Kennung des steuernden Agents, oder "" wenn keiner da ist.
func (r *RemoteController) AgentID() string {
	conn, ok := r.conn()
	if !ok {
		return ""
	}
	return conn.AgentID()
}

// call schickt ein Kommando an den zuständigen Agent.
func (r *RemoteController) call(method string, params interface{}) (json.RawMessage, error) {
	conn, ok := r.conn()
	if !ok {
		return nil, ErrNoAgent
	}

	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()

	result, err := conn.Call(ctx, method, params)
	if err != nil {
		if ctx.Err() != nil {
			return nil, fmt.Errorf("agent %s did not respond within %s", conn.AgentID(), r.timeout)
		}
		return nil, err
	}

	return result, nil
}

// callInto schickt ein Kommando und dekodiert das Ergebnis nach out.
func callInto[T any](r *RemoteController, method string, params interface{}) (T, error) {
	var out T

	raw, err := r.call(method, params)
	if err != nil {
		return out, err
	}

	if len(raw) == 0 {
		return out, nil
	}

	if err := json.Unmarshal(raw, &out); err != nil {
		return out, fmt.Errorf("failed to decode agent response for %s: %w", method, err)
	}

	return out, nil
}

// GetStatus returns the overall OBS status as reported by the agent
func (r *RemoteController) GetStatus() (*OBSStatus, error) {
	if _, ok := r.conn(); !ok {
		// Kein Agent verbunden ist ein normaler Betriebszustand, kein Fehler —
		// die UI soll "nicht verbunden" anzeigen können.
		return &OBSStatus{Connected: false}, nil
	}

	status, err := callInto[*OBSStatus](r, agentlink.MethodOBSStatus, nil)
	if err != nil {
		return nil, err
	}

	if status == nil {
		return &OBSStatus{Connected: false}, nil
	}

	return status, nil
}

// GetScenes returns all available scenes
func (r *RemoteController) GetScenes() ([]Scene, error) {
	return callInto[[]Scene](r, agentlink.MethodOBSScenes, nil)
}

// GetCurrentScene returns the current active scene
func (r *RemoteController) GetCurrentScene() (string, error) {
	return callInto[string](r, agentlink.MethodOBSCurrentScene, nil)
}

// SwitchScene switches to a different scene
func (r *RemoteController) SwitchScene(sceneName string) error {
	_, err := r.call(agentlink.MethodOBSSwitchScene, agentlink.SwitchSceneParams{
		SceneName: sceneName,
	})
	return err
}

// GetSources returns all sources in the given scene
func (r *RemoteController) GetSources(sceneName string) ([]Source, error) {
	return callInto[[]Source](r, agentlink.MethodOBSSources, agentlink.SourcesParams{
		SceneName: sceneName,
	})
}

// ToggleSourceVisibility toggles the visibility of a source
func (r *RemoteController) ToggleSourceVisibility(sceneName, sourceName string, visible bool) error {
	_, err := r.call(agentlink.MethodOBSToggleSource, agentlink.ToggleSourceParams{
		SceneName:  sceneName,
		SourceName: sourceName,
		Visible:    visible,
	})
	return err
}

// StartStreaming starts the OBS stream
func (r *RemoteController) StartStreaming() error {
	_, err := r.call(agentlink.MethodOBSStreamStart, nil)
	return err
}

// StopStreaming stops the OBS stream
func (r *RemoteController) StopStreaming() error {
	_, err := r.call(agentlink.MethodOBSStreamStop, nil)
	return err
}

// StartRecording starts OBS recording
func (r *RemoteController) StartRecording() error {
	_, err := r.call(agentlink.MethodOBSRecordStart, nil)
	return err
}

// StopRecording stops OBS recording
func (r *RemoteController) StopRecording() error {
	_, err := r.call(agentlink.MethodOBSRecordStop, nil)
	return err
}

// PauseRecording pauses OBS recording
func (r *RemoteController) PauseRecording() error {
	_, err := r.call(agentlink.MethodOBSRecordPause, nil)
	return err
}

// ResumeRecording resumes OBS recording
func (r *RemoteController) ResumeRecording() error {
	_, err := r.call(agentlink.MethodOBSRecordResume, nil)
	return err
}
