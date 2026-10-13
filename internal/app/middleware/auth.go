package middleware

import (
	"errors"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	goredis "github.com/go-redis/redis/v8"

	"glottochronology/internal/app/ds"
	"glottochronology/internal/app/redis"
	"glottochronology/internal/app/role"
)

const jwtPrefix = "Bearer "

func JWTAuth(redisClient *redis.Client, allowedRoles ...role.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if !strings.HasPrefix(authHeader, jwtPrefix) {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		tokenStr := authHeader[len(jwtPrefix):]

		inBL, err := redisClient.CheckJWTInBlacklist(c.Request.Context(), tokenStr)
		if err != nil && !errors.Is(err, goredis.Nil) {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		if inBL {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		claims := &ds.JWTClaims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("JWT_SECRET")), nil
		})
		if err != nil || !token.Valid {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		if len(allowedRoles) > 0 {
			ok := false
			for _, r := range allowedRoles {
				if claims.Role == r {
					ok = true
					break
				}
			}
			if !ok {
				c.AbortWithStatus(http.StatusForbidden)
				return
			}
		}

		c.Set("userID", claims.UserID)
		c.Set("userLogin", claims.Login)
		c.Set("userRole", claims.Role)
		c.Next()
	}
}

func CurrentUserID(c *gin.Context) uint {
	if v, ok := c.Get("userID"); ok {
		if id, ok := v.(uint); ok {
			return id
		}
	}
	return 0
}
