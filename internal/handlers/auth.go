package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"gotasker/internal/auth"
	"gotasker/internal/models"

	"github.com/golang-jwt/jwt/v5"
)

//step 1 - Convert the incoming JSON into GO - Done
//step 2 - Normalize email and pass, check empty and check password len - Done
//step 3 - Implement HashPassword  - Done
//step 4 - Write the details to DB return ID and Email - Done
//step 5 - check dulipcate email and failed to create user - Done
//step 6 - writeJson to client - Done

func RegisterHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.RegisterRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			WriteJson(w, http.StatusBadRequest, map[string]string{"error": "Invalid JSON"})
			return
		}

		req.Email = strings.TrimSpace(strings.ToLower(req.Email))
		req.Password = strings.TrimSpace(req.Password)

		if req.Email == "" || req.Password == "" {
			WriteJson(w, http.StatusBadRequest, map[string]string{
				"error": "email and password are required",
			})
			return
		}

		if len(req.Password) < 8 {
			WriteJson(w, http.StatusBadRequest, map[string]string{
				"error": "The must contain more than 8 characters",
			})
			return
		}

		hashedPassword, err := auth.HashPassword(req.Password)
		if err != nil {
			WriteJson(w, http.StatusInternalServerError, map[string]string{
				"error": "failed to secure password",
			})
			return
		}

		var user models.UserResponse
		err = db.QueryRow(`
			INSERT INTO users (email, password_hash)
			VALUES ($1, $2)
			RETURNING id, email
		`, req.Email, hashedPassword).Scan(&user.ID, &user.Email)

		if err != nil {
			if strings.Contains(err.Error(), "duplicate key") {
				WriteJson(w, http.StatusConflict, map[string]string{
					"error": "email already registered",
				})
				return
			}
			WriteJson(w, http.StatusInternalServerError, map[string]string{
				"error": "failed to create user",
			})
			return
		}

		WriteJson(w, http.StatusCreated, user)
	}
}

func LoginHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.LoginRequest

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			WriteJson(w, http.StatusBadRequest, map[string]string{"error": "Invalid JSON"})
			return
		}

		req.Email = strings.TrimSpace(strings.ToLower(req.Email))
		req.Password = strings.TrimSpace(req.Password)

		if req.Email == "" || req.Password == "" {
			WriteJson(w, http.StatusBadRequest, map[string]string{
				"error": "email and password are required",
			})
			return
		}

		var user models.LoginAuth
		err := db.QueryRow(`
			SELECT id, email, password_hash
			FROM users
			WHERE email = $1
		`, req.Email).Scan(&user.ID, &user.Email, &user.PasswordHash)

		if err != nil {
			WriteJson(w, http.StatusUnauthorized, map[string]string{
				"error": "invalid credentials",
			})
			return

		}

		if err := auth.ComparePassword(user.PasswordHash, req.Password); err != nil {
			WriteJson(w, http.StatusUnauthorized, map[string]string{
				"error": "invalid credentials",
			})
			return
		}

		//w.WriteHeader(http.StatusCreated)

		//============== JWT CREATION ==============
		claims := auth.JWTClaims{
			UserID: user.ID,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

		tonkenstring, err := token.SignedString(auth.JwtSecret)
		if err != nil {
			WriteJson(w, http.StatusInternalServerError, map[string]string{
				"error": "failed to generate token",
			})
			return
		}
		WriteJson(w, http.StatusOK, map[string]string{
			"token": tonkenstring,
		})

	}
}
