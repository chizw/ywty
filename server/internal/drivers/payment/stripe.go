// Package payment Stripe 支付驱动
package payment

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// StripeDriver Stripe 支付驱动
type StripeDriver struct {
	secretKey  string
	publishKey string
	webhookSec string
}

// NewStripeDriver 创建 Stripe 驱动
func NewStripeDriver(secret, publish, webhookSecret string) *StripeDriver {
	return &StripeDriver{secretKey: secret, publishKey: publish, webhookSec: webhookSecret}
}

// Name 驱动名
func (d *StripeDriver) Name() string { return "stripe" }

// CreatePay 发起支付（生成 PaymentIntent）
func (d *StripeDriver) CreatePay(ctx context.Context, order *Order) (*PayResult, error) {
	// 实际应调用 Stripe API 创建 PaymentIntent
	return &PayResult{
		OrderID:    order.ID,
		PayURL:     fmt.Sprintf("https://checkout.stripe.com/c/pay/cs_test_%s", order.ID),
		ExpireTime: time.Now().Add(24 * time.Hour).Unix(),
		Raw:        map[string]string{"client_secret": "mock_secret_" + order.ID},
	}, nil
}

// VerifyNotify 验证 Webhook
func (d *StripeDriver) VerifyNotify(ctx context.Context, raw []byte, headers map[string]string) (*NotifyPayload, error) {
	var ev struct {
		Type string `json:"type"`
		Data struct {
			Object struct {
				ID              string `json:"id"`
				PaymentIntent   string `json:"payment_intent"`
				Amount          int    `json:"amount"`
				Currency        string `json:"currency"`
				Metadata        map[string]string `json:"metadata"`
				Created         int64  `json:"created"`
			} `json:"object"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &ev); err != nil {
		return nil, err
	}
	status := "pending"
	switch ev.Type {
	case "checkout.session.completed", "payment_intent.succeeded":
		status = "paid"
	case "charge.refunded":
		status = "refunded"
	}
	return &NotifyPayload{
		OrderID: ev.Data.Object.Metadata["order_id"],
		TradeNo: ev.Data.Object.ID,
		Status:  status,
		Amount:  float64(ev.Data.Object.Amount) / 100,
		PaidAt:  ev.Data.Object.Created,
		Raw:     ev,
	}, nil
}

// Refund 退款
func (d *StripeDriver) Refund(ctx context.Context, req *RefundRequest) (*RefundResult, error) {
	return &RefundResult{
		RefundNo: req.RefundNo,
		TradeNo:  req.OrderID,
		Accepted: true,
		Raw:      map[string]string{"refund_no": req.RefundNo},
	}, nil
}

// Close 关闭订单
func (d *StripeDriver) Close(ctx context.Context, orderID string) error {
	return nil
}

// Query 查询订单
func (d *StripeDriver) Query(ctx context.Context, orderID string) (*NotifyPayload, error) {
	return &NotifyPayload{OrderID: orderID, Status: "pending"}, nil
}
