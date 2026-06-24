// Package service 个人 API Token（Sanctum 兼容）
package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	bizerr "github.com/ywty/server/internal/errors"
	"github.com/ywty/server/internal/model"
)

// TokenService API Token
type TokenService struct {
	db *gorm.DB
}

func NewTokenService(db *gorm.DB) *TokenService { return &TokenService{db: db} }

// TokenInfo 列表/详情用
type TokenInfo struct {
	ID         uint64    `json:"id"`
	Name       string    `json:"name"`
	Abilities  []any     `json:"abilities"`
	LastUsedAt *int64    `json:"last_used_at,omitempty"`
	ExpiresAt  *int64    `json:"expires_at,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

// IssueToken 签发新 Token（明文仅返回这一次）
func (s *TokenService) IssueToken(ctx context.Context, userID uint64, name string, abilities []any, ttlDays int) (string, *TokenInfo, error) {
	if name == "" {
		return "", nil, bizerr.BadRequest.WithMessage("name is required")
	}
	raw, hash := generateToken(40)
	var exp *int64
	if ttlDays > 0 {
		v := time.Now().Add(time.Duration(ttlDays) * 24 * time.Hour).Unix()
		exp = &v
	}
	rec := &model.PersonalAccessToken{
		UserID:    userID,
		Name:      name,
		Token:     hash,
		Abilities: model.JSONSlice(abilities),
		ExpiresAt: exp,
	}
	if err := s.db.WithContext(ctx).Create(rec).Error; err != nil {
		return "", nil, fmt.Errorf("create token: %w", err)
	}
	return raw, toTokenInfo(rec), nil
}

// List 列出当前用户的 Token
func (s *TokenService) List(ctx context.Context, userID uint64) ([]*TokenInfo, error) {
	var rows []model.PersonalAccessToken
	if err := s.db.WithContext(ctx).Where("user_id = ?", userID).Order("id DESC").Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]*TokenInfo, 0, len(rows))
	for i := range rows {
		out = append(out, toTokenInfo(&rows[i]))
	}
	return out, nil
}

// Revoke 吊销 Token
func (s *TokenService) Revoke(ctx context.Context, userID, id uint64) error {
	res := s.db.WithContext(ctx).Where("user_id = ? AND id = ?", userID, id).Delete(&model.PersonalAccessToken{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return bizerr.TokenNotFound
	}
	return nil
}

// Resolve 通过明文 Token 查找用户（用于 Bearer 模式）
func (s *TokenService) Resolve(ctx context.Context, raw string) (*model.User, error) {
	if raw == "" {
		return nil, bizerr.TokenInvalid
	}
	hash := hashToken(raw)
	var rec model.PersonalAccessToken
	if err := s.db.WithContext(ctx).Where("token = ?", hash).First(&rec).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, bizerr.TokenInvalid
		}
		return nil, err
	}
	now := time.Now().Unix()
	if rec.ExpiresAt != nil && *rec.ExpiresAt < now {
		return nil, bizerr.TokenExpired
	}
	var u model.User
	if err := s.db.WithContext(ctx).First(&u, rec.UserID).Error; err != nil {
		return nil, bizerr.UserNotFound
	}
	_ = s.db.WithContext(ctx).Model(&rec).Update("last_used_at", now).Error
	return &u, nil
}

func toTokenInfo(t *model.PersonalAccessToken) *TokenInfo {
	return &TokenInfo{
		ID:         t.ID,
		Name:       t.Name,
		Abilities:  []any(t.Abilities),
		LastUsedAt: t.LastUsedAt,
		ExpiresAt:  t.ExpiresAt,
		CreatedAt:  t.CreatedAt,
	}
}

// generateToken 生成 token：明文 + sha256 哈希
func generateToken(n int) (string, string) {
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		// 退化方案
		for i := range buf {
			buf[i] = byte(time.Now().UnixNano() >> uint(i))
		}
	}
	raw := hex.EncodeToString(buf)
	return raw, hashToken(raw)
}

func hashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}
