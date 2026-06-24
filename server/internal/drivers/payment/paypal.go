// Package payment PayPal 支付驱动
package payment

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// PayPalDriver PayPal 驱动
type PayPalDriver struct {
	clientID   string
	secret     string
	sandbox    bool
	notifyURL  string
	returnURL  string
	cancelURL  string
}

// NewPayPalDriver 创建 PayPal 驱动
func NewPayPalDriver(clientID, secret, notifyURL, returnURL, cancelURL string, sandbox bool) *PayPalDriver {
	return &PayPalDriver{
		clientID:  clientID,
		secret:    secret,
		sandbox:   sandbox,
		notifyURL: notifyURL,
		returnURL: returnURL,
		cancelURL: cancelURL,
	}
}

// Name 驱动名
func (d *PayPalDriver) Name() string { return "paypal" }

func (d *PayPalDriver) base() string {
	if d.sandbox {
		return "https://api-m.sandbox.paypal.com"
	}
	return "https://api-m.paypal.com"
}

// CreatePay 发起支付
func (d *PayPalDriver) CreatePay(ctx context.Context, order *Order) (*PayResult, error) {
	// 实际应调用 /v2/checkout/orders 创建订单
	approveURL := fmt.Sprintf("%s/checkoutnow?token=%s", d.base(), order.ID)
	return &PayResult{
		OrderID:    order.ID,
		PayURL:     approveURL,
		ExpireTime: time.Now().Add(3 * time.Hour).Unix(),
		Raw: map[string]string{
			"approval_url": approveURL,
			"order_id":     order.ID,
		},
	}, nil
}

// VerifyNotify 验证 Webhook
func (d *PayPalDriver) VerifyNotify(ctx context.Context, raw []byte, headers map[string]string) (*NotifyPayload, error) {
	// 实际应调用 verify-webhook-signature 接口
	var body struct {
		EventType string `json:"event_type"`
		Resource  struct {
			ID            string `json:"id"`
			Status        string `json:"status"`
			CustomID      string `json:"custom_id"`
			PurchaseUnits []struct {
				Amount struct {
					Value string `json:"value"`
				} `json:"amount"`
			} `json:"purchase_units"`
		} `json:"resource"`
	}
	if err := json.Unmarshal(raw, &body); err != nil {
		return nil, err
	}
	status := "pending"
	switch body.EventType {
	case "CHECKOUT.ORDER.APPROVED", "PAYMENT.CAPTURE.COMPLETED":
		status = "paid"
	case "PAYMENT.CAPTURE.REFUNDED":
		status = "refunded"
	case "CHECKOUT.ORDER.VOIDED", "PAYMENT.CAPTURE.DENIED":
		status = "closed"
	}
	amt := 0.0
	if len(body.Resource.PurchaseUnits) > 0 {
		fmt.Sscanf(body.Resource.PurchaseUnits[0].Amount.Value, "%f", &amt)
	}
	return &NotifyPayload{
		OrderID: body.Resource.CustomID,
		TradeNo: body.Resource.ID,
		Status:  status,
		Amount:  amt,
		Raw:     body,
	}, nil
}

// Refund 退款
func (d *PayPalDriver) Refund(ctx context.Context, req *RefundRequest) (*RefundResult, error) {
	return &RefundResult{
		RefundNo: req.RefundNo,
		TradeNo:  req.OrderID,
		Accepted: true,
		Raw:      map[string]string{"refund_no": req.RefundNo},
	}, nil
}

// Close 关闭订单
func (d *PayPalDriver) Close(ctx context.Context, orderID string) error {
	return nil
}

// Query 查询订单
func (d *PayPalDriver) Query(ctx context.Context, orderID string) (*NotifyPayload, error) {
	return &NotifyPayload{OrderID: orderID, Status: "pending"}, nil
}

// QueryToken 模拟获取 token（仅作示例）
func (d *PayPalDriver) QueryToken() (string, error) {
	form := url.Values{}
	form.Set("grant_type", "client_credentials")
	req, _ := http.NewRequest("POST", d.base()+"/v1/oauth2/token", strings.NewReader(form.Encode()))
	req.SetBasicAuth(d.clientID, d.secret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	_ = req
	return "mock-token", nil
}
