package middleware

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/TheAmgadX/gopher-chat/internal/utils"
	"github.com/golang-jwt/jwt/v5"
)

// --- Best Practice: Define constants for magic strings ---
const (
	authHeaderKey = "Authorization"
	authScheme    = "bearer"
)

type contextKey string

const UserClaimsKey contextKey = "userClaims"

var JWT_KEY = []byte(os.Getenv("JWT_SECRET_KEY"))

type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func GenerateJWT(username string) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)

	claims := &Claims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			// Also add IssuedAt for good practice
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(JWT_KEY)

	return tokenString, err
}

func ValidJWT(tokenStr string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return JWT_KEY, nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid token")
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Use the constant for the header key
		auth := r.Header.Get(authHeaderKey)

		if auth == "" {
			utils.WriteJsonErrors(w, "missing Authorization header", http.StatusUnauthorized)
			return
		}

		parts := strings.SplitN(auth, " ", 2)

		// Use the constant for the scheme and compare case-insensitively
		if len(parts) != 2 || strings.ToLower(parts[0]) != authScheme {
			utils.WriteJsonErrors(w, "invalid Authorization header format", http.StatusUnauthorized)
			return
		}

		claims, err := ValidJWT(parts[1])
		if err != nil {
			// The error from ValidJWT is now safe to show the user
			utils.WriteJsonErrors(w, err.Error(), http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserClaimsKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
