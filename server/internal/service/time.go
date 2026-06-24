// Package service 通用辅助
package service

import "time"

// timeNow 提供测试 hook（生产直接返回 time.Now）
var timeNow = func() time.Time { return time.Now() }
