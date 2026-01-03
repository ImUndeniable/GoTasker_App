package handlers

import (
	"sync"
	"time"

	"gotasker/internal/models"
)

// Global state variables
var (
	TasksMu   sync.Mutex
	StartedAt = time.Now()
	Tasks     = []models.Task{
		{ID: 1, Title: "Learn Go Basics", Done: true, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()},
		{ID: 2, Title: "Setup GoTasker App", Done: false, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()},
		{ID: 3, Title: "Learn HTTP Status Code", Done: false, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()},
	}
)
