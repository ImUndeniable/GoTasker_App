package main

import (
	"log"
	"net/http"
	"time"

	"github.com/go-chi/cors"
	"github.com/redis/go-redis/v9"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"

	"gotasker/internal/auth"
	"gotasker/internal/handlers"
	customMiddleware "gotasker/internal/middleware"
	internalRedis "gotasker/internal/redis"

	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib"
)

var redisClient *redis.Client

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

	//Redis
	redisClient, err = internalRedis.InitRedis()
	if err != nil {
		log.Printf("Redis unavailable, continuing without cache: %v", err)
		redisClient = nil
	}

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	//CORS - Frontend <-> Backend Conection
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173"},
		AllowedMethods:   []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-API-Key"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	// custom logging for request
	r.Use(customMiddleware.LoggingMiddleWare)

	// public routes

	r.Get("/health", handlers.HealthHandler)
	r.Get("/tasks", handlers.GetTasksHandler)
	r.Post("/register", handlers.RegisterHandler(db))
	r.Post("/login", handlers.LoginHandler(db))

	// protected routes group
	r.Group(func(r chi.Router) {
		r.Use(auth.JWTMiddleware)
		r.Get("/tasksdb", handlers.GetTasksHandlerDB(db, redisClient))
		r.Get("/tasksdb/{id}", handlers.GetTaskbyIDHandlerDB(db))
		r.Post("/tasksdb", handlers.CreateTaskHandlerDB(db, redisClient))
		r.Patch("/tasksdb/{id}", handlers.PatchTaskHandlerDB(db, redisClient))
		r.Delete("/tasksdb/{id}", handlers.DeleteTaskHandlerDB(db, redisClient))
		r.Get("/tasks/{id}", handlers.GetTaskByIDHandler)
		r.Post("/tasks", handlers.CreateTaskHandler)
		r.Patch("/tasks/{id}", handlers.PatchTaskHandler)
		r.Delete("/tasks/{id}", handlers.DeleteTaskHandler)

	})

	// Health Checkup of the API
	log.Println("Server is running on port 8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
