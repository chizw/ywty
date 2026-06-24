// Package handler 单页 HTTP 接口
package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/ywty/server/internal/auth"
	bizerr "github.com/ywty/server/internal/errors"
	"github.com/ywty/server/internal/response"
	"github.com/ywty/server/internal/service"
)

// PageHandler 单页接口
type PageHandler struct {
	svc *service.PageService
}

func NewPageHandler(svc *service.PageService) *PageHandler { return &PageHandler{svc: svc} }

// ListPublic GET /api/v1/pages
func (h *PageHandler) ListPublic(c *gin.Context) {
	rows, err := h.svc.ListPublic(c.Request.Context())
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, rows)
}

// GetPublic GET /api/v1/pages/:slug
func (h *PageHandler) GetPublic(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		response.FailCode(c, bizerr.BadRequest.WithMessage("slug is required"))
		return
	}
	p, err := h.svc.GetBySlug(c.Request.Context(), slug)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, p)
}

// AdminList GET /api/v1/admin/pages
func (h *PageHandler) AdminList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	if page <= 0 {
		page = 1
	}
	if perPage <= 0 || perPage > 100 {
		perPage = 20
	}
	rows, total, err := h.svc.AdminList(c.Request.Context(), page, perPage)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Page(c, rows, response.PageMeta{
		CurrentPage: page,
		PerPage:     perPage,
		Total:       total,
		LastPage:    calcLastPage(total, perPage),
	})
}

// AdminCreate POST /api/v1/admin/pages
func (h *PageHandler) AdminCreate(c *gin.Context) {
	var req service.CreatePageReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, bizerr.BadRequest.WithMessage(err.Error()))
		return
	}
	p, err := h.svc.AdminCreate(c.Request.Context(), req)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, p)
}

// AdminGet GET /api/v1/admin/pages/:id
func (h *PageHandler) AdminGet(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.FailCode(c, bizerr.BadRequest.WithMessage("invalid id"))
		return
	}
	p, err := h.svc.AdminGet(c.Request.Context(), id)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, p)
}

// AdminUpdate PATCH /api/v1/admin/pages/:id
func (h *PageHandler) AdminUpdate(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.FailCode(c, bizerr.BadRequest.WithMessage("invalid id"))
		return
	}
	var req service.UpdatePageReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, bizerr.BadRequest.WithMessage(err.Error()))
		return
	}
	p, err := h.svc.AdminUpdate(c.Request.Context(), id, req)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, p)
}

// AdminDelete DELETE /api/v1/admin/pages/:id
func (h *PageHandler) AdminDelete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.FailCode(c, bizerr.BadRequest.WithMessage("invalid id"))
		return
	}
	if err := h.svc.AdminDelete(c.Request.Context(), id); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, gin.H{"id": id, "deleted": true})
}

func (h *PageHandler) Mount(rg *gin.RouterGroup, mw gin.HandlerFunc) {
	adminOnly := auth.AdminOnly()

	// 公开
	pg := rg.Group("/pages")
	pg.GET("", h.ListPublic)
	pg.GET("/:slug", h.GetPublic)

	// 管理
	ag := rg.Group("/admin/pages", mw, adminOnly)
	ag.GET("", h.AdminList)
	ag.POST("", h.AdminCreate)
	ag.GET("/:id", h.AdminGet)
	ag.PATCH("/:id", h.AdminUpdate)
	ag.DELETE("/:id", h.AdminDelete)
}
