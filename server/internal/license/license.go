// Package license License 管理
package license

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/ywty/server/internal/logger"
	"go.uber.org/zap"
)

// LicenseType License 类型
const (
	LicenseTypeFree       = "free"       // 免费版
	LicenseTypePro        = "pro"        // 专业版
	LicenseTypeEnterprise = "enterprise" // 企业版
)

// LicenseStatus License 状态
const (
	LicenseStatusActive  = "active"  // 激活
	LicenseStatusExpired = "expired" // 已过期
	LicenseStatusRevoked = "revoked" // 已撤销
)

// License License 信息
type License struct {
	ID         uint64    `json:"id"`
	Key        string    `json:"key"`         // License 密钥
	Type       string    `json:"type"`        // 类型
	Status     string    `json:"status"`      // 状态
	MaxUsers   int       `json:"max_users"`   // 最大用户数
	MaxStorage int64     `json:"max_storage"` // 最大存储空间(字节)
	Features   []string  `json:"features"`    // 启用的功能
	ExpiresAt  time.Time `json:"expires_at"`  // 过期时间
	CreatedAt  time.Time `json:"created_at"`
}

// Service License 服务
type Service struct {
	db *gorm.DB
}

// NewService 创建 License 服务
func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

// Get 获取当前 License
func (s *Service) Get(ctx context.Context) (*License, error) {
	var license License
	err := s.db.WithContext(ctx).First(&license).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 返回默认免费版
			return s.defaultFreeLicense(), nil
		}
		return nil, err
	}
	return &license, nil
}

// Activate 激活 License
func (s *Service) Activate(ctx context.Context, key string) (*License, error) {
	// TODO: 实际项目中需要验证 License 密钥的有效性
	// 这里简化处理，直接创建一条记录

	license := &License{
		Key:        key,
		Type:       LicenseTypePro,
		Status:     LicenseStatusActive,
		MaxUsers:   1000,
		MaxStorage: 100 * 1024 * 1024 * 1024, // 100GB
		Features:   []string{"advanced_search", "batch_upload", "custom_theme"},
		ExpiresAt:  time.Now().AddDate(1, 0, 0), // 1年后过期
	}

	var existing License
	err := s.db.WithContext(ctx).First(&existing).Error
	if err == nil {
		// 更新现有记录
		license.ID = existing.ID
		if err := s.db.WithContext(ctx).Save(license).Error; err != nil {
			return nil, err
		}
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		// 创建新记录
		if err := s.db.WithContext(ctx).Create(license).Error; err != nil {
			return nil, err
		}
	} else {
		return nil, err
	}

	logger.L.Info("license activated",
		zap.String("type", license.Type),
		zap.Time("expires_at", license.ExpiresAt),
	)

	return license, nil
}

// Check 检查 License 状态
func (s *Service) Check(ctx context.Context) (bool, string, error) {
	license, err := s.Get(ctx)
	if err != nil {
		return false, "", err
	}

	if license.Status != LicenseStatusActive {
		return false, "license_not_active", nil
	}

	if time.Now().After(license.ExpiresAt) {
		return false, "license_expired", nil
	}

	return true, "ok", nil
}

// defaultFreeLicense 默认免费版 License
func (s *Service) defaultFreeLicense() *License {
	return &License{
		Type:       LicenseTypeFree,
		Status:     LicenseStatusActive,
		MaxUsers:   100,
		MaxStorage: 10 * 1024 * 1024 * 1024, // 10GB
		Features:   []string{},
		ExpiresAt:  time.Now().AddDate(100, 0, 0), // 永不过期
	}
}

// TableName 指定表名
func (License) TableName() string { return "licenses" }
