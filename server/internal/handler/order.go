// Package handler 订单 HTTP 接口
package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/ywty/server/internal/auth"
	bizerr "github.com/ywty/server/internal/errors"
	"github.com/ywty/server/internal/payment"
	"github.com/ywty/server/internal/response"
	"github.com/ywty/server/internal/service"
)

// OrderHandler 订单接口
type OrderHandler struct {
	svc *service.OrderService
}

// NewOrderHandler 创建订单处理器
func NewOrderHandler(svc *service.OrderService) *OrderHandler {
	return &OrderHandler{svc: svc}
}

// Create POST /api/v1/orders
func (h *OrderHandler) Create(c *gin.Context) {
	uid, ok := auth.CurrentUserID(c)
	if !ok {
		response.FailCode(c, bizerr.Unauthorized)
		return
	}
	var req service.CreateOrderReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, bizerr.BadRequest.WithMessage(err.Error()))
		return
	}
	order, err := h.svc.CreatePlanOrder(c.Request.Context(), uid, req)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, order)
}

// List GET /api/v1/orders
func (h *OrderHandler) List(c *gin.Context) {
	uid, ok := auth.CurrentUserID(c)
	if !ok {
		response.FailCode(c, bizerr.Unauthorized)
		return
	}
	page, _ := strconv.Atoi(c.Query("page"))
	perPage, _ := strconv.Atoi(c.Query("per_page"))
	rows, err := h.svc.List(c.Request.Context(), uid, page, perPage)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, rows)
}

// Get GET /api/v1/orders/:id
func (h *OrderHandler) Get(c *gin.Context) {
	uid, ok := auth.CurrentUserID(c)
	if !ok {
		response.FailCode(c, bizerr.Unauthorized)
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.FailCode(c, bizerr.BadRequest.WithMessage("invalid order id"))
		return
	}
	order, err := h.svc.Get(c.Request.Context(), uid, id)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, order)
}

// Cancel POST /api/v1/orders/:id/cancel
func (h *OrderHandler) Cancel(c *gin.Context) {
	uid, ok := auth.CurrentUserID(c)
	if !ok {
		response.FailCode(c, bizerr.Unauthorized)
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.FailCode(c, bizerr.BadRequest.WithMessage("invalid order id"))
		return
	}
	if err := h.svc.Cancel(c.Request.Context(), uid, id); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, gin.H{"id": id, "canceled": true})
}

// Pay POST /api/v1/orders/:id/pay
func (h *OrderHandler) Pay(c *gin.Context) {
	uid, ok := auth.CurrentUserID(c)
	if !ok {
		response.FailCode(c, bizerr.Unauthorized)
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.FailCode(c, bizerr.BadRequest.WithMessage("invalid order id"))
		return
	}
	res, err := h.svc.Pay(c.Request.Context(), uid, id, c.ClientIP())
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, res)
}

// Notify POST /api/v1/orders/notify
func (h *OrderHandler) Notify(c *gin.Context) {
	payMethod := c.Query("pay_method")
	if payMethod == "" {
		payMethod = c.PostForm("pay_method")
	}
	if payMethod == "" {
		var jsonData map[string]any
		if err := c.ShouldBindJSON(&jsonData); err == nil {
			if v, ok := jsonData["pay_method"].(string); ok {
				payMethod = v
			}
		}
	}
	if payMethod == "" {
		response.FailCode(c, bizerr.BadRequest.WithMessage("pay_method required"))
		return
	}

	driver, err := payment.Get(payMethod, nil)
	if err != nil {
		response.FailCode(c, bizerr.PaymentChannelInvalid)
		return
	}

	res, err := driver.VerifyCallback(c.Request.Context(), c.Request, nil)
	if err != nil || !res.Paid {
		response.FailCode(c, bizerr.PaymentFailed.WithMessage("notify verify failed"))
		return
	}

	if err := h.svc.HandlePaid(c.Request.Context(), res.TradeNo, res.OutTradeNo, payMethod); err != nil {
		response.Fail(c, err)
		return
	}
	c.String(200, "OK")
}

// Mount 挂载路由
func (h *OrderHandler) Mount(rg *gin.RouterGroup, mw gin.HandlerFunc) {
	// 支付回调公开访问
	rg.POST("/orders/notify", h.Notify)

	g := rg.Group("/orders", mw)
	g.POST("", h.Create)
	g.GET("", h.List)
	g.GET("/:id", h.Get)
	g.POST("/:id/cancel", h.Cancel)
	g.POST("/:id/pay", h.Pay)
}
