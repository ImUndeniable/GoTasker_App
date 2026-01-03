package main

import (
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"

	"gotasker/internal/handlers"
	customMiddleware "gotasker/internal/middleware"

	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	// DB Connection
	db, err := sql.Open("pgx", "postgres://gotasker:gotasker@localhost:5432/gotasker?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		log.Fatal("Cannot connect to Database", err)
	}
	defer db.Close()

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// custom logging for request
	r.Use(customMiddleware.LoggingMiddleWare)

	// public routes
	r.Get("/tasksdb", handlers.GetTasksHandlerDB(db))
	r.Get("/tasksdb/{id}", handlers.GetTaskbyIDHandlerDB(db))
	r.Get("/", handlers.WelcomeHandler)
	r.Get("/health", handlers.HealthHandler)
	r.Get("/tasks", handlers.GetTasksHandler)
	r.Get("/tasks/{id}", handlers.GetTaskByIDHandler)

	// protected routes group
	r.Group(func(r chi.Router) {
		r.Use(customMiddleware.AuthMiddleware)
		r.Post("/tasksdb", handlers.CreateTaskHandlerDB(db))
		r.Patch("/tasksdb/{id}", handlers.PatchTaskHandlerDB(db))
		r.Post("/tasks", handlers.CreateTaskHandler)
		r.Patch("/tasks/{id}", handlers.PatchTaskHandler)
		r.Delete("/tasks/{id}", handlers.DeleteTaskHandler)
		r.Delete("/tasksdb/{id}", handlers.DeleteTaskHandlerDB(db))
	})

	// Health Checkup of the API
	log.Println("Server is running on port 8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
