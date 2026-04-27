package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/moontv/server/internal/model"
	"github.com/moontv/server/internal/repository"
	"github.com/moontv/server/pkg/response"
)

type AdminHandler struct{}

func (h *AdminHandler) Stats(c *gin.Context) {
	now := time.Now()
	response.OK(c, gin.H{
		"total_users":       repository.CountUsers(),
		"active_users":      repository.CountActiveUsers(),
		"total_sources":     repository.CountSources(),
		"api_calls_today":   repository.CountAPICallsSince(time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())),
		"api_calls_7days":   repository.CountAPICallsSince(now.AddDate(0, 0, -7)),
	})
}

func (h *AdminHandler) ListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	users, total, err := repository.ListUsers(page, pageSize)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, response.ErrInternal, "failed to list users")
		return
	}

	response.OK(c, gin.H{
		"users": users,
		"total": total,
		"page":  page,
	})
}

func (h *AdminHandler) BanUser(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	user, err := repository.GetUserByID(uint(id))
	if err != nil {
		response.Fail(c, http.StatusNotFound, response.ErrNotFound, "user not found")
		return
	}

	var req struct {
		Banned bool `json:"banned"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.ErrBadRequest, "invalid request")
		return
	}

	user.Banned = req.Banned
	if err := repository.UpdateUser(user); err != nil {
		response.Fail(c, http.StatusInternalServerError, response.ErrInternal, "failed to update user")
		return
	}
	response.OK(c, user)
}

func (h *AdminHandler) DeleteUser(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)

	adminID := getUserID(c)
	if uint(id) == adminID {
		response.Fail(c, http.StatusBadRequest, response.ErrBadRequest, "cannot delete yourself")
		return
	}

	if err := repository.DeleteUser(uint(id)); err != nil {
		response.Fail(c, http.StatusInternalServerError, response.ErrInternal, "failed to delete user")
		return
	}
	response.OK(c, nil)
}

func (h *AdminHandler) GenerateInvites(c *gin.Context) {
	var req struct {
		Count      int `json:"count" binding:"required,min=1,max=50"`
		ExpireDays int `json:"expire_days"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.ErrBadRequest, "invalid request")
		return
	}

	adminID := getUserID(c)
	codes, err := repository.GenerateInviteCodes(adminID, req.Count, req.ExpireDays)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, response.ErrInternal, "failed to generate invite codes")
		return
	}
	response.OK(c, codes)
}

func (h *AdminHandler) ListInvites(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	codes, total, err := repository.ListInvites(page, pageSize)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, response.ErrInternal, "failed to list invites")
		return
	}

	response.OK(c, gin.H{
		"invites": codes,
		"total":   total,
		"page":    page,
	})
}

func (h *AdminHandler) DeleteInvite(c *gin.Context) {
	code := c.Param("code")
	if err := repository.DeleteInvite(code); err != nil {
		response.Fail(c, http.StatusInternalServerError, response.ErrInternal, "failed to delete invite")
		return
	}
	response.OK(c, nil)
}

// Global source management

func (h *AdminHandler) ListGlobalSources(c *gin.Context) {
	sources, err := repository.GetGlobalSources()
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, response.ErrInternal, "failed to get sources")
		return
	}
	response.OK(c, sources)
}

func (h *AdminHandler) CreateGlobalSource(c *gin.Context) {
	var req createSourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.ErrBadRequest, "invalid request")
		return
	}

	if err := validateSourceURL(req.APIUrl); err != nil {
		response.Fail(c, http.StatusBadRequest, response.ErrBadRequest, "invalid api_url: "+err.Error())
		return
	}
	if req.DetailUrl != "" {
		if err := validateSourceURL(req.DetailUrl); err != nil {
			response.Fail(c, http.StatusBadRequest, response.ErrBadRequest, "invalid detail_url: "+err.Error())
			return
		}
	}

	if _, err := repository.GetGlobalSourceByKey(req.Key); err == nil {
		response.Fail(c, http.StatusConflict, response.ErrDuplicate, "source key already exists")
		return
	}

	source := &model.Source{
		Key:       req.Key,
		Name:      req.Name,
		APIUrl:    req.APIUrl,
		DetailUrl: req.DetailUrl,
	}
	if err := repository.CreateSource(source); err != nil {
		response.Fail(c, http.StatusInternalServerError, response.ErrInternal, "failed to create source")
		return
	}
	response.OK(c, source)
}

func (h *AdminHandler) UpdateGlobalSource(c *gin.Context) {
	key := c.Param("key")

	source, err := repository.GetGlobalSourceByKey(key)
	if err != nil {
		response.Fail(c, http.StatusNotFound, response.ErrNotFound, "source not found")
		return
	}

	var req updateSourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.ErrBadRequest, "invalid request")
		return
	}

	if req.Name != nil {
		source.Name = *req.Name
	}
	if req.APIUrl != nil {
		if err := validateSourceURL(*req.APIUrl); err != nil {
			response.Fail(c, http.StatusBadRequest, response.ErrBadRequest, "invalid api_url: "+err.Error())
			return
		}
		source.APIUrl = *req.APIUrl
	}
	if req.DetailUrl != nil {
		if *req.DetailUrl != "" {
			if err := validateSourceURL(*req.DetailUrl); err != nil {
				response.Fail(c, http.StatusBadRequest, response.ErrBadRequest, "invalid detail_url: "+err.Error())
				return
			}
		}
		source.DetailUrl = *req.DetailUrl
	}
	if req.Disabled != nil {
		source.Disabled = *req.Disabled
	}

	if err := repository.UpdateSource(source); err != nil {
		response.Fail(c, http.StatusInternalServerError, response.ErrInternal, "failed to update source")
		return
	}
	response.OK(c, source)
}

func (h *AdminHandler) DeleteGlobalSource(c *gin.Context) {
	key := c.Param("key")
	if err := repository.DeleteGlobalSource(key); err != nil {
		response.Fail(c, http.StatusInternalServerError, response.ErrInternal, "failed to delete source")
		return
	}
	response.OK(c, nil)
}

func (h *AdminHandler) SortGlobalSources(c *gin.Context) {
	var req sortRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.ErrBadRequest, "invalid request")
		return
	}

	if err := repository.UpdateGlobalSourceSortOrder(req.Keys); err != nil {
		response.Fail(c, http.StatusInternalServerError, response.ErrInternal, "failed to update sort order")
		return
	}
	response.OK(c, nil)
}
