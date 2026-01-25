package api

import (
	"context"
	"log"
	"net/http"
	"time"

	"kit.workmate/gaming-agent/internal/config"
)

type Server struct {
	httpServer *http.Server
}

func New(addr string, handler http.Handler) *Server {
	return &Server{
		httpServer: &http.Server{
			Addr:         addr,
			Handler:      handler,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 5 * time.Second,
		},
	}
}

func NewWithConfig(addr string, handler http.Handler, timeouts config.TimeoutConfig) *Server {
	return &Server{
		httpServer: &http.Server{
			Addr:         addr,
			Handler:      handler,
			ReadTimeout:  timeouts.Read,
			WriteTimeout: timeouts.Write,
		},
	}
}

func (s *Server) Start() {
	go func() {
		log.Printf("API listening on %s", s.httpServer.Addr)
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()
}

func (s *Server) Shutdown(ctx context.Context) error {
	log.Println("Shutting down API server")
	return s.httpServer.Shutdown(ctx)
}
