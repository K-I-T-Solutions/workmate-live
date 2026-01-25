package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"kit.workmate/live-portal/internal/api/handlers"
)

type Handlers struct {
	Auth      *handlers.AuthHandler
	Agent     *handlers.AgentHandler
	WebSocket *handlers.WebSocketHandler
	OBS       *handlers.OBSHandler
	Twitch    *handlers.TwitchHandler
	YouTube   *handlers.YouTubeHandler
}

func Routes(h *Handlers) http.Handler {
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
		// Auth endpoints
		r.Route("/auth", func(r chi.Router) {
			r.Post("/login", h.Auth.Login)
			r.Post("/logout", h.Auth.Logout)
			r.Get("/verify", h.Auth.Verify)
		})

		// Agent proxy endpoints (Phase 2)
		r.Route("/agent", func(r chi.Router) {
			r.Get("/status", h.Agent.GetStatus)
			r.Get("/capabilities", h.Agent.GetCapabilities)
			r.Get("/info", h.Agent.GetInfo)
		})

		// OBS control endpoints (Phase 3)
		r.Route("/obs", func(r chi.Router) {
			r.Get("/status", h.OBS.GetStatus)
			r.Get("/scenes", h.OBS.GetScenes)
			r.Post("/scenes/switch", h.OBS.SwitchScene)
			r.Get("/sources", h.OBS.GetSources)
			r.Post("/sources/toggle", h.OBS.ToggleSource)

			// Streaming control
			r.Post("/streaming/start", h.OBS.StartStreaming)
			r.Post("/streaming/stop", h.OBS.StopStreaming)

			// Recording control
			r.Post("/recording/start", h.OBS.StartRecording)
			r.Post("/recording/stop", h.OBS.StopRecording)
			r.Post("/recording/pause", h.OBS.PauseRecording)
			r.Post("/recording/resume", h.OBS.ResumeRecording)
		})

		// Twitch endpoints (Phase 4)
		r.Route("/twitch", func(r chi.Router) {
			r.Get("/status", h.Twitch.GetStatus)
			r.Get("/stats", h.Twitch.GetStats)
			r.Patch("/stream", h.Twitch.UpdateStream)
		})

		// YouTube endpoints
		r.Route("/youtube", func(r chi.Router) {
			r.Get("/status", h.YouTube.GetStatus)
			r.Get("/stats", h.YouTube.GetStats)
			r.Patch("/stream", h.YouTube.UpdateStream)
		})
	})

	// WebSocket endpoint (Phase 2)
	r.Get("/ws", h.WebSocket.HandleWebSocket)

	return r
}
