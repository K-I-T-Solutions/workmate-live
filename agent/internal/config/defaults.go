package config

import "time"

// Default returns a configuration with sensible defaults
func Default() *Config {
	return &Config{
		Server: ServerConfig{
			Address: "127.0.0.1",
			Port:    8787,
			Timeouts: TimeoutConfig{
				Read:     5 * time.Second,
				Write:    5 * time.Second,
				Shutdown: 5 * time.Second,
			},
		},
		Health: HealthConfig{
			PollingInterval: 2 * time.Second,
			Checks: ChecksConfig{
				GPU:   true,
				Audio: true,
				Video: true,
				OBS:   true,
			},
		},
		Portal: PortalConfig{
			Enabled:       false,
			URL:           "",
			APIKey:        "",
			Timeout:       10 * time.Second,
			RetryAttempts: 3,
			RetryDelay:    5 * time.Second,
		},
	}
}
