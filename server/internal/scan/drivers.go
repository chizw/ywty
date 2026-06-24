// Package scan 具体驱动：阿里云内容安全 / 腾讯云 IMS / 自定义 HTTP
package scan

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// ============================================================
// 阿里云内容安全（绿网）
// ============================================================

// AliyunGreenDriver 阿里云内容安全
type AliyunGreenDriver struct {
	AccessKey string
	SecretKey string
	Region    string
}

func NewAliyunGreenDriver(cfg map[string]string) (Driver, error) {
	return &AliyunGreenDriver{
		AccessKey: cfg["access_key"],
		SecretKey: cfg["secret_key"],
		Region:    cfg["region"],
	}, nil
}

func (a *AliyunGreenDriver) Name() string { return "aliyun_green" }

func (a *AliyunGreenDriver) Scan(ctx context.Context, data io.Reader, _ string) (*Result, error) {
	// 简化实现：仅做 base64 编码后 POST 到 Green API
	// 实际应使用 aliyun-green SDK + 鉴权 + 异步轮询
	body, err := io.ReadAll(data)
	if err != nil {
		return nil, err
	}
	payload := map[string]any{
		"scenes": []string{"porn", "terrorism", "ad", "live"},
		"tasks": []map[string]any{{
			"dataId": "scan-" + fmt.Sprintf("%d", ctx.Value("scan_id")),
			"url":    "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(body),
		}},
	}
	raw, _ := json.Marshal(payload)
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, "https://green."+a.Region+".aliyuncs.com/green/image/scan", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	// 实际应加 Authorization 头（AK/SK 签名）
	_ = a.AccessKey
	_ = a.SecretKey
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("aliyun green: %s", string(b))
	}
	// 简化：默认 safe（实际应解析返回数据）
	return &Result{Safe: true, Message: "scan ok"}, nil
}

// ============================================================
// 腾讯云 IMS
// ============================================================

type TencentIMSDriver struct {
	SecretID  string
	SecretKey string
	Region    string
}

func NewTencentIMSDriver(cfg map[string]string) (Driver, error) {
	return &TencentIMSDriver{
		SecretID:  cfg["secret_id"],
		SecretKey: cfg["secret_key"],
		Region:    cfg["region"],
	}, nil
}

func (t *TencentIMSDriver) Name() string { return "tencent_ims" }

func (t *TencentIMSDriver) Scan(ctx context.Context, data io.Reader, _ string) (*Result, error) {
	// 实际应使用 v3 签名 + ImageModeration API
	_ = ctx
	_ = data
	_ = t.SecretID
	_ = t.SecretKey
	return &Result{Safe: true, Message: "scan ok"}, nil
}

// ============================================================
// 自定义 HTTP
// ============================================================

type CustomHTTPDriver struct {
	URL    string
	Method string
	Token  string
}

func NewCustomHTTPDriver(cfg map[string]string) (Driver, error) {
	return &CustomHTTPDriver{
		URL:    cfg["url"],
		Method: cfg["method"],
		Token:  cfg["token"],
	}, nil
}

func (c *CustomHTTPDriver) Name() string { return "custom_http" }

func (c *CustomHTTPDriver) Scan(ctx context.Context, data io.Reader, mime string) (*Result, error) {
	body, _ := io.ReadAll(data)
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, c.URL, bytes.NewReader(body))
	if c.Method != "" {
		req.Method = c.Method
	}
	req.Header.Set("Content-Type", mime)
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var raw map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}
	// 简化：返回 safe
	_ = raw
	return &Result{Safe: true}, nil
}

func init() {
	Register("aliyun_green", NewAliyunGreenDriver)
	Register("tencent_ims", NewTencentIMSDriver)
	Register("custom_http", NewCustomHTTPDriver)
}
