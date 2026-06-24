// Package handler 用户相关 HTTP 接口
package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/ywty/server/internal/auth"
	bizerr "github.com/ywty/server/internal/errors"
	"github.com/ywty/server/internal/response"
	"github.com/ywty/server/internal/service"
)

// UserHandler 用户中心接口
type UserHandler struct {
	svc *service.UserService
}

func NewUserHandler(svc *service.UserService) *UserHandler { return &UserHandler{svc: svc} }

// GetProfile GET /api/v1/user/profile
func (h *UserHandler) GetProfile(c *gin.Context) {
	uid, ok := auth.CurrentUserID(c)
	if !ok {
		response.FailCode(c, bizerr.Unauthorized)
		return
	}
	p, err := h.svc.GetProfile(c.Request.Context(), uid)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, p)
}

// UpdateProfile PATCH /api/v1/user/profile
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	uid, ok := auth.CurrentUserID(c)
	if !ok {
		response.FailCode(c, bizerr.Unauthorized)
		return
	}
	var req service.UpdateProfileReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, bizerr.BadRequest.WithMessage(err.Error()))
		return
	}
	p, err := h.svc.UpdateProfile(c.Request.Context(), uid, req)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, p)
}

// ChangePassword POST /api/v1/user/change-password
func (h *UserHandler) ChangePassword(c *gin.Context) {
	uid, ok := auth.CurrentUserID(c)
	if !ok {
		response.FailCode(c, bizerr.Unauthorized)
		return
	}
	var req struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=6"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, bizerr.BadRequest.WithMessage(err.Error()))
		return
	}
	if err := h.svc.ChangePassword(c.Request.Context(), uid, req.OldPassword, req.NewPassword); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, gin.H{"changed": true})
}

// ChangeEmail POST /api/v1/user/change-email
func (h *UserHandler) ChangeEmail(c *gin.Context) {
	uid, ok := auth.CurrentUserID(c)
	if !ok {
		response.FailCode(c, bizerr.Unauthorized)
		return
	}
	var req struct {
		NewEmail string `json:"new_email" binding:"required,email"`
		Code     string `json:"code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, bizerr.BadRequest.WithMessage(err.Error()))
		return
	}
	if err := h.svc.ChangeEmail(c.Request.Context(), uid, req.NewEmail, req.Code); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, gin.H{"changed": true})
}

// ChangePhone POST /api/v1/user/change-phone
func (h *UserHandler) ChangePhone(c *gin.Context) {
	uid, ok := auth.CurrentUserID(c)
	if !ok {
		response.FailCode(c, bizerr.Unauthorized)
		return
	}
	var req struct {
		NewPhone string `json:"new_phone" binding:"required"`
		Code     string `json:"code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, bizerr.BadRequest.WithMessage(err.Error()))
		return
	}
	if err := h.svc.ChangePhone(c.Request.Context(), uid, req.NewPhone, req.Code); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, gin.H{"changed": true})
}

// Mount 挂载路由
func (h *UserHandler) Mount(rg *gin.RouterGroup, mw gin.HandlerFunc) {
	g := rg.Group("/user")
	g.GET("/profile", mw, h.GetProfile)
	g.PATCH("/profile", mw, h.UpdateProfile)
	g.POST("/change-password", mw, h.ChangePassword)
	g.POST("/change-email", mw, h.ChangeEmail)
	g.POST("/change-phone", mw, h.ChangePhone)
}

// _ strconv 编译期占位（接口签名可能需要解析）
var _ = strconv.Atoi
