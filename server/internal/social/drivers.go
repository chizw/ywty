// Package social 通用 OAuth2 驱动（GitHub / Google / 微信 / QQ / 钉钉 / Gitee / 微博）
package social

import (
	"context"
	"fmt"
	"net/url"
)

// ============================================================
// GitHub
// ============================================================

type GitHubDriver struct {
	ClientID     string
	ClientSecret string
	Redirect     string
}

func NewGitHubDriver(cfg map[string]string) (Driver, error) {
	return &GitHubDriver{
		ClientID:     cfg["client_id"],
		ClientSecret: cfg["client_secret"],
		Redirect:     cfg["redirect"],
	}, nil
}

func (g *GitHubDriver) Name() string { return "github" }

func (g *GitHubDriver) AuthorizeURL(state, _ string) string {
	v := url.Values{}
	v.Set("client_id", g.ClientID)
	v.Set("redirect_uri", g.Redirect)
	v.Set("scope", "user:email read:user")
	v.Set("state", state)
	return "https://github.com/login/oauth/authorize?" + v.Encode()
}

func (g *GitHubDriver) ExchangeCode(ctx context.Context, code, redirect string) (*Token, error) {
	v := url.Values{}
	v.Set("client_id", g.ClientID)
	v.Set("client_secret", g.ClientSecret)
	v.Set("code", code)
	v.Set("redirect_uri", firstNonEmpty(redirect, g.Redirect))
	var raw map[string]any
	if err := HTTPPost(ctx, "https://github.com/login/oauth/access_token", v, nil, &raw); err != nil {
		return nil, err
	}
	return &Token{
		AccessToken: asString(raw["access_token"]),
		Scope:       asString(raw["scope"]),
		TokenType:   asString(raw["token_type"]),
		Raw:         raw,
	}, nil
}

func (g *GitHubDriver) GetUserInfo(ctx context.Context, t *Token) (*User, error) {
	var raw map[string]any
	if err := HTTPGet(ctx, "https://api.github.com/user", map[string]string{
		"Authorization": "Bearer " + t.AccessToken,
	}, &raw); err != nil {
		return nil, err
	}
	// emails
	var emails []map[string]any
	_ = HTTPGet(ctx, "https://api.github.com/user/emails", map[string]string{
		"Authorization": "Bearer " + t.AccessToken,
	}, &emails)
	email := asString(raw["email"])
	if email == "" && len(emails) > 0 {
		email = asString(emails[0]["email"])
	}
	return &User{
		OpenID: fmt.Sprintf("%v", raw["id"]),
		Nick:   asString(raw["login"]),
		Name:   asString(raw["name"]),
		Email:  email,
		Avatar: asString(raw["avatar_url"]),
		Raw:    raw,
	}, nil
}

// ============================================================
// Google
// ============================================================

type GoogleDriver struct {
	ClientID     string
	ClientSecret string
	Redirect     string
}

func NewGoogleDriver(cfg map[string]string) (Driver, error) {
	return &GoogleDriver{
		ClientID:     cfg["client_id"],
		ClientSecret: cfg["client_secret"],
		Redirect:     cfg["redirect"],
	}, nil
}

func (g *GoogleDriver) Name() string { return "google" }

func (g *GoogleDriver) AuthorizeURL(state, _ string) string {
	v := url.Values{}
	v.Set("client_id", g.ClientID)
	v.Set("redirect_uri", g.Redirect)
	v.Set("response_type", "code")
	v.Set("scope", "openid email profile")
	v.Set("access_type", "offline")
	v.Set("state", state)
	return "https://accounts.google.com/o/oauth2/v2/auth?" + v.Encode()
}

func (g *GoogleDriver) ExchangeCode(ctx context.Context, code, redirect string) (*Token, error) {
	v := url.Values{}
	v.Set("client_id", g.ClientID)
	v.Set("client_secret", g.ClientSecret)
	v.Set("code", code)
	v.Set("grant_type", "authorization_code")
	v.Set("redirect_uri", firstNonEmpty(redirect, g.Redirect))
	var raw map[string]any
	if err := HTTPPost(ctx, "https://oauth2.googleapis.com/token", v, nil, &raw); err != nil {
		return nil, err
	}
	return &Token{
		AccessToken:  asString(raw["access_token"]),
		RefreshToken: asString(raw["refresh_token"]),
		ExpiresIn:    int(asFloat(raw["expires_in"])),
		TokenType:    asString(raw["token_type"]),
		Raw:          raw,
	}, nil
}

func (g *GoogleDriver) GetUserInfo(ctx context.Context, t *Token) (*User, error) {
	var raw map[string]any
	if err := HTTPGet(ctx, "https://openidconnect.googleapis.com/v1/userinfo", map[string]string{
		"Authorization": "Bearer " + t.AccessToken,
	}, &raw); err != nil {
		return nil, err
	}
	return &User{
		OpenID: asString(raw["sub"]),
		Nick:   asString(raw["given_name"]),
		Name:   asString(raw["name"]),
		Email:  asString(raw["email"]),
		Avatar: asString(raw["picture"]),
		Raw:    raw,
	}, nil
}

// ============================================================
// 微信
// ============================================================

type WeChatDriver struct {
	AppID     string
	AppSecret string
	Redirect  string
}

func NewWeChatDriver(cfg map[string]string) (Driver, error) {
	return &WeChatDriver{
		AppID:     cfg["app_id"],
		AppSecret: cfg["app_secret"],
		Redirect:  cfg["redirect"],
	}, nil
}

func (w *WeChatDriver) Name() string { return "wechat" }

func (w *WeChatDriver) AuthorizeURL(state, _ string) string {
	v := url.Values{}
	v.Set("appid", w.AppID)
	v.Set("redirect_uri", w.Redirect)
	v.Set("response_type", "code")
	v.Set("scope", "snsapi_login")
	v.Set("state", state)
	return "https://open.weixin.qq.com/connect/qrconnect?" + v.Encode() + "#wechat_redirect"
}

func (w *WeChatDriver) ExchangeCode(ctx context.Context, code, _ string) (*Token, error) {
	v := url.Values{}
	v.Set("appid", w.AppID)
	v.Set("secret", w.AppSecret)
	v.Set("code", code)
	v.Set("grant_type", "authorization_code")
	var raw map[string]any
	if err := HTTPGet(ctx, "https://api.weixin.qq.com/sns/oauth2/access_token?"+v.Encode(), nil, &raw); err != nil {
		return nil, err
	}
	return &Token{
		AccessToken: asString(raw["access_token"]),
		RefreshToken: asString(raw["refresh_token"]),
		ExpiresIn:    int(asFloat(raw["expires_in"])),
		Raw:          raw,
	}, nil
}

func (w *WeChatDriver) GetUserInfo(ctx context.Context, t *Token) (*User, error) {
	v := url.Values{}
	v.Set("access_token", t.AccessToken)
	v.Set("openid", asString(t.Raw["openid"]))
	v.Set("lang", "zh_CN")
	var raw map[string]any
	if err := HTTPGet(ctx, "https://api.weixin.qq.com/sns/userinfo?"+v.Encode(), nil, &raw); err != nil {
		return nil, err
	}
	return &User{
		OpenID: asString(raw["openid"]),
		UnionID: asString(raw["unionid"]),
		Nick:   asString(raw["nickname"]),
		Name:   asString(raw["nickname"]),
		Avatar: asString(raw["headimgurl"]),
		Raw:    raw,
	}, nil
}

// ============================================================
// QQ
// ============================================================

type QQDriver struct {
	AppID     string
	AppKey    string
	Redirect  string
}

func NewQQDriver(cfg map[string]string) (Driver, error) {
	return &QQDriver{
		AppID:    cfg["app_id"],
		AppKey:   cfg["app_key"],
		Redirect: cfg["redirect"],
	}, nil
}

func (q *QQDriver) Name() string { return "qq" }

func (q *QQDriver) AuthorizeURL(state, _ string) string {
	v := url.Values{}
	v.Set("response_type", "code")
	v.Set("client_id", q.AppID)
	v.Set("redirect_uri", q.Redirect)
	v.Set("state", state)
	return "https://graph.qq.com/oauth2.0/authorize?" + v.Encode()
}

func (q *QQDriver) ExchangeCode(ctx context.Context, code, _ string) (*Token, error) {
	v := url.Values{}
	v.Set("grant_type", "authorization_code")
	v.Set("client_id", q.AppID)
	v.Set("client_secret", q.AppKey)
	v.Set("code", code)
	v.Set("redirect_uri", q.Redirect)
	var raw map[string]any
	if err := HTTPGet(ctx, "https://graph.qq.com/oauth2.0/token?"+v.Encode(), nil, &raw); err != nil {
		return nil, err
	}
	return &Token{
		AccessToken: asString(raw["access_token"]),
		ExpiresIn:   int(asFloat(raw["expires_in"])),
		Raw:         raw,
	}, nil
}

func (q *QQDriver) GetUserInfo(ctx context.Context, t *Token) (*User, error) {
	// 简化：直接返回 OpenID，openid 已在 token raw 中
	openid := asString(t.Raw["openid"])
	return &User{OpenID: openid, Raw: t.Raw}, nil
}

// ============================================================
// 钉钉
// ============================================================

type DingTalkDriver struct {
	AppID     string
	AppSecret string
	Redirect  string
}

func NewDingTalkDriver(cfg map[string]string) (Driver, error) {
	return &DingTalkDriver{
		AppID:     cfg["app_id"],
		AppSecret: cfg["app_secret"],
		Redirect:  cfg["redirect"],
	}, nil
}

func (d *DingTalkDriver) Name() string { return "dingtalk" }

func (d *DingTalkDriver) AuthorizeURL(state, _ string) string {
	v := url.Values{}
	v.Set("appid", d.AppID)
	v.Set("response_type", "code")
	v.Set("scope", "snsapi_login")
	v.Set("state", state)
	v.Set("redirect_uri", d.Redirect)
	return "https://oapi.dingtalk.com/connect/oauth2/sns_authorize?" + v.Encode()
}

func (d *DingTalkDriver) ExchangeCode(ctx context.Context, code, _ string) (*Token, error) {
	v := url.Values{}
	v.Set("accessKey", d.AppID)
	v.Set("secret", d.AppSecret)
	v.Set("code", code)
	v.Set("grant_type", "authorization_code")
	var raw map[string]any
	if err := HTTPGet(ctx, "https://api.dingtalk.com/v1.0/oauth2/userAccessToken?"+v.Encode(), nil, &raw); err != nil {
		return nil, err
	}
	return &Token{
		AccessToken: asString(raw["accessToken"]),
		ExpiresIn:   int(asFloat(raw["expireIn"])),
		Raw:         raw,
	}, nil
}

func (d *DingTalkDriver) GetUserInfo(ctx context.Context, t *Token) (*User, error) {
	var raw map[string]any
	if err := HTTPGet(ctx, "https://api.dingtalk.com/v1.0/contact/users/me", map[string]string{
		"Authorization": "Bearer " + t.AccessToken,
	}, &raw); err != nil {
		return nil, err
	}
	return &User{
		OpenID: asString(raw["unionId"]),
		Nick:   asString(raw["nick"]),
		Name:   asString(raw["name"]),
		Avatar: asString(raw["avatarUrl"]),
		Raw:    raw,
	}, nil
}

// ============================================================
// Gitee
// ============================================================

type GiteeDriver struct {
	ClientID     string
	ClientSecret string
	Redirect     string
}

func NewGiteeDriver(cfg map[string]string) (Driver, error) {
	return &GiteeDriver{
		ClientID:     cfg["client_id"],
		ClientSecret: cfg["client_secret"],
		Redirect:     cfg["redirect"],
	}, nil
}

func (g *GiteeDriver) Name() string { return "gitee" }

func (g *GiteeDriver) AuthorizeURL(state, _ string) string {
	v := url.Values{}
	v.Set("client_id", g.ClientID)
	v.Set("redirect_uri", g.Redirect)
	v.Set("response_type", "code")
	v.Set("scope", "user_info emails")
	v.Set("state", state)
	return "https://gitee.com/oauth/authorize?" + v.Encode()
}

func (g *GiteeDriver) ExchangeCode(ctx context.Context, code, _ string) (*Token, error) {
	v := url.Values{}
	v.Set("grant_type", "authorization_code")
	v.Set("code", code)
	v.Set("client_id", g.ClientID)
	v.Set("client_secret", g.ClientSecret)
	v.Set("redirect_uri", g.Redirect)
	var raw map[string]any
	if err := HTTPPost(ctx, "https://gitee.com/oauth/token", v, nil, &raw); err != nil {
		return nil, err
	}
	return &Token{
		AccessToken: asString(raw["access_token"]),
		TokenType:   asString(raw["token_type"]),
		Raw:         raw,
	}, nil
}

func (g *GiteeDriver) GetUserInfo(ctx context.Context, t *Token) (*User, error) {
	var raw map[string]any
	if err := HTTPGet(ctx, "https://gitee.com/api/v5/user", map[string]string{
		"Authorization": "Bearer " + t.AccessToken,
	}, &raw); err != nil {
		return nil, err
	}
	return &User{
		OpenID: fmt.Sprintf("%v", raw["id"]),
		Nick:   asString(raw["login"]),
		Name:   asString(raw["name"]),
		Email:  asString(raw["email"]),
		Avatar: asString(raw["avatar_url"]),
		Raw:    raw,
	}, nil
}

// ============================================================
// 微博
// ============================================================

type WeiboDriver struct {
	ClientID     string
	ClientSecret string
	Redirect     string
}

func NewWeiboDriver(cfg map[string]string) (Driver, error) {
	return &WeiboDriver{
		ClientID:     cfg["client_id"],
		ClientSecret: cfg["client_secret"],
		Redirect:     cfg["redirect"],
	}, nil
}

func (w *WeiboDriver) Name() string { return "weibo" }

func (w *WeiboDriver) AuthorizeURL(state, _ string) string {
	v := url.Values{}
	v.Set("client_id", w.ClientID)
	v.Set("response_type", "code")
	v.Set("redirect_uri", w.Redirect)
	v.Set("state", state)
	return "https://api.weibo.com/oauth2/authorize?" + v.Encode()
}

func (w *WeiboDriver) ExchangeCode(ctx context.Context, code, _ string) (*Token, error) {
	v := url.Values{}
	v.Set("client_id", w.ClientID)
	v.Set("client_secret", w.ClientSecret)
	v.Set("grant_type", "authorization_code")
	v.Set("code", code)
	v.Set("redirect_uri", w.Redirect)
	var raw map[string]any
	if err := HTTPPost(ctx, "https://api.weibo.com/oauth2/access_token", v, nil, &raw); err != nil {
		return nil, err
	}
	return &Token{
		AccessToken: asString(raw["access_token"]),
		Raw:         raw,
	}, nil
}

func (w *WeiboDriver) GetUserInfo(ctx context.Context, t *Token) (*User, error) {
	uid := asString(t.Raw["uid"])
	var raw map[string]any
	if err := HTTPGet(ctx, "https://api.weibo.com/2/users/show.json?access_token="+t.AccessToken+"&uid="+uid, nil, &raw); err != nil {
		return nil, err
	}
	return &User{
		OpenID: uid,
		Nick:   asString(raw["screen_name"]),
		Name:   asString(raw["name"]),
		Avatar: asString(raw["avatar_large"]),
		Raw:    raw,
	}, nil
}

// 工具
func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

func asString(v any) string {
	s, _ := v.(string)
	return s
}

func asFloat(v any) float64 {
	switch n := v.(type) {
	case float64:
		return n
	case int:
		return float64(n)
	case int64:
		return float64(n)
	}
	return 0
}

func init() {
	Register("github", NewGitHubDriver)
	Register("google", NewGoogleDriver)
	Register("wechat", NewWeChatDriver)
	Register("qq", NewQQDriver)
	Register("dingtalk", NewDingTalkDriver)
	Register("gitee", NewGiteeDriver)
	Register("weibo", NewWeiboDriver)
}
