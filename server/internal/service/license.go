// Package service License 管理服务
package service

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/ywty/server/internal/license"
)

// LicenseService License 管理服务
type LicenseService struct {
	db *gorm.DB
}

// NewLicenseService 创建 License 管理服务
func NewLicenseService(db *gorm.DB) *LicenseService {
	return &LicenseService{db: db}
}

// GetLicense 获取当前 License
func (s *LicenseService) GetLicense(ctx context.Context) (*license.License, error) {
	var l license.License
	if err := s.db.WithContext(ctx).First(&l).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 返回默认免费版
			return &license.License{
				Type:       license.LicenseTypeFree,
				Status:     license.LicenseStatusActive,
				MaxUsers:   100,
				MaxStorage: 10 * 1024 * 1024 * 1024, // 10GB
				Features:   []string{},
				ExpiresAt:  time.Now().AddDate(100, 0, 0),
			}, nil
		}
		return nil, err
	}
	return &l, nil
}

// ActivateLicense 激活 License
func (s *LicenseService) ActivateLicense(ctx context.Context, key string) (*license.License, error) {
	// TODO: 实际项目中需要验证 License 密钥的有效性（调用远程验证服务）
	// 这里简化处理，直接创建一条记录

	l := &license.License{
		Key:        key,
		Type:       license.LicenseTypePro,
		Status:     license.LicenseStatusActive,
		MaxUsers:   1000,
		MaxStorage: 100 * 1024 * 1024 * 1024, // 100GB
		Features:   []string{"advanced_search", "batch_upload", "custom_theme"},
		ExpiresAt:  time.Now().AddDate(1, 0, 0), // 1年后过期
	}

	// 检查是否已存在
	var existing license.License
	err := s.db.WithContext(ctx).First(&existing).Error
	if err == nil {
		// 更新现有记录
		l.ID = existing.ID
		if err := s.db.WithContext(ctx).Save(l).Error; err != nil {
			return nil, err
		}
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		// 创建新记录
		if err := s.db.WithContext(ctx).Create(l).Error; err != nil {
			return nil, err
		}
	} else {
		return nil, err
	}

	return l, nil
}

// CheckLicense 检查 License 状态
func (s *LicenseService) CheckLicense(ctx context.Context) (bool, string, error) {
	l, err := s.GetLicense(ctx)
	if err != nil {
		return false, "", err
	}

	if l.Status != license.LicenseStatusActive {
		return false, "license_not_active", nil
	}

	if time.Now().After(l.ExpiresAt) {
		return false, "license_expired", nil
	}

	return true, "ok", nil
}
