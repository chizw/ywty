// Package handler 验证码 HTTP 接口
package handler

import (
	"github.com/gin-gonic/gin"

	bizerr "github.com/ywty/server/internal/errors"
	"github.com/ywty/server/internal/response"
	"github.com/ywty/server/internal/service"
)

type VerifyCodeHandler struct {
	svc *service.VerifyCodeService
}

func NewVerifyCodeHandler(svc *service.VerifyCodeService) *VerifyCodeHandler {
	return &VerifyCodeHandler{svc: svc}
}

// Send POST /api/v1/verify-codes
// body: { channel: "email"|"sms", account: "...", event: "register|reset_password|..." }
func (h *VerifyCodeHandler) Send(c *gin.Context) {
	var req struct {
		Channel string `json:"channel" binding:"required,oneof=email sms"`
		Account string `json:"account" binding:"required"`
		Event   string `json:"event" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, bizerr.BadRequest.WithMessage(err.Error()))
		return
	}
	if err := h.svc.Send(c.Request.Context(), req.Channel, req.Account, req.Event, c.ClientIP()); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, gin.H{"sent": true})
}

func (h *VerifyCodeHandler) Mount(rg *gin.RouterGroup) {
	rg.POST("/verify-codes", h.Send)
}
