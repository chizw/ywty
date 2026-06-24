package payment

import (
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/ywty/server/internal/model"
)

// firstNonEmpty 返回第一个非空字符串
func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

// parseAmount 将元字符串转换为分
func parseAmount(s string) uint {
	if s == "" {
		return 0
	}
	var f float64
	_, _ = fmt.Sscanf(s, "%f", &f)
	return uint(f * 100)
}

// parseAmountFen 将分字符串转换为分
func parseAmountFen(s string) uint {
	if s == "" {
		return 0
	}
	var n uint
	_, _ = fmt.Sscanf(s, "%d", &n)
	return n
}

// orderSubject 从订单或配置中提取商品标题
func orderSubject(order model.Order, cfg map[string]string) string {
	if cfg != nil && cfg["subject"] != "" {
		return cfg["subject"]
	}
	if planSnap, ok := order.Snapshot["plan"].(map[string]any); ok {
		if name, ok := planSnap["name"].(string); ok && name != "" {
			return name
		}
	}
	return "订单支付"
}

// buildForm 构造自动提交的 HTML form
func buildForm(action string, params url.Values) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf(`<form action="%s" method="POST" id="payment_form">`, action))
	for k, vs := range params {
		for _, v := range vs {
			b.WriteString(fmt.Sprintf(`<input type="hidden" name="%s" value="%s" />`, k, v))
		}
	}
	b.WriteString(`</form><script>document.getElementById("payment_form").submit();</script>`)
	return b.String()
}

// epaySign 简易签名（仅用于占位演示，生产请使用真实签算）
func epaySign(params url.Values, key string) string {
	h := md5.New()
	_, _ = io.WriteString(h, params.Encode()+key)
	return fmt.Sprintf("%x", h.Sum(nil))
}

// --- log 驱动（测试用）---

type logDriver struct{}

func newLogDriver(_ map[string]string) (Driver, error) { return &logDriver{}, nil }
func (d *logDriver) Name() string                       { return "log" }
func (d *logDriver) CreatePayment(ctx context.Context, order model.Order, cfg map[string]string) (*CreateResult, error) {
	returnURL := firstNonEmpty(cfg["return_url"], "/orders/result")
	data := fmt.Sprintf("%s?trade_no=%s&log_paid=1", returnURL, order.TradeNo)
	return &CreateResult{Type: "url", Data: data}, nil
}
func (d *logDriver) VerifyCallback(ctx context.Context, req *http.Request, cfg map[string]string) (*CallbackResult, error) {
	_ = req.ParseForm()
	tradeNo := firstNonEmpty(req.FormValue("trade_no"), req.URL.Query().Get("trade_no"))
	paid := req.FormValue("log_paid") == "1" || req.URL.Query().Get("log_paid") == "1"
	if !paid {
		return &CallbackResult{Paid: false}, nil
	}
	return &CallbackResult{TradeNo: tradeNo, Paid: true, Amount: 0}, nil
}
func (d *logDriver) Query(ctx context.Context, order model.Order, cfg map[string]string) (*QueryResult, error) {
	return &QueryResult{Paid: true, Amount: order.Amount}, nil
}

// --- alipay 驱动 ---

type alipayDriver struct {
	appID     string
	gateway   string
	returnURL string
	notifyURL string
}

func newAlipayDriver(cfg map[string]string) (Driver, error) {
	return &alipayDriver{
		appID:     cfg["app_id"],
		gateway:   cfg["gateway"],
		returnURL: cfg["return_url"],
		notifyURL: cfg["notify_url"],
	}, nil
}
func (d *alipayDriver) Name() string { return "alipay" }
func (d *alipayDriver) CreatePayment(ctx context.Context, order model.Order, cfg map[string]string) (*CreateResult, error) {
	gateway := firstNonEmpty(cfg["gateway"], d.gateway, "https://openapi.alipay.com/gateway.do")
	returnURL := firstNonEmpty(cfg["return_url"], d.returnURL)
	notifyURL := firstNonEmpty(cfg["notify_url"], d.notifyURL)
	appID := firstNonEmpty(cfg["app_id"], d.appID, "demo")

	params := url.Values{}
	params.Set("app_id", appID)
	params.Set("method", "alipay.trade.page.pay")
	params.Set("out_trade_no", order.TradeNo)
	params.Set("total_amount", fmt.Sprintf("%.2f", float64(order.Amount)/100))
	params.Set("subject", orderSubject(order, cfg))
	params.Set("return_url", returnURL)
	params.Set("notify_url", notifyURL)
	params.Set("sign", "demo-sign")

	return &CreateResult{Type: "url", Data: gateway + "?" + params.Encode()}, nil
}
func (d *alipayDriver) VerifyCallback(ctx context.Context, req *http.Request, cfg map[string]string) (*CallbackResult, error) {
	_ = req.ParseForm()
	sign := req.FormValue("sign")
	if sign == "" {
		return &CallbackResult{Paid: false}, nil
	}
	return &CallbackResult{
		TradeNo:    req.FormValue("out_trade_no"),
		OutTradeNo: req.FormValue("trade_no"),
		Paid:       true,
		Amount:     parseAmount(req.FormValue("total_amount")),
	}, nil
}
func (d *alipayDriver) Query(ctx context.Context, order model.Order, cfg map[string]string) (*QueryResult, error) {
	return &QueryResult{Paid: false, Amount: order.Amount}, nil
}

// --- wechat 驱动 ---

type wechatDriver struct {
	appID     string
	mchID     string
	apiKey    string
	notifyURL string
}

func newWechatDriver(cfg map[string]string) (Driver, error) {
	return &wechatDriver{
		appID:     cfg["app_id"],
		mchID:     cfg["mch_id"],
		apiKey:    cfg["api_key"],
		notifyURL: cfg["notify_url"],
	}, nil
}
func (d *wechatDriver) Name() string { return "wechat" }
func (d *wechatDriver) CreatePayment(ctx context.Context, order model.Order, cfg map[string]string) (*CreateResult, error) {
	appID := firstNonEmpty(cfg["app_id"], d.appID, "demo")
	notifyURL := firstNonEmpty(cfg["notify_url"], d.notifyURL)
	codeURL := fmt.Sprintf("weixin://wxpay/bizpayurl?pr=%s&appid=%s&notify=%s", order.TradeNo, appID, url.QueryEscape(notifyURL))
	return &CreateResult{Type: "qrcode", Data: codeURL}, nil
}
func (d *wechatDriver) VerifyCallback(ctx context.Context, req *http.Request, cfg map[string]string) (*CallbackResult, error) {
	_ = req.ParseForm()
	if req.FormValue("return_code") == "SUCCESS" || req.URL.Query().Get("return_code") == "SUCCESS" {
		return &CallbackResult{
			TradeNo:    req.FormValue("out_trade_no"),
			OutTradeNo: req.FormValue("transaction_id"),
			Paid:       true,
			Amount:     parseAmountFen(req.FormValue("total_fee")),
		}, nil
	}
	return &CallbackResult{Paid: false}, nil
}
func (d *wechatDriver) Query(ctx context.Context, order model.Order, cfg map[string]string) (*QueryResult, error) {
	return &QueryResult{Paid: false, Amount: order.Amount}, nil
}

// --- paypal 驱动 ---

type paypalDriver struct {
	clientID  string
	secret    string
	returnURL string
	cancelURL string
	sandbox   bool
}

func newPaypalDriver(cfg map[string]string) (Driver, error) {
	return &paypalDriver{
		clientID:  cfg["client_id"],
		secret:    cfg["secret"],
		returnURL: cfg["return_url"],
		cancelURL: cfg["cancel_url"],
		sandbox:   cfg["sandbox"] == "true",
	}, nil
}
func (d *paypalDriver) Name() string { return "paypal" }
func (d *paypalDriver) CreatePayment(ctx context.Context, order model.Order, cfg map[string]string) (*CreateResult, error) {
	base := "https://www.paypal.com"
	if d.sandbox || cfg["sandbox"] == "true" {
		base = "https://www.sandbox.paypal.com"
	}
	returnURL := firstNonEmpty(cfg["return_url"], d.returnURL)
	cancelURL := firstNonEmpty(cfg["cancel_url"], d.cancelURL)
	data := fmt.Sprintf("%s/checkoutnow?token=%s&return_url=%s&cancel_url=%s", base, order.TradeNo, url.QueryEscape(returnURL), url.QueryEscape(cancelURL))
	return &CreateResult{Type: "url", Data: data}, nil
}
func (d *paypalDriver) VerifyCallback(ctx context.Context, req *http.Request, cfg map[string]string) (*CallbackResult, error) {
	payerID := req.URL.Query().Get("PayerID")
	token := req.URL.Query().Get("token")
	if payerID == "" {
		return &CallbackResult{Paid: false}, nil
	}
	return &CallbackResult{TradeNo: token, OutTradeNo: payerID, Paid: true, Amount: 0}, nil
}
func (d *paypalDriver) Query(ctx context.Context, order model.Order, cfg map[string]string) (*QueryResult, error) {
	return &QueryResult{Paid: false, Amount: order.Amount}, nil
}

// --- epay 聚合支付驱动 ---

type epayDriver struct {
	apiURL    string
	pid       string
	key       string
	returnURL string
	notifyURL string
}

func newEpayDriver(cfg map[string]string) (Driver, error) {
	return &epayDriver{
		apiURL:    cfg["api_url"],
		pid:       cfg["pid"],
		key:       cfg["key"],
		returnURL: cfg["return_url"],
		notifyURL: cfg["notify_url"],
	}, nil
}
func (d *epayDriver) Name() string { return "epay" }
func (d *epayDriver) CreatePayment(ctx context.Context, order model.Order, cfg map[string]string) (*CreateResult, error) {
	apiURL := firstNonEmpty(cfg["api_url"], d.apiURL, "https://epay.example.com/submit.php")
	returnURL := firstNonEmpty(cfg["return_url"], d.returnURL)
	notifyURL := firstNonEmpty(cfg["notify_url"], d.notifyURL)
	pid := firstNonEmpty(cfg["pid"], d.pid, "demo")
	key := firstNonEmpty(cfg["key"], d.key, "demo-key")

	params := url.Values{}
	params.Set("pid", pid)
	params.Set("out_trade_no", order.TradeNo)
	params.Set("name", orderSubject(order, cfg))
	params.Set("money", fmt.Sprintf("%.2f", float64(order.Amount)/100))
	params.Set("notify_url", notifyURL)
	params.Set("return_url", returnURL)
	params.Set("sign", epaySign(params, key))
	params.Set("sign_type", "MD5")

	return &CreateResult{Type: "form", Data: buildForm(apiURL, params)}, nil
}
func (d *epayDriver) VerifyCallback(ctx context.Context, req *http.Request, cfg map[string]string) (*CallbackResult, error) {
	_ = req.ParseForm()
	if req.FormValue("sign") == "" {
		return &CallbackResult{Paid: false}, nil
	}
	return &CallbackResult{
		TradeNo:    req.FormValue("out_trade_no"),
		OutTradeNo: req.FormValue("trade_no"),
		Paid:       true,
		Amount:     parseAmount(req.FormValue("money")),
	}, nil
}
func (d *epayDriver) Query(ctx context.Context, order model.Order, cfg map[string]string) (*QueryResult, error) {
	return &QueryResult{Paid: false, Amount: order.Amount}, nil
}

// --- stripe 驱动 ---

type stripeDriver struct {
	secretKey  string
	publicKey  string
	successURL string
	cancelURL  string
}

func newStripeDriver(cfg map[string]string) (Driver, error) {
	return &stripeDriver{
		secretKey:  cfg["secret_key"],
		publicKey:  cfg["public_key"],
		successURL: cfg["success_url"],
		cancelURL:  cfg["cancel_url"],
	}, nil
}
func (d *stripeDriver) Name() string { return "stripe" }
func (d *stripeDriver) CreatePayment(ctx context.Context, order model.Order, cfg map[string]string) (*CreateResult, error) {
	publicKey := firstNonEmpty(cfg["public_key"], d.publicKey, "pk_demo")
	successURL := firstNonEmpty(cfg["success_url"], d.successURL)
	cancelURL := firstNonEmpty(cfg["cancel_url"], d.cancelURL)
	data := fmt.Sprintf("https://checkout.stripe.com/pay/cs_%s?pk=%s&success=%s&cancel=%s",
		order.TradeNo, publicKey, url.QueryEscape(successURL), url.QueryEscape(cancelURL))
	return &CreateResult{Type: "url", Data: data}, nil
}
func (d *stripeDriver) VerifyCallback(ctx context.Context, req *http.Request, cfg map[string]string) (*CallbackResult, error) {
	sessionID := req.URL.Query().Get("session_id")
	if sessionID == "" {
		return &CallbackResult{Paid: false}, nil
	}
	return &CallbackResult{OutTradeNo: sessionID, Paid: true, Amount: 0}, nil
}
func (d *stripeDriver) Query(ctx context.Context, order model.Order, cfg map[string]string) (*QueryResult, error) {
	return &QueryResult{Paid: false, Amount: order.Amount}, nil
}

func init() {
	Register("log", newLogDriver)
	Register("alipay", newAlipayDriver)
	Register("wechat", newWechatDriver)
	Register("paypal", newPaypalDriver)
	Register("epay", newEpayDriver)
	Register("stripe", newStripeDriver)
}
