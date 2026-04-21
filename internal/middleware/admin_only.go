package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/moontv/server/pkg/response"
)

func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists || role != "admin" {
			response.Fail(c, http.StatusForbidden, response.ErrForbidden, "admin only")
			c.Abort()
			return
		}
		c.Next()
	}
}
