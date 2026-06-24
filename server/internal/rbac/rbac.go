// Package rbac 提供 Casbin RBAC 鉴权能力
package rbac

import (
	"fmt"

	"github.com/casbin/casbin/v2"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"gorm.io/gorm"

	"github.com/ywty/server/internal/logger"
	"go.uber.org/zap"
)

// Init 使用 gorm-adapter 加载 casbin 模型并返回 Enforcer
func Init(db *gorm.DB) (*casbin.Enforcer, error) {
	adapter, err := gormadapter.NewAdapterByDB(db)
	if err != nil {
		return nil, fmt.Errorf("casbin gorm adapter: %w", err)
	}
	enforcer, err := casbin.NewEnforcer("configs/casbin_model.conf", adapter)
	if err != nil {
		return nil, fmt.Errorf("casbin new enforcer: %w", err)
	}
	if err := enforcer.LoadPolicy(); err != nil {
		return nil, fmt.Errorf("casbin load policy: %w", err)
	}
	return enforcer, nil
}

// AutoSeed 初始化默认策略（幂等：已存在的策略不会重复添加）
func AutoSeed(e *casbin.Enforcer) error {
	if e == nil {
		return nil
	}

	// 默认策略：p, sub, obj, act
	policies := [][]string{
		// admin 拥有全部权限
		{"role:admin", "/*", "*"},
		// user 策略（路径与实际路由一致：/user 单数，其余复数）
		{"role:user", "/api/v1/user/*", "GET|POST|PATCH|DELETE"},
		{"role:user", "/api/v1/photos/*", "GET|POST|PATCH|DELETE"},
		{"role:user", "/api/v1/albums/*", "GET|POST|PATCH|DELETE"},
		{"role:user", "/api/v1/tags/*", "GET|POST|DELETE"},
		{"role:user", "/api/v1/shares/*", "GET|POST|PATCH|DELETE"},
		{"role:user", "/api/v1/likes/*", "GET|POST"},
		{"role:user", "/api/v1/reports/*", "POST"},
		{"role:user", "/api/v1/tokens/*", "GET|POST|DELETE"},
		{"role:user", "/api/v1/oauth/*", "GET|POST|DELETE"},
		{"role:user", "/api/v1/capacity", "GET"},
		{"role:user", "/api/v1/storage/sign", "GET"},
		// guest 策略
		{"role:guest", "/api/v1/auth/*", "POST|GET"},
		{"role:guest", "/api/v1/verify-codes/*", "POST"},
		{"role:guest", "/api/v1/captcha/*", "GET|POST"},
		{"role:guest", "/api/v1/plans", "GET"},
		{"role:guest", "/api/v1/oauth/:provider/authorize", "GET"},
		{"role:guest", "/api/v1/oauth/:provider/callback", "GET"},
	}

	for _, p := range policies {
		exists, err := e.HasPolicy(p[0], p[1], p[2])
		if err != nil {
			return fmt.Errorf("casbin has policy: %w", err)
		}
		if exists {
			continue
		}
		if _, err := e.AddPolicy(p[0], p[1], p[2]); err != nil {
			return fmt.Errorf("casbin add policy %v: %w", p, err)
		}
	}

	logger.L.Info("casbin policies seeded", zap.Int("count", len(policies)))
	return nil
}
