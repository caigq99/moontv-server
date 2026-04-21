package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/moontv/server/internal/model"
	"github.com/moontv/server/internal/repository"
	"github.com/moontv/server/pkg/response"
)

type SourceHandler struct{}

type createSourceRequest struct {
	Key       string `json:"key" binding:"required"`
	Name      string `json:"name" binding:"required"`
	APIUrl    string `json:"api_url" binding:"required,url"`
	DetailUrl string `json:"detail_url,omitempty"`
}

type updateSourceRequest struct {
	Name      *string `json:"name,omitempty"`
	APIUrl    *string `json:"api_url,omitempty"`
	DetailUrl *string `json:"detail_url,omitempty"`
	Disabled  *bool   `json:"disabled,omitempty"`
}

type sortRequest struct {
	Keys []string `json:"keys" binding:"required"`
}

func (h *SourceHandler) List(c *gin.Context) {
	userID := getUserID(c)
	sources, err := repository.GetSourcesByUserID(userID)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, response.ErrInternal, "failed to get sources")
		return
	}
	response.OK(c, sources)
}

func (h *SourceHandler) Create(c *gin.Context) {
	userID := getUserID(c)
	var req createSourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.ErrBadRequest, "invalid request")
		return
	}

	if _, err := repository.GetSourceByKey(userID, req.Key); err == nil {
		response.Fail(c, http.StatusConflict, response.ErrDuplicate, "source key already exists")
		return
	}

	source := &model.Source{
		UserID:    &userID,
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

func (h *SourceHandler) Update(c *gin.Context) {
	userID := getUserID(c)
	key := c.Param("key")

	source, err := repository.GetSourceByKey(userID, key)
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
		source.APIUrl = *req.APIUrl
	}
	if req.DetailUrl != nil {
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

func (h *SourceHandler) Delete(c *gin.Context) {
	userID := getUserID(c)
	key := c.Param("key")

	if err := repository.DeleteSource(userID, key); err != nil {
		response.Fail(c, http.StatusInternalServerError, response.ErrInternal, "failed to delete source")
		return
	}
	response.OK(c, nil)
}

func (h *SourceHandler) Sort(c *gin.Context) {
	userID := getUserID(c)
	var req sortRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.ErrBadRequest, "invalid request")
		return
	}

	if err := repository.UpdateSourceSortOrder(userID, req.Keys); err != nil {
		response.Fail(c, http.StatusInternalServerError, response.ErrInternal, "failed to update sort order")
		return
	}
	response.OK(c, nil)
}
