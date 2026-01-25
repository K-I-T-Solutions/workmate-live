package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"kit.workmate/live-agent/internal/api"
	"kit.workmate/live-agent/internal/config"
	"kit.workmate/live-agent/internal/health"
)

func main() {
	configPath := flag.String("config", "", "path to config file")
	flag.Parse()

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// Initialize components with config
	cache := health.NewCache()
	poller := health.NewPoller(cache, cfg.Health.PollingInterval, cfg.Health.Checks)
	poller.Start()

	handler := api.Routes(cache)
	server := api.NewWithConfig(cfg.Server.Addr(), handler, cfg.Server.Timeouts)
	server.Start()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Println("stopping agent")

	poller.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.Timeouts.Shutdown)
	defer cancel()
	_ = server.Shutdown(ctx)
}
