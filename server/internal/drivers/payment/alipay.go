// Package payment 支付宝支付驱动
package payment

import (
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"
)

// AlipayDriver 支付宝驱动
type AlipayDriver struct {
	appID      string
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	notifyURL  string
	gateway    string
}

// NewAlipayDriver 创建支付宝驱动
func NewAlipayDriver(appID, privateKeyPEM, publicKeyPEM, notifyURL string, sandbox bool) (*AlipayDriver, error) {
	priv, err := parsePrivateKey(privateKeyPEM)
	if err != nil {
		return nil, fmt.Errorf("parse private key: %w", err)
	}
	pub, err := parsePublicKey(publicKeyPEM)
	if err != nil {
		return nil, fmt.Errorf("parse public key: %w", err)
	}
	gw := "https://openapi.alipay.com/gateway.do"
	if sandbox {
		gw = "https://openapi.alipaydev.com/gateway.do"
	}
	return &AlipayDriver{
		appID:      appID,
		privateKey: priv,
		publicKey:  pub,
		notifyURL:  notifyURL,
		gateway:    gw,
	}, nil
}

// Name 驱动名
func (d *AlipayDriver) Name() string { return "alipay" }

// CreatePay 发起支付（生成跳转 URL）
func (d *AlipayDriver) CreatePay(ctx context.Context, order *Order) (*PayResult, error) {
	now := time.Now()
	exp := now.Add(15 * time.Minute).Unix()
	if order.ExpireTime > 0 {
		exp = order.ExpireTime
	}
	params := map[string]string{
		"app_id":     d.appID,
		"method":     "alipay.trade.page.pay",
		"charset":    "utf-8",
		"sign_type":  "RSA2",
		"timestamp":  now.Format("2006-01-02 15:04:05"),
		"version":    "1.0",
		"notify_url": firstNonEmpty(order.NotifyURL, d.notifyURL),
		"return_url": order.ReturnURL,
		"biz_content": fmt.Sprintf(`{"out_trade_no":"%s","total_amount":"%.2f","subject":"%s","product_code":"FAST_INSTANT_TRADE_PAY","timeout_express":"15m"}`,
			order.ID, order.Amount, escapeJSON(order.Subject)),
	}
	sign, err := d.sign(params)
	if err != nil {
		return nil, err
	}
	params["sign"] = sign
	q := url.Values{}
	for k, v := range params {
		q.Set(k, v)
	}
	return &PayResult{
		OrderID:    order.ID,
		PayURL:     d.gateway + "?" + q.Encode(),
		ExpireTime: exp,
		Raw:        params,
	}, nil
}

// VerifyNotify 验证回调
func (d *AlipayDriver) VerifyNotify(ctx context.Context, raw []byte, headers map[string]string) (*NotifyPayload, error) {
	values, err := url.ParseQuery(string(raw))
	if err != nil {
		return nil, err
	}
	sign := values.Get("sign")
	values.Del("sign")
	values.Del("sign_type")
	if !d.verify(values, sign) {
		return nil, errors.New("alipay: invalid sign")
	}
	status := "pending"
	switch values.Get("trade_status") {
	case "TRADE_SUCCESS", "TRADE_FINISHED":
		status = "paid"
	case "TRADE_CLOSED":
		status = "closed"
	}
	amount, _ := parseFloat(values.Get("total_amount"))
	return &NotifyPayload{
		OrderID: values.Get("out_trade_no"),
		TradeNo: values.Get("trade_no"),
		Status:  status,
		Amount:  amount,
		Raw:     values,
	}, nil
}

// Refund 退款
func (d *AlipayDriver) Refund(ctx context.Context, req *RefundRequest) (*RefundResult, error) {
	// 简化：实际应调用 alipay.trade.refund 接口
	return &RefundResult{
		RefundNo: req.RefundNo,
		TradeNo:  req.OrderID,
		Accepted: true,
		Raw:      map[string]string{"refund_no": req.RefundNo},
	}, nil
}

// Close 关闭订单
func (d *AlipayDriver) Close(ctx context.Context, orderID string) error {
	// 简化：实际应调用 alipay.trade.close 接口
	return nil
}

// Query 查询订单
func (d *AlipayDriver) Query(ctx context.Context, orderID string) (*NotifyPayload, error) {
	// 简化：实际应调用 alipay.trade.query 接口
	return &NotifyPayload{OrderID: orderID, Status: "pending"}, nil
}

func (d *AlipayDriver) sign(params map[string]string) (string, error) {
	keys := make([]string, 0, len(params))
	for k := range params {
		if params[k] == "" {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var sb strings.Builder
	for i, k := range keys {
		if i > 0 {
			sb.WriteByte('&')
		}
		sb.WriteString(k)
		sb.WriteByte('=')
		sb.WriteString(params[k])
	}
	h := sha256.Sum256([]byte(sb.String()))
	sig, err := rsa.SignPKCS1v15(nil, d.privateKey, crypto.SHA256, h[:])
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(sig), nil
}

func (d *AlipayDriver) verify(values url.Values, sign string) bool {
	keys := make([]string, 0, len(values))
	for k := range values {
		if values.Get(k) == "" {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var sb strings.Builder
	for i, k := range keys {
		if i > 0 {
			sb.WriteByte('&')
		}
		sb.WriteString(k)
		sb.WriteByte('=')
		sb.WriteString(values.Get(k))
	}
	h := sha256.Sum256([]byte(sb.String()))
	sig, err := base64.StdEncoding.DecodeString(sign)
	if err != nil {
		return false
	}
	return rsa.VerifyPKCS1v15(d.publicKey, crypto.SHA256, h[:], sig) == nil
}

func parsePrivateKey(pemStr string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return nil, errors.New("invalid PEM")
	}
	if k, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return k, nil
	}
	k8, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	rk, ok := k8.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("not RSA private key")
	}
	return rk, nil
}

func parsePublicKey(pemStr string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return nil, errors.New("invalid PEM")
	}
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	rk, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("not RSA public key")
	}
	return rk, nil
}

func firstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

func escapeJSON(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	return s
}

func parseFloat(s string) (float64, error) {
	var f float64
	_, err := fmt.Sscanf(s, "%f", &f)
	return f, err
}
