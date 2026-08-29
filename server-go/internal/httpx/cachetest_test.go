package httpx_test

import (
	"github.com/chizw/ywty/server-go/internal/cache"
	"gorm.io/gorm"
)

// putCode 直接写验证码进缓存（绕过队列的 SMTP 依赖）。
func putCode(gdb *gorm.DB, key, code string) error {
	cache.New(gdb).Put(key, code, 900)
	return nil
}
