package config

import "time"

// Default returns a configuration with sensible defaults
func Default() *Config {
	return &Config{
		Server: ServerConfig{
			Address: "0.0.0.0",
			Port:    8080,
			Timeouts: TimeoutConfig{
				Read:     10 * time.Second,
				Write:    10 * time.Second,
				Shutdown: 5 * time.Second,
			},
		},
		Auth: AuthConfig{
			JWTSecret:     "change-this-secret-in-production",
			TokenDuration: 24 * time.Hour,
			DefaultUser: DefaultUserConfig{
				Username: "admin",
				Password: "changeme",
			},
		},
		Agent: AgentConfig{
			URL:             "http://127.0.0.1:9999",
			PollingInterval: 3 * time.Second,
			Timeout:         5 * time.Second,
		},
		OBS: OBSConfig{
			Host:           "localhost",
			Port:           4455,
			Password:       "WebSocket2025!",
			AutoReconnect:  true,
			ReconnectDelay: 5 * time.Second,
		},
		Twitch: TwitchConfig{
			Enabled:      false,
			ClientID:     "",
			ClientSecret: "",
			Channel:      "",
			OAuthToken:   "",
		},
		YouTube: YouTubeConfig{
			Enabled:      false,
			APIKey:       "",
			ChannelID:    "",
			ClientID:     "",
			ClientSecret: "",
		},
		Storage: StorageConfig{
			Type: "sqlite",
			Path: "./portal.db",
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "text",
		},
	}
}
