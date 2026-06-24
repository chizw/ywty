// Package service OAuth 三方登录服务
package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/ywty/server/internal/auth"
	bizerr "github.com/ywty/server/internal/errors"
	"github.com/ywty/server/internal/model"
	"github.com/ywty/server/internal/social"
)

// OAuthService 三方登录
type OAuthService struct {
	db     *gorm.DB
	issuer *auth.Issuer
}

func NewOAuthService(db *gorm.DB, issuer *auth.Issuer) *OAuthService {
	return &OAuthService{db: db, issuer: issuer}
}

// OAuthInfo 对外展示用
type OAuthInfo struct {
	ID        uint64 `json:"id"`
	DriverID  uint64 `json:"driver_id"`
	OpenID    string `json:"openid"`
	Avatar    string `json:"avatar"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	Nickname  string `json:"nickname"`
	CreatedAt int64  `json:"created_at"`
}

// BindRequest 绑定参数
type BindRequest struct {
	DriverID uint64 `json:"driver_id" binding:"required"`
	OpenID   string `json:"openid" binding:"required"`
	Avatar   string `json:"avatar"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	Nickname string `json:"nickname"`
}

// Bind 把三方账号绑定到当前用户
func (s *OAuthService) Bind(ctx context.Context, userID uint64, req BindRequest) (*model.OAuth, error) {
	driverID := req.DriverID
	openID := strings.TrimSpace(req.OpenID)
	if openID == "" {
		return nil, bizerr.BadRequest.WithMessage("openid is required")
	}

	// 检查 driver 是否存在
	var driver model.Driver
	if err := s.db.WithContext(ctx).First(&driver, driverID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, bizerr.BadRequest.WithMessage("driver not found")
		}
		return nil, err
	}
	if driver.Type != model.DriverTypeOAuth {
		return nil, bizerr.BadRequest.WithMessage("driver is not oauth type")
	}

	// 同 openid 不能重复绑定到不同用户
	var existing model.OAuth
	err := s.db.WithContext(ctx).Where("driver_id = ? AND open_id = ?", driverID, openID).First(&existing).Error
	if err == nil {
		if existing.UserID != userID {
			return nil, bizerr.AlreadyExists.WithMessage("this oauth account is bound to another user")
		}
		// 已绑定到当前用户：更新资料
		updates := map[string]any{
			"avatar":   req.Avatar,
			"email":    req.Email,
			"name":     req.Name,
			"nickname": req.Nickname,
		}
		if err := s.db.WithContext(ctx).Model(&existing).Updates(updates).Error; err != nil {
			return nil, fmt.Errorf("update oauth: %w", err)
		}
		return &existing, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	rec := &model.OAuth{
		DriverID: driverID,
		UserID:   userID,
		OpenID:   openID,
		Avatar:   req.Avatar,
		Email:    req.Email,
		Name:     req.Name,
		Nickname: req.Nickname,
		Raw:      model.JSONMap{},
	}
	if err := s.db.WithContext(ctx).Create(rec).Error; err != nil {
		return nil, fmt.Errorf("create oauth: %w", err)
	}
	return rec, nil
}

// List 列出当前用户的所有三方绑定
func (s *OAuthService) List(ctx context.Context, userID uint64) ([]*OAuthInfo, error) {
	var rows []model.OAuth
	if err := s.db.WithContext(ctx).Where("user_id = ?", userID).Order("id DESC").Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]*OAuthInfo, 0, len(rows))
	for i := range rows {
		out = append(out, toOAuthInfo(&rows[i]))
	}
	return out, nil
}

// Unbind 解绑
func (s *OAuthService) Unbind(ctx context.Context, userID, id uint64) error {
	res := s.db.WithContext(ctx).Where("id = ? AND user_id = ?", id, userID).Delete(&model.OAuth{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return bizerr.ResourceNotFound
	}
	return nil
}

// FindByOpenID 通过三方标识查找用户（用于三方登录回调）
func (s *OAuthService) FindByOpenID(ctx context.Context, driverID uint64, openID string) (*model.User, error) {
	openID = strings.TrimSpace(openID)
	if openID == "" {
		return nil, bizerr.BadRequest.WithMessage("openid is required")
	}
	var oa model.OAuth
	if err := s.db.WithContext(ctx).Where("driver_id = ? AND open_id = ?", driverID, openID).First(&oa).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, bizerr.UserNotFound
		}
		return nil, err
	}
	var u model.User
	if err := s.db.WithContext(ctx).First(&u, oa.UserID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, bizerr.UserNotFound
		}
		return nil, err
	}
	return &u, nil
}

func toOAuthInfo(o *model.OAuth) *OAuthInfo {
	return &OAuthInfo{
		ID:        o.ID,
		DriverID:  o.DriverID,
		OpenID:    o.OpenID,
		Avatar:    o.Avatar,
		Email:     o.Email,
		Name:      o.Name,
		Nickname:  o.Nickname,
		CreatedAt: o.CreatedAt.Unix(),
	}
}

// GetDriverByProvider 按 provider 名称查找 OAuth 驱动配置
func (s *OAuthService) GetDriverByProvider(ctx context.Context, provider string) (*model.Driver, error) {
	provider = strings.TrimSpace(strings.ToLower(provider))
	if provider == "" {
		return nil, bizerr.BadRequest.WithMessage("provider is required")
	}
	var driver model.Driver
	if err := s.db.WithContext(ctx).
		Where("name = ? AND type = ?", provider, model.DriverTypeOAuth).
		First(&driver).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, bizerr.BadRequest.WithMessage("oauth driver not found: " + provider)
		}
		return nil, err
	}
	return &driver, nil
}

// AuthorizeURL 构造三方授权跳转 URL
// 返回: url（跳转地址）、state（CSRF 随机串）
func (s *OAuthService) AuthorizeURL(ctx context.Context, provider string) (string, string, error) {
	driver, err := s.GetDriverByProvider(ctx, provider)
	if err != nil {
		return "", "", err
	}
	d, err := social.Get(provider, optionsToStringMap(driver.Options))
	if err != nil {
		return "", "", bizerr.BadRequest.WithMessage(err.Error())
	}
	state := randomToken(16)
	return d.AuthorizeURL(state, ""), state, nil
}

// LoginOrRegister OAuth 回调：换 token、拉用户信息、登录或自动注册
// driverID: DB 中的驱动 ID；provider: 社交驱动名（github/google/wechat...）；code: 授权码
func (s *OAuthService) LoginOrRegister(ctx context.Context, driverID uint64, provider, code string) (*TokenResponse, error) {
	provider = strings.TrimSpace(strings.ToLower(provider))
	code = strings.TrimSpace(code)
	if provider == "" || code == "" {
		return nil, bizerr.BadRequest.WithMessage("provider and code are required")
	}

	var driver model.Driver
	if err := s.db.WithContext(ctx).First(&driver, driverID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, bizerr.BadRequest.WithMessage("driver not found")
		}
		return nil, err
	}
	if driver.Type != model.DriverTypeOAuth {
		return nil, bizerr.BadRequest.WithMessage("driver is not oauth type")
	}

	d, err := social.Get(provider, optionsToStringMap(driver.Options))
	if err != nil {
		return nil, bizerr.BadRequest.WithMessage(err.Error())
	}

	token, err := d.ExchangeCode(ctx, code, "")
	if err != nil {
		return nil, bizerr.Internal.WithMessage("exchange code failed: " + err.Error())
	}

	info, err := d.GetUserInfo(ctx, token)
	if err != nil {
		return nil, bizerr.Internal.WithMessage("get user info failed: " + err.Error())
	}
	if strings.TrimSpace(info.OpenID) == "" {
		return nil, bizerr.Internal.WithMessage("openid is empty from provider")
	}

	// 已绑定：直接登录
	if u, err := s.FindByOpenID(ctx, driverID, info.OpenID); err == nil {
		return s.issueTokens(u)
	} else if !errors.Is(err, bizerr.UserNotFound) {
		return nil, err
	}

	// 未绑定：自动注册新用户
	username := buildOAuthUsername(provider, info.OpenID)
	email := strings.TrimSpace(info.Email)
	if email == "" {
		email = fmt.Sprintf("%s_%s@oauth.local", provider, info.OpenID)
	}
	if len(email) > 255 {
		email = email[:255]
	}

	pwd := randomToken(24)
	hashed, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user := model.User{
		Username: username,
		Email:    email,
		Name:     firstNonEmptyStr(info.Nick, info.Name, username),
		Avatar:   info.Avatar,
		Password: string(hashed),
		IsAdmin:  false,
		Status:   model.UserStatusNormal,
		Options:  model.JSONMap{},
	}
	if err := s.db.WithContext(ctx).Create(&user).Error; err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	// 绑定默认角色组
	var defaultGroup model.Group
	if err := s.db.WithContext(ctx).Where("is_default = ?", true).First(&defaultGroup).Error; err == nil {
		_ = s.db.WithContext(ctx).Create(&model.UserGroup{
			UserID:  user.ID,
			GroupID: defaultGroup.ID,
			From:    model.GroupFromSystem,
		}).Error
	}

	// 绑定 OAuth 记录
	oa := &model.OAuth{
		DriverID: driverID,
		UserID:   user.ID,
		OpenID:   info.OpenID,
		Avatar:   info.Avatar,
		Email:    info.Email,
		Name:     info.Name,
		Nickname: info.Nick,
		Raw:      model.JSONMap{},
	}
	if info.Raw != nil {
		oa.Raw = model.JSONMap(info.Raw)
	}
	if err := s.db.WithContext(ctx).Create(oa).Error; err != nil {
		return nil, fmt.Errorf("create oauth binding: %w", err)
	}

	return s.issueTokens(&user)
}

// issueTokens 颁发 access + refresh token
func (s *OAuthService) issueTokens(user *model.User) (*TokenResponse, error) {
	if s.issuer == nil {
		return nil, errors.New("jwt issuer is not configured")
	}
	access, exp, err := s.issuer.Issue(user.ID, user.Username, user.IsAdmin)
	if err != nil {
		return nil, fmt.Errorf("issue access: %w", err)
	}
	refresh, _, err := s.issuer.IssueRefresh(user.ID)
	if err != nil {
		return nil, fmt.Errorf("issue refresh: %w", err)
	}
	return &TokenResponse{
		AccessToken:  access,
		RefreshToken: refresh,
		TokenType:    "Bearer",
		ExpiresAt:    exp,
		User:         *toUserInfo(user),
	}, nil
}

// optionsToStringMap 把 JSONMap(map[string]any) 转为 map[string]string
func optionsToStringMap(opts model.JSONMap) map[string]string {
	out := make(map[string]string, len(opts))
	for k, v := range opts {
		switch s := v.(type) {
		case string:
			out[k] = s
		case fmt.Stringer:
			out[k] = s.String()
		default:
			if v != nil {
				out[k] = fmt.Sprintf("%v", v)
			}
		}
	}
	return out
}

// buildOAuthUsername 用 provider + openid 生成用户名（截断到 32 字符）
func buildOAuthUsername(provider, openID string) string {
	name := provider + "_" + openID
	if len(name) > 32 {
		name = name[:32]
	}
	return name
}

func firstNonEmptyStr(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

// randomToken 生成 URL 安全的随机字符串
func randomToken(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	max := big.NewInt(int64(len(letters)))
	for i := range b {
		idx, _ := rand.Int(rand.Reader, max)
		b[i] = letters[idx.Int64()]
	}
	return string(b)
}
