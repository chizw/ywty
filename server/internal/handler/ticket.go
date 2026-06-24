// Package handler 工单 HTTP 接口
package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/ywty/server/internal/auth"
	bizerr "github.com/ywty/server/internal/errors"
	"github.com/ywty/server/internal/response"
	"github.com/ywty/server/internal/service"
)

// TicketHandler 工单接口
type TicketHandler struct {
	svc *service.TicketService
}

func NewTicketHandler(svc *service.TicketService) *TicketHandler { return &TicketHandler{svc: svc} }

// List GET /api/v1/tickets
func (h *TicketHandler) List(c *gin.Context) {
	uid, ok := auth.CurrentUserID(c)
	if !ok {
		response.FailCode(c, bizerr.Unauthorized)
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	rows, total, err := h.svc.List(c.Request.Context(), uid, page, perPage)
	if err != nil {
		response.Fail(c, err)
		return
	}
	if page <= 0 {
		page = 1
	}
	if perPage <= 0 || perPage > 100 {
		perPage = 20
	}
	response.Page(c, rows, response.PageMeta{
		CurrentPage: page,
		PerPage:     perPage,
		Total:       total,
		LastPage:    calcLastPage(total, perPage),
	})
}

// Create POST /api/v1/tickets
func (h *TicketHandler) Create(c *gin.Context) {
	uid, ok := auth.CurrentUserID(c)
	if !ok {
		response.FailCode(c, bizerr.Unauthorized)
		return
	}
	var req service.CreateTicketReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, bizerr.BadRequest.WithMessage(err.Error()))
		return
	}
	ticket, err := h.svc.Create(c.Request.Context(), uid, req)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, ticket)
}

// Get GET /api/v1/tickets/:id
func (h *TicketHandler) Get(c *gin.Context) {
	uid, ok := auth.CurrentUserID(c)
	if !ok {
		response.FailCode(c, bizerr.Unauthorized)
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.FailCode(c, bizerr.BadRequest.WithMessage("invalid id"))
		return
	}
	ticket, replies, err := h.svc.Get(c.Request.Context(), uid, id)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, gin.H{"ticket": ticket, "replies": replies})
}

// Reply POST /api/v1/tickets/:id/replies
func (h *TicketHandler) Reply(c *gin.Context) {
	uid, ok := auth.CurrentUserID(c)
	if !ok {
		response.FailCode(c, bizerr.Unauthorized)
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.FailCode(c, bizerr.BadRequest.WithMessage("invalid id"))
		return
	}
	var req struct {
		Content string `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, bizerr.BadRequest.WithMessage(err.Error()))
		return
	}
	reply, err := h.svc.Reply(c.Request.Context(), uid, id, req.Content)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, reply)
}

// Close POST /api/v1/tickets/:id/close
func (h *TicketHandler) Close(c *gin.Context) {
	uid, ok := auth.CurrentUserID(c)
	if !ok {
		response.FailCode(c, bizerr.Unauthorized)
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.FailCode(c, bizerr.BadRequest.WithMessage("invalid id"))
		return
	}
	if err := h.svc.Close(c.Request.Context(), uid, id); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, gin.H{"id": id, "closed": true})
}

// AdminList GET /api/v1/admin/tickets
func (h *TicketHandler) AdminList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	status := c.Query("status")
	level := c.Query("level")
	if page <= 0 {
		page = 1
	}
	if perPage <= 0 || perPage > 100 {
		perPage = 20
	}
	rows, total, err := h.svc.AdminList(c.Request.Context(), page, perPage, status, level)
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

// AdminGet GET /api/v1/admin/tickets/:id
func (h *TicketHandler) AdminGet(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.FailCode(c, bizerr.BadRequest.WithMessage("invalid id"))
		return
	}
	ticket, replies, err := h.svc.AdminGet(c.Request.Context(), id)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, gin.H{"ticket": ticket, "replies": replies})
}

// AdminReply POST /api/v1/admin/tickets/:id/replies
func (h *TicketHandler) AdminReply(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.FailCode(c, bizerr.BadRequest.WithMessage("invalid id"))
		return
	}
	var req struct {
		Content string `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, bizerr.BadRequest.WithMessage(err.Error()))
		return
	}
	reply, err := h.svc.AdminReply(c.Request.Context(), id, req.Content)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, reply)
}

// AdminUpdateStatus PATCH /api/v1/admin/tickets/:id/status
func (h *TicketHandler) AdminUpdateStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.FailCode(c, bizerr.BadRequest.WithMessage("invalid id"))
		return
	}
	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, bizerr.BadRequest.WithMessage(err.Error()))
		return
	}
	if err := h.svc.AdminUpdateStatus(c.Request.Context(), id, req.Status); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, gin.H{"id": id, "status": req.Status})
}

func (h *TicketHandler) Mount(rg *gin.RouterGroup, mw gin.HandlerFunc) {
	adminOnly := auth.AdminOnly()

	// 用户侧
	g := rg.Group("/tickets", mw)
	g.GET("", h.List)
	g.POST("", h.Create)
	g.GET("/:id", h.Get)
	g.POST("/:id/replies", h.Reply)
	g.POST("/:id/close", h.Close)

	// 管理侧
	ag := rg.Group("/admin/tickets", mw, adminOnly)
	ag.GET("", h.AdminList)
	ag.GET("/:id", h.AdminGet)
	ag.POST("/:id/replies", h.AdminReply)
	ag.PATCH("/:id/status", h.AdminUpdateStatus)
}
