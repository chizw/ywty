package payment

import (
	"fmt"
	"sort"
)

// Factory 支付驱动工厂
type Factory func(map[string]string) (Driver, error)

var registry = map[string]Factory{}

// Register 注册支付驱动工厂
func Register(name string, factory Factory) {
	registry[name] = factory
}

// Get 按名称获取驱动实例
func Get(name string, cfg map[string]string) (Driver, error) {
	f, ok := registry[name]
	if !ok {
		return nil, fmt.Errorf("payment driver %s not found", name)
	}
	return f(cfg)
}

// Drivers 返回已注册驱动名称列表（排序后）
func Drivers() []string {
	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
