package rbac

import (
	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"

	"github.com/ywty/server/internal/auth"
	bizerr "github.com/ywty/server/internal/errors"
	"github.com/ywty/server/internal/response"
)

// 角色常量
const (
	RoleAdmin = "role:admin"
	RoleUser  = "role:user"
	RoleGuest = "role:guest"
)

// Middleware 返回 Casbin 鉴权中间件
// 需在 auth.Middleware 之后使用，依据 context 中的用户身份确定角色
func Middleware(enforcer *casbin.Enforcer) gin.HandlerFunc {
	return func(c *gin.Context) {
		if enforcer == nil {
			c.Next()
			return
		}

		role := resolveRole(c)

		ok, err := enforcer.Enforce(role, c.Request.URL.Path, c.Request.Method)
		if err != nil {
			response.FailCode(c, bizerr.Internal.WithMessage("casbin enforce: "+err.Error()))
			c.Abort()
			return
		}
		if !ok {
			response.FailCode(c, bizerr.Forbidden)
			c.Abort()
			return
		}
		c.Next()
	}
}

// resolveRole 依据 context 中的用户身份解析角色
func resolveRole(c *gin.Context) string {
	uid, ok := auth.CurrentUserID(c)
	if !ok || uid == 0 {
		return RoleGuest
	}
	if auth.CurrentIsAdmin(c) {
		return RoleAdmin
	}
	return RoleUser
}
