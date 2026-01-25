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

	if err := c.Health.Validate(); err != nil {
		return fmt.Errorf("health config: %w", err)
	}

	if err := c.Portal.Validate(); err != nil {
		return fmt.Errorf("portal config: %w", err)
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

func (h *HealthConfig) Validate() error {
	if h.PollingInterval <= 0 {
		return errors.New("polling interval must be positive")
	}

	return nil
}

func (p *PortalConfig) Validate() error {
	if !p.Enabled {
		return nil // Skip validation if disabled
	}

	if p.URL == "" {
		return errors.New("portal URL required when enabled")
	}

	// Either API key or credentials required
	hasAPIKey := p.APIKey != ""
	hasCredentials := p.Credentials.Username != "" && p.Credentials.Password != ""

	if !hasAPIKey && !hasCredentials {
		return errors.New("either api_key or credentials required")
	}

	return nil
}
