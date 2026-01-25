package health

import (
	"log"
	"time"

	"kit.workmate/gaming-agent/internal/config"
)

type Poller struct {
	cache    *Cache
	interval time.Duration
	checks   config.ChecksConfig
	stop     chan struct{}
}

func NewPoller(cache *Cache, interval time.Duration, checks config.ChecksConfig) *Poller {
	return &Poller{
		cache:    cache,
		interval: interval,
		checks:   checks,
		stop:     make(chan struct{}),
	}
}

func (p *Poller) Start() {
	go func() {
		ticker := time.NewTicker(p.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				status, err := Collect(p.checks)
				if err != nil {
					log.Printf("status collect failed: %v", err)
					continue
				}
				p.cache.Set(status)

			case <-p.stop:
				return
			}
		}
	}()
}

func (p *Poller) Stop() {
	close(p.stop)
}
