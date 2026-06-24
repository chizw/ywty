// Package notify SMS 驱动集合
package notify

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// AliyunSMSer 阿里云短信
type AliyunSMSer struct {
	AccessKey string
	SecretKey string
	Sign      string
	Region    string
}

func NewAliyunSMSer(ak, sk, sign, region string) *AliyunSMSer {
	if region == "" {
		region = "cn-hangzhou"
	}
	return &AliyunSMSer{AccessKey: ak, SecretKey: sk, Sign: sign, Region: region}
}

func (a *AliyunSMSer) Name() string { return "aliyun" }

func (a *AliyunSMSer) Send(ctx context.Context, s SMS) error {
	// 公共参数
	params := url.Values{}
	params.Set("AccessKeyId", a.AccessKey)
	params.Set("Timestamp", time.Now().UTC().Format("2006-01-02T15:04:05Z"))
	params.Set("Format", "JSON")
	params.Set("SignatureMethod", "HMAC-SHA1")
	params.Set("SignatureVersion", "1.0")
	params.Set("SignatureNonce", strconv.FormatInt(time.Now().UnixNano(), 10))
	// 业务参数
	params.Set("Action", "SendSms")
	params.Set("Version", "2017-05-25")
	params.Set("PhoneNumbers", s.To)
	params.Set("SignName", a.Sign)
	// TemplateCode 应从 sms.Body 中解析约定格式 "TPL:code|vars"
	tpl, vars := parseSMSBody(s.Body)
	params.Set("TemplateCode", tpl)
	params.Set("TemplateParam", vars)
	// 排序 + 签名
	sign := a.signParams(params)
	// 完整 URL
	apiURL := "https://dysmsapi.aliyuncs.com/?Signature=" + sign + "&" + params.Encode()
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return fmt.Errorf("aliyun sms: %s", string(body))
	}
	return nil
}

func (a *AliyunSMSer) signParams(params url.Values) string {
	// 1. 排序
	var keys []string
	for k := range params {
		keys = append(keys, k)
	}
	// sort.Strings 不可用：手写快排
	quickSort(keys, 0, len(keys)-1)
	// 2. 拼接
	var sb strings.Builder
	sb.WriteString("GET&%2F&")
	encoded := make([]string, 0, len(keys))
	for _, k := range keys {
		encoded = append(encoded, url.QueryEscape(k)+"="+url.QueryEscape(params.Get(k)))
	}
	sb.WriteString(url.QueryEscape(strings.Join(encoded, "&")))
	// 3. HMAC-SHA1
	mac := hmac.New(sha1.New, []byte(a.SecretKey+"&"))
	mac.Write([]byte(sb.String()))
	return url.QueryEscape(base64.StdEncoding.EncodeToString(mac.Sum(nil)))
}

func quickSort(arr []string, lo, hi int) {
	if lo >= hi {
		return
	}
	p := arr[lo]
	i, j := lo, hi
	for i < j {
		for i < j && arr[j] >= p {
			j--
		}
		if i < j {
			arr[i] = arr[j]
			i++
		}
		for i < j && arr[i] <= p {
			i++
		}
		if i < j {
			arr[j] = arr[i]
			j--
		}
	}
	arr[i] = p
	quickSort(arr, lo, i-1)
	quickSort(arr, i+1, hi)
}

// parseSMSBody 把 "TPL:SMS_xxx|{...}" 解析为 tpl + json vars
func parseSMSBody(body string) (string, string) {
	if !strings.HasPrefix(body, "TPL:") {
		return "", "{}"
	}
	rest := body[4:]
	parts := strings.SplitN(rest, "|", 2)
	tpl := parts[0]
	if len(parts) == 1 {
		return tpl, "{}"
	}
	vars := parts[1]
	// 尝试 JSON 化：若已经是 JSON 直接用，否则包装
	if strings.HasPrefix(vars, "{") {
		return tpl, vars
	}
	wrapped := `{"code":"` + vars + `"}`
	return tpl, wrapped
}

// ============================================================
// Tencent SMS
// ============================================================

// TencentSMSer 腾讯云短信
type TencentSMSer struct {
	SecretID  string
	SecretKey string
	AppID     string
	Sign      string
	Region    string
}

func NewTencentSMSer(secretID, secretKey, appID, sign, region string) *TencentSMSer {
	if region == "" {
		region = "ap-guangzhou"
	}
	return &TencentSMSer{SecretID: secretID, SecretKey: secretKey, AppID: appID, Sign: sign, Region: region}
}

func (t *TencentSMSer) Name() string { return "tencent" }

func (t *TencentSMSer) Send(ctx context.Context, s SMS) error {
	tpl, vars := parseSMSBody(s.Body)
	payload := map[string]any{
		"PhoneNumberSet":   []string{"+86" + s.To},
		"SmsSdkAppId":      t.AppID,
		"SignName":         t.Sign,
		"TemplateId":       tpl,
		"TemplateParamSet": []string{vars},
	}
	body, _ := json.Marshal(payload)
	// 简化版 TC3-HMAC-SHA256 签名（生产应严格按 v3 签名流程）
	host := "sms.tencentcloudapi.com"
	service := "sms"
	action := "SendSms"
	version := "2021-01-11"
	timestamp := time.Now().Unix()
	date := time.Unix(timestamp, 0).UTC().Format("2006-01-02")

	canonicalRequest := strings.Join([]string{
		"POST", "/", "", "content-type:application/json; charset=utf-8\nhost:" + host + "\n", "content-type;host", sha256Hex([]byte(body)),
	}, "\n")
	credentialScope := date + "/" + service + "/tc3_request"
	stringToSign := "TC3-HMAC-SHA256\n" + strconv.FormatInt(timestamp, 10) + "\n" + credentialScope + "\n" + sha256Hex([]byte(canonicalRequest))

	secretDate := hmacSHA256Hex("TC3"+t.SecretKey, date)
	secretService := hmacSHA256Hex(secretDate, service)
	secretSigning := hmacSHA256Hex(secretService, "tc3_request")
	signature := hmacSHA256Hex(secretSigning, stringToSign)

	authorization := "TC3-HMAC-SHA256 Credential=" + t.SecretID + "/" + credentialScope +
		", SignedHeaders=content-type;host, Signature=" + signature

	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, "https://"+host, bytes.NewReader(body))
	req.Header.Set("Authorization", authorization)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Host", host)
	req.Header.Set("X-TC-Action", action)
	req.Header.Set("X-TC-Version", version)
	req.Header.Set("X-TC-Timestamp", strconv.FormatInt(timestamp, 10))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return fmt.Errorf("tencent sms: %s", string(respBody))
	}
	return nil
}

func sha256Hex(b []byte) string {
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}

func hmacSHA256Hex(key, data string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

// ============================================================
// Twilio SMS
// ============================================================

// TwilioSMSer Twilio 短信
type TwilioSMSer struct {
	AccountSID string
	AuthToken  string
	From       string
}

func NewTwilioSMSer(sid, token, from string) *TwilioSMSer {
	return &TwilioSMSer{AccountSID: sid, AuthToken: token, From: from}
}

func (t *TwilioSMSer) Name() string { return "twilio" }

func (t *TwilioSMSer) Send(ctx context.Context, s SMS) error {
	apiURL := fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json", t.AccountSID)
	form := url.Values{}
	form.Set("To", s.To)
	form.Set("From", t.From)
	form.Set("Body", s.Body)
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(t.AccountSID, t.AuthToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("twilio sms: %s", string(body))
	}
	return nil
}

// ============================================================
// Qiniu SMS（占位实现）
// ============================================================

type QiniuSMSer struct{}

func NewQiniuSMSer() *QiniuSMSer   { return &QiniuSMSer{} }
func (q *QiniuSMSer) Name() string { return "qiniu" }
func (q *QiniuSMSer) Send(_ context.Context, s SMS) error {
	// 实际应调用七牛短信 API
	_ = s
	return fmt.Errorf("qiniu sms driver is not configured")
}

// ============================================================
// Aliyun DirectMail
// ============================================================

// AliyunDirectMailer 阿里云邮件推送
type AliyunDirectMailer struct {
	AccessKey string
	SecretKey string
	Region    string
	From      string
	FromName  string
}

func NewAliyunDirectMailer(ak, sk, region, from, fromName string) *AliyunDirectMailer {
	if region == "" {
		region = "cn-hangzhou"
	}
	return &AliyunDirectMailer{AccessKey: ak, SecretKey: sk, Region: region, From: from, FromName: fromName}
}

func (a *AliyunDirectMailer) Name() string { return "aliyun_directmail" }

func (a *AliyunDirectMailer) Send(ctx context.Context, m Mail) error {
	// 简化：与 SMS 类似的 v1 签名流程
	subject := m.Subject
	body := m.Text
	if body == "" {
		body = m.HTML
	}
	tos := strings.Join(m.To, ",")

	params := url.Values{}
	params.Set("AccessKeyId", a.AccessKey)
	params.Set("Timestamp", time.Now().UTC().Format("2006-01-02T15:04:05Z"))
	params.Set("Format", "JSON")
	params.Set("SignatureMethod", "HMAC-SHA1")
	params.Set("SignatureVersion", "1.0")
	params.Set("SignatureNonce", strconv.FormatInt(time.Now().UnixNano(), 10))
	params.Set("Action", "SingleSendMail")
	params.Set("Version", "2015-11-23")
	params.Set("RegionId", a.Region)
	params.Set("AccountName", a.From)
	params.Set("FromAlias", a.FromName)
	params.Set("AddressType", "1")
	params.Set("ToAddress", tos)
	params.Set("Subject", subject)
	params.Set("TextBody", body)

	// 排序
	var keys []string
	for k := range params {
		keys = append(keys, k)
	}
	quickSort(keys, 0, len(keys)-1)
	var sb strings.Builder
	sb.WriteString("GET&%2F&")
	encoded := make([]string, 0, len(keys))
	for _, k := range keys {
		encoded = append(encoded, url.QueryEscape(k)+"="+url.QueryEscape(params.Get(k)))
	}
	sb.WriteString(url.QueryEscape(strings.Join(encoded, "&")))

	mac := hmac.New(sha1.New, []byte(a.SecretKey+"&"))
	mac.Write([]byte(sb.String()))
	sign := url.QueryEscape(base64.StdEncoding.EncodeToString(mac.Sum(nil)))

	apiURL := "https://dm.aliyuncs.com/?Signature=" + sign + "&" + params.Encode()
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("aliyun directmail: %s", string(body))
	}
	return nil
}

// 注册驱动
func init() {
	RegisterSMS("aliyun", func() SMSer { return &AliyunSMSer{} })
	RegisterSMS("tencent", func() SMSer { return &TencentSMSer{} })
	RegisterSMS("twilio", func() SMSer { return &TwilioSMSer{} })
	RegisterSMS("qiniu", func() SMSer { return &QiniuSMSer{} })
	RegisterMail("aliyun_directmail", func() Mailer { return &AliyunDirectMailer{} })
}

// 编译期接口校验
var (
	_ SMSer  = (*AliyunSMSer)(nil)
	_ SMSer  = (*TencentSMSer)(nil)
	_ SMSer  = (*TwilioSMSer)(nil)
	_ SMSer  = (*QiniuSMSer)(nil)
	_ Mailer = (*AliyunDirectMailer)(nil)
)
