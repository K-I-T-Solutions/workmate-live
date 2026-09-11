package agent

import (
	"sync"
	"time"
)

// StatusCache hält den zuletzt von jedem Agent gepushten Status. Verbindet sich
// ein Agent über den Link (statt per HTTP gepollt zu werden), ist das die
// einzige Statusquelle.
type StatusCache struct {
	mu      sync.RWMutex
	entries map[string]cacheEntry
}

type cacheEntry struct {
	status    *Status
	updatedAt time.Time
}

// NewStatusCache creates an empty status cache
func NewStatusCache() *StatusCache {
	return &StatusCache{entries: make(map[string]cacheEntry)}
}

// Set speichert den Status eines Agents.
func (c *StatusCache) Set(agentID string, status *Status) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries[agentID] = cacheEntry{status: status, updatedAt: time.Now()}
}

// Get liefert den Status eines bestimmten Agents.
func (c *StatusCache) Get(agentID string) (*Status, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, ok := c.entries[agentID]
	if !ok {
		return nil, false
	}

	return entry.status, true
}

// Latest liefert den zuletzt aktualisierten Status über alle Agents hinweg.
// Nützlich für die Single-Agent-Sicht der bestehenden API.
func (c *StatusCache) Latest() (*Status, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var (
		newest cacheEntry
		found  bool
	)

	for _, entry := range c.entries {
		if !found || entry.updatedAt.After(newest.updatedAt) {
			newest = entry
			found = true
		}
	}

	if !found {
		return nil, false
	}

	return newest.status, true
}

// Remove löscht den Eintrag eines Agents, etwa nach dem Verbindungsende.
func (c *StatusCache) Remove(agentID string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.entries, agentID)
}
