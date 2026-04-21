package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/moontv/server/internal/database"
	"github.com/moontv/server/internal/model"
	"github.com/moontv/server/pkg/apikey"
	"github.com/moontv/server/pkg/response"
)

func APIKeyAuth(secret []byte, prefix string) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.GetHeader("X-API-Key")
		if key == "" {
			key = c.Query("apikey")
		}
		if key == "" {
			response.Fail(c, http.StatusUnauthorized, response.ErrNoAPIKey, "missing api key")
			c.Abort()
			return
		}

		userID, err := apikey.Validate(secret, key, prefix)
		if err != nil {
			response.Fail(c, http.StatusUnauthorized, response.ErrInvalidAPIKey, "invalid api key")
			c.Abort()
			return
		}

		var user model.User
		if err := database.DB.First(&user, userID).Error; err != nil {
			response.Fail(c, http.StatusUnauthorized, response.ErrInvalidAPIKey, "user not found")
			c.Abort()
			return
		}

		if user.Banned {
			response.Fail(c, http.StatusForbidden, response.ErrBanned, "account banned")
			c.Abort()
			return
		}

		if user.APIKeyCipher == "" {
			response.Fail(c, http.StatusUnauthorized, response.ErrInvalidAPIKey, "api key revoked")
			c.Abort()
			return
		}

		c.Set("user_id", user.ID)
		c.Set("username", user.Username)
		c.Set("role", user.Role)
		c.Next()
	}
}
