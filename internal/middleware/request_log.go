package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/moontv/server/internal/database"
	"github.com/moontv/server/internal/model"
)

func RequestLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		uid, exists := c.Get("user_id")
		if !exists {
			return
		}

		log := model.APICallLog{
			UserID:    uid.(uint),
			Endpoint:  c.Request.Method + " " + c.FullPath(),
			IP:        c.ClientIP(),
			CreatedAt: time.Now(),
		}
		database.DB.Create(&log)
	}
}
