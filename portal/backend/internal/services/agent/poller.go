package agent

import (
	"log"
	"time"
)

type StatusCallback func(*Status)

type Poller struct {
	client   *Client
	interval time.Duration
	callback StatusCallback
	stop     chan struct{}
}

func NewPoller(client *Client, interval time.Duration, callback StatusCallback) *Poller {
	return &Poller{
		client:   client,
		interval: interval,
		callback: callback,
		stop:     make(chan struct{}),
	}
}

func (p *Poller) Start() {
	go func() {
		ticker := time.NewTicker(p.interval)
		defer ticker.Stop()

		// Fetch immediately on start
		p.fetchAndNotify()

		for {
			select {
			case <-ticker.C:
				p.fetchAndNotify()
			case <-p.stop:
				return
			}
		}
	}()
}

func (p *Poller) fetchAndNotify() {
	status, err := p.client.GetStatus()
	if err != nil {
		log.Printf("Failed to fetch agent status: %v", err)
		return
	}

	if p.callback != nil {
		p.callback(status)
	}
}

func (p *Poller) Stop() {
	close(p.stop)
}
