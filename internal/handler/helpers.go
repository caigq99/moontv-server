package handler

import "github.com/gin-gonic/gin"

func getUserID(c *gin.Context) uint {
	id, _ := c.Get("user_id")
	return id.(uint)
}
