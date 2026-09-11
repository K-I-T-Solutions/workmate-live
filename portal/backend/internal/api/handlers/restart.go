package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"syscall"
	"time"
)

type RestartHandler struct{}

func NewRestartHandler() *RestartHandler {
	return &RestartHandler{}
}

func (h *RestartHandler) Restart(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "restarting"})

	go func() {
		time.Sleep(500 * time.Millisecond)
		log.Println("Restart requested via API, sending SIGTERM...")
		syscall.Kill(os.Getpid(), syscall.SIGTERM)
	}()
}
