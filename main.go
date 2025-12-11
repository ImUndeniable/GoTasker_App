package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

var tasksMu sync.Mutex
var startedAt = time.Now()

// struct for statusWriter
type statusWriter struct {
	http.ResponseWriter
	status int
	size   int
}

// methods for logging the final status and size with proper wrapping.
func (w *statusWriter) WriterHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *statusWriter) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	n, err := w.ResponseWriter.Write(b)
	w.size += n
	return n, err
}

// Logging Middleware
func LoggingMiddleWare(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sw := &statusWriter{ResponseWriter: w}
		next.ServeHTTP(sw, r)
		latency := time.Since(start)
		ip := r.RemoteAddr
		status := sw.status
		if status == 0 {
			status = http.StatusOK
		}
		log.Printf("%s %s %s %d %dB %s", r.Method, r.URL.Path, ip, status, sw.size, latency)
	})
}

// Simple API Key Auth Middlewares

type ctxKey string

const ctxKeyAPIUser ctxKey = "api_user"

// API Key validation
func validateAPIKey(key string) bool {
	expected := os.Getenv("GOTASKER_API_KEY")
	if expected == "" {
		expected = "dev-secet-key"
	}
	return key == expected
}

// Auth Middleware
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get("X-API-Key")
		if apiKey == "" || !validateAPIKey(apiKey) {
			writeJson(w, http.StatusUnauthorized, map[string]string{"error": "Unauthorized Key"})
			return
		}
		ctx := context.WithValue(r.Context(), ctxKeyAPIUser, "api-key-user")
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

type HealthResponse struct {
	Status        string `json:"status"`
	UptimeSeconds int64  `json:"uptime_seconds"`
	TasksCount    int    `json:"tasks_count"`
}

type Task struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Done      bool      `json:"done"`
	CreatedAt time.Time `json:"created_at"`
	UpadtedAt time.Time `json:"updated_at",omitempty`
}

type CreateTaskRequest struct {
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

type UpdateTaskRequest struct {
	Title *string `json:"title,omitempty"`
	Done  *bool   `json:"done,omitempty"`
}

var tasks = []Task{
	{ID: 1, Title: "Learn Go Basics", Done: true, CreatedAt: time.Now().UTC(), UpadtedAt: time.Now().UTC()},
	{ID: 2, Title: "Setup GoTasker App", Done: false, CreatedAt: time.Now().UTC(), UpadtedAt: time.Now().UTC()},
	{ID: 3, Title: "Learn HTTP Status Code", Done: false, CreatedAt: time.Now().UTC(), UpadtedAt: time.Now().UTC()},
}

// Handlers

func welcomeHander(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte("Welcome to GoTasker 🚀"))
}

func healthHander(w http.ResponseWriter, r *http.Request) {
	tasksMu.Lock()
	count := len(tasks)
	tasksMu.Unlock()

	resp := HealthResponse{
		Status:        "OK",
		UptimeSeconds: int64(time.Since(startedAt).Seconds()),
		TasksCount:    count,
	}

	writeJson(w, http.StatusOK, resp)
}

func getTasksHander(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))            // search keyword
	doneParam := strings.TrimSpace(r.URL.Query().Get("done")) // true or false
	//Pagination
	limit := strings.TrimSpace(r.URL.Query().Get("limit"))   // get the limit value
	offset := strings.TrimSpace(r.URL.Query().Get("offset")) // offset value

	var doneFilter *bool
	if doneParam != "" {
		val, err := strconv.ParseBool(doneParam)
		if err != nil {
			writeJson(w, http.StatusBadRequest, map[string]string{"error": "Invalid bool type"})
			return
		}

		doneFilter = &val
	}

	out := make([]Task, 0, len(tasks))
	for _, t := range tasks {
		if doneFilter != nil && t.Done != *doneFilter {
			continue
		}

		if q != "" && !strings.Contains(strings.ToLower(t.Title), strings.ToLower(q)) {
			continue
		}
		out = append(out, t)
	}
	//writeJson(w, http.StatusOK, out)

	//Pagination

	//Limit
	total := len(out)
	limitVal := total // default: return all filtered tasks
	offsetVal := 0    // default: start at 0

	const maxLimit = 100

	if limit != "" {
		val, err := strconv.Atoi(limit)
		if err != nil || val <= 0 {
			writeJson(w, http.StatusBadRequest, map[string]string{"error": "Invalid limit value"})
			return
		}

		if val > maxLimit {
			val = maxLimit
		}
		limitVal = val

	}

	// Offset
	if offset != "" {
		val, err := strconv.Atoi(offset)
		if err != nil || val < 0 {
			writeJson(w, http.StatusBadRequest, map[string]string{"error": "Invalid offset value"})
			return
		}
		offsetVal = val
	}

	// If Offset is beyond avaialable items, return empty list
	if offsetVal >= total {
		writeJson(w, http.StatusOK, []Task{})
		return
	}

	// Compute and index safely
	end := offsetVal + limitVal
	if end > total {
		end = total
	}

	// Slice and respond
	paged := out[offsetVal:end]
	writeJson(w, http.StatusOK, paged)

	//1 Read query parameters - Done
	//2 Parse done filter if provided - Done
	//3 Build filtered list - Done
	//3 a Filter by done = true/fasle if provided - Done
	//3 b Filter by q in title (case - insensitive) if provided - Done

}

func getTaskByIDHandler(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id") // get the id from the url

	id, err := strconv.Atoi(idParam) // convert the id to an integer
	if err != nil {
		writeJson(w, http.StatusBadRequest, map[string]string{"error": "Invalid task ID"})
		return
	}
	// how to optimize
	for _, task := range tasks {
		if task.ID == id {
			writeJson(w, http.StatusOK, task)
			return
		}
	}
	// if the task is not found, return a 404 error
	writeJson(w, http.StatusNotFound, map[string]string{"error": "Task not found"})

}

func createTaskHandler(w http.ResponseWriter, r *http.Request) {
	var req CreateTaskRequest

	// Decode the Json Body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJson(w, http.StatusBadRequest, map[string]string{"error": "Invalid Json"})
		return
	}

	// Basic Validation
	if req.Title == "" {
		writeJson(w, http.StatusBadRequest, map[string]string{"error": "Title is empty"})
		return
	}

	now := time.Now().UTC()

	newTask := Task{
		ID:        nextID(),
		Title:     req.Title,
		Done:      req.Done,
		CreatedAt: now,
		UpadtedAt: now,
	}

	tasks = append(tasks, newTask)

	w.Header().Set("Location", "/tasks/"+strconv.Itoa(newTask.ID))
	writeJson(w, http.StatusCreated, newTask)

}

func patchTaskHandler(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		writeJson(w, http.StatusBadRequest, map[string]string{"Error": "Invalid Request"})
		return
	}

	var req UpdateTaskRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJson(w, http.StatusBadRequest, map[string]string{"error": "Invalid Json"})
		return
	}

	if req.Title != nil {
		trimmed := strings.TrimSpace(*req.Title)
		req.Title = &trimmed
	}

	tasksMu.Lock()
	defer tasksMu.Unlock()

	// rewrite
	for i := range tasks {
		if tasks[i].ID == id {
			updated := false
			if req.Title != nil {
				if *req.Title == "" {
					writeJson(w, http.StatusBadRequest, map[string]string{"error": "Title Cannot be Empty"})
					return
				}
				tasks[i].Title = *req.Title
				updated = true
			}
			if req.Done != nil {
				tasks[i].Done = *req.Done
				updated = true
			}

			if updated {
				tasks[i].UpadtedAt = time.Now().UTC()
			}

			writeJson(w, http.StatusOK, tasks[i])
			return
		}

		writeJson(w, http.StatusNotFound, map[string]string{"error": " Tasks not found"})

	}

}

func deleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	IdParam := chi.URLParam(r, "id")
	id, err := strconv.Atoi(IdParam)
	if err != nil {
		writeJson(w, http.StatusBadRequest, map[string]string{"error": "ID not found"})
	}

	tasksMu.Lock()
	defer tasksMu.Unlock()

	for i := range tasks {
		if tasks[i].ID == id {
			tasks = append(tasks[:i], tasks[i+1:]...)
			w.WriteHeader(http.StatusNoContent)
			return
		}
	}

	writeJson(w, http.StatusNotFound, map[string]string{"error": "Task not found"})

}

func main() {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// custom logging for request
	r.Use(LoggingMiddleWare)

	// public routes
	r.Get("/", welcomeHander)
	r.Get("/health", healthHander)
	r.Get("/tasks", getTasksHander)
	r.Get("/tasks/{id}", getTaskByIDHandler)

	// protected routes group
	r.Group(func(r chi.Router) {
		r.Use(AuthMiddleware)
		r.Post("/tasks", createTaskHandler)
		r.Patch("/tasks/{id}", patchTaskHandler)
		r.Delete("/tasks/{id}", deleteTaskHandler)
	})

	// Health Checkup of the API

	log.Println("Server is running on port 8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatalf("server failed: %v", err)
	}

	// PATCH Step by Step

	// 1. Parse ID - done
	// 2. Decode request body into DTO with pointer variable  fields - done
	// 3. Normalize title if provided - done
	// 4. Lock shared state for safe mutation -done
	// 5. Find task and apply updates (only fields that were providied) - done
	// 6. Not found - done

	// Delete Step by Step

	// 1. Parse ID - done
	// 2. Lock shared state for safe mutation - done
	// 3. Remove the tasks at index - done
	// 4. Not found
}

func writeJson(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func nextID() int {
	max := 0
	for _, t := range tasks {
		if t.ID > max {
			max = t.ID
		}
	}
	return max + 1
}

// git pushed
