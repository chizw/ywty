// Package middleware 限流中间件（基于内存 token bucket）
package middleware

import (
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	bizerr "github.com/ywty/server/internal/errors"
	"github.com/ywty/server/internal/response"
)

// bucket 单 IP 限流桶
type bucket struct {
	tokens   float64
	lastFill time.Time
}

// RateLimiter 限流器
type RateLimiter struct {
	mu       sync.Mutex
	buckets  map[string]*bucket
	rate     float64       // 每秒补充 token 数
	capacity float64       // 桶容量（最大突发）
	ttl      time.Duration // 桶空闲清理时间
}

// NewRateLimiter 构造限流器
// perMinute: 每分钟允许的请求数
func NewRateLimiter(perMinute int) *RateLimiter {
	if perMinute <= 0 {
		perMinute = 60
	}
	rl := &RateLimiter{
		buckets:  make(map[string]*bucket),
		rate:     float64(perMinute) / 60.0,
		capacity: float64(perMinute),
		ttl:      10 * time.Minute,
	}
	go rl.gc()
	return rl
}

// Allow 是否允许
func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	b, ok := rl.buckets[key]
	if !ok {
		rl.buckets[key] = &bucket{tokens: rl.capacity - 1, lastFill: now}
		return true
	}
	elapsed := now.Sub(b.lastFill).Seconds()
	b.tokens += elapsed * rl.rate
	if b.tokens > rl.capacity {
		b.tokens = rl.capacity
	}
	b.lastFill = now
	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}

func (rl *RateLimiter) gc() {
	t := time.NewTicker(time.Minute)
	defer t.Stop()
	for range t.C {
		now := time.Now()
		rl.mu.Lock()
		for k, b := range rl.buckets {
			if now.Sub(b.lastFill) > rl.ttl {
				delete(rl.buckets, k)
			}
		}
		rl.mu.Unlock()
	}
}

// Middleware 限流中间件，按 client IP 限流
func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !rl.Allow(c.ClientIP()) {
			response.FailCode(c, bizerr.TooManyRequests)
			c.Abort()
			return
		}
		c.Next()
	}
}
