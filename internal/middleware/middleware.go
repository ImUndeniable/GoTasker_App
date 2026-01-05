package middleware

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"
)

// --- Logging Logic ---

type statusWriter struct {
	http.ResponseWriter
	status int
	size   int
}

func (w *statusWriter) WriteHeader(code int) {
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

// --- Auth Logic ---

type ctxKey string

const CtxKeyAPIUser ctxKey = "api_user"

func validateAPIKey(key string) bool {
	expected := os.Getenv("GOTASKER_API_KEY")
	log.Printf("DEBUG → Expected key: [%s], Received key: [%s]", expected, key)
	if expected == "" {
		expected = "dev-secret-key"
	}
	return key == expected
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get("X-API-Key")
		if apiKey == "" || !validateAPIKey(apiKey) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized Key"})
			return
		}
		ctx := context.WithValue(r.Context(), CtxKeyAPIUser, "api-key-user")
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
