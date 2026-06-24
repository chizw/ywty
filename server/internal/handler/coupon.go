// Package handler 优惠券 HTTP 接口
package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/ywty/server/internal/auth"
	bizerr "github.com/ywty/server/internal/errors"
	"github.com/ywty/server/internal/response"
	"github.com/ywty/server/internal/service"
)

// CouponHandler 优惠券接口
type CouponHandler struct {
	svc *service.CouponService
}

// NewCouponHandler 创建优惠券处理器
func NewCouponHandler(svc *service.CouponService) *CouponHandler {
	return &CouponHandler{svc: svc}
}

// CouponValidateReq 优惠券校验请求
type CouponValidateReq struct {
	Code   string `json:"code" binding:"required"`
	Amount uint   `json:"amount"`
}

// AdminList GET /api/v1/admin/coupons
func (h *CouponHandler) AdminList(c *gin.Context) {
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
		LastPage:    calcLastPage(total, perPage),
	})
}

// AdminCreate POST /api/v1/admin/coupons
func (h *CouponHandler) AdminCreate(c *gin.Context) {
	var req service.AdminCouponReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, bizerr.BadRequest.WithMessage(err.Error()))
		return
	}
	coupon, err := h.svc.Create(c.Request.Context(), req)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, coupon)
}

// AdminGetDetail GET /api/v1/admin/coupons/:id
func (h *CouponHandler) AdminGetDetail(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.FailCode(c, bizerr.BadRequest.WithMessage("invalid id"))
		return
	}
	coupon, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, coupon)
}

// AdminUpdate PATCH /api/v1/admin/coupons/:id
func (h *CouponHandler) AdminUpdate(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.FailCode(c, bizerr.BadRequest.WithMessage("invalid id"))
		return
	}
	var req service.AdminCouponReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, bizerr.BadRequest.WithMessage(err.Error()))
		return
	}
	coupon, err := h.svc.Update(c.Request.Context(), id, req)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, coupon)
}

// AdminDelete DELETE /api/v1/admin/coupons/:id
func (h *CouponHandler) AdminDelete(c *gin.Context) {
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

// Validate POST /api/v1/coupons/validate
func (h *CouponHandler) Validate(c *gin.Context) {
	var req CouponValidateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, bizerr.BadRequest.WithMessage(err.Error()))
		return
	}
	coupon, discount, err := h.svc.Validate(c.Request.Context(), req.Code, req.Amount)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, gin.H{
		"coupon":   coupon,
		"discount": discount,
	})
}

// Mount 注册路由
func (h *CouponHandler) Mount(rg *gin.RouterGroup, mw gin.HandlerFunc) {
	// 用户侧校验接口：无需鉴权
	public := rg.Group("/coupons")
	public.POST("/validate", h.Validate)

	// 管理接口：需鉴权 + 管理员
	admin := rg.Group("/admin/coupons", mw, auth.AdminOnly())
	admin.GET("", h.AdminList)
	admin.POST("", h.AdminCreate)
	admin.GET("/:id", h.AdminGetDetail)
	admin.PATCH("/:id", h.AdminUpdate)
	admin.DELETE("/:id", h.AdminDelete)
}
