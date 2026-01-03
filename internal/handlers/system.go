package handlers

import (
	"net/http"
	"time"

	"gotasker/internal/models"
)

func WelcomeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte("Welcome to GoTasker 🚀"))
}

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	TasksMu.Lock()
	count := len(Tasks)
	TasksMu.Unlock()

	resp := models.HealthResponse{
		Status:        "OK",
		UptimeSeconds: int64(time.Since(StartedAt).Seconds()),
		TasksCount:    count,
	}

	WriteJson(w, http.StatusOK, resp)
}
