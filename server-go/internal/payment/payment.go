// Package payment 支付驱动 SPI（对齐 app/Contracts/PaymentAbstract + Drivers/Payment）。
// M4 实现 EPay（RSA-SHA256 签名）；Alipay/WeChat/UniPay/PayPal 在 M5 接入。
package payment

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"
)

// CreateOrderDTO 下单参数（对齐 App\DTOs\CreateOrderDto）。
type CreateOrderDTO struct {
	OutTradeNo string
	Subject    string
	Amount     int64 // 分
	ClientIP   string
}

// Driver 支付驱动接口。
type Driver interface {
	// CreateOrder 发起支付，返回结果（如 {action: jump, url: ...}）。
	CreateOrder(dto CreateOrderDTO, channel, method string) (map[string]any, error)
	// VerifyNotify 校验回调请求。
	VerifyNotify(req *http.Request) (bool, error)
	// VerifyReturnBody 回调成功后的响应体（如 EPay 返回 "success"）。
	VerifyReturnBody() string
}

// New 按提供者创建驱动；config 为 drivers.options JSON + notify_url 等附加项。
func New(provider string, config map[string]any) (Driver, error) {
	switch provider {
	case "epay":
		return NewEPay(config), nil
	case "alipay", "wechat", "unipay", "paypal":
		return nil, fmt.Errorf("支付驱动 %s 尚未接入，请联系管理员", provider)
	}
	return nil, fmt.Errorf("未知的支付提供者: %s", provider)
}

// ---------- 签名工具（对齐 EPayService::getSignContent） ----------

// SignContent 参数排序拼接：跳过数组值、空值、sign/sign_type。
func SignContent(params map[string]any) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	pairs := make([]string, 0, len(keys))
	for _, k := range keys {
		v := params[k]
		if k == "sign" || k == "sign_type" {
			continue
		}
		switch val := v.(type) {
		case map[string]any, []any, nil:
			continue
		case string:
			if strings.TrimSpace(val) == "" {
				continue
			}
			pairs = append(pairs, k+"="+val)
		case float64:
			pairs = append(pairs, k+"="+trimFloat(val))
		default:
			pairs = append(pairs, fmt.Sprintf("%s=%v", k, v))
		}
	}
	return strings.Join(pairs, "&")
}

func trimFloat(f float64) string {
	s := fmt.Sprintf("%f", f)
	s = strings.TrimRight(s, "0")
	s = strings.TrimRight(s, ".")
	return s
}

// ---------- EPay ----------

type EPay struct {
	APIURL             string
	PID                string
	PlatformPublicKey  string
	MerchantPrivateKey string
}

func NewEPay(config map[string]any) *EPay {
	return &EPay{
		APIURL:             strings.TrimRight(strConfig(config, "api_url"), "/"),
		PID:                strConfig(config, "pid"),
		PlatformPublicKey:  strConfig(config, "platform_public_key"),
		MerchantPrivateKey: strConfig(config, "merchant_private_key"),
	}
}

func strConfig(config map[string]any, key string) string {
	if v, ok := config[key].(string); ok {
		return v
	}
	return ""
}

// buildRequestParams 复制参数并附加签名。
func (e *EPay) buildRequestParams(params map[string]any) map[string]any {
	out := make(map[string]any, len(params)+2)
	for k, v := range params {
		out[k] = v
	}
	out["sign"] = e.getSign(out)
	out["sign_type"] = "RSA"
	return out
}

func (e *EPay) getSign(params map[string]any) string {
	dataToSign := SignContent(params)
	return e.rsaPrivateSign(dataToSign)
}

// PayLink 跳转支付链接（getPayLink）。
func (e *EPay) PayLink(params map[string]any) string {
	requestURL := e.APIURL + "/api/pay/submit"
	rp := e.buildRequestParams(params)
	return requestURL + "?" + encodeQuery(rp)
}

// encodeQuery http_build_query 语义（跳过数组/空值——签名时已过滤，但 URL 保留全部标量）。
func encodeQuery(params map[string]any) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	pairs := make([]string, 0, len(keys))
	for _, k := range keys {
		switch v := params[k].(type) {
		case string:
			pairs = append(pairs, urlEscape(k)+"="+urlEscape(v))
		case float64:
			pairs = append(pairs, urlEscape(k)+"="+urlEscape(trimFloat(v)))
		default:
			pairs = append(pairs, urlEscape(k)+"="+urlEscape(fmt.Sprintf("%v", v)))
		}
	}
	return strings.Join(pairs, "&")
}

func urlEscape(s string) string {
	replacer := strings.NewReplacer(
		"%", "%25", "&", "%26", "=", "%3D", "+", "%2B", " ", "%20", "?", "%3F", "#", "%23",
	)
	return replacer.Replace(s)
}

// rsaPrivateSign 商户私钥 RSA-SHA256 签名（PKCS1v15）。
func (e *EPay) rsaPrivateSign(data string) string {
	key, err := parsePrivateKey(e.MerchantPrivateKey)
	if err != nil {
		return ""
	}
	sum := sha256.Sum256([]byte(data))
	sig, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, sum[:])
	if err != nil {
		return ""
	}
	return base64.StdEncoding.EncodeToString(sig)
}

// Verify 平台公钥验签（verify）。
func (e *EPay) Verify(data map[string]any) bool {
	sign, _ := data["sign"].(string)
	if len(data) == 0 || sign == "" {
		return false
	}
	pub, err := parsePublicKey(e.PlatformPublicKey)
	if err != nil {
		return false
	}
	sum := sha256.Sum256([]byte(SignContent(data)))
	sig, err := base64.StdEncoding.DecodeString(sign)
	if err != nil {
		return false
	}
	return rsa.VerifyPKCS1v15(pub, crypto.SHA256, sum[:], sig) == nil
}

// CreateOrder 实现 Driver（对齐 EPayPayment::createOrder）。
func (e *EPay) CreateOrder(dto CreateOrderDTO, channel, method string) (map[string]any, error) {
	methods := map[string]string{"web": "jump", "mp": "jsapi", "mini": "applet"}
	if m, ok := methods[method]; ok {
		method = m
	}
	order := map[string]any{
		"out_trade_no": dto.OutTradeNo,
		"name":         dto.Subject,
		"money":        fmt.Sprintf("%.2f", float64(dto.Amount)/100),
	}
	if channel == "unified" {
		order["type"] = channel
		url := e.PayLink(order)
		return map[string]any{"action": "jump", "url": url}, nil
	}
	order["method"] = method
	order["type"] = channel
	order["clientip"] = dto.ClientIP
	result, err := e.apiPay(order)
	if err != nil {
		return nil, err
	}
	payType, _ := result["pay_type"].(string)
	payInfo, _ := result["pay_info"].(string)
	if payType == "jump" {
		return map[string]any{"action": "jump", "url": payInfo}, nil
	}
	return map[string]any{"action": payType, "content": payInfo}, nil
}

// apiPay 调用 api/pay/create（execute）。
func (e *EPay) apiPay(params map[string]any) (map[string]any, error) {
	client := &http.Client{Timeout: 15 * time.Second}
	rp := e.buildRequestParams(params)
	form := encodeQuery(rp)
	req, err := http.NewRequest(http.MethodPost, e.APIURL+"/api/pay/create",
		strings.NewReader(form))
	if err != nil {
		return nil, errors.New("请求外部支付服务失败")
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.New("请求外部支付服务失败")
	}
	defer func() { _ = resp.Body.Close() }()
	var data map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, errors.New("请求失败")
	}
	if code, _ := data["code"].(float64); code == 0 {
		if !e.Verify(data) {
			return nil, errors.New("返回数据验签失败")
		}
		return data, nil
	}
	msg, _ := data["msg"].(string)
	if msg == "" {
		msg = "请求失败"
	}
	return nil, errors.New(msg)
}

// VerifyNotify 实现 Driver（对齐 EPayPayment::verifyNotify）。
func (e *EPay) VerifyNotify(req *http.Request) (bool, error) {
	if err := req.ParseForm(); err != nil {
		return false, nil
	}
	data := map[string]any{}
	for k, v := range req.Form {
		if len(v) == 1 {
			data[k] = v[0]
		}
	}
	if !e.Verify(data) {
		return false, nil
	}
	if data["trade_status"] != "TRADE_SUCCESS" {
		return false, nil
	}
	return true, nil
}

// VerifyReturnBody 实现 Driver。
func (e *EPay) VerifyReturnBody() string { return "success" }

// ---------- 密钥解析 ----------

func parsePrivateKey(key string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(formatPEM(key, "PRIVATE KEY")))
	if block != nil {
		if k, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
			return k, nil
		}
		if k8, err := x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
			if rk, ok := k8.(*rsa.PrivateKey); ok {
				return rk, nil
			}
		}
	}
	return nil, errors.New("私钥格式错误")
}

func parsePublicKey(key string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(formatPEM(key, "PUBLIC KEY")))
	if block == nil {
		return nil, errors.New("公钥格式错误")
	}
	if k, err := x509.ParsePKIXPublicKey(block.Bytes); err == nil {
		if rsaKey, ok := k.(*rsa.PublicKey); ok {
			return rsaKey, nil
		}
		return nil, errors.New("公钥不是 RSA 类型")
	}
	return x509.ParsePKCS1PublicKey(block.Bytes)
}

// formatPEM 补全 PEM 头尾与换行（PHP formatPrivateKey/formatPublicKey 等价）。
func formatPEM(key, typ string) string {
	if strings.Contains(key, "-----BEGIN") {
		return key
	}
	clean := strings.NewReplacer("\r", "", "\n", "", " ", "").Replace(key)
	var b strings.Builder
	b.WriteString("-----BEGIN " + typ + "-----\n")
	for len(clean) > 64 {
		b.WriteString(clean[:64] + "\n")
		clean = clean[64:]
	}
	b.WriteString(clean + "\n-----END " + typ + "-----\n")
	return b.String()
}
