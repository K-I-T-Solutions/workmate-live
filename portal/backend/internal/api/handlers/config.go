package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"kit.workmate/live-portal/internal/config"
)

type ConfigHandler struct {
	cfg *config.Config
	mu  sync.Mutex
}

func NewConfigHandler(cfg *config.Config) *ConfigHandler {
	return &ConfigHandler{
		cfg: cfg,
	}
}

// GetConfig returns the configuration with masked secrets
func (h *ConfigHandler) GetConfig(w http.ResponseWriter, r *http.Request) {
	h.mu.Lock()
	safe := h.cfg.ToSafe()
	h.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(safe)
}

// maskedSecret ist der Platzhalter, den GetConfig statt echter Geheimnisse
// ausliefert. Kommt er zurueck, war das Feld in der UI unveraendert.
const maskedSecret = "***"

type updateConfigRequest struct {
	Section string          `json:"section"`
	Data    json.RawMessage `json:"data"`
}

// UpdateConfig updates a specific section of the configuration
func (h *ConfigHandler) UpdateConfig(w http.ResponseWriter, r *http.Request) {
	var req updateConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	// Auf einer Kopie arbeiten: wird die Konfiguration abgelehnt, darf die
	// laufende Instanz nicht mit ungueltigen Werten zurueckbleiben.
	// Config enthaelt ausschliesslich Wertetypen, eine flache Kopie genuegt.
	updated := *h.cfg

	switch req.Section {
	case "obs":
		var data config.OBSConfig
		if err := json.Unmarshal(req.Data, &data); err != nil {
			http.Error(w, fmt.Sprintf("Invalid OBS config: %v", err), http.StatusBadRequest)
			return
		}
		// Preserve existing secrets if masked values sent
		if data.Password == maskedSecret {
			data.Password = h.cfg.OBS.Password
		}
		updated.OBS = data

	case "twitch":
		var data config.TwitchConfig
		if err := json.Unmarshal(req.Data, &data); err != nil {
			http.Error(w, fmt.Sprintf("Invalid Twitch config: %v", err), http.StatusBadRequest)
			return
		}
		// Preserve existing secrets if masked values sent
		if data.ClientSecret == maskedSecret {
			data.ClientSecret = h.cfg.Twitch.ClientSecret
		}
		if data.OAuthToken == maskedSecret {
			data.OAuthToken = h.cfg.Twitch.OAuthToken
		}
		updated.Twitch = data

	case "youtube":
		var data config.YouTubeConfig
		if err := json.Unmarshal(req.Data, &data); err != nil {
			http.Error(w, fmt.Sprintf("Invalid YouTube config: %v", err), http.StatusBadRequest)
			return
		}
		// Preserve existing secrets if masked values sent
		if data.APIKey == maskedSecret {
			data.APIKey = h.cfg.YouTube.APIKey
		}
		if data.ClientSecret == maskedSecret {
			data.ClientSecret = h.cfg.YouTube.ClientSecret
		}
		updated.YouTube = data

	case "auth":
		var data struct {
			DefaultUser config.DefaultUserConfig `json:"default_user"`
		}
		if err := json.Unmarshal(req.Data, &data); err != nil {
			http.Error(w, fmt.Sprintf("Invalid auth config: %v", err), http.StatusBadRequest)
			return
		}
		// Preserve existing password if masked value sent
		if data.DefaultUser.Password == maskedSecret {
			data.DefaultUser.Password = h.cfg.Auth.DefaultUser.Password
		}
		updated.Auth.DefaultUser = data.DefaultUser

	case "agent":
		var data config.AgentConfig
		if err := json.Unmarshal(req.Data, &data); err != nil {
			http.Error(w, fmt.Sprintf("Invalid agent config: %v", err), http.StatusBadRequest)
			return
		}
		// Preserve existing secrets if masked values sent
		if data.APIKey == maskedSecret {
			data.APIKey = h.cfg.Agent.APIKey
		}
		updated.Agent = data

	default:
		http.Error(w, fmt.Sprintf("Unknown or protected section: %s", req.Section), http.StatusBadRequest)
		return
	}

	// Erst pruefen, dann uebernehmen: eine ungueltige Kombination (etwa
	// obs.mode=agent ohne Agent-Schluessel) wuerde den naechsten Start
	// verhindern.
	if err := updated.Validate(); err != nil {
		http.Error(w, fmt.Sprintf("Invalid configuration: %v", err), http.StatusBadRequest)
		return
	}

	*h.cfg = updated

	if err := h.cfg.Save(); err != nil {
		http.Error(w, fmt.Sprintf("Failed to save config: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
