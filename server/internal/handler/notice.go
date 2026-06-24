// Package handler 公告 HTTP 接口
package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/ywty/server/internal/auth"
	bizerr "github.com/ywty/server/internal/errors"
	"github.com/ywty/server/internal/response"
	"github.com/ywty/server/internal/service"
)

// NoticeHandler 公告接口
type NoticeHandler struct {
	svc *service.NoticeService
}

func NewNoticeHandler(svc *service.NoticeService) *NoticeHandler { return &NoticeHandler{svc: svc} }

// ListPublic GET /api/v1/notices
func (h *NoticeHandler) ListPublic(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	if page <= 0 {
		page = 1
	}
	if perPage <= 0 || perPage > 100 {
		perPage = 20
	}
	rows, total, err := h.svc.ListPublic(c.Request.Context(), page, perPage)
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

// GetPublic GET /api/v1/notices/:id
func (h *NoticeHandler) GetPublic(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.FailCode(c, bizerr.BadRequest.WithMessage("invalid id"))
		return
	}
	n, err := h.svc.GetPublic(c.Request.Context(), id)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, n)
}

// AdminList GET /api/v1/admin/notices
func (h *NoticeHandler) AdminList(c *gin.Context) {
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

// AdminCreate POST /api/v1/admin/notices
func (h *NoticeHandler) AdminCreate(c *gin.Context) {
	var req service.CreateNoticeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, bizerr.BadRequest.WithMessage(err.Error()))
		return
	}
	n, err := h.svc.AdminCreate(c.Request.Context(), req)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, n)
}

// AdminUpdate PATCH /api/v1/admin/notices/:id
func (h *NoticeHandler) AdminUpdate(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.FailCode(c, bizerr.BadRequest.WithMessage("invalid id"))
		return
	}
	var req service.UpdateNoticeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, bizerr.BadRequest.WithMessage(err.Error()))
		return
	}
	n, err := h.svc.AdminUpdate(c.Request.Context(), id, req)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, n)
}

// AdminDelete DELETE /api/v1/admin/notices/:id
func (h *NoticeHandler) AdminDelete(c *gin.Context) {
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

func (h *NoticeHandler) Mount(rg *gin.RouterGroup, mw gin.HandlerFunc) {
	adminOnly := auth.AdminOnly()

	// 公开
	pg := rg.Group("/notices")
	pg.GET("", h.ListPublic)
	pg.GET("/:id", h.GetPublic)

	// 管理
	ag := rg.Group("/admin/notices", mw, adminOnly)
	ag.GET("", h.AdminList)
	ag.POST("", h.AdminCreate)
	ag.PATCH("/:id", h.AdminUpdate)
	ag.DELETE("/:id", h.AdminDelete)
}
