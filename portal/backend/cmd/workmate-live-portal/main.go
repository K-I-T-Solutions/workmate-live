package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"kit.workmate/live-portal/internal/agentlink"
	"kit.workmate/live-portal/internal/api"
	"kit.workmate/live-portal/internal/api/handlers"
	"kit.workmate/live-portal/internal/auth"
	"kit.workmate/live-portal/internal/automation"
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

	// Kontext für die langlebigen Hintergrunddienste
	ctx, cancel := context.WithCancel(context.Background())

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

	// Initialize WebSocket hub (Browser-Clients)
	hub := websocket.NewHub()
	go hub.Run()

	// Initialize automation engine
	autoStore := automation.NewStore(cfg.Automation.RulesFile)
	autoEngine := automation.NewEngine(autoStore, hub)
	go autoEngine.Run()

	// handleOBSEvent verteilt ein OBS-Event an Browser-Clients und die
	// Automation-Engine — unabhängig davon, ob es aus der direkten
	// OBS-Verbindung oder von einem Agent stammt.
	//
	// agentID benennt die Herkunft. Ohne sie sind bei mehreren verbundenen
	// Agents weder Anzeige noch Regeln dem richtigen Rechner zuzuordnen.
	handleOBSEvent := func(agentID string, event map[string]interface{}) {
		hub.Broadcast(websocket.Message{
			Type: websocket.MessageTypeOBSEvent,
			Data: websocket.WithAgentID(agentID, event),
		})

		if ev, ok := automation.FromOBSEvent(agentID, event); ok {
			autoEngine.Send(ev)
		}
	}

	// Status-Cache für Agents, die ihren Status über den Link pushen
	statusCache := agent.NewStatusCache()

	// Agent-Link: Agents bauen die Verbindung zum Portal auf. Damit lässt sich
	// OBS auf einem entfernten Rechner steuern, ohne dort einen Port zu öffnen.
	var (
		linkHub     *agentlink.Hub
		linkHandler *agentlink.Handler
	)
	if cfg.AgentLinkEnabled() {
		linkHub = agentlink.NewHub(cfg.Agent.PrimaryID)
		linkHub.SetEventHandler(func(agentID, event string, data json.RawMessage) {
			switch event {
			case agentlink.EventAgentStatus:
				var status agent.Status
				if err := json.Unmarshal(data, &status); err != nil {
					log.Printf("Invalid agent status from %s: %v", agentID, err)
					return
				}
				statusCache.Set(agentID, &status)
				hub.Broadcast(websocket.Message{
					Type: websocket.MessageTypeAgentStatus,
					Data: agent.StatusMessage{Status: &status, AgentID: agentID},
				})

			case agentlink.EventOBS:
				var obsEvent map[string]interface{}
				if err := json.Unmarshal(data, &obsEvent); err != nil {
					log.Printf("Invalid OBS event from %s: %v", agentID, err)
					return
				}
				handleOBSEvent(agentID, obsEvent)
			}
		})

		linkHandler = agentlink.NewHandler(linkHub, cfg.Agent.APIKey)
		log.Println("Agent link enabled at /ws/agent")
	}

	// Agent-Polling per HTTP — nur sinnvoll, wenn der Agent direkt erreichbar
	// ist. Im Link-Betrieb hinter NAT bleibt agent.url leer.
	var poller *agent.Poller
	var agentClient *agent.Client
	if cfg.Agent.URL != "" {
		agentClient = agent.NewClient(cfg.Agent.URL, cfg.Agent.Timeout)
		// Der gepollte Agent ist über die URL adressiert und meldet seine
		// Kennung nicht selbst — hier steht die konfigurierte, ersatzweise
		// die des lokalen Betriebs.
		polledID := cfg.Agent.PrimaryID
		if polledID == "" {
			polledID = websocket.LocalAgentID
		}

		poller = agent.NewPoller(agentClient, cfg.Agent.PollingInterval, func(status *agent.Status) {
			hub.Broadcast(websocket.Message{
				Type: websocket.MessageTypeAgentStatus,
				Data: agent.StatusMessage{Status: status, AgentID: polledID},
			})
		})
		poller.Start()
	} else {
		log.Println("No agent URL configured, relying on agent link for status")
	}

	// OBS-Steuerung: direkt zum OBS-WebSocket oder über einen Agent.
	var obsController obs.Controller
	switch cfg.OBS.Mode {
	case config.OBSModeAgent:
		obsController = obs.NewRemoteController(linkHub, cfg.Agent.CommandTimeout)
		log.Println("OBS control routed through agent link")

	default:
		obsDirect := obs.NewClient(cfg.OBS.Host, cfg.OBS.Port, cfg.OBS.Password, cfg.OBS.ReconnectDelay)
		// Im Direktmodus gibt es keinen Agent — die Ereignisse werden als
		// lokal gekennzeichnet.
		obsDirect.SetEventCallback(func(event map[string]interface{}) {
			handleOBSEvent(websocket.LocalAgentID, event)
		})
		go obsDirect.Run(ctx)
		obsController = obsDirect
		log.Printf("OBS control connecting directly to %s:%d", cfg.OBS.Host, cfg.OBS.Port)
	}

	// Der Executor ist immer gesetzt: fehlt die Verbindung, meldet die Aktion
	// das zur Laufzeit, statt die Regel stumm ins Leere laufen zu lassen.
	autoEngine.SetOBSExecutor(obsController)

	// Initialize Twitch client (if enabled)
	var twitchClient *twitch.Client
	if cfg.Twitch.Enabled {
		twitchClient = twitch.NewClient(
			cfg.Twitch.ClientID,
			cfg.Twitch.ClientSecret,
			cfg.Twitch.Channel,
			cfg.Twitch.OAuthToken,
			cfg.Twitch.CommandsFile,
		)

		if err := twitchClient.Connect(); err != nil {
			log.Printf("Warning: Failed to connect to Twitch: %v", err)
			log.Println("Twitch features will be unavailable")
		} else {
			log.Println("Successfully connected to Twitch")
			autoEngine.SetTwitchExecutor(twitchClient)

			twitchClient.SetEventCallback(func(event interface{}) {
				eventMap := event.(map[string]interface{})
				eventType := eventMap["type"].(string)

				var msgType string
				if eventType == "chat_message" {
					msgType = websocket.MessageTypeTwitchChat
				} else if eventType == "eventsub_event" {
					msgType = websocket.MessageTypeTwitchEvent
				} else if eventType == "command_executed" {
					msgType = websocket.MessageTypeTwitchCommand
				} else {
					return
				}

				hub.Broadcast(websocket.Message{
					Type: msgType,
					Data: eventMap["data"],
				})

				// Route Twitch EventSub events to automation engine
				if eventType == "eventsub_event" {
					if esEvent, ok := eventMap["data"].(*twitch.EventSubEvent); ok {
						autoEngine.Send(twitchEventToAutomation(esEvent))
					}
				}
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
		Auth:       handlers.NewAuthHandler(userStore, jwtService),
		Agent:      handlers.NewAgentHandler(agentClient, statusCache, linkHub),
		WebSocket:  handlers.NewWebSocketHandler(hub),
		OBS:        handlers.NewOBSHandler(obsController),
		Twitch:     handlers.NewTwitchHandler(twitchClient),
		YouTube:    handlers.NewYouTubeHandler(youtubeClient),
		Config:     handlers.NewConfigHandler(cfg, userStore),
		Restart:    handlers.NewRestartHandler(),
		Automation: handlers.NewAutomationHandler(autoStore, autoEngine),
		AgentLink:  linkHandler,
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

	// Beendet OBS-Reconnect-Loop und alle weiteren Hintergrunddienste
	cancel()

	// Stop automation engine
	autoEngine.Stop()

	// Stop poller
	if poller != nil {
		poller.Stop()
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
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.Server.Timeouts.Shutdown)
	defer shutdownCancel()
	_ = server.Shutdown(shutdownCtx)

	log.Println("Portal server stopped")
}

// twitchEventToAutomation maps a Twitch EventSub event to an automation.Event with template vars
func twitchEventToAutomation(esEvent *twitch.EventSubEvent) automation.Event {
	vars := map[string]string{}

	switch esEvent.Type {
	case "follow":
		if data, ok := esEvent.Data.(twitch.FollowEvent); ok {
			vars["user"] = data.UserName
		}
		return automation.Event{Type: "twitch:follow", Vars: vars}
	case "subscribe":
		if data, ok := esEvent.Data.(twitch.SubscribeEvent); ok {
			vars["user"] = data.UserName
			vars["tier"] = data.Tier
		}
		return automation.Event{Type: "twitch:subscribe", Vars: vars}
	case "raid":
		if data, ok := esEvent.Data.(twitch.RaidEvent); ok {
			vars["raider"] = data.FromUserName
			vars["viewers"] = strconv.Itoa(data.Viewers)
		}
		return automation.Event{Type: "twitch:raid", Vars: vars}
	}

	return automation.Event{Type: "twitch:" + esEvent.Type, Vars: vars}
}
