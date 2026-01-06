package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const UserIDContextKey contextKey = "user_id"

func JWTMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		//1. Read Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "missing authorization header", http.StatusUnauthorized)
			return
		}

		//2. Expect: Bearer <token>
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "invalid authorization header", http.StatusUnauthorized)
			return
		}

		tokenString := parts[1]

		//3. Parse & Verify token
		token, err := jwt.ParseWithClaims(
			tokenString,
			&JWTClaims{},
			func(t *jwt.Token) (interface{}, error) {
				return JwtSecret, nil
			},
		)

		if err != nil || !token.Valid {
			http.Error(w, "invalid or expired token", http.StatusUnauthorized)
			return
		}

		//4. Extract claims
		claims, ok := token.Claims.(*JWTClaims)
		if !ok {
			http.Error(w, "invalid token claims", http.StatusUnauthorized)
			return
		}

		//5. Store user_id in context
		ctx := context.WithValue(r.Context(), UserIDContextKey, claims.UserID)

		//6. Continue request
		next.ServeHTTP(w, r.WithContext(ctx))
	})

}

func userIDContextKey(ctx context.Context) (int64, bool) {
	userID, ok := ctx.Value(UserIDContextKey).(int64)
	return userID, ok
}
