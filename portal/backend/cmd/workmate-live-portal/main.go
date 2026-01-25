package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"kit.workmate/live-portal/internal/api"
	"kit.workmate/live-portal/internal/api/handlers"
	"kit.workmate/live-portal/internal/auth"
	"kit.workmate/live-portal/internal/config"
	"kit.workmate/live-portal/internal/services/agent"
	"kit.workmate/live-portal/internal/services/obs"
	"kit.workmate/live-portal/internal/services/twitch"
	"kit.workmate/live-portal/internal/services/youtube"
	"kit.workmate/live-portal/internal/storage"
	"kit.workmate/live-portal/internal/websocket"
)

func main() {
	configPath := flag.String("config", "", "path to config file")
	flag.Parse()

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize user storage
	userStore, err := storage.NewUserStore(cfg.Storage.Path)
	if err != nil {
		log.Fatalf("Failed to initialize user store: %v", err)
	}
	defer userStore.Close()

	// Ensure default user exists
	if err := userStore.EnsureDefaultUser(cfg.Auth.DefaultUser.Username, cfg.Auth.DefaultUser.Password); err != nil {
		log.Fatalf("Failed to ensure default user: %v", err)
	}
	log.Printf("Default user ready: %s", cfg.Auth.DefaultUser.Username)

	// Initialize JWT service
	jwtService := auth.NewJWTService(cfg.Auth.JWTSecret, cfg.Auth.TokenDuration)

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

	// Initialize YouTube client (if enabled)
	var youtubeClient *youtube.Client
	if cfg.YouTube.Enabled {
		youtubeClient = youtube.NewClient(
			cfg.YouTube.APIKey,
			cfg.YouTube.ChannelID,
			cfg.YouTube.ClientID,
			cfg.YouTube.ClientSecret,
		)

		if err := youtubeClient.Connect(); err != nil {
			log.Printf("Warning: Failed to connect to YouTube: %v", err)
			log.Println("YouTube features will be unavailable")
		} else {
			log.Println("Successfully connected to YouTube")

			// Set event callback to broadcast YouTube events via WebSocket
			youtubeClient.SetEventCallback(func(event interface{}) {
				eventMap := event.(map[string]interface{})
				eventType := eventMap["type"].(string)

				if eventType == "chat_message" {
					hub.Broadcast(websocket.Message{
						Type: websocket.MessageTypeYouTubeChat,
						Data: eventMap["data"],
					})
				}
			})
		}
	}

	// Initialize handlers
	h := &api.Handlers{
		Auth:      handlers.NewAuthHandler(userStore, jwtService),
		Agent:     handlers.NewAgentHandler(agentClient),
		WebSocket: handlers.NewWebSocketHandler(hub),
		OBS:       handlers.NewOBSHandler(obsClient),
		Twitch:    handlers.NewTwitchHandler(twitchClient),
		YouTube:   handlers.NewYouTubeHandler(youtubeClient),
	}

	// Setup routes with JWT middleware
	handler := api.Routes(h, jwtService)

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

	// Disconnect from YouTube
	if youtubeClient != nil {
		if err := youtubeClient.Disconnect(); err != nil {
			log.Printf("Error disconnecting from YouTube: %v", err)
		}
	}

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.Timeouts.Shutdown)
	defer cancel()
	_ = server.Shutdown(ctx)

	log.Println("Portal server stopped")
}
