package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the entire portal configuration
type Config struct {
	Server  ServerConfig  `yaml:"server"`
	Auth    AuthConfig    `yaml:"auth"`
	Agent   AgentConfig   `yaml:"agent"`
	OBS     OBSConfig     `yaml:"obs"`
	Twitch  TwitchConfig  `yaml:"twitch"`
	YouTube YouTubeConfig `yaml:"youtube"`
	Storage StorageConfig `yaml:"storage"`
	Logging LoggingConfig `yaml:"logging"`
}

type ServerConfig struct {
	Address  string        `yaml:"address"`
	Port     int           `yaml:"port"`
	Timeouts TimeoutConfig `yaml:"timeouts"`
}

type TimeoutConfig struct {
	Read     time.Duration `yaml:"read"`
	Write    time.Duration `yaml:"write"`
	Shutdown time.Duration `yaml:"shutdown"`
}

type AuthConfig struct {
	JWTSecret     string            `yaml:"jwt_secret"`
	TokenDuration time.Duration     `yaml:"token_duration"`
	DefaultUser   DefaultUserConfig `yaml:"default_user"`
}

type DefaultUserConfig struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type AgentConfig struct {
	URL             string        `yaml:"url"`
	PollingInterval time.Duration `yaml:"polling_interval"`
	Timeout         time.Duration `yaml:"timeout"`
}

type OBSConfig struct {
	Host           string        `yaml:"host"`
	Port           int           `yaml:"port"`
	Password       string        `yaml:"password"`
	AutoReconnect  bool          `yaml:"auto_reconnect"`
	ReconnectDelay time.Duration `yaml:"reconnect_delay"`
}

type TwitchConfig struct {
	Enabled      bool   `yaml:"enabled"`
	ClientID     string `yaml:"client_id"`
	ClientSecret string `yaml:"client_secret"`
	Channel      string `yaml:"channel"`
	OAuthToken   string `yaml:"oauth_token"`
}

type YouTubeConfig struct {
	Enabled      bool   `yaml:"enabled"`
	APIKey       string `yaml:"api_key"`
	ChannelID    string `yaml:"channel_id"`
	ClientID     string `yaml:"client_id"`
	ClientSecret string `yaml:"client_secret"`
}

type StorageConfig struct {
	Type string `yaml:"type"`
	Path string `yaml:"path"`
}

type LoggingConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
}

// Load attempts to load configuration from a file path
// Falls back to defaults if file doesn't exist
func Load(path string) (*Config, error) {
	// If no path specified, search default locations
	if path == "" {
		path = findConfigFile()
	}

	// If still no file found, use defaults
	if path == "" {
		log.Println("No config file found, using defaults")
		return Default(), nil
	}

	log.Printf("Loading config from: %s", path)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			log.Println("Config file not found, using defaults")
			return Default(), nil
		}
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	cfg := Default()
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	// Validate after loading
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return cfg, nil
}

// findConfigFile searches for config in standard locations
func findConfigFile() string {
	locations := []string{
		"./config/portal.yaml",
		"./portal.yaml",
		filepath.Join(os.Getenv("HOME"), ".config", "workmate-portal", "portal.yaml"),
		"/etc/workmate-portal/portal.yaml",
	}

	for _, loc := range locations {
		if _, err := os.Stat(loc); err == nil {
			return loc
		}
	}

	return ""
}

// Addr returns the full server address (host:port)
func (s ServerConfig) Addr() string {
	return fmt.Sprintf("%s:%d", s.Address, s.Port)
}
