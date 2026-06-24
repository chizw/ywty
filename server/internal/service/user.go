// Package service 用户服务（资料/密码/邮箱/手机）
package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	bizerr "github.com/ywty/server/internal/errors"
	"github.com/ywty/server/internal/model"
)

// UserService 用户服务
type UserService struct {
	db *gorm.DB
}

func NewUserService(db *gorm.DB) *UserService { return &UserService{db: db} }

// ProfileDTO 用户资料
type ProfileDTO struct {
	ID           uint64    `json:"id"`
	Username     string    `json:"username"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	Phone        string    `json:"phone"`
	Avatar       string    `json:"avatar"`
	Location     string    `json:"location"`
	URL          string    `json:"url"`
	Company      string    `json:"company"`
	CompanyTitle string    `json:"company_title"`
	Tagline      string    `json:"tagline"`
	Bio          string    `json:"bio"`
	Interests    []any     `json:"interests"`
	Socials      []any     `json:"socials"`
	IsAdmin      bool      `json:"is_admin"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
}

// GetProfile 获取资料
func (s *UserService) GetProfile(ctx context.Context, userID uint64) (*ProfileDTO, error) {
	var u model.User
	if err := s.db.WithContext(ctx).First(&u, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, bizerr.UserNotFound
		}
		return nil, err
	}
	return toProfile(&u), nil
}

// UpdateProfileReq 更新资料
type UpdateProfileReq struct {
	Name         string `json:"name"`
	Avatar       string `json:"avatar"`
	Location     string `json:"location"`
	URL          string `json:"url"`
	Company      string `json:"company"`
	CompanyTitle string `json:"company_title"`
	Tagline      string `json:"tagline"`
	Bio          string `json:"bio"`
	Interests    []any  `json:"interests"`
	Socials      []any  `json:"socials"`
}

// UpdateProfile 更新资料（仅允许本人）
func (s *UserService) UpdateProfile(ctx context.Context, userID uint64, req UpdateProfileReq) (*ProfileDTO, error) {
	updates := map[string]any{
		"name":          strings.TrimSpace(req.Name),
		"avatar":        req.Avatar,
		"location":      req.Location,
		"url":           req.URL,
		"company":       req.Company,
		"company_title": req.CompanyTitle,
		"tagline":       req.Tagline,
		"bio":           req.Bio,
	}
	if req.Interests != nil {
		updates["interests"] = model.JSONSlice(req.Interests)
	}
	if req.Socials != nil {
		updates["socials"] = model.JSONSlice(req.Socials)
	}
	if err := s.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", userID).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("update profile: %w", err)
	}
	return s.GetProfile(ctx, userID)
}

// ChangePassword 修改密码（已登录）
func (s *UserService) ChangePassword(ctx context.Context, userID uint64, oldPwd, newPwd string) error {
	if len(newPwd) < 6 {
		return bizerr.BadRequest.WithMessage("new password too short")
	}
	var u model.User
	if err := s.db.WithContext(ctx).First(&u, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return bizerr.UserNotFound
		}
		return err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(oldPwd)); err != nil {
		return bizerr.OldPwdWrong
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(newPwd), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}
	if err := s.db.WithContext(ctx).Model(&u).Update("password", string(hashed)).Error; err != nil {
		return err
	}
	// 吊销该用户所有 API Token（安全考虑）
	_ = s.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&model.PersonalAccessToken{}).Error
	return nil
}

// ChangeEmail 更换邮箱
// newEmail 必须是未使用的新邮箱
// code 必须与 model.VerifyEventChangeEmail 一致
func (s *UserService) ChangeEmail(ctx context.Context, userID uint64, newEmail, code string) error {
	newEmail = strings.TrimSpace(strings.ToLower(newEmail))
	if newEmail == "" {
		return bizerr.BadRequest.WithMessage("email is required")
	}
	var count int64
	if err := s.db.WithContext(ctx).Model(&model.User{}).Where("email = ?", newEmail).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return bizerr.EmailBound
	}
	if err := NewVerifyCodeService(s.db).Verify(ctx, model.VerifyChannelEmail, newEmail, model.VerifyEventChangeEmail, code); err != nil {
		return err
	}
	now := time.Now().Unix()
	if err := s.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", userID).Updates(map[string]any{
		"email":             newEmail,
		"email_verified_at": now,
	}).Error; err != nil {
		return err
	}
	return nil
}

// ChangePhone 更换手机
func (s *UserService) ChangePhone(ctx context.Context, userID uint64, newPhone, code string) error {
	newPhone = strings.TrimSpace(newPhone)
	if newPhone == "" {
		return bizerr.BadRequest.WithMessage("phone is required")
	}
	var count int64
	if err := s.db.WithContext(ctx).Model(&model.User{}).Where("phone = ?", newPhone).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return bizerr.PhoneBound
	}
	if err := NewVerifyCodeService(s.db).Verify(ctx, model.VerifyChannelSMS, newPhone, model.VerifyEventChangePhone, code); err != nil {
		return err
	}
	now := time.Now().Unix()
	if err := s.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", userID).Updates(map[string]any{
		"phone":             newPhone,
		"phone_verified_at": now,
	}).Error; err != nil {
		return err
	}
	return nil
}

func toProfile(u *model.User) *ProfileDTO {
	p := &ProfileDTO{
		ID:           u.ID,
		Username:     u.Username,
		Name:         u.Name,
		Email:        u.Email,
		Avatar:       u.Avatar,
		Location:     u.Location,
		URL:          u.URL,
		Company:      u.Company,
		CompanyTitle: u.CompanyTitle,
		Tagline:      u.Tagline,
		Bio:          u.Bio,
		Interests:    []any(u.Interests),
		Socials:      []any(u.Socials),
		IsAdmin:      u.IsAdmin,
		Status:       u.Status,
		CreatedAt:    u.CreatedAt,
	}
	if u.Phone != nil {
		p.Phone = *u.Phone
	}
	return p
}
