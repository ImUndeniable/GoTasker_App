package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"gotasker/internal/models"

	"github.com/go-chi/chi"
)

func GetTasksHandler(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	doneParam := strings.TrimSpace(r.URL.Query().Get("done"))
	limit := strings.TrimSpace(r.URL.Query().Get("limit"))
	offset := strings.TrimSpace(r.URL.Query().Get("offset"))

	var doneFilter *bool
	if doneParam != "" {
		val, err := strconv.ParseBool(doneParam)
		if err != nil {
			WriteJson(w, http.StatusBadRequest, map[string]string{"error": "Invalid bool type"})
			return
		}
		doneFilter = &val
	}

	out := make([]models.Task, 0, len(Tasks))
	for _, t := range Tasks {
		if doneFilter != nil && t.Done != *doneFilter {
			continue
		}
		if q != "" && !strings.Contains(strings.ToLower(t.Title), strings.ToLower(q)) {
			continue
		}
		out = append(out, t)
	}

	total := len(out)
	limitVal := total
	offsetVal := 0
	const maxLimit = 100

	if limit != "" {
		val, err := strconv.Atoi(limit)
		if err != nil || val <= 0 {
			WriteJson(w, http.StatusBadRequest, map[string]string{"error": "Invalid limit value"})
			return
		}
		if val > maxLimit {
			val = maxLimit
		}
		limitVal = val
	}

	if offset != "" {
		val, err := strconv.Atoi(offset)
		if err != nil || val < 0 {
			WriteJson(w, http.StatusBadRequest, map[string]string{"error": "Invalid offset value"})
			return
		}
		offsetVal = val
	}

	if offsetVal >= total {
		WriteJson(w, http.StatusOK, []models.Task{})
		return
	}

	end := offsetVal + limitVal
	if end > total {
		end = total
	}

	paged := out[offsetVal:end]
	WriteJson(w, http.StatusOK, paged)
}

func GetTaskByIDHandler(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		WriteJson(w, http.StatusBadRequest, map[string]string{"error": "Invalid task ID"})
		return
	}
	for _, task := range Tasks {
		if task.ID == id {
			WriteJson(w, http.StatusOK, task)
			return
		}
	}
	WriteJson(w, http.StatusNotFound, map[string]string{"error": "Task not found"})
}

func CreateTaskHandler(w http.ResponseWriter, r *http.Request) {
	var req models.CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJson(w, http.StatusBadRequest, map[string]string{"error": "Invalid Json"})
		return
	}
	if req.Title == "" {
		WriteJson(w, http.StatusBadRequest, map[string]string{"error": "Title is empty"})
		return
	}

	now := time.Now().UTC()
	newTask := models.Task{
		ID:        NextID(),
		Title:     req.Title,
		Done:      req.Done,
		CreatedAt: now,
		UpdatedAt: now,
	}

	Tasks = append(Tasks, newTask)
	w.Header().Set("Location", "/tasks/"+strconv.Itoa(newTask.ID))
	WriteJson(w, http.StatusCreated, newTask)
}

func PatchTaskHandler(w http.ResponseWriter, r *http.Request) {
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

	TasksMu.Lock()
	defer TasksMu.Unlock()

	for i := range Tasks {
		if Tasks[i].ID == id {
			updated := false
			if req.Title != nil {
				if *req.Title == "" {
					WriteJson(w, http.StatusBadRequest, map[string]string{"error": "Title Cannot be Empty"})
					return
				}
				Tasks[i].Title = *req.Title
				updated = true
			}
			if req.Done != nil {
				Tasks[i].Done = *req.Done
				updated = true
			}
			if updated {
				Tasks[i].UpdatedAt = time.Now().UTC()
			}
			WriteJson(w, http.StatusOK, Tasks[i])
			return
		}
	}
	WriteJson(w, http.StatusNotFound, map[string]string{"error": " Tasks not found"})
}

func DeleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	IdParam := chi.URLParam(r, "id")
	id, err := strconv.Atoi(IdParam)
	if err != nil {
		WriteJson(w, http.StatusBadRequest, map[string]string{"error": "ID not found"})
		return
	}

	TasksMu.Lock()
	defer TasksMu.Unlock()

	for i := range Tasks {
		if Tasks[i].ID == id {
			Tasks = append(Tasks[:i], Tasks[i+1:]...)
			w.WriteHeader(http.StatusNoContent)
			return
		}
	}
	WriteJson(w, http.StatusNotFound, map[string]string{"error": "Task not found"})
}
