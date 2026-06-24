// Package payment 支付驱动
package payment

import (
	"context"
	"errors"
	"fmt"
)

// Order 订单
type Order struct {
	ID          string  // 订单号
	Subject     string  // 订单标题
	Amount      float64 // 金额（元）
	Currency    string  // 货币（CNY/USD/JPY）
	UserID      string
	ExpireTime  int64 // 过期时间戳
	NotifyURL   string
	ReturnURL   string
	Metadata    map[string]string
}

// PayResult 支付发起结果
type PayResult struct {
	OrderID     string
	PayURL      string // 跳转支付 URL
	FormHTML    string // 表单提交 HTML（用于扫码支付）
	QRCode      string // 二维码内容
	PrepayID    string // 预支付 ID
	ExpireTime  int64
	Raw         any
}

// NotifyPayload 回调通知载荷
type NotifyPayload struct {
	OrderID     string
	TradeNo     string // 第三方流水号
	Status      string // paid/closed/refunded
	Amount      float64
	PaidAt      int64
	Raw         any
}

// RefundRequest 退款请求
type RefundRequest struct {
	OrderID   string
	RefundNo  string
	Amount    float64
	Reason    string
}

// RefundResult 退款结果
type RefundResult struct {
	RefundNo string
	TradeNo  string
	Accepted bool
	Raw      any
}

// Driver 支付驱动接口
type Driver interface {
	// Name 驱动名
	Name() string
	// CreatePay 发起支付
	CreatePay(ctx context.Context, order *Order) (*PayResult, error)
	// VerifyNotify 验证回调
	VerifyNotify(ctx context.Context, raw []byte, headers map[string]string) (*NotifyPayload, error)
	// Refund 退款
	Refund(ctx context.Context, req *RefundRequest) (*RefundResult, error)
	// Close 关闭订单
	Close(ctx context.Context, orderID string) error
	// Query 查询订单状态
	Query(ctx context.Context, orderID string) (*NotifyPayload, error)
}

// Factory 驱动工厂
type Factory struct {
	drivers map[string]Driver
}

// NewFactory 创建工厂
func NewFactory() *Factory {
	return &Factory{drivers: make(map[string]Driver)}
}

// Register 注册驱动
func (f *Factory) Register(d Driver) {
	f.drivers[d.Name()] = d
}

// Get 获取驱动
func (f *Factory) Get(name string) (Driver, error) {
	d, ok := f.drivers[name]
	if !ok {
		return nil, fmt.Errorf("payment driver %q not found", name)
	}
	return d, nil
}

// List 列出所有驱动
func (f *Factory) List() []string {
	out := make([]string, 0, len(f.drivers))
	for k := range f.drivers {
		out = append(out, k)
	}
	return out
}

// ErrUnsupported 暂不支持的错误
var ErrUnsupported = errors.New("payment driver not supported in this build")
