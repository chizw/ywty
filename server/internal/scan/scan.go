// Package scan 图片违规扫描驱动抽象
package scan

import (
	"context"
	"errors"
	"io"
)

// Result 扫描结果
type Result struct {
	Safe     bool     `json:"safe"`     // 是否安全
	Labels   []string `json:"labels"`   // 命中标签（如 porn, ad, ...）
	Severity int      `json:"severity"` // 0 安全, 1 警告, 2 违规, 3 严重违规
	Message  string   `json:"message"`
}

// Driver 扫描驱动接口
type Driver interface {
	Name() string
	// Scan 扫描图片数据
	Scan(ctx context.Context, data io.Reader, mime string) (*Result, error)
}

// Factory 驱动工厂
type Factory func(cfg map[string]string) (Driver, error)

var registry = map[string]Factory{}

// Register 注册
func Register(name string, f Factory) { registry[name] = f }

// Get 构造
func Get(name string, cfg map[string]string) (Driver, error) {
	f, ok := registry[name]
	if !ok {
		return nil, errors.New("unsupported scan driver: " + name)
	}
	return f(cfg)
}

// Drivers 列出
func Drivers() []string {
	out := make([]string, 0, len(registry))
	for k := range registry {
		out = append(out, k)
	}
	return out
}

// NoopDriver 占位驱动（始终返回 safe）
type NoopDriver struct{}

func NewNoopDriver(_ map[string]string) (Driver, error) { return &NoopDriver{}, nil }
func (n *NoopDriver) Name() string                      { return "noop" }
func (n *NoopDriver) Scan(_ context.Context, _ io.Reader, _ string) (*Result, error) {
	return &Result{Safe: true}, nil
}

func init() {
	Register("noop", NewNoopDriver)
}
