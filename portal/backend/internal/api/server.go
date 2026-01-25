package api

import (
	"context"
	"log"
	"net/http"

	"kit.workmate/gaming-portal/internal/config"
)

type Server struct {
	httpServer *http.Server
}

func New(cfg config.ServerConfig, handler http.Handler) *Server {
	return &Server{
		httpServer: &http.Server{
			Addr:         cfg.Addr(),
			Handler:      handler,
			ReadTimeout:  cfg.Timeouts.Read,
			WriteTimeout: cfg.Timeouts.Write,
		},
	}
}

func (s *Server) Start() {
	go func() {
		log.Printf("Portal server listening on %s", s.httpServer.Addr)
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()
}

func (s *Server) Shutdown(ctx context.Context) error {
	log.Println("Shutting down portal server")
	return s.httpServer.Shutdown(ctx)
}
