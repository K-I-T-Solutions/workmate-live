package commands

import (
	"sync"
	"time"
)

// CooldownManager tracks per-user command cooldowns
type CooldownManager struct {
	cooldowns map[string]time.Time
	mu        sync.Mutex
}

// NewCooldownManager creates a new cooldown manager
func NewCooldownManager() *CooldownManager {
	return &CooldownManager{
		cooldowns: make(map[string]time.Time),
	}
}

// IsOnCooldown checks if a command is on cooldown for a user.
// If not on cooldown, it sets the cooldown automatically and returns false.
func (cm *CooldownManager) IsOnCooldown(command, user string, duration time.Duration) bool {
	if duration == 0 {
		return false
	}

	cm.mu.Lock()
	defer cm.mu.Unlock()

	key := command + ":" + user
	if expiry, exists := cm.cooldowns[key]; exists {
		if time.Now().Before(expiry) {
			return true
		}
	}

	cm.cooldowns[key] = time.Now().Add(duration)
	return false
}
