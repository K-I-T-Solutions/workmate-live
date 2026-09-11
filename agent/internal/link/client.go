package link

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"kit.workmate/live-agent/internal/buildinfo"
	"kit.workmate/live-agent/internal/config"
	"kit.workmate/live-agent/internal/health"
	"kit.workmate/live-agent/internal/services/obs"
)

const (
	// writeWait ist das Limit für einen einzelnen Schreibvorgang.
	writeWait = 10 * time.Second
	// pongWait ist die Frist, innerhalb derer das Portal auf einen Ping antworten muss.
	pongWait = 60 * time.Second
	// pingInterval muss kleiner als pongWait sein.
	pingInterval = 25 * time.Second
	// maxBackoff deckelt die Wartezeit zwischen Reconnect-Versuchen.
	maxBackoff = 2 * time.Minute
	// statusInterval bestimmt, wie oft der Agent seinen Health-Status pusht.
	statusInterval = 3 * time.Second
	// outboundBuffer ist die Sendepuffergröße pro Verbindung.
	outboundBuffer = 64
)

// OBSController ist der Teil des OBS-Clients, den der Link benötigt.
type OBSController interface {
	IsConnected() bool
	GetStatus() (*obs.OBSStatus, error)
	GetScenes() ([]obs.Scene, error)
	GetCurrentScene() (string, error)
	SwitchScene(sceneName string) error
	GetSources(sceneName string) ([]obs.Source, error)
	ToggleSourceVisibility(sceneName, sourceName string, visible bool) error
	StartStreaming() error
	StopStreaming() error
	StartRecording() error
	StopRecording() error
	PauseRecording() error
	ResumeRecording() error
}

// Client hält eine ausgehende WebSocket-Verbindung zum Portal offen. Das Portal
// schickt darüber Kommandos, der Agent führt sie gegen das lokale OBS aus und
// pusht Events zurück. Da der Agent die Verbindung aufbaut, braucht der
// Streaming-Rechner keinen eingehenden Port.
type Client struct {
	cfg     config.PortalConfig
	agentID string
	obs     OBSController
	cache   *health.Cache

	// obsLocal gibt an, ob dieser Agent überhaupt eine OBS-Instanz steuert.
	obsLocal bool

	mu       sync.Mutex
	outbound chan Frame
}

// NewClient creates a new portal link client
func NewClient(cfg config.PortalConfig, agentID string, obsCtrl OBSController, cache *health.Cache) *Client {
	return &Client{
		cfg:      cfg,
		agentID:  agentID,
		obs:      obsCtrl,
		cache:    cache,
		obsLocal: obsCtrl != nil,
	}
}

// Run hält die Portal-Verbindung bis zum Abbruch des Kontexts offen.
// Blockiert; per Goroutine starten.
func (c *Client) Run(ctx context.Context) {
	endpoint, err := c.endpoint()
	if err != nil {
		log.Printf("link: invalid portal URL: %v — link disabled", err)
		return
	}

	log.Printf("link: connecting to portal at %s", c.cfg.URL)

	attempt := 0
	for {
		if err := c.session(ctx, endpoint); err != nil {
			if ctx.Err() != nil {
				return
			}
			attempt++
			delay := c.backoff(attempt)
			log.Printf("link: %v (retry %d in %s)", err, attempt, delay)

			select {
			case <-ctx.Done():
				return
			case <-time.After(delay):
			}
			continue
		}

		// Saubere Trennung: sofort neu verbinden.
		if ctx.Err() != nil {
			return
		}
		attempt = 0
		select {
		case <-ctx.Done():
			return
		case <-time.After(c.retryDelay()):
		}
	}
}

// endpoint baut die WebSocket-URL aus der konfigurierten Portal-URL.
func (c *Client) endpoint() (string, error) {
	u, err := url.Parse(strings.TrimSpace(c.cfg.URL))
	if err != nil {
		return "", err
	}

	switch u.Scheme {
	case "http", "ws":
		u.Scheme = "ws"
	case "https", "wss":
		u.Scheme = "wss"
	default:
		return "", fmt.Errorf("unsupported scheme %q (use http:// or https://)", u.Scheme)
	}

	if u.Host == "" {
		return "", fmt.Errorf("missing host in portal URL")
	}

	u.Path = strings.TrimSuffix(u.Path, "/") + "/ws/agent"
	u.RawQuery = ""
	u.Fragment = ""

	return u.String(), nil
}

func (c *Client) retryDelay() time.Duration {
	if c.cfg.RetryDelay > 0 {
		return c.cfg.RetryDelay
	}
	return 5 * time.Second
}

// backoff verdoppelt die Wartezeit pro Fehlversuch bis zu retry_attempts
// Verdopplungen, gedeckelt auf maxBackoff.
func (c *Client) backoff(attempt int) time.Duration {
	base := c.retryDelay()

	maxDoublings := c.cfg.RetryAttempts
	if maxDoublings < 0 {
		maxDoublings = 0
	}
	if attempt-1 < maxDoublings {
		maxDoublings = attempt - 1
	}

	delay := base
	for i := 0; i < maxDoublings; i++ {
		delay *= 2
		if delay >= maxBackoff {
			return maxBackoff
		}
	}

	return delay
}

// session baut eine Verbindung auf und bedient sie, bis sie abreißt.
func (c *Client) session(ctx context.Context, endpoint string) error {
	dialer := websocket.Dialer{
		HandshakeTimeout: c.cfg.Timeout,
	}

	header := http.Header{}
	if c.cfg.APIKey != "" {
		header.Set("X-Agent-Key", c.cfg.APIKey)
	}
	header.Set("X-Agent-ID", c.agentID)

	conn, resp, err := dialer.DialContext(ctx, endpoint, header)
	if err != nil {
		if resp != nil {
			return fmt.Errorf("dial failed: %w (HTTP %d)", err, resp.StatusCode)
		}
		return fmt.Errorf("dial failed: %w", err)
	}
	defer conn.Close()

	log.Printf("link: connected to portal")

	sessionCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	outbound := make(chan Frame, outboundBuffer)
	c.mu.Lock()
	c.outbound = outbound
	c.mu.Unlock()

	defer func() {
		c.mu.Lock()
		c.outbound = nil
		c.mu.Unlock()
	}()

	// OBS-Events auf die Verbindung leiten.
	if setter, ok := c.obs.(interface {
		SetEventCallback(func(map[string]interface{}))
	}); ok {
		setter.SetEventCallback(func(event map[string]interface{}) {
			c.send(Frame{Type: FrameTypeEvent, Event: EventOBS, Data: mustJSON(event)})
		})
		defer setter.SetEventCallback(nil)
	}

	var wg sync.WaitGroup
	wg.Add(2)
	go func() { defer wg.Done(); c.writePump(sessionCtx, cancel, conn, outbound) }()
	go func() { defer wg.Done(); c.statusPump(sessionCtx) }()

	c.send(Frame{Type: FrameTypeHello, Data: mustJSON(c.hello())})

	err = c.readPump(sessionCtx, conn)

	cancel()
	wg.Wait()

	if err != nil {
		return fmt.Errorf("link closed: %w", err)
	}
	return nil
}

func (c *Client) hello() Hello {
	// Direkt vom System lesen statt aus dem Health-Cache: beim ersten Verbinden
	// hat der Poller unter Umständen noch keinen Status geschrieben.
	hostname, err := os.Hostname()
	if err != nil {
		if status := c.cache.Get(); status != nil {
			hostname = status.Hostname
		}
	}

	return Hello{
		AgentID:   c.agentID,
		Hostname:  hostname,
		Version:   buildinfo.Version,
		Commit:    buildinfo.Commit,
		OBSLocal:  c.obsLocal,
		OBSActive: c.obsLocal && c.obs.IsConnected(),
	}
}

// send stellt einen Frame in den Sendepuffer. Ist der Puffer voll oder besteht
// keine Verbindung, wird der Frame verworfen — Events sind nicht kritisch genug,
// um den Aufrufer zu blockieren.
func (c *Client) send(f Frame) {
	c.mu.Lock()
	out := c.outbound
	c.mu.Unlock()

	if out == nil {
		return
	}

	select {
	case out <- f:
	default:
		log.Printf("link: outbound buffer full, dropping %s frame", f.Type)
	}
}

func (c *Client) writePump(ctx context.Context, cancel context.CancelFunc, conn *websocket.Conn, outbound <-chan Frame) {
	ticker := time.NewTicker(pingInterval)
	defer ticker.Stop()
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			_ = conn.SetWriteDeadline(time.Now().Add(writeWait))
			_ = conn.WriteMessage(websocket.CloseMessage,
				websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			return

		case frame := <-outbound:
			_ = conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := conn.WriteJSON(frame); err != nil {
				log.Printf("link: write failed: %v", err)
				return
			}

		case <-ticker.C:
			_ = conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// statusPump pusht den Health-Status periodisch ans Portal, damit dieses den
// Agent nicht mehr über HTTP pollen muss.
func (c *Client) statusPump(ctx context.Context) {
	ticker := time.NewTicker(statusInterval)
	defer ticker.Stop()

	// Sofort einmal senden, damit das Portal nach einem Reconnect nicht bis
	// zum ersten Tick ohne Status dasteht.
	c.pushStatus()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.pushStatus()
		}
	}
}

func (c *Client) pushStatus() {
	if status := c.cache.Get(); status != nil {
		c.send(Frame{Type: FrameTypeEvent, Event: EventAgentStatus, Data: mustJSON(status)})
	}
}

func (c *Client) readPump(ctx context.Context, conn *websocket.Conn) error {
	_ = conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error {
		return conn.SetReadDeadline(time.Now().Add(pongWait))
	})

	for {
		var frame Frame
		if err := conn.ReadJSON(&frame); err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return err
		}

		_ = conn.SetReadDeadline(time.Now().Add(pongWait))

		if frame.Type != FrameTypeCommand {
			continue
		}

		// Nebenläufig ausführen, damit ein hängender OBS-Call den Lesepfad
		// nicht blockiert.
		go c.handleCommand(frame)
	}
}

func (c *Client) handleCommand(frame Frame) {
	result, err := c.execute(frame)

	resp := Frame{ID: frame.ID, Type: FrameTypeResponse}
	if err != nil {
		resp.Error = err.Error()
	} else {
		resp.OK = true
		resp.Result = result
	}

	c.send(resp)
}

func (c *Client) execute(frame Frame) (json.RawMessage, error) {
	if c.obs == nil {
		return nil, fmt.Errorf("agent does not control an OBS instance")
	}

	switch frame.Method {
	case MethodOBSStatus:
		status, err := c.obs.GetStatus()
		return encode(status, err)

	case MethodOBSScenes:
		scenes, err := c.obs.GetScenes()
		return encode(scenes, err)

	case MethodOBSCurrentScene:
		scene, err := c.obs.GetCurrentScene()
		return encode(scene, err)

	case MethodOBSSwitchScene:
		var p SwitchSceneParams
		if err := json.Unmarshal(frame.Params, &p); err != nil {
			return nil, fmt.Errorf("invalid params: %w", err)
		}
		return nil, c.obs.SwitchScene(p.SceneName)

	case MethodOBSSources:
		var p SourcesParams
		if err := json.Unmarshal(frame.Params, &p); err != nil {
			return nil, fmt.Errorf("invalid params: %w", err)
		}
		sources, err := c.obs.GetSources(p.SceneName)
		return encode(sources, err)

	case MethodOBSToggleSource:
		var p ToggleSourceParams
		if err := json.Unmarshal(frame.Params, &p); err != nil {
			return nil, fmt.Errorf("invalid params: %w", err)
		}
		return nil, c.obs.ToggleSourceVisibility(p.SceneName, p.SourceName, p.Visible)

	case MethodOBSStreamStart:
		return nil, c.obs.StartStreaming()
	case MethodOBSStreamStop:
		return nil, c.obs.StopStreaming()
	case MethodOBSRecordStart:
		return nil, c.obs.StartRecording()
	case MethodOBSRecordStop:
		return nil, c.obs.StopRecording()
	case MethodOBSRecordPause:
		return nil, c.obs.PauseRecording()
	case MethodOBSRecordResume:
		return nil, c.obs.ResumeRecording()
	}

	return nil, fmt.Errorf("unknown method: %s", frame.Method)
}

func encode(v interface{}, err error) (json.RawMessage, error) {
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("failed to encode result: %w", err)
	}

	return data, nil
}

func mustJSON(v interface{}) json.RawMessage {
	data, err := json.Marshal(v)
	if err != nil {
		return json.RawMessage(`null`)
	}
	return data
}
