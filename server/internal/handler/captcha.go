// Package handler 图形验证码 HTTP 接口
package handler

import (
	"github.com/gin-gonic/gin"

	bizerr "github.com/ywty/server/internal/errors"
	"github.com/ywty/server/internal/response"
	"github.com/ywty/server/internal/service"
)

// CaptchaHandler 图形验证码接口
type CaptchaHandler struct {
	svc *service.CaptchaService
}

// NewCaptchaHandler 创建图形验证码处理器
func NewCaptchaHandler(svc *service.CaptchaService) *CaptchaHandler {
	return &CaptchaHandler{svc: svc}
}

// Get GET /api/v1/captcha
// 生成图形验证码，返回 base64 图片 + captcha_id
func (h *CaptchaHandler) Get(c *gin.Context) {
	id, _, imageBase64 := h.svc.Generate()
	if imageBase64 == "" {
		response.FailCode(c, bizerr.Internal.WithMessage("生成验证码失败"))
		return
	}
	response.Success(c, gin.H{
		"captcha_id": id,
		"image":      imageBase64,
	})
}

// Verify POST /api/v1/captcha/verify
// 校验验证码（内部接口，供注册/登录前验证）
// body: { "captcha_id": "xxx", "code": "xxxx" }
func (h *CaptchaHandler) Verify(c *gin.Context) {
	var req struct {
		CaptchaID string `json:"captcha_id" binding:"required"`
		Code      string `json:"code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, bizerr.BadRequest.WithMessage(err.Error()))
		return
	}
	if !h.svc.Verify(req.CaptchaID, req.Code) {
		response.FailCode(c, bizerr.CaptchaIncorrect)
		return
	}
	response.Success(c, gin.H{"verified": true})
}

// Mount 注册路由（公开接口，无需鉴权）
func (h *CaptchaHandler) Mount(rg *gin.RouterGroup) {
	g := rg.Group("/captcha")
	g.GET("", h.Get)
	g.POST("/verify", h.Verify)
}
