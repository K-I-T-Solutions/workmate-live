package obs

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"

	"kit.workmate/live-portal/internal/agentlink"
)

// connectAgent hängt einen minimalen Agent an einen Hub.
func connectAgent(t *testing.T, srv *httptest.Server, hub *agentlink.Hub, agentID string) {
	t.Helper()

	header := http.Header{}
	header.Set("X-Agent-Key", testKey)
	header.Set("X-Agent-ID", agentID)

	ws, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(srv.URL, "http"), header)
	if err != nil {
		t.Fatalf("dial failed for %s: %v", agentID, err)
	}
	t.Cleanup(func() { ws.Close() })

	hello, _ := json.Marshal(agentlink.Hello{AgentID: agentID, OBSLocal: true, OBSActive: true})
	if err := ws.WriteJSON(agentlink.Frame{Type: agentlink.FrameTypeHello, Data: hello}); err != nil {
		t.Fatalf("hello failed: %v", err)
	}

	// Frames schlucken, damit die Verbindung nicht am Lesefehler stirbt.
	go func() {
		for {
			var f agentlink.Frame
			if err := ws.ReadJSON(&f); err != nil {
				return
			}
		}
	}()

	// Auf die verarbeitete Anmeldung warten, nicht nur auf die Registrierung:
	// Hello wird asynchron gelesen, vorher steht OBSActive noch nicht.
	deadline := time.Now().Add(2 * time.Second)
	for {
		if conn, ok := hub.Get(agentID); ok && conn.Hello().AgentID == agentID {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("agent %s did not register", agentID)
		}
		time.Sleep(5 * time.Millisecond)
	}
}

func TestLinkResolverAddressesSpecificAgent(t *testing.T) {
	hub := agentlink.NewHub("")
	srv := httptest.NewServer(agentlink.NewHandler(hub, testKey))
	t.Cleanup(srv.Close)

	connectAgent(t, srv, hub, "cisco")
	connectAgent(t, srv, hub, "barry")

	resolver := NewLinkResolver(hub, time.Second)

	for _, want := range []string{"cisco", "barry"} {
		ctrl, err := resolver.ControllerFor(want)
		if err != nil {
			t.Fatalf("ControllerFor(%q) failed: %v", want, err)
		}

		remote, ok := ctrl.(*RemoteController)
		if !ok {
			t.Fatalf("expected a RemoteController for %q", want)
		}
		if got := remote.AgentID(); got != want {
			t.Errorf("controller addresses %q, want %q", got, want)
		}
	}
}

// Ohne Kennung gilt der primäre Agent — so verhalten sich die bisherigen
// Routen unverändert.
func TestEmptyAgentIDUsesPrimary(t *testing.T) {
	hub := agentlink.NewHub("barry")
	srv := httptest.NewServer(agentlink.NewHandler(hub, testKey))
	t.Cleanup(srv.Close)

	connectAgent(t, srv, hub, "cisco")
	connectAgent(t, srv, hub, "barry")

	ctrl, err := NewLinkResolver(hub, time.Second).ControllerFor("")
	if err != nil {
		t.Fatalf("ControllerFor(\"\") failed: %v", err)
	}

	if got := ctrl.(*RemoteController).AgentID(); got != "barry" {
		t.Errorf("primary resolved to %q, want %q", got, "barry")
	}
}

// Ein Tippfehler im Agentnamen muss sofort auffallen, nicht erst als
// Zeitüberschreitung beim Kommando.
func TestUnknownAgentIsRejectedEarly(t *testing.T) {
	hub := agentlink.NewHub("")
	srv := httptest.NewServer(agentlink.NewHandler(hub, testKey))
	t.Cleanup(srv.Close)

	connectAgent(t, srv, hub, "cisco")

	_, err := NewLinkResolver(hub, time.Second).ControllerFor("ciso")
	if !errors.Is(err, ErrAgentNotFound) {
		t.Fatalf("error = %v, want ErrAgentNotFound", err)
	}
	if !strings.Contains(err.Error(), "ciso") {
		t.Errorf("error %q should name the agent that was asked for", err)
	}
}

// Verbindet sich ein Agent neu, muss der Controller die neue Verbindung
// nutzen — deshalb wird pro Aufruf aufgelöst statt einmalig.
func TestControllerFollowsReconnect(t *testing.T) {
	hub := agentlink.NewHub("")
	srv := httptest.NewServer(agentlink.NewHandler(hub, testKey))
	t.Cleanup(srv.Close)

	connectAgent(t, srv, hub, "cisco")

	ctrl, err := NewLinkResolver(hub, time.Second).ControllerFor("cisco")
	if err != nil {
		t.Fatalf("ControllerFor failed: %v", err)
	}

	if !ctrl.IsConnected() {
		t.Fatal("expected the agent to be connected")
	}

	// Neu verbinden — der Hub verdrängt dabei die alte Verbindung.
	old, _ := hub.Get("cisco")
	connectAgent(t, srv, hub, "cisco")

	deadline := time.Now().Add(2 * time.Second)
	for {
		if current, ok := hub.Get("cisco"); ok && current != old {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("reconnect did not replace the connection")
		}
		time.Sleep(5 * time.Millisecond)
	}

	if !ctrl.IsConnected() {
		t.Error("controller lost the agent after reconnect")
	}
}

func TestDirectResolverRejectsAgentAddressing(t *testing.T) {
	client := &RemoteController{} // Platzhalter, wird nicht aufgerufen
	resolver := NewDirectResolver(client, "local")

	// Ohne Kennung und unter der eigenen Kennung: derselbe Controller.
	for _, id := range []string{"", "local"} {
		got, err := resolver.ControllerFor(id)
		if err != nil {
			t.Fatalf("ControllerFor(%q) failed: %v", id, err)
		}
		if got != Controller(client) {
			t.Errorf("ControllerFor(%q) returned a different controller", id)
		}
	}

	// Ein fremder Agent ist im Direktmodus nicht ansprechbar.
	if _, err := resolver.ControllerFor("barry"); !errors.Is(err, ErrAgentNotAddressable) {
		t.Fatalf("error = %v, want ErrAgentNotAddressable", err)
	}
}
