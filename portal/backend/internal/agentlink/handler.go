package agentlink

import (
	"crypto/subtle"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

// Handler nimmt eingehende Agent-Verbindungen entgegen. Agents authentifizieren
// sich mit einem gemeinsamen API-Key — nicht mit dem Benutzer-JWT, da die
// Verbindung unabhängig von einer Benutzersitzung besteht.
type Handler struct {
	hub      *Hub
	apiKey   string
	upgrader websocket.Upgrader
}

// NewHandler creates the /ws/agent handler. Ein leerer apiKey deaktiviert den
// Endpunkt, damit ein unkonfiguriertes Portal nicht offen steht.
func NewHandler(hub *Hub, apiKey string) *Handler {
	return &Handler{
		hub:    hub,
		apiKey: apiKey,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  4096,
			WriteBufferSize: 4096,
			// Agents sind keine Browser — es gibt keinen sinnvollen Origin
			// zu prüfen. Zugang regelt ausschließlich der API-Key.
			CheckOrigin: func(r *http.Request) bool { return true },
		},
	}
}

// Enabled meldet, ob der Agent-Endpunkt nutzbar konfiguriert ist.
func (h *Handler) Enabled() bool {
	return h.apiKey != ""
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !h.Enabled() {
		http.Error(w, "Agent link not configured", http.StatusServiceUnavailable)
		return
	}

	if !h.authorized(r) {
		log.Printf("agentlink: rejected connection from %s (bad key)", r.RemoteAddr)
		http.Error(w, "Invalid agent key", http.StatusUnauthorized)
		return
	}

	agentID := r.Header.Get("X-Agent-ID")
	if agentID == "" {
		agentID = r.URL.Query().Get("agent_id")
	}
	if agentID == "" {
		http.Error(w, "Missing agent id", http.StatusBadRequest)
		return
	}

	ws, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		// Upgrade schreibt die Fehlerantwort bereits selbst.
		log.Printf("agentlink: upgrade failed for %s: %v", agentID, err)
		return
	}

	conn := newConn(ws, agentID)
	h.hub.register(conn)

	go conn.writePump()

	// readPump blockiert bis zum Verbindungsende; danach aufräumen.
	go func() {
		defer h.hub.unregister(conn)
		conn.readPump(h.hub.eventHandler())
	}()
}

// authorized prüft den API-Key aus Header oder Query in konstanter Zeit.
func (h *Handler) authorized(r *http.Request) bool {
	key := r.Header.Get("X-Agent-Key")
	if key == "" {
		key = r.URL.Query().Get("key")
	}

	return subtle.ConstantTimeCompare([]byte(key), []byte(h.apiKey)) == 1
}
