package obs

import (
	"errors"
	"fmt"
	"time"

	"kit.workmate/live-portal/internal/agentlink"
)

// ErrAgentNotFound meldet, dass der angesprochene Agent nicht verbunden ist.
//
// Bewusst getrennt von ErrNoAgent: "der Rechner, den du meinst, ist weg" ist
// etwas anderes als "es ist überhaupt keiner da", und die Oberfläche soll das
// unterscheiden können.
var ErrAgentNotFound = errors.New("agent not connected")

// ErrAgentNotAddressable meldet, dass in dieser Betriebsart kein bestimmter
// Agent angesprochen werden kann.
var ErrAgentNotAddressable = errors.New("agent addressing requires obs.mode: agent")

// Resolver bildet eine Agent-Kennung auf den zuständigen OBS-Controller ab.
// Ein leerer agentID bedeutet immer: der primäre Agent.
type Resolver interface {
	ControllerFor(agentID string) (Controller, error)
}

// LinkResolver adressiert Agents über den Link.
type LinkResolver struct {
	hub     *agentlink.Hub
	timeout time.Duration
}

// NewLinkResolver creates a resolver backed by the agent link.
func NewLinkResolver(hub *agentlink.Hub, timeout time.Duration) *LinkResolver {
	return &LinkResolver{hub: hub, timeout: timeout}
}

func (r *LinkResolver) ControllerFor(agentID string) (Controller, error) {
	// Ohne Angabe gilt der primäre Agent — so verhalten sich die bisherigen
	// Routen unverändert weiter.
	if agentID == "" {
		return NewRemoteController(r.hub, r.timeout), nil
	}

	// Frühe Prüfung, damit ein Tippfehler im Agentnamen sofort als solcher
	// erkennbar ist und nicht erst als Zeitüberschreitung beim Kommando.
	if _, ok := r.hub.Get(agentID); !ok {
		return nil, fmt.Errorf("%w: %s", ErrAgentNotFound, agentID)
	}

	return NewRemoteControllerFor(r.hub, agentID, r.timeout), nil
}

// DirectResolver steht für die direkte OBS-Verbindung des Portals
// (obs.mode: direct). Dort gibt es genau eine Instanz und keine Agents.
type DirectResolver struct {
	client  Controller
	agentID string
}

// NewDirectResolver creates a resolver for the portal's own OBS connection.
// agentID ist die Kennung, unter der diese Instanz ansprechbar ist.
func NewDirectResolver(client Controller, agentID string) *DirectResolver {
	return &DirectResolver{client: client, agentID: agentID}
}

func (r *DirectResolver) ControllerFor(agentID string) (Controller, error) {
	if agentID == "" || agentID == r.agentID {
		return r.client, nil
	}

	return nil, fmt.Errorf("%w (asked for %q)", ErrAgentNotAddressable, agentID)
}
