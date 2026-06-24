package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
)

// CORS 跨域中间件
func CORS(allowOrigins []string, allowMethods []string, allowHeaders []string,
	exposeHeaders []string, allowCredentials bool, maxAge int) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		allowed := false
		for _, o := range allowOrigins {
			if o == "*" || o == origin {
				allowed = true
				break
			}
		}
		if allowed {
			if len(allowOrigins) == 1 && allowOrigins[0] == "*" {
				c.Header("Access-Control-Allow-Origin", "*")
			} else {
				c.Header("Access-Control-Allow-Origin", origin)
				c.Header("Vary", "Origin")
			}
			c.Header("Access-Control-Allow-Methods", joinCSV(allowMethods))
			c.Header("Access-Control-Allow-Headers", joinCSV(allowHeaders))
			c.Header("Access-Control-Expose-Headers", joinCSV(exposeHeaders))
			if allowCredentials {
				c.Header("Access-Control-Allow-Credentials", "true")
			}
			c.Header("Access-Control-Max-Age", itoa(maxAge))
		}

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}

// RequestID 请求 ID
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := c.GetHeader("X-Request-Id")
		if rid == "" {
			rid = randomID()
		}
		c.Set("request_id", rid)
		c.Header("X-Request-Id", rid)
		c.Next()
	}
}

// AccessLog 访问日志
func AccessLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		// 实际写入 zap logger 留给调用方
		_ = start
	}
}

func joinCSV(s []string) string {
	out := ""
	for i, v := range s {
		if i > 0 {
			out += ", "
		}
		out += v
	}
	return out
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		b[i] = '-'
	}
	return string(b[i:])
}

func randomID() string {
	// 简单实现：基于时间戳
	return time.Now().Format("20060102150405") + "-" + randomHex(8)
}

func randomHex(n int) string {
	const hex = "0123456789abcdef"
	ns := time.Now().UnixNano()
	out := make([]byte, n)
	for i := 0; i < n; i++ {
		out[i] = hex[int(uint64(ns)>>uint(i*4))&0xf]
	}
	return string(out)
}
