package handler

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/moontv/server/internal/middleware"
	"github.com/moontv/server/internal/repository"
	"github.com/moontv/server/pkg/response"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	JWTSecret string
}

type loginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type registerRequest struct {
	Username   string `json:"username" binding:"required,min=3,max=32"`
	Password   string `json:"password" binding:"required,min=6"`
	InviteCode string `json:"invite_code" binding:"required"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.ErrBadRequest, "invalid request")
		return
	}

	user, err := repository.GetUserByUsername(req.Username)
	if err != nil {
		response.Fail(c, http.StatusUnauthorized, response.ErrUnauthorized, "invalid credentials")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		response.Fail(c, http.StatusUnauthorized, response.ErrUnauthorized, "invalid credentials")
		return
	}

	if user.Banned {
		response.Fail(c, http.StatusForbidden, response.ErrBanned, "account banned")
		return
	}

	claims := middleware.JWTClaims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(h.JWTSecret))
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, response.ErrInternal, "failed to generate token")
		return
	}

	response.OK(c, gin.H{
		"token":    tokenStr,
		"user_id":  user.ID,
		"username": user.Username,
		"role":     user.Role,
	})
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.ErrBadRequest, "invalid request")
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, response.ErrInternal, "failed to hash password")
		return
	}

	user, err := repository.RegisterUser(req.Username, string(hash), req.InviteCode)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrInviteUsed):
			response.Fail(c, http.StatusBadRequest, response.ErrInviteInvalid, err.Error())
		case errors.Is(err, repository.ErrInviteExpired):
			response.Fail(c, http.StatusBadRequest, response.ErrInviteInvalid, err.Error())
		case errors.Is(err, repository.ErrUsernameExists):
			response.Fail(c, http.StatusConflict, response.ErrDuplicate, err.Error())
		default:
			response.Fail(c, http.StatusBadRequest, response.ErrInviteInvalid, "invalid invite code")
		}
		return
	}

	response.OK(c, gin.H{
		"user_id":  user.ID,
		"username": user.Username,
	})
}
