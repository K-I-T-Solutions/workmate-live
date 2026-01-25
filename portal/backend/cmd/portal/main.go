package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"kit.workmate/gaming-portal/internal/api"
	"kit.workmate/gaming-portal/internal/api/handlers"
	"kit.workmate/gaming-portal/internal/config"
	"kit.workmate/gaming-portal/internal/services/agent"
	"kit.workmate/gaming-portal/internal/services/obs"
	"kit.workmate/gaming-portal/internal/services/twitch"
	"kit.workmate/gaming-portal/internal/websocket"
)

func main() {
	configPath := flag.String("config", "", "path to config file")
	flag.Parse()

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize WebSocket hub
	hub := websocket.NewHub()
	go hub.Run()

	// Initialize agent client
	agentClient := agent.NewClient(cfg.Agent.URL, cfg.Agent.Timeout)

	// Initialize agent poller with callback to broadcast status via WebSocket
	poller := agent.NewPoller(agentClient, cfg.Agent.PollingInterval, func(status *agent.Status) {
		hub.Broadcast(websocket.Message{
			Type: websocket.MessageTypeAgentStatus,
			Data: status,
		})
	})
	poller.Start()

	// Initialize OBS client
	obsClient := obs.NewClient(cfg.OBS.Host, cfg.OBS.Port, cfg.OBS.Password)

	// Try to connect to OBS (non-blocking)
	if err := obsClient.Connect(); err != nil {
		log.Printf("Warning: Failed to connect to OBS: %v", err)
		log.Println("OBS features will be unavailable until connection is established")
	} else {
		log.Println("Successfully connected to OBS")

		// Set event callback to broadcast OBS events via WebSocket
		obsClient.SetEventCallback(func(event interface{}) {
			hub.Broadcast(websocket.Message{
				Type: websocket.MessageTypeOBSEvent,
				Data: event,
			})
		})
	}

	// Initialize Twitch client (if enabled)
	var twitchClient *twitch.Client
	if cfg.Twitch.Enabled {
		twitchClient = twitch.NewClient(
			cfg.Twitch.ClientID,
			cfg.Twitch.ClientSecret,
			cfg.Twitch.Channel,
			cfg.Twitch.OAuthToken,
		)

		if err := twitchClient.Connect(); err != nil {
			log.Printf("Warning: Failed to connect to Twitch: %v", err)
			log.Println("Twitch features will be unavailable")
		} else {
			log.Println("Successfully connected to Twitch")

			// Set event callback to broadcast Twitch events via WebSocket
			twitchClient.SetEventCallback(func(event interface{}) {
				eventMap := event.(map[string]interface{})
				eventType := eventMap["type"].(string)

				var msgType string
				if eventType == "chat_message" {
					msgType = websocket.MessageTypeTwitchChat
				} else if eventType == "eventsub_event" {
					msgType = websocket.MessageTypeTwitchEvent
				} else {
					return
				}

				hub.Broadcast(websocket.Message{
					Type: msgType,
					Data: eventMap["data"],
				})
			})
		}
	}

	// Initialize handlers
	h := &api.Handlers{
		Agent:     handlers.NewAgentHandler(agentClient),
		WebSocket: handlers.NewWebSocketHandler(hub),
		OBS:       handlers.NewOBSHandler(obsClient),
		Twitch:    handlers.NewTwitchHandler(twitchClient),
	}

	// Setup routes
	handler := api.Routes(h)

	// Create and start server
	server := api.New(cfg.Server, handler)
	server.Start()

	// Wait for interrupt signal
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Println("Stopping portal server")

	// Stop poller
	poller.Stop()

	// Disconnect from OBS
	if err := obsClient.Disconnect(); err != nil {
		log.Printf("Error disconnecting from OBS: %v", err)
	}

	// Disconnect from Twitch
	if twitchClient != nil {
		if err := twitchClient.Disconnect(); err != nil {
			log.Printf("Error disconnecting from Twitch: %v", err)
		}
	}

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.Timeouts.Shutdown)
	defer cancel()
	_ = server.Shutdown(ctx)

	log.Println("Portal server stopped")
}
