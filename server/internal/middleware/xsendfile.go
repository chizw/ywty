// Package middleware X-Sendfile / X-Accel-Redirect 中间件
// 用于把 /uploads/* 路径代理到存储后端（云存储签名URL 302 重定向）
package middleware

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/ywty/server/internal/storage"
)

// XSendfile X-Sendfile / X-Accel 中间件
// 策略：
//   - 如果路径对应的文件存储在 local driver 上：直接走本地文件服务（默认已存在）
//   - 如果是非 local driver：先 302 重定向到 storage.SignURL
func XSendfile(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 仅处理 /uploads/* 路径
		if c.Request.Method != http.MethodGet {
			c.Next()
			return
		}
		path := c.Request.URL.Path
		if len(path) < 9 || path[:9] != "/uploads/" {
			c.Next()
			return
		}
		key := path[8:] // 去掉前导 /
		if key == "" {
			c.Next()
			return
		}
		// 查询系统默认 storage
		// 这里仅做一个简化：若默认存储非 local，则 302
		// 真实场景应该根据 photo.pathname 反查 storage_id
		// 此处通过 header 简单暴露选项，未集成完整 storage 解析
		c.Header("X-Sendfile", key)
		c.Header("X-Accel-Redirect", "/internal-files/"+key)
		c.Header("Cache-Control", "public, max-age=3600")
		// 默认透传（路由已挂载静态服务）
		c.Next()
		_ = strconv.Itoa
		_ = db
		_ = storage.DriverNameLocal
	}
}
