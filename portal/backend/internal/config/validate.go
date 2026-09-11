package config

import (
	"errors"
	"fmt"
)

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if err := c.Server.Validate(); err != nil {
		return fmt.Errorf("server config: %w", err)
	}

	if err := c.Auth.Validate(); err != nil {
		return fmt.Errorf("auth config: %w", err)
	}

	if err := c.Agent.Validate(); err != nil {
		return fmt.Errorf("agent config: %w", err)
	}

	if err := c.OBS.Validate(); err != nil {
		return fmt.Errorf("obs config: %w", err)
	}

	if err := c.Twitch.Validate(); err != nil {
		return fmt.Errorf("twitch config: %w", err)
	}

	if err := c.Storage.Validate(); err != nil {
		return fmt.Errorf("storage config: %w", err)
	}

	// Der Agent-Modus setzt voraus, dass sich Agents überhaupt verbinden
	// dürfen — ohne API-Key ist der Link-Endpunkt deaktiviert.
	if c.OBS.Mode == OBSModeAgent && c.Agent.APIKey == "" {
		return fmt.Errorf("obs mode %q requires agent.api_key to be set", OBSModeAgent)
	}

	return nil
}

func (s *ServerConfig) Validate() error {
	if s.Port < 1 || s.Port > 65535 {
		return errors.New("port must be between 1 and 65535")
	}

	if s.Timeouts.Read <= 0 {
		return errors.New("read timeout must be positive")
	}

	if s.Timeouts.Write <= 0 {
		return errors.New("write timeout must be positive")
	}

	if s.Timeouts.Shutdown <= 0 {
		return errors.New("shutdown timeout must be positive")
	}

	return nil
}

func (a *AuthConfig) Validate() error {
	if a.JWTSecret == "" {
		return errors.New("JWT secret required")
	}

	if len(a.JWTSecret) < 16 {
		return errors.New("JWT secret must be at least 16 characters")
	}

	if a.TokenDuration <= 0 {
		return errors.New("token duration must be positive")
	}

	if a.DefaultUser.Username == "" {
		return errors.New("default username required")
	}

	if a.DefaultUser.Password == "" {
		return errors.New("default password required")
	}

	return nil
}

func (a *AgentConfig) Validate() error {
	// Eine leere URL ist zulässig: verbindet sich der Agent selbst über den
	// Link, ist er von außen gar nicht erreichbar und wird nicht gepollt.
	if a.URL == "" && a.APIKey == "" {
		return errors.New("agent URL required unless api_key enables the agent link")
	}

	if a.PollingInterval <= 0 {
		return errors.New("polling interval must be positive")
	}

	if a.Timeout <= 0 {
		return errors.New("timeout must be positive")
	}

	if a.CommandTimeout < 0 {
		return errors.New("command timeout must not be negative")
	}

	return nil
}

func (o *OBSConfig) Validate() error {
	switch o.Mode {
	case "", OBSModeDirect, OBSModeAgent:
		// ok — leer wird beim Laden auf "direct" normalisiert
	default:
		return fmt.Errorf("mode must be %q or %q, got %q", OBSModeDirect, OBSModeAgent, o.Mode)
	}

	// Im Agent-Modus verbindet nicht das Portal, sondern der Agent zu OBS —
	// Host und Port aus diesem Block sind dann bedeutungslos.
	if o.Mode == OBSModeAgent {
		return nil
	}

	if o.Port < 1 || o.Port > 65535 {
		return errors.New("port must be between 1 and 65535")
	}

	if o.AutoReconnect && o.ReconnectDelay <= 0 {
		return errors.New("reconnect delay must be positive when auto-reconnect is enabled")
	}

	return nil
}

func (t *TwitchConfig) Validate() error {
	// Only validate if Twitch is enabled
	if !t.Enabled {
		return nil
	}

	if t.ClientID == "" {
		return errors.New("client_id required when Twitch is enabled")
	}

	// Client Secret is optional when using token generators
	// Only required if implementing OAuth flow

	if t.Channel == "" {
		return errors.New("channel required when Twitch is enabled")
	}

	if t.OAuthToken == "" {
		return errors.New("oauth_token required when Twitch is enabled")
	}

	return nil
}

func (s *StorageConfig) Validate() error {
	if s.Type == "" {
		return errors.New("storage type required")
	}

	if s.Type == "sqlite" && s.Path == "" {
		return errors.New("storage path required for SQLite")
	}

	return nil
}
