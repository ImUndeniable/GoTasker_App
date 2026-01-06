package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"gotasker/internal/auth"
	"gotasker/internal/models"

	"github.com/go-chi/chi"
)

func GetTasksHandlerDB(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
			WHERE user_id = $1
			ORDER BY created_at DESC`, userID)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		tasks := make([]models.Task, 0)

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
			tasks = append(tasks, t)
		}

		if err := rows.Err(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
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

func CreateTaskHandlerDB(db *sql.DB) http.HandlerFunc {
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
		WriteJson(w, http.StatusCreated, task)
	}
}

func PatchTaskHandlerDB(db *sql.DB) http.HandlerFunc {
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
		WriteJson(w, http.StatusNotFound, map[string]string{"error": "Task not found"})

	}
}

func DeleteTaskHandlerDB(db *sql.DB) http.HandlerFunc {
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
		w.WriteHeader(http.StatusNoContent)

	}
}
