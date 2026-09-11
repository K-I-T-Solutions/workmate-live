package commands

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

// CustomCommand represents a user-created chat command
type CustomCommand struct {
	Name        string        `yaml:"name" json:"name"`
	Description string        `yaml:"description" json:"description"`
	Response    string        `yaml:"response" json:"response"`
	ModOnly     bool          `yaml:"mod_only" json:"mod_only"`
	Cooldown    time.Duration `yaml:"cooldown" json:"cooldown"`
	Enabled     bool          `yaml:"enabled" json:"enabled"`
	Count       int           `yaml:"count" json:"count"`
}

// commandsFile is the YAML structure for the commands file
type commandsFile struct {
	Commands []CustomCommand `yaml:"commands"`
}

// CommandStore manages custom commands with YAML persistence
type CommandStore struct {
	commands []CustomCommand
	filePath string
	mu       sync.RWMutex
}

// builtinNames contains the names of all built-in commands
var builtinNames = map[string]bool{
	"ping":   true,
	"uptime": true,
	"help":   true,
	"say":    true,
}

// IsBuiltinCommand checks if a command name is a built-in command
func IsBuiltinCommand(name string) bool {
	return builtinNames[name]
}

// NewCommandStore creates a new CommandStore and loads existing commands
func NewCommandStore(filePath string) *CommandStore {
	s := &CommandStore{
		filePath: filePath,
	}
	if err := s.Load(); err != nil {
		log.Printf("Warning: Could not load commands file: %v", err)
	}
	return s
}

// Load reads the YAML file and loads commands
func (s *CommandStore) Load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			s.commands = nil
			return nil
		}
		return fmt.Errorf("reading commands file: %w", err)
	}

	var f commandsFile
	if err := yaml.Unmarshal(data, &f); err != nil {
		return fmt.Errorf("parsing commands file: %w", err)
	}

	s.commands = f.Commands
	return nil
}

// Save writes all commands to the YAML file
func (s *CommandStore) Save() error {
	f := commandsFile{Commands: s.commands}
	data, err := yaml.Marshal(&f)
	if err != nil {
		return fmt.Errorf("marshaling commands: %w", err)
	}

	if err := os.WriteFile(s.filePath, data, 0644); err != nil {
		return fmt.Errorf("writing commands file: %w", err)
	}
	return nil
}

// List returns all custom commands
func (s *CommandStore) List() []CustomCommand {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]CustomCommand, len(s.commands))
	copy(result, s.commands)
	return result
}

// Get finds a custom command by name
func (s *CommandStore) Get(name string) (*CustomCommand, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for i := range s.commands {
		if s.commands[i].Name == name {
			cmd := s.commands[i]
			return &cmd, true
		}
	}
	return nil, false
}

// Add adds a new custom command
func (s *CommandStore) Add(cmd CustomCommand) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if IsBuiltinCommand(cmd.Name) {
		return fmt.Errorf("command '%s' conflicts with a built-in command", cmd.Name)
	}

	for _, c := range s.commands {
		if c.Name == cmd.Name {
			return fmt.Errorf("command '%s' already exists", cmd.Name)
		}
	}

	s.commands = append(s.commands, cmd)
	return s.Save()
}

// Update updates an existing custom command
func (s *CommandStore) Update(name string, cmd CustomCommand) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.commands {
		if s.commands[i].Name == name {
			// Preserve count if not explicitly set
			if cmd.Count == 0 {
				cmd.Count = s.commands[i].Count
			}
			cmd.Name = name // Ensure name stays the same
			s.commands[i] = cmd
			return s.Save()
		}
	}
	return fmt.Errorf("command '%s' not found", name)
}

// Delete removes a custom command
func (s *CommandStore) Delete(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.commands {
		if s.commands[i].Name == name {
			s.commands = append(s.commands[:i], s.commands[i+1:]...)
			return s.Save()
		}
	}
	return fmt.Errorf("command '%s' not found", name)
}

// IncrementCount increments the usage counter for a command
func (s *CommandStore) IncrementCount(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.commands {
		if s.commands[i].Name == name {
			s.commands[i].Count++
			_ = s.Save()
			return
		}
	}
}
