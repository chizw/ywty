// Package handler 分享 HTTP 接口
package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/ywty/server/internal/auth"
	bizerr "github.com/ywty/server/internal/errors"
	"github.com/ywty/server/internal/response"
	"github.com/ywty/server/internal/service"
)

type ShareHandler struct {
	svc *service.ShareService
}

func NewShareHandler(svc *service.ShareService) *ShareHandler { return &ShareHandler{svc: svc} }

// List GET /api/v1/shares
func (h *ShareHandler) List(c *gin.Context) {
	uid, _ := auth.CurrentUserID(c)
	rows, err := h.svc.List(c.Request.Context(), uid)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, rows)
}

// Create POST /api/v1/shares
func (h *ShareHandler) Create(c *gin.Context) {
	uid, _ := auth.CurrentUserID(c)
	var req service.CreateShareReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, bizerr.BadRequest.WithMessage(err.Error()))
		return
	}
	sh, err := h.svc.Create(c.Request.Context(), uid, req)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, sh)
}

// Delete DELETE /api/v1/shares/:id
func (h *ShareHandler) Delete(c *gin.Context) {
	uid, _ := auth.CurrentUserID(c)
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.svc.Delete(c.Request.Context(), uid, id); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, gin.H{"id": id, "deleted": true})
}

// Update PATCH /api/v1/shares/:id
func (h *ShareHandler) Update(c *gin.Context) {
	uid, _ := auth.CurrentUserID(c)
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req service.UpdateShareReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, bizerr.BadRequest.WithMessage(err.Error()))
		return
	}
	sh, err := h.svc.Update(c.Request.Context(), uid, id, req)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, sh)
}

// View GET /s/:slug 公开访问
func (h *ShareHandler) View(c *gin.Context) {
	slug := c.Param("slug")
	sh, items, err := h.svc.GetBySlug(c.Request.Context(), slug)
	if err != nil {
		response.Fail(c, err)
		return
	}
	// 校验密码
	if sh.Content != "" {
		var pwd struct {
			Password string `json:"password"`
		}
		_ = c.ShouldBindQuery(&pwd)
		if sh.Content != pwd.Password {
			response.Success(c, gin.H{
				"share":    sh,
				"need_pwd": true,
			})
			return
		}
	}
	response.Success(c, gin.H{
		"share": sh,
		"items": items,
	})
}

// ShareView 提供给 router 使用的顶层 handler
func ShareView(svc *service.ShareService) gin.HandlerFunc {
	h := &ShareHandler{svc: svc}
	return h.View
}

func (h *ShareHandler) Mount(rg *gin.RouterGroup, mw gin.HandlerFunc) {
	g := rg.Group("/shares")
	g.GET("", mw, h.List)
	g.POST("", mw, h.Create)
	g.PATCH("/:id", mw, h.Update)
	g.DELETE("/:id", mw, h.Delete)
}
