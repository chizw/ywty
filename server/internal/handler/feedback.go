// Package handler 意见反馈
package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/ywty/server/internal/auth"
	bizerr "github.com/ywty/server/internal/errors"
	"github.com/ywty/server/internal/response"
	"github.com/ywty/server/internal/service"
)

// FeedbackHandler 反馈接口
type FeedbackHandler struct {
	svc *service.FeedbackService
}

// NewFeedbackHandler 创建反馈 Handler
func NewFeedbackHandler(svc *service.FeedbackService) *FeedbackHandler {
	return &FeedbackHandler{svc: svc}
}

// Create POST /api/v1/feedback（公开）
func (h *FeedbackHandler) Create(c *gin.Context) {
	var req service.CreateFeedbackReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, bizerr.BadRequest.WithMessage(err.Error()))
		return
	}
	f, err := h.svc.Create(c.Request.Context(), c.ClientIP(), req)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, f)
}

// AdminList GET /api/v1/admin/feedbacks
func (h *FeedbackHandler) AdminList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	if page <= 0 {
		page = 1
	}
	if perPage <= 0 || perPage > 100 {
		perPage = 20
	}
	rows, total, err := h.svc.List(c.Request.Context(), page, perPage)
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

// Delete DELETE /api/v1/admin/feedbacks/:id
func (h *FeedbackHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, gin.H{"id": id, "deleted": true})
}

// Mount 挂载路由
func (h *FeedbackHandler) Mount(rg *gin.RouterGroup, mw gin.HandlerFunc) {
	// 公开提交
	rg.POST("/feedback", h.Create)

	// 后台管理
	adminOnly := auth.AdminOnly()
	g := rg.Group("/admin/feedbacks", mw, adminOnly)
	g.GET("", h.AdminList)
	g.DELETE("/:id", h.Delete)
}
