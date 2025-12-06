package auth

import "github.com/golang-jwt/jwt/v5"

// Your secret signing key
var JWTSecret = []byte("supersecret-change-this")

// Claims structure
type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}
