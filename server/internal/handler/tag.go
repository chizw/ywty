// Package handler 标签 HTTP 接口
package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/ywty/server/internal/auth"
	bizerr "github.com/ywty/server/internal/errors"
	"github.com/ywty/server/internal/response"
	"github.com/ywty/server/internal/service"
)

type TagHandler struct {
	svc *service.TagService
}

func NewTagHandler(svc *service.TagService) *TagHandler { return &TagHandler{svc: svc} }

// List GET /api/v1/tags
func (h *TagHandler) List(c *gin.Context) {
	rows, err := h.svc.List(c.Request.Context())
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, rows)
}

// Create POST /api/v1/tags
func (h *TagHandler) Create(c *gin.Context) {
	var req struct {
		Name string `json:"name" binding:"required,max=255"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, bizerr.BadRequest.WithMessage(err.Error()))
		return
	}
	t, err := h.svc.Create(c.Request.Context(), req.Name)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, t)
}

// Delete DELETE /api/v1/tags/:id
func (h *TagHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, gin.H{"id": id, "deleted": true})
}

// Attach POST /api/v1/tags/attach
func (h *TagHandler) Attach(c *gin.Context) {
	uid, _ := auth.CurrentUserID(c)
	var req struct {
		Names      []string `json:"names"`
		TagID      uint64   `json:"tag_id"`
		TargetType string   `json:"target_type" binding:"required"`
		TargetID   uint64   `json:"target_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, bizerr.BadRequest.WithMessage(err.Error()))
		return
	}
	if len(req.Names) == 0 && req.TagID == 0 {
		response.FailCode(c, bizerr.BadRequest.WithMessage("names or tag_id required"))
		return
	}
	if len(req.Names) > 0 {
		tags, err := h.svc.AttachByNames(c.Request.Context(), uid, req.Names, req.TargetType, req.TargetID)
		if err != nil {
			response.Fail(c, err)
			return
		}
		response.Success(c, gin.H{"tags": tags})
		return
	}
	if err := h.svc.Attach(c.Request.Context(), uid, req.TagID, req.TargetType, req.TargetID); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, gin.H{"tag_id": req.TagID, "attached": true})
}

// Detach POST /api/v1/tags/detach
func (h *TagHandler) Detach(c *gin.Context) {
	uid, _ := auth.CurrentUserID(c)
	var req struct {
		TagID      uint64 `json:"tag_id" binding:"required"`
		TargetType string `json:"target_type" binding:"required"`
		TargetID   uint64 `json:"target_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, bizerr.BadRequest.WithMessage(err.Error()))
		return
	}
	if err := h.svc.Detach(c.Request.Context(), uid, req.TagID, req.TargetType, req.TargetID); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, gin.H{"tag_id": req.TagID, "detached": true})
}

// ListForTarget GET /api/v1/tags?target_type=&target_id=
func (h *TagHandler) ListForTarget(c *gin.Context) {
	tt := c.Query("target_type")
	tid, _ := strconv.ParseUint(c.Query("target_id"), 10, 64)
	if tt == "" || tid == 0 {
		// 无 target 时退化为全局列表
		h.List(c)
		return
	}
	tags, err := h.svc.ListForTarget(c.Request.Context(), tt, tid)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, tags)
}

func (h *TagHandler) Mount(rg *gin.RouterGroup, mw gin.HandlerFunc) {
	g := rg.Group("/tags")
	g.GET("", mw, h.ListForTarget)
	g.POST("", mw, h.Create)
	g.DELETE("/:id", mw, h.Delete)
	g.POST("/attach", mw, h.Attach)
	g.POST("/detach", mw, h.Detach)
}
