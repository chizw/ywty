// Package handler 相册 HTTP 接口
package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/ywty/server/internal/auth"
	bizerr "github.com/ywty/server/internal/errors"
	"github.com/ywty/server/internal/response"
	"github.com/ywty/server/internal/service"
)

// AlbumHandler 相册接口
type AlbumHandler struct {
	svc *service.AlbumService
}

func NewAlbumHandler(svc *service.AlbumService) *AlbumHandler { return &AlbumHandler{svc: svc} }

// List GET /api/v1/albums
func (h *AlbumHandler) List(c *gin.Context) {
	uid, _ := auth.CurrentUserID(c)
	rows, err := h.svc.List(c.Request.Context(), uid)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, rows)
}

// Create POST /api/v1/albums
func (h *AlbumHandler) Create(c *gin.Context) {
	uid, ok := auth.CurrentUserID(c)
	if !ok {
		response.FailCode(c, bizerr.Unauthorized)
		return
	}
	var req service.CreateAlbumReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, bizerr.BadRequest.WithMessage(err.Error()))
		return
	}
	a, err := h.svc.Create(c.Request.Context(), uid, req)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, a)
}

// Get GET /api/v1/albums/:id
func (h *AlbumHandler) Get(c *gin.Context) {
	uid, _ := auth.CurrentUserID(c)
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	a, err := h.svc.Get(c.Request.Context(), id, uid)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, a)
}

// Update PATCH /api/v1/albums/:id
func (h *AlbumHandler) Update(c *gin.Context) {
	uid, _ := auth.CurrentUserID(c)
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req service.UpdateAlbumReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, bizerr.BadRequest.WithMessage(err.Error()))
		return
	}
	a, err := h.svc.Update(c.Request.Context(), id, uid, req)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, a)
}

// Delete DELETE /api/v1/albums/:id
func (h *AlbumHandler) Delete(c *gin.Context) {
	uid, _ := auth.CurrentUserID(c)
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.svc.Delete(c.Request.Context(), id, uid); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, gin.H{"id": id, "deleted": true})
}

// AddPhoto POST /api/v1/albums/:id/photos
func (h *AlbumHandler) AddPhoto(c *gin.Context) {
	uid, _ := auth.CurrentUserID(c)
	albumID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req struct {
		PhotoID uint64 `json:"photo_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, bizerr.BadRequest.WithMessage(err.Error()))
		return
	}
	if err := h.svc.AddPhoto(c.Request.Context(), albumID, req.PhotoID, uid); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, gin.H{"album_id": albumID, "photo_id": req.PhotoID})
}

// RemovePhoto DELETE /api/v1/albums/:id/photos/:photo_id
func (h *AlbumHandler) RemovePhoto(c *gin.Context) {
	uid, _ := auth.CurrentUserID(c)
	albumID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	photoID, _ := strconv.ParseUint(c.Param("photo_id"), 10, 64)
	if err := h.svc.RemovePhoto(c.Request.Context(), albumID, photoID, uid); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, gin.H{"album_id": albumID, "photo_id": photoID, "removed": true})
}

// ListPhotos GET /api/v1/albums/:id/photos
func (h *AlbumHandler) ListPhotos(c *gin.Context) {
	uid, _ := auth.CurrentUserID(c)
	albumID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	rows, err := h.svc.ListPhotos(c.Request.Context(), albumID, uid)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, rows)
}

func (h *AlbumHandler) Mount(rg *gin.RouterGroup, mw gin.HandlerFunc) {
	g := rg.Group("/albums")
	g.GET("", mw, h.List)
	g.POST("", mw, h.Create)
	g.GET("/:id", mw, h.Get)
	g.PATCH("/:id", mw, h.Update)
	g.DELETE("/:id", mw, h.Delete)
	g.GET("/:id/photos", mw, h.ListPhotos)
	g.POST("/:id/photos", mw, h.AddPhoto)
	g.DELETE("/:id/photos/:photo_id", mw, h.RemovePhoto)
}
