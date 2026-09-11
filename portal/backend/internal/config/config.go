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
	Server     ServerConfig     `yaml:"server" json:"server"`
	Auth       AuthConfig       `yaml:"auth" json:"auth"`
	Agent      AgentConfig      `yaml:"agent" json:"agent"`
	OBS        OBSConfig        `yaml:"obs" json:"obs"`
	Twitch     TwitchConfig     `yaml:"twitch" json:"twitch"`
	YouTube    YouTubeConfig    `yaml:"youtube" json:"youtube"`
	Automation AutomationConfig `yaml:"automation" json:"automation"`
	Storage    StorageConfig    `yaml:"storage" json:"storage"`
	Logging    LoggingConfig    `yaml:"logging" json:"logging"`

	filePath string `yaml:"-" json:"-"`
}

// AutomationConfig holds settings for the automation engine
type AutomationConfig struct {
	RulesFile string `yaml:"rules_file" json:"rules_file"`
}

type ServerConfig struct {
	Address  string        `yaml:"address" json:"address"`
	Port     int           `yaml:"port" json:"port"`
	Timeouts TimeoutConfig `yaml:"timeouts" json:"timeouts"`
}

type TimeoutConfig struct {
	Read     time.Duration `yaml:"read" json:"read"`
	Write    time.Duration `yaml:"write" json:"write"`
	Shutdown time.Duration `yaml:"shutdown" json:"shutdown"`
}

type AuthConfig struct {
	JWTSecret     string            `yaml:"jwt_secret" json:"jwt_secret"`
	TokenDuration time.Duration     `yaml:"token_duration" json:"token_duration"`
	DefaultUser   DefaultUserConfig `yaml:"default_user" json:"default_user"`
}

type DefaultUserConfig struct {
	Username string `yaml:"username" json:"username"`
	Password string `yaml:"password" json:"password"`
}

type AgentConfig struct {
	URL             string        `yaml:"url" json:"url"`
	PollingInterval time.Duration `yaml:"polling_interval" json:"polling_interval"`
	Timeout         time.Duration `yaml:"timeout" json:"timeout"`

	// APIKey authentifiziert Agents am /ws/agent-Endpunkt. Leer bedeutet:
	// der Endpunkt ist deaktiviert.
	APIKey string `yaml:"api_key" json:"api_key"`

	// PrimaryID benennt den Agent, der OBS steuert. Leer bedeutet: den ersten
	// verbundenen Agent mit lokaler OBS-Instanz verwenden.
	PrimaryID string `yaml:"primary_id" json:"primary_id"`

	// CommandTimeout begrenzt, wie lange auf die Antwort eines Agents
	// auf ein OBS-Kommando gewartet wird.
	CommandTimeout time.Duration `yaml:"command_timeout" json:"command_timeout"`
}

// OBSMode bestimmt, wie das Portal OBS erreicht.
const (
	// OBSModeDirect: Portal verbindet sich selbst zum OBS-WebSocket.
	OBSModeDirect = "direct"
	// OBSModeAgent: Steuerung läuft über einen Agent, der die Verbindung zum
	// Portal aufbaut. Der Streaming-Rechner braucht keinen offenen Port.
	OBSModeAgent = "agent"
)

type OBSConfig struct {
	// Mode ist "direct" (Standard) oder "agent".
	Mode           string        `yaml:"mode" json:"mode"`
	Host           string        `yaml:"host" json:"host"`
	Port           int           `yaml:"port" json:"port"`
	Password       string        `yaml:"password" json:"password"`
	AutoReconnect  bool          `yaml:"auto_reconnect" json:"auto_reconnect"`
	ReconnectDelay time.Duration `yaml:"reconnect_delay" json:"reconnect_delay"`
}

type TwitchConfig struct {
	Enabled      bool   `yaml:"enabled" json:"enabled"`
	ClientID     string `yaml:"client_id" json:"client_id"`
	ClientSecret string `yaml:"client_secret" json:"client_secret"`
	Channel      string `yaml:"channel" json:"channel"`
	OAuthToken   string `yaml:"oauth_token" json:"oauth_token"`
	CommandsFile string `yaml:"commands_file" json:"commands_file"`
}

type YouTubeConfig struct {
	Enabled      bool   `yaml:"enabled" json:"enabled"`
	APIKey       string `yaml:"api_key" json:"api_key"`
	ChannelID    string `yaml:"channel_id" json:"channel_id"`
	ClientID     string `yaml:"client_id" json:"client_id"`
	ClientSecret string `yaml:"client_secret" json:"client_secret"`
}

type StorageConfig struct {
	Type string `yaml:"type" json:"type"`
	Path string `yaml:"path" json:"path"`
}

type LoggingConfig struct {
	Level  string `yaml:"level" json:"level"`
	Format string `yaml:"format" json:"format"`
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

	cfg.filePath = path
	cfg.normalize()

	// Validate after loading
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return cfg, nil
}

// normalize füllt optionale Felder mit ihren Standardwerten, damit der
// restliche Code nicht überall Sonderfälle behandeln muss.
func (c *Config) normalize() {
	if c.OBS.Mode == "" {
		c.OBS.Mode = OBSModeDirect
	}

	if c.Agent.CommandTimeout <= 0 {
		c.Agent.CommandTimeout = 10 * time.Second
	}
}

// AgentLinkEnabled meldet, ob Agents sich verbinden dürfen.
func (c *Config) AgentLinkEnabled() bool {
	return c.Agent.APIKey != ""
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

// Save writes the current configuration back to the YAML file
func (c *Config) Save() error {
	if c.filePath == "" {
		return fmt.Errorf("no config file path set, cannot save")
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	if err := os.WriteFile(c.filePath, data, 0644); err != nil {
		return fmt.Errorf("writing config file: %w", err)
	}

	log.Printf("Config saved to: %s", c.filePath)
	return nil
}

// maskSecret replaces a secret string with "***" if non-empty
func maskSecret(s string) string {
	if s == "" {
		return ""
	}
	return "***"
}

// ToSafe returns a copy of the config with sensitive values masked
func (c *Config) ToSafe() Config {
	safe := *c

	// Mask auth secrets
	safe.Auth.JWTSecret = maskSecret(c.Auth.JWTSecret)
	safe.Auth.DefaultUser.Password = maskSecret(c.Auth.DefaultUser.Password)

	// Mask OBS password
	safe.OBS.Password = maskSecret(c.OBS.Password)

	// Mask the shared agent key
	safe.Agent.APIKey = maskSecret(c.Agent.APIKey)

	// Mask Twitch secrets
	safe.Twitch.ClientSecret = maskSecret(c.Twitch.ClientSecret)
	safe.Twitch.OAuthToken = maskSecret(c.Twitch.OAuthToken)

	// Mask YouTube secrets
	safe.YouTube.APIKey = maskSecret(c.YouTube.APIKey)
	safe.YouTube.ClientSecret = maskSecret(c.YouTube.ClientSecret)

	return safe
}
