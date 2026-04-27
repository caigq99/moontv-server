package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/moontv/server/internal/repository"
	"github.com/moontv/server/pkg/response"
	"gorm.io/gorm"
)

type JWTClaims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

func JWTAuth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
			response.Fail(c, http.StatusUnauthorized, response.ErrUnauthorized, "missing token")
			c.Abort()
			return
		}

		tokenStr := auth[7:]
		claims := &JWTClaims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
			return []byte(secret), nil
		})
		if err != nil || !token.Valid {
			response.Fail(c, http.StatusUnauthorized, response.ErrInvalidToken, "invalid token")
			c.Abort()
			return
		}

		user, err := repository.GetUserByID(claims.UserID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				response.Fail(c, http.StatusUnauthorized, response.ErrInvalidToken, "invalid token")
			} else {
				response.Fail(c, http.StatusInternalServerError, response.ErrInternal, "failed to load user")
			}
			c.Abort()
			return
		}

		if user.Banned {
			response.Fail(c, http.StatusForbidden, response.ErrBanned, "account banned")
			c.Abort()
			return
		}

		if user.Role != claims.Role {
			response.Fail(c, http.StatusUnauthorized, response.ErrInvalidToken, "stale token")
			c.Abort()
			return
		}

		c.Set("user_id", user.ID)
		c.Set("username", user.Username)
		c.Set("role", user.Role)
		c.Next()
	}
}
