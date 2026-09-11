package automation

import (
	"fmt"
	"log"
	"os"
	"sync"

	"gopkg.in/yaml.v3"
)

type rulesFile struct {
	Rules []Rule `yaml:"rules"`
}

// Store manages automation rules with YAML persistence
type Store struct {
	rules    []Rule
	filePath string
	mu       sync.RWMutex
}

// NewStore creates a Store and loads existing rules from filePath
func NewStore(filePath string) *Store {
	s := &Store{filePath: filePath}
	if err := s.load(); err != nil {
		log.Printf("Automation: Could not load rules file: %v", err)
	}
	return s
}

func (s *Store) load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			s.rules = nil
			return nil
		}
		return fmt.Errorf("reading rules file: %w", err)
	}

	var f rulesFile
	if err := yaml.Unmarshal(data, &f); err != nil {
		return fmt.Errorf("parsing rules file: %w", err)
	}

	s.rules = f.Rules
	log.Printf("Automation: Loaded %d rule(s) from %s", len(s.rules), s.filePath)
	return nil
}

func (s *Store) save() error {
	f := rulesFile{Rules: s.rules}
	data, err := yaml.Marshal(&f)
	if err != nil {
		return fmt.Errorf("marshaling rules: %w", err)
	}
	if err := os.WriteFile(s.filePath, data, 0644); err != nil {
		return fmt.Errorf("writing rules file: %w", err)
	}
	return nil
}

// List returns a copy of all rules
func (s *Store) List() []Rule {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]Rule, len(s.rules))
	copy(result, s.rules)
	return result
}

// Get returns a rule by name
func (s *Store) Get(name string) (*Rule, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for i := range s.rules {
		if s.rules[i].Name == name {
			r := s.rules[i]
			return &r, true
		}
	}
	return nil, false
}

// Add adds a new rule (returns error if name already exists)
func (s *Store) Add(rule Rule) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, r := range s.rules {
		if r.Name == rule.Name {
			return fmt.Errorf("rule '%s' already exists", rule.Name)
		}
	}

	s.rules = append(s.rules, rule)
	return s.save()
}

// Update replaces an existing rule
func (s *Store) Update(name string, rule Rule) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.rules {
		if s.rules[i].Name == name {
			rule.Name = name
			s.rules[i] = rule
			return s.save()
		}
	}
	return fmt.Errorf("rule '%s' not found", name)
}

// Delete removes a rule by name
func (s *Store) Delete(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.rules {
		if s.rules[i].Name == name {
			s.rules = append(s.rules[:i], s.rules[i+1:]...)
			return s.save()
		}
	}
	return fmt.Errorf("rule '%s' not found", name)
}
