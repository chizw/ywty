package auth

import (
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	bizerr "github.com/ywty/server/internal/errors"
	"github.com/ywty/server/internal/response"
)

// 上下文键
const (
	CtxUserID   = "ywty.user_id"
	CtxUsername = "ywty.username"
	CtxIsAdmin  = "ywty.is_admin"
)

// Middleware JWT 鉴权中间件
// 同时支持 Authorization: Bearer <token> 与 X-Token: <token>
func Middleware(db *gorm.DB, issuer *Issuer) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractToken(c)
		if token == "" {
			response.FailCode(c, bizerr.Unauthorized)
			c.Abort()
			return
		}
		claims, err := issuer.Parse(token)
		if err != nil {
			response.FailCode(c, bizerr.TokenInvalid)
			c.Abort()
			return
		}
		// 可在此处加用户存在性 / 状态校验（避免被禁用户仍能访问）
		c.Set(CtxUserID, claims.UserID)
		c.Set(CtxUsername, claims.Username)
		c.Set(CtxIsAdmin, claims.IsAdmin)
		c.Next()
	}
}

// AdminOnly 仅管理员可访问（在 Middleware 之后）
func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		isAdmin, ok := c.Get(CtxIsAdmin)
		if !ok || !isAdmin.(bool) {
			response.FailCode(c, bizerr.Forbidden)
			c.Abort()
			return
		}
		c.Next()
	}
}

// OptionalAuth 可选鉴权：有 token 解析无 token 放行
func OptionalAuth(issuer *Issuer) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractToken(c)
		if token == "" {
			c.Next()
			return
		}
		if claims, err := issuer.Parse(token); err == nil {
			c.Set(CtxUserID, claims.UserID)
			c.Set(CtxUsername, claims.Username)
			c.Set(CtxIsAdmin, claims.IsAdmin)
		}
		c.Next()
	}
}

// extractToken 提取 Bearer / X-Token 中的 token
func extractToken(c *gin.Context) string {
	if h := c.GetHeader("Authorization"); h != "" {
		if strings.HasPrefix(h, "Bearer ") {
			return strings.TrimPrefix(h, "Bearer ")
		}
		return h
	}
	if t := c.GetHeader("X-Token"); t != "" {
		return t
	}
	if t := c.Query("token"); t != "" {
		return t
	}
	return ""
}

// CurrentUserID 获取当前用户 ID
func CurrentUserID(c *gin.Context) (uint64, bool) {
	v, ok := c.Get(CtxUserID)
	if !ok {
		return 0, false
	}
	return v.(uint64), true
}

// CurrentIsAdmin 判断当前用户是否为管理员
func CurrentIsAdmin(c *gin.Context) bool {
	v, ok := c.Get(CtxIsAdmin)
	if !ok {
		return false
	}
	return v.(bool)
}
