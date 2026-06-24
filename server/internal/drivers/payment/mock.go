// Package payment 模拟支付驱动（用于开发环境）
package payment

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"
)

// MockDriver 模拟支付（开发/测试用）
type MockDriver struct{}

// NewMockDriver 创建模拟支付驱动
func NewMockDriver() *MockDriver { return &MockDriver{} }

// Name 驱动名
func (d *MockDriver) Name() string { return "mock" }

// CreatePay 模拟发起
func (d *MockDriver) CreatePay(ctx context.Context, order *Order) (*PayResult, error) {
	b := make([]byte, 8)
	rand.Read(b)
	return &PayResult{
		OrderID:    order.ID,
		PayURL:     "/mock-pay?order=" + order.ID,
		ExpireTime: time.Now().Add(15 * time.Minute).Unix(),
		Raw:        map[string]string{"token": hex.EncodeToString(b)},
	}, nil
}

// VerifyNotify 模拟验证（始终通过）
func (d *MockDriver) VerifyNotify(ctx context.Context, raw []byte, headers map[string]string) (*NotifyPayload, error) {
	return &NotifyPayload{
		OrderID: "mock-order",
		TradeNo: "mock-trade",
		Status:  "paid",
		Amount:  0,
		PaidAt:  time.Now().Unix(),
		Raw:     map[string]string{"mock": "true"},
	}, nil
}

// Refund 模拟退款
func (d *MockDriver) Refund(ctx context.Context, req *RefundRequest) (*RefundResult, error) {
	return &RefundResult{RefundNo: req.RefundNo, TradeNo: req.OrderID, Accepted: true}, nil
}

// Close 模拟关闭
func (d *MockDriver) Close(ctx context.Context, orderID string) error { return nil }

// Query 模拟查询
func (d *MockDriver) Query(ctx context.Context, orderID string) (*NotifyPayload, error) {
	return &NotifyPayload{OrderID: orderID, Status: "paid"}, nil
}
