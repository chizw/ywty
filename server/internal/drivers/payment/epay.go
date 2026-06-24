// Package payment 彩虹聚合支付驱动
package payment

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"
)

// EPayDriver 彩虹聚合支付
type EPayDriver struct {
	pid       string
	key       string
	apiURL    string
	notifyURL string
	returnURL string
	type2name map[string]string
}

// NewEPayDriver 创建 EPay 驱动
func NewEPayDriver(pid, key, apiURL, notifyURL, returnURL string) *EPayDriver {
	return &EPayDriver{
		pid:       pid,
		key:       key,
		apiURL:    apiURL,
		notifyURL: notifyURL,
		returnURL: returnURL,
		type2name: map[string]string{
			"alipay": "alipay",
			"wxpay":  "wxpay",
			"qqpay":  "qqpay",
		},
	}
}

// Name 驱动名
func (d *EPayDriver) Name() string { return "epay" }

// CreatePay 发起支付
func (d *EPayDriver) CreatePay(ctx context.Context, order *Order) (*PayResult, error) {
	typeStr := "alipay"
	if order.Metadata != nil {
		if t, ok := order.Metadata["type"]; ok {
			typeStr = t
		}
	}
	params := map[string]string{
		"pid":          d.pid,
		"type":         typeStr,
		"out_trade_no": order.ID,
		"notify_url":   firstNonEmpty(order.NotifyURL, d.notifyURL),
		"return_url":   firstNonEmpty(order.ReturnURL, d.returnURL),
		"name":         order.Subject,
		"money":        fmt.Sprintf("%.2f", order.Amount),
	}
	sign := d.sign(params)
	params["sign"] = sign
	params["sign_type"] = "MD5"
	q := url.Values{}
	for k, v := range params {
		q.Set(k, v)
	}
	return &PayResult{
		OrderID:    order.ID,
		PayURL:     d.apiURL + "/submit.php?" + q.Encode(),
		ExpireTime: time.Now().Add(15 * time.Minute).Unix(),
		Raw:        params,
	}, nil
}

// VerifyNotify 验证回调
func (d *EPayDriver) VerifyNotify(ctx context.Context, raw []byte, headers map[string]string) (*NotifyPayload, error) {
	values, err := url.ParseQuery(string(raw))
	if err != nil {
		return nil, err
	}
	sign := values.Get("sign")
	values.Del("sign")
	values.Del("sign_type")
	if !d.verify(values, sign) {
		return nil, fmt.Errorf("epay: invalid sign")
	}
	status := "pending"
	if values.Get("trade_status") == "TRADE_SUCCESS" {
		status = "paid"
	}
	amt, _ := parseFloat(values.Get("money"))
	return &NotifyPayload{
		OrderID: values.Get("out_trade_no"),
		TradeNo: values.Get("trade_no"),
		Status:  status,
		Amount:  amt,
		Raw:     values,
	}, nil
}

// Refund 退款
func (d *EPayDriver) Refund(ctx context.Context, req *RefundRequest) (*RefundResult, error) {
	return &RefundResult{RefundNo: req.RefundNo, TradeNo: req.OrderID, Accepted: true}, nil
}

// Close 关闭订单
func (d *EPayDriver) Close(ctx context.Context, orderID string) error { return nil }

// Query 查询订单
func (d *EPayDriver) Query(ctx context.Context, orderID string) (*NotifyPayload, error) {
	return &NotifyPayload{OrderID: orderID, Status: "pending"}, nil
}

func (d *EPayDriver) sign(params map[string]string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		if params[k] == "" {
			continue
		}
		if k == "sign" || k == "sign_type" {
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
	sb.WriteString(d.key)
	sum := md5.Sum([]byte(sb.String()))
	return hex.EncodeToString(sum[:])
}

func (d *EPayDriver) verify(values url.Values, sign string) bool {
	keys := make([]string, 0, len(values))
	for k := range values {
		if values.Get(k) == "" {
			continue
		}
		if k == "sign" || k == "sign_type" {
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
	sb.WriteString(d.key)
	sum := md5.Sum([]byte(sb.String()))
	return strings.EqualFold(hex.EncodeToString(sum[:]), sign)
}
