// Package handler 套餐 HTTP 接口
package handler

import (
	"context"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/ywty/server/internal/auth"
	bizerr "github.com/ywty/server/internal/errors"
	"github.com/ywty/server/internal/response"
	"github.com/ywty/server/internal/service"
)

// PlanHandler 套餐接口
type PlanHandler struct {
	svc *service.PlanService
}

// NewPlanHandler 创建套餐处理器
func NewPlanHandler(svc *service.PlanService) *PlanHandler {
	return &PlanHandler{svc: svc}
}

func (h *PlanHandler) planContext(c *gin.Context) (context.Context, error) {
	ctx := c.Request.Context()
	isAdmin := auth.CurrentIsAdmin(c)
	ctx = service.WithPlanAdmin(ctx, isAdmin)
	if !isAdmin {
		userID, _ := auth.CurrentUserID(c)
		groupIDs, err := h.svc.ResolveViewerGroupIDs(ctx, userID)
		if err != nil {
			return nil, err
		}
		ctx = service.WithPlanGroupIDs(ctx, groupIDs)
	}
	return ctx, nil
}

// ListPublic GET /api/v1/plans
func (h *PlanHandler) ListPublic(c *gin.Context) {
	ctx, err := h.planContext(c)
	if err != nil {
		response.Fail(c, err)
		return
	}
	rows, err := h.svc.List(ctx, false)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, rows)
}

// GetPublic GET /api/v1/plans/:id
func (h *PlanHandler) GetPublic(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.FailCode(c, bizerr.BadRequest.WithMessage("invalid id"))
		return
	}
	ctx, err := h.planContext(c)
	if err != nil {
		response.Fail(c, err)
		return
	}
	detail, err := h.svc.Get(ctx, id)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, gin.H{
		"plan":       detail.Plan,
		"prices":     detail.Prices,
		"capacities": detail.Capacities,
		"group_ids":  detail.GroupIDs,
	})
}

// AdminList GET /api/v1/admin/plans
func (h *PlanHandler) AdminList(c *gin.Context) {
	ctx := service.WithPlanAdmin(c.Request.Context(), true)
	rows, err := h.svc.List(ctx, true)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, rows)
}

// AdminCreate POST /api/v1/admin/plans
func (h *PlanHandler) AdminCreate(c *gin.Context) {
	var req service.AdminPlanReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, bizerr.BadRequest.WithMessage(err.Error()))
		return
	}
	plan, err := h.svc.Create(c.Request.Context(), req)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, plan)
}

// AdminGetDetail GET /api/v1/admin/plans/:id
func (h *PlanHandler) AdminGetDetail(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.FailCode(c, bizerr.BadRequest.WithMessage("invalid id"))
		return
	}
	ctx := service.WithPlanAdmin(c.Request.Context(), true)
	detail, err := h.svc.Get(ctx, id)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, gin.H{
		"plan":       detail.Plan,
		"prices":     detail.Prices,
		"capacities": detail.Capacities,
		"group_ids":  detail.GroupIDs,
	})
}

// AdminUpdate PATCH /api/v1/admin/plans/:id
func (h *PlanHandler) AdminUpdate(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.FailCode(c, bizerr.BadRequest.WithMessage("invalid id"))
		return
	}
	var req service.AdminPlanReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, bizerr.BadRequest.WithMessage(err.Error()))
		return
	}
	plan, err := h.svc.Update(c.Request.Context(), id, req)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, plan)
}

// AdminDelete DELETE /api/v1/admin/plans/:id
func (h *PlanHandler) AdminDelete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.FailCode(c, bizerr.BadRequest.WithMessage("invalid id"))
		return
	}
	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, gin.H{"id": id, "deleted": true})
}

// AdminToggleUp POST /api/v1/admin/plans/:id/toggle
func (h *PlanHandler) AdminToggleUp(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.FailCode(c, bizerr.BadRequest.WithMessage("invalid id"))
		return
	}
	if err := h.svc.ToggleUp(c.Request.Context(), id); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, gin.H{"id": id, "toggled": true})
}

// Mount 注册路由
func (h *PlanHandler) Mount(rg *gin.RouterGroup, mw gin.HandlerFunc) {
	// 公开接口：无需鉴权，支持游客访问
	public := rg.Group("/plans")
	public.GET("", h.ListPublic)
	public.GET("/:id", h.GetPublic)

	// 管理接口：需鉴权 + 管理员
	admin := rg.Group("/admin/plans", mw, auth.AdminOnly())
	admin.GET("", h.AdminList)
	admin.POST("", h.AdminCreate)
	admin.GET("/:id", h.AdminGetDetail)
	admin.PATCH("/:id", h.AdminUpdate)
	admin.DELETE("/:id", h.AdminDelete)
	admin.POST("/:id/toggle", h.AdminToggleUp)
}
