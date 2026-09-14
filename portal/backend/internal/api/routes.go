package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"kit.workmate/live-portal/internal/agentlink"
	"kit.workmate/live-portal/internal/api/handlers"
	"kit.workmate/live-portal/internal/auth"
)

type Handlers struct {
	Auth       *handlers.AuthHandler
	Agent      *handlers.AgentHandler
	WebSocket  *handlers.WebSocketHandler
	OBS        *handlers.OBSHandler
	Twitch     *handlers.TwitchHandler
	YouTube    *handlers.YouTubeHandler
	Config     *handlers.ConfigHandler
	Restart    *handlers.RestartHandler
	Automation *handlers.AutomationHandler

	// AgentLink bedient eingehende Agent-Verbindungen. Nil, wenn kein
	// Agent-API-Key konfiguriert ist.
	AgentLink *agentlink.Handler
}

func Routes(h *Handlers, jwtService *auth.JWTService) http.Handler {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{
			"http://localhost:5173",
			"http://localhost:5174",
			"http://localhost:5175",
			"http://localhost:5176",
			"http://localhost:8080",
			"http://192.168.178.47:5174",
			"http://192.168.178.47:5175",
			"http://192.168.178.47:5176",
			"http://192.168.178.100:5174",
			"http://192.168.178.100:5175",
			"http://192.168.178.100:5176",
		},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "ok",
			"service": "workmate-portal",
		})
	})

	// API routes
	r.Route("/api", func(r chi.Router) {
		// Public auth endpoints (no JWT required)
		r.Post("/auth/login", h.Auth.Login)
		r.Get("/auth/verify", h.Auth.Verify)

		// Protected routes group (JWT required)
		r.Group(func(r chi.Router) {
			// Apply JWT middleware to all routes in this group
			r.Use(auth.Middleware(jwtService))

			// Protected auth endpoints
			r.Post("/auth/logout", h.Auth.Logout)

			// Agent proxy endpoints
			r.Get("/agent/status", h.Agent.GetStatus)
			r.Get("/agent/capabilities", h.Agent.GetCapabilities)
			r.Get("/agent/info", h.Agent.GetInfo)

			// Über den Link verbundene Agents
			r.Get("/agents", h.Agent.ListAgents)

			// OBS-Steuerung. Die Routen gibt es zweimal: mit Agent-Kennung
			// unter /agents/{agent}/obs/... und ohne unter /obs/..., wo
			// dann der primäre Agent angesprochen wird.
			r.Route("/agents/{agent}/obs", func(r chi.Router) { obsRoutes(r, h.OBS) })
			r.Route("/obs", func(r chi.Router) { obsRoutes(r, h.OBS) })

			// Twitch endpoints
			r.Get("/twitch/status", h.Twitch.GetStatus)
			r.Get("/twitch/stats", h.Twitch.GetStats)
			r.Patch("/twitch/stream", h.Twitch.UpdateStream)

			// Twitch chat command endpoints
			r.Get("/twitch/commands", h.Twitch.ListCommands)
			r.Post("/twitch/commands", h.Twitch.CreateCommand)
			r.Put("/twitch/commands/{name}", h.Twitch.UpdateCommand)
			r.Delete("/twitch/commands/{name}", h.Twitch.DeleteCommand)
			r.Post("/twitch/commands/exec", h.Twitch.ExecuteCommand)
			r.Post("/twitch/chat/send", h.Twitch.SendMessage)

			// YouTube endpoints
			r.Get("/youtube/status", h.YouTube.GetStatus)
			r.Get("/youtube/stats", h.YouTube.GetStats)
			r.Patch("/youtube/stream", h.YouTube.UpdateStream)

			// Config endpoints
			r.Get("/config", h.Config.GetConfig)
			r.Patch("/config", h.Config.UpdateConfig)

			// Restart endpoint
			r.Post("/restart", h.Restart.Restart)

			// Automation rule endpoints
			r.Get("/automation/rules", h.Automation.ListRules)
			r.Post("/automation/rules", h.Automation.CreateRule)
			r.Put("/automation/rules/{name}", h.Automation.UpdateRule)
			r.Delete("/automation/rules/{name}", h.Automation.DeleteRule)
			r.Post("/automation/rules/{name}/test", h.Automation.TestRule)
		})
	})

	// WebSocket endpoint (protected with query param auth)
	r.With(auth.WebSocketMiddleware(jwtService)).Get("/ws", h.WebSocket.HandleWebSocket)

	// Agent link endpoint. Agents authentifizieren sich mit dem gemeinsamen
	// API-Key, nicht mit einem Benutzer-JWT.
	if h.AgentLink != nil {
		r.Get("/ws/agent", h.AgentLink.ServeHTTP)
	}

	return r
}

// obsRoutes hängt die OBS-Endpunkte an einen Router. Beide Aufrufer teilen
// sich dieselben Handler; welcher Agent gemeint ist, liest der Handler aus
// dem Routenparameter {agent} — fehlt er, gilt der primäre.
func obsRoutes(r chi.Router, h *handlers.OBSHandler) {
	r.Get("/status", h.GetStatus)
	r.Get("/scenes", h.GetScenes)
	r.Post("/scenes/switch", h.SwitchScene)
	r.Get("/sources", h.GetSources)
	r.Post("/sources/toggle", h.ToggleSource)
	r.Post("/streaming/start", h.StartStreaming)
	r.Post("/streaming/stop", h.StopStreaming)
	r.Post("/recording/start", h.StartRecording)
	r.Post("/recording/stop", h.StopRecording)
	r.Post("/recording/pause", h.PauseRecording)
	r.Post("/recording/resume", h.ResumeRecording)
}
