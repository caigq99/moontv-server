package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Code    int    `json:"code"`
	Data    any    `json:"data"`
	Message string `json:"message"`
}

func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Response{Code: 0, Data: data, Message: "ok"})
}

func Fail(c *gin.Context, httpStatus int, code int, msg string) {
	c.JSON(httpStatus, Response{Code: code, Data: nil, Message: msg})
}

// Error codes
const (
	ErrUnauthorized    = 40001
	ErrInvalidToken    = 40002
	ErrBanned          = 40003
	ErrNoAPIKey        = 40004
	ErrInvalidAPIKey   = 40005
	ErrForbidden       = 40006
	ErrBadRequest      = 40101
	ErrInvalidParam    = 40102
	ErrDuplicate       = 40103
	ErrInviteInvalid   = 40104
	ErrNotFound        = 40401
	ErrInternal        = 50001
)
