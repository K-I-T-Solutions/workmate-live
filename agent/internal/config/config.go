package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the entire agent configuration
type Config struct {
	Server ServerConfig `yaml:"server"`
	Health HealthConfig `yaml:"health"`
	OBS    OBSConfig    `yaml:"obs"`
	Portal PortalConfig `yaml:"portal"`
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

type HealthConfig struct {
	PollingInterval time.Duration `yaml:"polling_interval"`
	Checks          ChecksConfig  `yaml:"checks"`
}

type ChecksConfig struct {
	GPU   bool `yaml:"gpu"`
	Audio bool `yaml:"audio"`
	Video bool `yaml:"video"`
	OBS   bool `yaml:"obs"`
}

// OBSConfig steuert die Verbindung zur lokal laufenden OBS-Instanz.
// Ist sie aktiv, kann das Portal OBS über die Link-Verbindung fernsteuern,
// ohne den OBS-WebSocket-Port nach außen zu öffnen.
type OBSConfig struct {
	Enabled        bool          `yaml:"enabled"`
	Host           string        `yaml:"host"`
	Port           int           `yaml:"port"`
	Password       string        `yaml:"password"`
	ReconnectDelay time.Duration `yaml:"reconnect_delay"`
}

type PortalConfig struct {
	Enabled bool   `yaml:"enabled"`
	URL     string `yaml:"url"`
	APIKey  string `yaml:"api_key"`
	// AgentID identifiziert diesen Agent gegenüber dem Portal.
	// Leer bedeutet: Hostname verwenden.
	AgentID       string            `yaml:"agent_id"`
	Credentials   CredentialsConfig `yaml:"credentials"`
	Timeout       time.Duration     `yaml:"timeout"`
	RetryAttempts int               `yaml:"retry_attempts"`
	RetryDelay    time.Duration     `yaml:"retry_delay"`
}

// Addr returns the OBS WebSocket address (host:port)
func (o OBSConfig) Addr() string {
	return fmt.Sprintf("%s:%d", o.Host, o.Port)
}

type CredentialsConfig struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
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

// findConfigFile searches for config in standard locations.
//
// Die Suche ist plattformabhängig: os.UserConfigDir liefert unter Linux
// ~/.config, unter macOS ~/Library/Application Support und unter Windows
// %AppData%. Der Unix-Pfad wird zusätzlich geprüft, weil ~/.config auch auf
// macOS gebräuchlich ist. /etc entfällt unter Windows — dort gibt es keinen
// entsprechenden Ort, und $HOME ist häufig gar nicht gesetzt.
func findConfigFile() string {
	locations := []string{"./config.yaml"}

	if dir, err := os.UserConfigDir(); err == nil && dir != "" {
		locations = append(locations, filepath.Join(dir, "workmate-agent", "config.yaml"))
	}

	if home, err := os.UserHomeDir(); err == nil && home != "" {
		locations = append(locations, filepath.Join(home, ".config", "workmate-agent", "config.yaml"))
	}

	if runtime.GOOS != "windows" {
		locations = append(locations, "/etc/workmate-agent/config.yaml")
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
