// Package service 业务服务层
package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/ywty/server/internal/auth"
	bizerr "github.com/ywty/server/internal/errors"
	"github.com/ywty/server/internal/model"
)

// 错误
var (
	ErrInvalidCredentials = errors.New("用户名或密码错误")
	ErrUserDisabled       = errors.New("账号已被停用")
	ErrUserExists         = errors.New("用户已存在")
	ErrInvalidVerifyCode  = errors.New("验证码错误或已过期")
)

// AuthService 认证服务
type AuthService struct {
	db     *gorm.DB
	issuer *auth.Issuer
}

// NewAuthService 创建认证服务
func NewAuthService(db *gorm.DB, issuer *auth.Issuer) *AuthService {
	return &AuthService{db: db, issuer: issuer}
}

// RegisterRequest 注册参数
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=32"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6,max=64"`
	Phone    string `json:"phone"`
}

// LoginRequest 登录参数
type LoginRequest struct {
	Account  string `json:"account" binding:"required"` // username / email / phone
	Password string `json:"password" binding:"required,min=6"`
}

// TokenResponse 登录成功返回
type TokenResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	TokenType    string    `json:"token_type"`
	ExpiresAt    time.Time `json:"expires_at"`
	User         UserInfo  `json:"user"`
}

// UserInfo 用户信息
type UserInfo struct {
	ID       uint64 `json:"id"`
	Username string `json:"username"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Avatar   string `json:"avatar"`
	IsAdmin  bool   `json:"is_admin"`
	Status   string `json:"status"`
}

// Register 注册新用户
func (s *AuthService) Register(ctx context.Context, req RegisterRequest) (*TokenResponse, error) {
	var existing int64
	q := s.db.WithContext(ctx).Model(&model.User{}).
		Where("username = ? OR email = ?", req.Username, req.Email)
	if req.Phone != "" {
		q = q.Or("phone = ?", req.Phone)
	}
	if err := q.Count(&existing).Error; err != nil {
		return nil, fmt.Errorf("check existing: %w", err)
	}
	if existing > 0 {
		return nil, ErrUserExists
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	var phonePtr *string
	if req.Phone != "" {
		p := req.Phone
		phonePtr = &p
	}
	user := model.User{
		Username: req.Username,
		Email:    req.Email,
		Phone:    phonePtr,
		Name:     req.Username,
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

	return s.issueTokens(&user)
}

// Login 账号密码登录
func (s *AuthService) Login(ctx context.Context, req LoginRequest) (*TokenResponse, error) {
	var user model.User
	account := strings.TrimSpace(req.Account)
	q := s.db.WithContext(ctx).Where("username = ?", account)
	if strings.Contains(account, "@") {
		q = s.db.WithContext(ctx).Where("email = ?", account)
	} else if len(account) > 0 && account[0] >= '0' && account[0] <= '9' {
		q = s.db.WithContext(ctx).Where("phone = ?", account)
	}
	if err := q.First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("find user: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	if user.Status != model.UserStatusNormal {
		return nil, ErrUserDisabled
	}

	return s.issueTokens(&user)
}

// Refresh 刷新 access token
func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (*TokenResponse, error) {
	// 使用专门的 refresh token 解析方法
	userID, err := s.issuer.ParseRefresh(refreshToken)
	if err != nil {
		return nil, ErrInvalidCredentials
	}
	var user model.User
	if err := s.db.WithContext(ctx).First(&user, userID).Error; err != nil {
		return nil, ErrInvalidCredentials
	}
	return s.issueTokens(&user)
}

// ResetPasswordByEmail 通过邮箱验证码重置密码
func (s *AuthService) ResetPasswordByEmail(ctx context.Context, email, code, newPwd string) error {
	email = strings.ToLower(strings.TrimSpace(email))
	if err := NewVerifyCodeService(s.db).Verify(ctx, model.VerifyChannelEmail, email, model.VerifyEventResetPassword, code); err != nil {
		return err
	}
	var u model.User
	if err := s.db.WithContext(ctx).Where("email = ?", email).First(&u).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return bizerr.UserNotFound
		}
		return err
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(newPwd), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return s.db.WithContext(ctx).Model(&u).Update("password", string(hashed)).Error
}

// ResetPasswordByPhone 通过短信验证码重置密码
func (s *AuthService) ResetPasswordByPhone(ctx context.Context, phone, code, newPwd string) error {
	phone = strings.TrimSpace(phone)
	if err := NewVerifyCodeService(s.db).Verify(ctx, model.VerifyChannelSMS, phone, model.VerifyEventResetPassword, code); err != nil {
		return err
	}
	var u model.User
	if err := s.db.WithContext(ctx).Where("phone = ?", phone).First(&u).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return bizerr.UserNotFound
		}
		return err
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(newPwd), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return s.db.WithContext(ctx).Model(&u).Update("password", string(hashed)).Error
}

// Me 获取当前用户信息
func (s *AuthService) Me(userID uint64) (*UserInfo, error) {
	var user model.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return nil, err
	}
	return toUserInfo(&user), nil
}

func (s *AuthService) issueTokens(user *model.User) (*TokenResponse, error) {
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

func toUserInfo(u *model.User) *UserInfo {
	return &UserInfo{
		ID:       u.ID,
		Username: u.Username,
		Name:     u.Name,
		Email:    u.Email,
		Avatar:   u.Avatar,
		IsAdmin:  u.IsAdmin,
		Status:   u.Status,
	}
}

// IsRecordNotFound 判断是否为记录不存在
func IsRecordNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}

// ToAPIError 把 service 错误转为业务错误
func ToAPIError(err error) *bizerr.Error {
	switch {
	case errors.Is(err, ErrInvalidCredentials):
		return bizerr.PasswordIncorrect
	case errors.Is(err, ErrUserDisabled):
		return bizerr.UserDisabled
	case errors.Is(err, ErrUserExists):
		return bizerr.UserExists
	case errors.Is(err, ErrInvalidVerifyCode):
		return bizerr.CodeIncorrect
	default:
		return bizerr.Internal.WithCause(err)
	}
}
