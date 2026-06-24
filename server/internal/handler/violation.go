// Package handler 违规记录
package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/ywty/server/internal/auth"
	bizerr "github.com/ywty/server/internal/errors"
	"github.com/ywty/server/internal/response"
	"github.com/ywty/server/internal/service"
)

// ViolationHandler 违规记录接口
type ViolationHandler struct {
	svc *service.ViolationService
}

// NewViolationHandler 创建违规记录 Handler
func NewViolationHandler(svc *service.ViolationService) *ViolationHandler {
	return &ViolationHandler{svc: svc}
}

// Create POST /api/v1/violations
func (h *ViolationHandler) Create(c *gin.Context) {
	uid, _ := auth.CurrentUserID(c)
	if uid == 0 {
		response.FailCode(c, bizerr.Unauthorized)
		return
	}
	var req struct {
		PhotoID uint64 `json:"photo_id" binding:"required"`
		Reason  string `json:"reason" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, bizerr.BadRequest.WithMessage(err.Error()))
		return
	}
	v, err := h.svc.Create(c.Request.Context(), uid, req.PhotoID, req.Reason)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, v)
}

// AdminList GET /api/v1/admin/violations
func (h *ViolationHandler) AdminList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	status := c.Query("status")
	if page <= 0 {
		page = 1
	}
	if perPage <= 0 || perPage > 100 {
		perPage = 20
	}
	rows, total, err := h.svc.List(c.Request.Context(), page, perPage, status)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Page(c, rows, response.PageMeta{
		CurrentPage: page,
		PerPage:     perPage,
		Total:       total,
		LastPage:    int((total + int64(perPage) - 1) / int64(perPage)),
	})
}

// UpdateStatus PATCH /api/v1/admin/violations/:id
func (h *ViolationHandler) UpdateStatus(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, bizerr.BadRequest.WithMessage(err.Error()))
		return
	}
	if err := h.svc.UpdateStatus(c.Request.Context(), id, req.Status); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, gin.H{"id": id, "status": req.Status})
}

// Mount 挂载路由
func (h *ViolationHandler) Mount(rg *gin.RouterGroup, mw gin.HandlerFunc) {
	g := rg.Group("/violations", mw)
	g.POST("", h.Create)

	adminOnly := auth.AdminOnly()
	admin := rg.Group("/admin/violations", mw, adminOnly)
	admin.GET("", h.AdminList)
	admin.PATCH("/:id", h.UpdateStatus)
}
