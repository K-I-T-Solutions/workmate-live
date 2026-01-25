package health

import (
	"sync"
	"time"
)

type Cache struct {
	mu           sync.RWMutex
	status       *Status
	capabilities Capabilities
	updatedAt    time.Time
}

func NewCache() *Cache {
	return &Cache{}
}

func (c *Cache) Set(s *Status) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.status = s
	c.capabilities = CollectCapabilities(s)
	c.updatedAt = time.Now()
}

func (c *Cache) Get() *Status {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.status
}

func (c *Cache) UpdatedAt() time.Time {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.updatedAt
}

func (c *Cache) Capabilities() Capabilities {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.capabilities
}
