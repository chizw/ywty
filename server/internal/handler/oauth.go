// Package handler OAuth 三方登录 HTTP 接口
package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/ywty/server/internal/auth"
	bizerr "github.com/ywty/server/internal/errors"
	"github.com/ywty/server/internal/response"
	"github.com/ywty/server/internal/service"
)

type OAuthHandler struct {
	svc *service.OAuthService
}

func NewOAuthHandler(svc *service.OAuthService) *OAuthHandler {
	return &OAuthHandler{svc: svc}
}

// List GET /api/v1/oauth
func (h *OAuthHandler) List(c *gin.Context) {
	uid, ok := auth.CurrentUserID(c)
	if !ok {
		response.FailCode(c, bizerr.Unauthorized)
		return
	}
	rows, err := h.svc.List(c.Request.Context(), uid)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, rows)
}

// Bind POST /api/v1/oauth
func (h *OAuthHandler) Bind(c *gin.Context) {
	uid, ok := auth.CurrentUserID(c)
	if !ok {
		response.FailCode(c, bizerr.Unauthorized)
		return
	}
	var req service.BindRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, bizerr.BadRequest.WithMessage(err.Error()))
		return
	}
	oa, err := h.svc.Bind(c.Request.Context(), uid, req)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, oa)
}

// Unbind DELETE /api/v1/oauth/:id
func (h *OAuthHandler) Unbind(c *gin.Context) {
	uid, ok := auth.CurrentUserID(c)
	if !ok {
		response.FailCode(c, bizerr.Unauthorized)
		return
	}
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.svc.Unbind(c.Request.Context(), uid, id); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, gin.H{"id": id, "unbound": true})
}

// FindByOpenID POST /api/v1/oauth/find
// 用于开发期 / 内部：给定 driver_id + openid 返回已绑定用户
func (h *OAuthHandler) FindByOpenID(c *gin.Context) {
	var req struct {
		DriverID uint64 `json:"driver_id" binding:"required"`
		OpenID   string `json:"openid" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, bizerr.BadRequest.WithMessage(err.Error()))
		return
	}
	u, err := h.svc.FindByOpenID(c.Request.Context(), req.DriverID, req.OpenID)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, gin.H{
		"id":       u.ID,
		"username": u.Username,
		"email":    u.Email,
	})
}

// Authorize GET /api/v1/oauth/:provider/authorize
// 返回三方授权跳转 URL（公开接口）
func (h *OAuthHandler) Authorize(c *gin.Context) {
	provider := c.Param("provider")
	url, state, err := h.svc.AuthorizeURL(c.Request.Context(), provider)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, gin.H{
		"url":   url,
		"state": state,
	})
}

// Callback GET /api/v1/oauth/:provider/callback
// OAuth 回调：用 code 换 token、拉用户信息、登录或自动注册
func (h *OAuthHandler) Callback(c *gin.Context) {
	provider := c.Param("provider")
	code := c.Query("code")
	if code == "" {
		response.FailCode(c, bizerr.BadRequest.WithMessage("code is required"))
		return
	}
	driver, err := h.svc.GetDriverByProvider(c.Request.Context(), provider)
	if err != nil {
		response.Fail(c, err)
		return
	}
	res, err := h.svc.LoginOrRegister(c.Request.Context(), driver.ID, provider, code)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, res)
}

func (h *OAuthHandler) Mount(rg *gin.RouterGroup, mw gin.HandlerFunc) {
	g := rg.Group("/oauth")
	// 公开：授权跳转与回调登录（无需鉴权）
	g.GET("/:provider/authorize", h.Authorize)
	g.GET("/:provider/callback", h.Callback)
	// 需鉴权
	g.GET("", mw, h.List)
	g.POST("", mw, h.Bind)
	g.DELETE("/:id", mw, h.Unbind)
	g.POST("/find", mw, h.FindByOpenID)
}
