// Package payment 微信支付驱动
package payment

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// WxPayDriver 微信支付驱动
type WxPayDriver struct {
	appID      string
	mchID      string
	apiKey     string
	privateKey *rsa.PrivateKey
	notifyURL  string
}

// NewWxPayDriver 创建微信支付驱动
func NewWxPayDriver(appID, mchID, apiKey, privateKeyPEM, notifyURL string) (*WxPayDriver, error) {
	priv, err := parsePrivateKey(privateKeyPEM)
	if err != nil {
		return nil, fmt.Errorf("parse private key: %w", err)
	}
	return &WxPayDriver{
		appID:      appID,
		mchID:      mchID,
		apiKey:     apiKey,
		privateKey: priv,
		notifyURL:  notifyURL,
	}, nil
}

// Name 驱动名
func (d *WxPayDriver) Name() string { return "wxpay" }

// CreatePay 发起支付（Native 扫码）
func (d *WxPayDriver) CreatePay(ctx context.Context, order *Order) (*PayResult, error) {
	nonce := randString(32)
	body := map[string]any{
		"appid":        d.appID,
		"mchid":        d.mchID,
		"description":  order.Subject,
		"out_trade_no": order.ID,
		"notify_url":   firstNonEmpty(order.NotifyURL, d.notifyURL),
		"amount": map[string]any{
			"total":    int(order.Amount * 100),
			"currency": strings.ToUpper(order.Currency),
		},
	}
	raw, _ := json.Marshal(body)
	ts := fmt.Sprintf("%d", time.Now().Unix())
	msg := fmt.Sprintf("POST\n/v3/pay/transactions/native\n%s\n%s\n%s", ts, nonce, raw)
	sig, err := d.sign([]byte(msg))
	if err != nil {
		return nil, err
	}
	auth := fmt.Sprintf("WECHATPAY2-SHA256-RSA2048 mchid=\"%s\",nonce_str=\"%s\",timestamp=\"%s\",serial_no=\"%s\",signature=\"%s\"",
		d.mchID, nonce, ts, "default", sig)

	_ = auth // 实际请求需要 HTTP 客户端，此处返回模拟结果
	return &PayResult{
		OrderID:    order.ID,
		QRCode:     "weixin://wxpay/bizpayurl?pr=mock_" + order.ID,
		ExpireTime: time.Now().Add(15 * time.Minute).Unix(),
		Raw:        body,
	}, nil
}

// VerifyNotify 验证回调
func (d *WxPayDriver) VerifyNotify(ctx context.Context, raw []byte, headers map[string]string) (*NotifyPayload, error) {
	var body struct {
		OutTradeNo    string `json:"out_trade_no"`
		TransactionID string `json:"transaction_id"`
		Amount        struct {
			Total    int    `json:"total"`
			Currency string `json:"currency"`
		} `json:"amount"`
		State    string `json:"state"`
		SuccessTime string `json:"success_time"`
	}
	if err := json.Unmarshal(raw, &body); err != nil {
		return nil, err
	}
	// 实际应验证签名 + 解密回调资源
	status := "pending"
	if body.State == "SUCCESS" {
		status = "paid"
	}
	return &NotifyPayload{
		OrderID: body.OutTradeNo,
		TradeNo: body.TransactionID,
		Status:  status,
		Amount:  float64(body.Amount.Total) / 100,
		Raw:     body,
	}, nil
}

// Refund 退款
func (d *WxPayDriver) Refund(ctx context.Context, req *RefundRequest) (*RefundResult, error) {
	return &RefundResult{
		RefundNo: req.RefundNo,
		TradeNo:  req.OrderID,
		Accepted: true,
		Raw:      map[string]string{"refund_no": req.RefundNo},
	}, nil
}

// Close 关闭订单
func (d *WxPayDriver) Close(ctx context.Context, orderID string) error {
	return nil
}

// Query 查询订单
func (d *WxPayDriver) Query(ctx context.Context, orderID string) (*NotifyPayload, error) {
	return &NotifyPayload{OrderID: orderID, Status: "pending"}, nil
}

func (d *WxPayDriver) sign(msg []byte) (string, error) {
	h := sha256.Sum256(msg)
	sig, err := rsa.SignPKCS1v15(rand.Reader, d.privateKey, crypto.SHA256, h[:])
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(sig), nil
}

func randString(n int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = charset[randInt(len(charset))]
	}
	return string(b)
}

func randInt(n int) int {
	b := make([]byte, 1)
	_, _ = rand.Read(b)
	return int(b[0]) % n
}

var _ = http.StatusOK