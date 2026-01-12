package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"gotasker/internal/auth"
	"gotasker/internal/models"
	cache "gotasker/internal/redis"

	"github.com/go-chi/chi"
	"github.com/redis/go-redis/v9"
)

func GetTasksHandlerDB(db *sql.DB, rdb *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// 1. Extract userID from JWT context
		userIDVal := r.Context().Value(auth.UserIDContextKey)
		if userIDVal == nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		userID, ok := userIDVal.(int64)
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		ctx := r.Context()

		// 2. Try Redis cache first
		if tasks, hit, err := cache.GetTasks(ctx, rdb, userID); err == nil && hit {
			log.Println("Redis cache HIT")
			WriteJson(w, http.StatusOK, tasks)
			return
		}
		log.Println("Redis cache MISS")

		// 3. Query DB
		rows, err := db.Query(`
			SELECT id, title, done, created_at, updated_at
			FROM tasks
			WHERE user_id = $1
			ORDER BY created_at DESC
		`, userID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		// 4. Build tasks slice
		tasks := make([]models.Task, 0)

		for rows.Next() {
			var t models.Task
			if err := rows.Scan(
				&t.ID,
				&t.Title,
				&t.Done,
				&t.CreatedAt,
				&t.UpdatedAt,
			); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			tasks = append(tasks, t)
		}

		if err := rows.Err(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// 5. Store result in Redis
		if err := cache.SetTasks(ctx, rdb, userID, tasks); err != nil {
			log.Printf("Redis SET failed: %v", err)
		}

		// 6. Return response ONCE
		WriteJson(w, http.StatusOK, tasks)
	}
}

func GetTaskbyIDHandlerDB(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idParam := chi.URLParam(r, "id")
		id, err := strconv.Atoi(idParam)
		if err != nil {
			WriteJson(w, http.StatusBadRequest, map[string]string{"error": "Invalid task ID"})
			return
		}

		userIDVal := r.Context().Value(auth.UserIDContextKey)
		if userIDVal == nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		userID, ok := userIDVal.(int64)
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		rows, err := db.Query(`
			SELECT id, title, done, created_at, updated_at
			FROM tasks
			WHERE user_id = $1 AND id = $2
			ORDER BY created_at DESC`, userID, id)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var t models.Task
			err := rows.Scan(
				&t.ID,
				&t.Title,
				&t.Done,
				&t.CreatedAt,
				&t.UpdatedAt,
			)

			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			WriteJson(w, http.StatusOK, t)
			return
		}
		WriteJson(w, http.StatusNotFound, map[string]string{"error": "Task not found"})
	}
}

func CreateTaskHandlerDB(db *sql.DB, rdb *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.CreateTaskRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			WriteJson(w, http.StatusBadRequest, map[string]string{"error": "Invalid Json"})
			return
		}

		if req.Title == "" {
			WriteJson(w, http.StatusBadRequest, map[string]string{"error": "Title is empty"})
			return
		}

		userIDVal := r.Context().Value(auth.UserIDContextKey)
		if userIDVal == nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		userID, ok := userIDVal.(int64)
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		var task models.Task

		err := db.QueryRow(`
			INSERT INTO tasks (user_id, title, done)
			VALUES ($1, $2, $3)
			RETURNING id, title, done, created_at, updated_at`, userID, req.Title, req.Done).Scan(
			&task.ID,
			&task.Title,
			&task.Done,
			&task.CreatedAt,
			&task.UpdatedAt,
		)

		if err != nil {
			WriteJson(w, http.StatusInternalServerError, map[string]string{
				"error": "failed to create task",
			})
			return
		}

		//Using Redis to delete the data
		ctx := r.Context()

		if err := cache.DeletTaks(ctx, rdb, userID); err != nil {
			log.Printf("Redis DEL failed: %v", err)
		}

		WriteJson(w, http.StatusCreated, task)
	}
}

func PatchTaskHandlerDB(db *sql.DB, rdb *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idParam := chi.URLParam(r, "id")
		id, err := strconv.Atoi(idParam)
		if err != nil {
			WriteJson(w, http.StatusBadRequest, map[string]string{"Error": "Invalid Request"})
			return
		}

		var req models.UpdateTaskRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			WriteJson(w, http.StatusBadRequest, map[string]string{"error": "Invalid Json"})
			return
		}

		if req.Title != nil {
			trimmed := strings.TrimSpace(*req.Title)
			req.Title = &trimmed
		}

		//DB logic

		userIDVal := r.Context().Value(auth.UserIDContextKey)
		if userIDVal == nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		userID, ok := userIDVal.(int64)
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		row, err := db.Query(`
			UPDATE tasks SET
            title = COALESCE($1, title),
            done = COALESCE($2, done),
            updated_at = NOW()
            WHERE
            id = $3
            AND user_id = $4
            RETURNING
        id, title, done, created_at, updated_at`, req.Title, req.Done, id, userID)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		for row.Next() {
			var t models.Task
			err := row.Scan(
				&t.ID,
				&t.Title,
				&t.Done,
				&t.CreatedAt,
				&t.UpdatedAt,
			)

			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			WriteJson(w, http.StatusOK, t)
			return
		}
		//Using Redis to delete the data
		ctx := r.Context()

		if err := cache.DeletTaks(ctx, rdb, userID); err != nil {
			log.Printf("Redis DEL failed: %v", err)
		}

		WriteJson(w, http.StatusNotFound, map[string]string{"error": "Task not found"})

	}
}

func DeleteTaskHandlerDB(db *sql.DB, rdb *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		IDParam := chi.URLParam(r, "id")
		id, err := strconv.Atoi(IDParam)
		if err != nil {
			WriteJson(w, http.StatusBadRequest, map[string]string{"error": "ID not found/invalid"})
			return
		}

		userIDVal := r.Context().Value(auth.UserIDContextKey)
		if userIDVal == nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		userID, ok := userIDVal.(int64)
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		//var task models.Task
		res, err := db.Exec(`DELETE FROM tasks
               WHERE id = $1
               AND user_id = $2`, id, userID)

		if err != nil {
			WriteJson(w, http.StatusInternalServerError, map[string]string{
				"error": "failed to create task",
			})
			return
		}
		rowsAffected, err := res.RowsAffected()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if rowsAffected == 0 {
			WriteJson(w, http.StatusNotFound, map[string]string{
				"error": "Task not found"})
			return
		}

		//Using Redis to delete the data
		ctx := r.Context()

		if err := cache.DeletTaks(ctx, rdb, userID); err != nil {
			log.Printf("Redis DEL failed: %v", err)
		}

		w.WriteHeader(http.StatusNoContent)

	}
}
