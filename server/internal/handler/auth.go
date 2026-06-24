// Package handler HTTP 处理器层（按业务模块拆文件）
package handler

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/ywty/server/internal/auth"
	bizerr "github.com/ywty/server/internal/errors"
	"github.com/ywty/server/internal/logger"
	"github.com/ywty/server/internal/response"
	"github.com/ywty/server/internal/service"
)

// AuthHandler 认证相关接口
type AuthHandler struct {
	svc *service.AuthService
	cap *CaptchaHandler
}

// NewAuthHandler 创建认证处理器
func NewAuthHandler(svc *service.AuthService, cap *CaptchaHandler) *AuthHandler {
	return &AuthHandler{svc: svc, cap: cap}
}

// Register POST /api/v1/auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	var req service.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, bizerr.BadRequest.WithMessage(err.Error()))
		return
	}
	res, err := h.svc.Register(c.Request.Context(), req)
	if err != nil {
		logger.L.Error("register failed", zap.Error(err))
		response.Fail(c, service.ToAPIError(err))
		return
	}
	response.Success(c, res)
}

// Login POST /api/v1/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req service.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, bizerr.BadRequest.WithMessage(err.Error()))
		return
	}
	res, err := h.svc.Login(c.Request.Context(), req)
	if err != nil {
		logger.L.Error("login failed", zap.Error(err))
		response.Fail(c, service.ToAPIError(err))
		return
	}
	response.Success(c, res)
}

// Refresh POST /api/v1/auth/refresh
func (h *AuthHandler) Refresh(c *gin.Context) {
	var body struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.FailCode(c, bizerr.BadRequest.WithMessage(err.Error()))
		return
	}
	res, err := h.svc.Refresh(c.Request.Context(), body.RefreshToken)
	if err != nil {
		response.Fail(c, service.ToAPIError(err))
		return
	}
	response.Success(c, res)
}

// Me GET /api/v1/auth/me
func (h *AuthHandler) Me(c *gin.Context) {
	uid, ok := auth.CurrentUserID(c)
	if !ok {
		response.FailCode(c, bizerr.Unauthorized)
		return
	}
	u, err := h.svc.Me(uid)
	if err != nil {
		if service.IsRecordNotFound(err) {
			response.FailCode(c, bizerr.UserNotFound)
			return
		}
		response.Fail(c, err)
		return
	}
	response.Success(c, u)
}

// Logout POST /api/v1/auth/logout
// JWT 模式下无服务端会话，客户端清除 token 即可
func (h *AuthHandler) Logout(c *gin.Context) {
	response.Success(c, gin.H{"message": "ok"})
}

// ResetPasswordByEmail POST /api/v1/auth/reset-password
func (h *AuthHandler) ResetPasswordByEmail(c *gin.Context) {
	var req struct {
		Email       string `json:"email" binding:"required,email"`
		Code        string `json:"code" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=6"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, bizerr.BadRequest.WithMessage(err.Error()))
		return
	}
	if err := h.svc.ResetPasswordByEmail(c.Request.Context(), req.Email, req.Code, req.NewPassword); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, gin.H{"reset": true})
}

// ResetPasswordByPhone POST /api/v1/auth/reset-password/phone
func (h *AuthHandler) ResetPasswordByPhone(c *gin.Context) {
	var req struct {
		Phone       string `json:"phone" binding:"required"`
		Code        string `json:"code" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=6"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, bizerr.BadRequest.WithMessage(err.Error()))
		return
	}
	if err := h.svc.ResetPasswordByPhone(c.Request.Context(), req.Phone, req.Code, req.NewPassword); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, gin.H{"reset": true})
}

// Mount 注册到路由
func (h *AuthHandler) Mount(rg *gin.RouterGroup, mw gin.HandlerFunc) {
	// 图形验证码（公开接口）
	if h.cap != nil {
		h.cap.Mount(rg)
	}
	g := rg.Group("/auth")
	g.POST("/register", h.Register)
	g.POST("/login", h.Login)
	g.POST("/refresh", h.Refresh)
	g.POST("/logout", h.Logout)
	g.POST("/reset-password", h.ResetPasswordByEmail)
	g.POST("/reset-password/phone", h.ResetPasswordByPhone)
	g.GET("/me", mw, h.Me)
}
