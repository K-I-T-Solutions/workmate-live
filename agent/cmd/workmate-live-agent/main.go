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
	"kit.workmate/live-agent/internal/link"
	"kit.workmate/live-agent/internal/services/obs"
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

	// Kontext für die langlebigen Hintergrunddienste (OBS-Verbindung, Portal-Link)
	ctx, cancel := context.WithCancel(context.Background())

	// OBS-Steuerung: verbindet zur lokalen OBS-Instanz, damit das Portal sie
	// über den Link fernsteuern kann.
	var obsClient *obs.Client
	if cfg.OBS.Enabled {
		obsClient = obs.NewClient(cfg.OBS.Host, cfg.OBS.Port, cfg.OBS.Password, cfg.OBS.ReconnectDelay)
		go obsClient.Run(ctx)
		log.Printf("obs control enabled for %s", cfg.OBS.Addr())
	}

	// Portal-Link: ausgehende Verbindung zum Portal, über die Kommandos
	// entgegengenommen und Events gepusht werden.
	if cfg.Portal.Enabled {
		var ctrl link.OBSController
		if obsClient != nil {
			ctrl = obsClient
		}

		linkClient := link.NewClient(cfg.Portal, agentID(cfg), ctrl, cache)
		go linkClient.Run(ctx)
	}

	handler := api.Routes(cache)
	server := api.NewWithConfig(cfg.Server.Addr(), handler, cfg.Server.Timeouts)
	server.Start()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Println("stopping agent")

	cancel()
	poller.Stop()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.Server.Timeouts.Shutdown)
	defer shutdownCancel()
	_ = server.Shutdown(shutdownCtx)
}

// agentID liefert die konfigurierte Agent-ID, ersatzweise den Hostnamen.
func agentID(cfg *config.Config) string {
	if cfg.Portal.AgentID != "" {
		return cfg.Portal.AgentID
	}

	if hostname, err := os.Hostname(); err == nil && hostname != "" {
		return hostname
	}

	return "workmate-agent"
}
