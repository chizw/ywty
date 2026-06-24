// Package payment 支付驱动抽象层
package payment

import (
	"context"
	"fmt"
	"net/http"

	"github.com/ywty/server/internal/model"
)

// CreateResult 发起支付返回结果
type CreateResult struct {
	Type string `json:"type"` // form / qrcode / url / app
	Data string `json:"data"` // form html / qrcode url / 跳转 url / app 参数
}

// CallbackResult 回调验签结果
type CallbackResult struct {
	TradeNo    string `json:"trade_no"`     // 系统订单号
	OutTradeNo string `json:"out_trade_no"` // 第三方订单号
	Paid       bool   `json:"paid"`         // 是否支付成功
	Amount     uint   `json:"amount"`       // 实际支付金额（分）
}

// QueryResult 主动查询订单状态结果
type QueryResult struct {
	Paid   bool `json:"paid"`
	Amount uint `json:"amount"`
}

// Driver 支付驱动接口
type Driver interface {
	Name() string
	// CreatePayment 发起支付，返回支付参数
	CreatePayment(ctx context.Context, order model.Order, cfg map[string]string) (*CreateResult, error)
	// VerifyCallback 验签并解析回调请求
	VerifyCallback(ctx context.Context, req *http.Request, cfg map[string]string) (*CallbackResult, error)
	// Query 主动查询订单状态
	Query(ctx context.Context, order model.Order, cfg map[string]string) (*QueryResult, error)
}

// StringMap 将 map[string]interface{} 配置项转换为 map[string]string
func StringMap(opts map[string]interface{}) map[string]string {
	if opts == nil {
		return nil
	}
	res := make(map[string]string, len(opts))
	for k, v := range opts {
		res[k] = fmt.Sprint(v)
	}
	return res
}
