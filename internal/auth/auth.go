package auth

import (
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
)

var JWT_KEY = []byte("amjad_secret_key")

type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

func GenerateJWT(username string) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)

	claims := &Claims{
		Username: username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(JWT_KEY)

	return tokenString, err
}

func ValidJWT(tokenStr string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return JWT_KEY, nil
	})

	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			return nil, fmt.Errorf("invalid token signature")
		}

		return nil, fmt.Errorf("couldn't parse token %v", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}
