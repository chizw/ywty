// Package handler 点赞/举报
package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/ywty/server/internal/auth"
	bizerr "github.com/ywty/server/internal/errors"
	"github.com/ywty/server/internal/response"
	"github.com/ywty/server/internal/service"
)

type LikeHandler struct {
	svc *service.LikeService
}

func NewLikeHandler(svc *service.LikeService) *LikeHandler { return &LikeHandler{svc: svc} }

// Toggle POST /api/v1/likes
func (h *LikeHandler) Toggle(c *gin.Context) {
	uid, _ := auth.CurrentUserID(c)
	var req struct {
		TargetType string `json:"target_type" binding:"required"`
		TargetID   uint64 `json:"target_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, bizerr.BadRequest.WithMessage(err.Error()))
		return
	}
	liked, err := h.svc.Like(c.Request.Context(), uid, req.TargetType, req.TargetID)
	if err != nil {
		response.Fail(c, err)
		return
	}
	count := h.svc.Count(c.Request.Context(), req.TargetType, req.TargetID)
	response.Success(c, gin.H{"liked": liked, "count": count})
}

// Status GET /api/v1/likes?target_type=&target_id=
func (h *LikeHandler) Status(c *gin.Context) {
	tt := c.Query("target_type")
	tid, _ := strconv.ParseUint(c.Query("target_id"), 10, 64)
	uid, _ := auth.CurrentUserID(c)
	count := h.svc.Count(c.Request.Context(), tt, tid)
	liked := false
	if uid > 0 {
		liked = h.svc.Liked(c.Request.Context(), uid, tt, tid)
	}
	response.Success(c, gin.H{"liked": liked, "count": count})
}

func (h *LikeHandler) Mount(rg *gin.RouterGroup, mw gin.HandlerFunc) {
	g := rg.Group("/likes")
	g.POST("", mw, h.Toggle)
	g.GET("", mw, h.Status)
}

// ========================= 举报 =========================

type ReportHandler struct {
	svc *service.ReportService
}

func NewReportHandler(svc *service.ReportService) *ReportHandler { return &ReportHandler{svc: svc} }

// Create POST /api/v1/reports
func (h *ReportHandler) Create(c *gin.Context) {
	uid, _ := auth.CurrentUserID(c)
	if uid == 0 {
		response.FailCode(c, bizerr.Unauthorized)
		return
	}
	var req service.CreateReportReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, bizerr.BadRequest.WithMessage(err.Error()))
		return
	}
	r, err := h.svc.Create(c.Request.Context(), uid, c.ClientIP(), req)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, r)
}

// AdminList GET /api/v1/admin/reports
func (h *ReportHandler) AdminList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	status := c.Query("status")
	if page <= 0 {
		page = 1
	}
	if perPage <= 0 || perPage > 100 {
		perPage = 20
	}
	rows, total, err := h.svc.AdminList(c.Request.Context(), page, perPage, status)
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

// UpdateStatus PATCH /api/v1/admin/reports/:id
func (h *ReportHandler) UpdateStatus(c *gin.Context) {
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

func (h *ReportHandler) Mount(rg *gin.RouterGroup, mw gin.HandlerFunc) {
	g := rg.Group("/reports")
	g.POST("", mw, h.Create)

	adminOnly := auth.AdminOnly()
	admin := rg.Group("/admin/reports", mw, adminOnly)
	admin.GET("", h.AdminList)
	admin.PATCH("/:id", h.UpdateStatus)
}
