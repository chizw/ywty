// Package cache 基于 cache 表（Laravel database cache store 同表结构）的缓存实现，
// 带进程内加速层。PHP 版写入的序列化值会被视为未命中（缓存是临时数据，无需互读）。
package cache

import (
	"crypto/sha1"
	"encoding/hex"
	"log/slog"
	"sync"
	"time"

	"gorm.io/gorm"
)

type memEntry struct {
	value    string
	expireAt time.Time
}

type Cache struct {
	db  *gorm.DB
	mem sync.Map // key -> memEntry
}

func New(gdb *gorm.DB) *Cache {
	return &Cache{db: gdb}
}

// keyHash 与 PHP 版 mews/captcha 的 get_cache_key 类似：长 key 做 sha1 截断。
func (c *Cache) keyHash(key string) string {
	if len(key) <= 128 {
		return key
	}
	sum := sha1.Sum([]byte(key))
	return key[:64] + hex.EncodeToString(sum[:])
}

func (c *Cache) Get(key string) (string, bool) {
	k := c.keyHash(key)
	if v, ok := c.mem.Load(k); ok {
		e := v.(memEntry)
		if time.Now().Before(e.expireAt) {
			return e.value, true
		}
		c.mem.Delete(k)
	}
	var val *string
	var expiration int64
	err := c.db.Raw("SELECT `value`, `expiration` FROM `cache` WHERE `key` = ?", k).
		Row().Scan(&val, &expiration)
	if err != nil || val == nil {
		return "", false
	}
	if expiration > 0 && time.Now().Unix() >= expiration {
		_ = c.db.Exec("DELETE FROM `cache` WHERE `key` = ?", k).Error
		return "", false
	}
	return *val, true
}

func (c *Cache) Put(key, value string, ttlSeconds int) {
	k := c.keyHash(key)
	exp := int64(0)
	if ttlSeconds > 0 {
		exp = time.Now().Add(time.Duration(ttlSeconds) * time.Second).Unix()
	}
	memExp := time.Now().Add(time.Hour) // expiration=0 表示永久，内存层给个有限上限
	if exp > 0 {
		memExp = time.Unix(exp, 0)
	}
	c.mem.Store(k, memEntry{value: value, expireAt: memExp})

	var stmt string
	if c.db.Dialector.Name() == "mysql" {
		stmt = "INSERT INTO `cache` (`key`, `value`, `expiration`) VALUES (?, ?, ?) " +
			"ON DUPLICATE KEY UPDATE `value` = VALUES(`value`), `expiration` = VALUES(`expiration`)"
	} else {
		stmt = "INSERT INTO `cache` (`key`, `value`, `expiration`) VALUES (?, ?, ?) " +
			"ON CONFLICT(`key`) DO UPDATE SET `value` = excluded.`value`, `expiration` = excluded.`expiration`"
	}
	if err := c.db.Exec(stmt, k, value, exp).Error; err != nil {
		slog.Warn("cache put 失败", "key", key, "err", err)
	}
}

// Pull 取出并删除。
func (c *Cache) Pull(key string) (string, bool) {
	v, ok := c.Get(key)
	if ok {
		c.Forget(key)
	}
	return v, ok
}

func (c *Cache) Forget(key string) {
	k := c.keyHash(key)
	c.mem.Delete(k)
	_ = c.db.Exec("DELETE FROM `cache` WHERE `key` = ?", k).Error
}

// GetInt 读取整型缓存值。
func (c *Cache) GetInt(key string) (int, bool) {
	v, ok := c.Get(key)
	if !ok {
		return 0, false
	}
	n := 0
	for _, r := range v {
		if r < '0' || r > '9' {
			return 0, false
		}
		n = n*10 + int(r-'0')
	}
	return n, true
}

// Increment 自增并返回新值（用于发送计数、限流）。
func (c *Cache) Increment(key string, ttlSeconds int) int {
	n, _ := c.GetInt(key)
	n++
	c.Put(key, itoa(n), ttlSeconds)
	return n
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}
