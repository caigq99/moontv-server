package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/moontv/server/internal/repository"
	"github.com/moontv/server/pkg/apikey"
	"github.com/moontv/server/pkg/response"
)

type APIKeyHandler struct {
	Secret []byte
}

func (h *APIKeyHandler) Generate(c *gin.Context) {
	userID := getUserID(c)

	user, err := repository.GetUserByID(userID)
	if err != nil {
		response.Fail(c, http.StatusNotFound, response.ErrNotFound, "user not found")
		return
	}

	if user.APIKeyCipher != "" {
		response.Fail(c, http.StatusConflict, response.ErrDuplicate, "api key already exists, revoke first")
		return
	}

	plainKey, cipher, err := apikey.Generate(h.Secret, userID)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, response.ErrInternal, "failed to generate api key")
		return
	}

	user.APIKeyCipher = cipher
	if err := repository.UpdateUser(user); err != nil {
		response.Fail(c, http.StatusInternalServerError, response.ErrInternal, "failed to save api key")
		return
	}

	response.OK(c, gin.H{"api_key": plainKey})
}

func (h *APIKeyHandler) Revoke(c *gin.Context) {
	userID := getUserID(c)

	user, err := repository.GetUserByID(userID)
	if err != nil {
		response.Fail(c, http.StatusNotFound, response.ErrNotFound, "user not found")
		return
	}

	user.APIKeyCipher = ""
	if err := repository.UpdateUser(user); err != nil {
		response.Fail(c, http.StatusInternalServerError, response.ErrInternal, "failed to revoke api key")
		return
	}

	response.OK(c, gin.H{"message": "api key revoked"})
}
