// Package handler 群组管理接口
package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	bizerr "github.com/ywty/server/internal/errors"
	"github.com/ywty/server/internal/response"
	"github.com/ywty/server/internal/service"
)

// GroupHandler 群组管理接口
type GroupHandler struct {
	svc *service.GroupService
}

// NewGroupHandler 创建群组管理接口
func NewGroupHandler(svc *service.GroupService) *GroupHandler {
	return &GroupHandler{svc: svc}
}

// Mount 挂载路由
func (h *GroupHandler) Mount(rg *gin.RouterGroup, mw gin.HandlerFunc) {
	g := rg.Group("/admin/groups")
	if mw != nil {
		g.Use(mw)
	}
	g.GET("", h.List)
	g.GET("/:id", h.Get)
	g.POST("", h.Create)
	g.PUT("/:id", h.Update)
	g.DELETE("/:id", h.Delete)
}

// Create 创建群组
// POST /api/v1/admin/groups
func (h *GroupHandler) Create(c *gin.Context) {
	var req service.CreateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, bizerr.BadRequest.WithMessage(err.Error()))
		return
	}

	group, err := h.svc.Create(c.Request.Context(), req)
	if err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, group)
}

// List 列出群组
// GET /api/v1/admin/groups
func (h *GroupHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	groups, total, err := h.svc.List(c.Request.Context(), page, pageSize)
	if err != nil {
		response.Fail(c, err)
		return
	}

	response.Page(c, groups, response.PageMeta{
		CurrentPage: page,
		PerPage:     pageSize,
		Total:       total,
		LastPage:    int((total + int64(pageSize) - 1) / int64(pageSize)),
	})
}

// Get 获取群组详情
// GET /api/v1/admin/groups/:id
func (h *GroupHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, bizerr.BadRequest.WithMessage("无效的群组 ID"))
		return
	}

	group, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, group)
}

// Update 更新群组
// PUT /api/v1/admin/groups/:id
func (h *GroupHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, bizerr.BadRequest.WithMessage("无效的群组 ID"))
		return
	}

	var req service.UpdateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, bizerr.BadRequest.WithMessage(err.Error()))
		return
	}

	group, err := h.svc.Update(c.Request.Context(), id, req)
	if err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, group)
}

// Delete 删除群组
// DELETE /api/v1/admin/groups/:id
func (h *GroupHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, bizerr.BadRequest.WithMessage("无效的群组 ID"))
		return
	}

	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, nil)
}
