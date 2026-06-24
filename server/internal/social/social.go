// Package social 三方登录驱动（OAuth2）
package social

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// User 三方账号资料
type User struct {
	OpenID  string
	UnionID string
	Nick    string
	Name    string
	Email   string
	Avatar  string
	Raw     map[string]any
}

// Driver 三方登录驱动接口
type Driver interface {
	Name() string

	// AuthorizeURL 构造授权跳转 URL
	// state: 防 CSRF 随机串
	AuthorizeURL(state, redirect string) string

	// ExchangeCode 拿授权码换 access token
	ExchangeCode(ctx context.Context, code, redirect string) (*Token, error)

	// GetUserInfo 拉取用户信息
	GetUserInfo(ctx context.Context, token *Token) (*User, error)
}

// Token 授权 token
type Token struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int
	Scope        string
	TokenType    string
	Raw          map[string]any
}

// Factory 驱动工厂
type Factory func(cfg map[string]string) (Driver, error)

var registry = map[string]Factory{}

// Register 注册驱动
func Register(name string, f Factory) { registry[name] = f }

// Get 构造驱动
func Get(name string, cfg map[string]string) (Driver, error) {
	f, ok := registry[name]
	if !ok {
		return nil, errors.New("unsupported social driver: " + name)
	}
	return f(cfg)
}

// Drivers 列出已注册
func Drivers() []string {
	out := make([]string, 0, len(registry))
	for k := range registry {
		out = append(out, k)
	}
	return out
}

// HTTPGet 简易 HTTP GET JSON
func HTTPGet(ctx context.Context, u string, headers map[string]string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("http %d: %s", resp.StatusCode, string(body))
	}
	if out == nil {
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

// HTTPPost 简易 HTTP POST 表单
func HTTPPost(ctx context.Context, u string, form url.Values, headers map[string]string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("http %d: %s", resp.StatusCode, string(body))
	}
	if out == nil {
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(out)
}
