package ds

import (
	"glottochronology/internal/app/role"
	"github.com/golang-jwt/jwt/v5"
)

type JWTClaims struct {
	jwt.RegisteredClaims
	UserID uint      `json:"user_id"`
	Login  string    `json:"login"`
	Role   role.Role `json:"role"`
}
