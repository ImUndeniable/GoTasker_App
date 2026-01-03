package handlers

import (
	"encoding/json"
	"net/http"
)

func WriteJson(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func NextID() int {
	max := 0
	for _, t := range Tasks {
		if t.ID > max {
			max = t.ID
		}
	}
	return max + 1
}
