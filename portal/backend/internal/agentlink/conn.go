package agentlink

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingInterval   = 25 * time.Second
	outboundBuffer = 64
)

// ErrAgentGone wird zurückgegeben, wenn die Verbindung während eines Calls endet.
var ErrAgentGone = fmt.Errorf("agent disconnected")

// Conn ist eine aktive Verbindung zu genau einem Agent. Sie korreliert
// ausgehende Kommandos mit den zugehörigen Antworten über die Frame-ID.
type Conn struct {
	ws       *websocket.Conn
	agentID  string
	hello    Hello
	joinedAt time.Time

	outbound chan Frame

	mu      sync.Mutex
	pending map[string]chan Frame
	closed  bool

	closeOnce sync.Once
	done      chan struct{}
}

func newConn(ws *websocket.Conn, agentID string) *Conn {
	return &Conn{
		ws:       ws,
		agentID:  agentID,
		joinedAt: time.Now(),
		outbound: make(chan Frame, outboundBuffer),
		pending:  make(map[string]chan Frame),
		done:     make(chan struct{}),
	}
}

// AgentID liefert die Kennung des verbundenen Agents.
func (c *Conn) AgentID() string { return c.agentID }

// Hello liefert die Selbstbeschreibung des Agents aus dem Verbindungsaufbau.
func (c *Conn) Hello() Hello {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.hello
}

// Info beschreibt den Agent für die API.
func (c *Conn) Info() AgentInfo {
	h := c.Hello()
	return AgentInfo{
		AgentID:   c.agentID,
		Hostname:  h.Hostname,
		Version:   h.Version,
		Commit:    h.Commit,
		OBSLocal:  h.OBSLocal,
		OBSActive: h.OBSActive,
		Connected: !c.isClosed(),
		Since:     c.joinedAt,
	}
}

func (c *Conn) isClosed() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.closed
}

// Call schickt ein Kommando an den Agent und wartet auf die Antwort.
func (c *Conn) Call(ctx context.Context, method string, params interface{}) (json.RawMessage, error) {
	id, err := newID()
	if err != nil {
		return nil, err
	}

	frame := Frame{ID: id, Type: FrameTypeCommand, Method: method}
	if params != nil {
		raw, err := json.Marshal(params)
		if err != nil {
			return nil, fmt.Errorf("failed to encode params: %w", err)
		}
		frame.Params = raw
	}

	replyCh := make(chan Frame, 1)

	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return nil, ErrAgentGone
	}
	c.pending[id] = replyCh
	c.mu.Unlock()

	defer func() {
		c.mu.Lock()
		delete(c.pending, id)
		c.mu.Unlock()
	}()

	select {
	case c.outbound <- frame:
	case <-c.done:
		return nil, ErrAgentGone
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	select {
	case reply := <-replyCh:
		if !reply.OK {
			if reply.Error == "" {
				return nil, fmt.Errorf("agent reported failure")
			}
			return nil, fmt.Errorf("%s", reply.Error)
		}
		return reply.Result, nil
	case <-c.done:
		return nil, ErrAgentGone
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// close beendet die Verbindung und weckt alle wartenden Calls.
func (c *Conn) close() {
	c.closeOnce.Do(func() {
		c.mu.Lock()
		c.closed = true
		c.mu.Unlock()

		close(c.done)
		_ = c.ws.Close()
	})
}

func (c *Conn) writePump() {
	ticker := time.NewTicker(pingInterval)
	defer ticker.Stop()
	defer c.close()

	for {
		select {
		case <-c.done:
			_ = c.ws.SetWriteDeadline(time.Now().Add(writeWait))
			_ = c.ws.WriteMessage(websocket.CloseMessage,
				websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			return

		case frame := <-c.outbound:
			_ = c.ws.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.ws.WriteJSON(frame); err != nil {
				log.Printf("agentlink: write to %s failed: %v", c.agentID, err)
				return
			}

		case <-ticker.C:
			_ = c.ws.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.ws.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// readPump verarbeitet eingehende Frames, bis die Verbindung endet.
func (c *Conn) readPump(onEvent EventHandler) {
	defer c.close()

	_ = c.ws.SetReadDeadline(time.Now().Add(pongWait))
	c.ws.SetPongHandler(func(string) error {
		return c.ws.SetReadDeadline(time.Now().Add(pongWait))
	})

	for {
		var frame Frame
		if err := c.ws.ReadJSON(&frame); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				log.Printf("agentlink: read from %s failed: %v", c.agentID, err)
			}
			return
		}

		_ = c.ws.SetReadDeadline(time.Now().Add(pongWait))

		switch frame.Type {
		case FrameTypeResponse:
			c.deliver(frame)

		case FrameTypeHello:
			var hello Hello
			if err := json.Unmarshal(frame.Data, &hello); err != nil {
				log.Printf("agentlink: invalid hello from %s: %v", c.agentID, err)
				continue
			}
			c.mu.Lock()
			c.hello = hello
			c.mu.Unlock()
			log.Printf("agentlink: agent %s ready (host=%s version=%s obs=%v)",
				c.agentID, hello.Hostname, hello.Version, hello.OBSLocal)

		case FrameTypeEvent:
			// Der Agent meldet über obs_event auch Verbindungswechsel — die
			// zwischengespeicherte Hello-Info entsprechend nachziehen.
			if frame.Event == EventOBS {
				c.trackOBSConnection(frame.Data)
			}
			if onEvent != nil {
				onEvent(c.agentID, frame.Event, frame.Data)
			}
		}
	}
}

// trackOBSConnection hält Hello.OBSActive aktuell, damit die Agentenliste den
// tatsächlichen OBS-Zustand zeigt.
func (c *Conn) trackOBSConnection(data json.RawMessage) {
	var payload struct {
		Type      string `json:"type"`
		Connected bool   `json:"connected"`
	}
	if err := json.Unmarshal(data, &payload); err != nil || payload.Type != "connection_changed" {
		return
	}

	c.mu.Lock()
	c.hello.OBSActive = payload.Connected
	c.mu.Unlock()
}

func (c *Conn) deliver(frame Frame) {
	c.mu.Lock()
	ch, ok := c.pending[frame.ID]
	c.mu.Unlock()

	if !ok {
		// Antwort auf einen bereits abgelaufenen Call — verwerfen.
		return
	}

	select {
	case ch <- frame:
	default:
	}
}

func newID() (string, error) {
	buf := make([]byte, 12)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("failed to generate request id: %w", err)
	}
	return hex.EncodeToString(buf), nil
}
