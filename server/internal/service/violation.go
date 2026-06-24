// Package service 违规记录
package service

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	bizerr "github.com/ywty/server/internal/errors"
	"github.com/ywty/server/internal/model"
)

// ViolationService 违规记录服务
type ViolationService struct {
	db *gorm.DB
}

// NewViolationService 创建违规记录服务
func NewViolationService(db *gorm.DB) *ViolationService {
	return &ViolationService{db: db}
}

// Create 创建违规记录
func (s *ViolationService) Create(ctx context.Context, userID, photoID uint64, reason string) (*model.Violation, error) {
	v := &model.Violation{
		Reason: reason,
		Status: model.ViolationStatusUnhandled,
	}
	if userID > 0 {
		v.UserID = userID
	}
	if photoID > 0 {
		v.PhotoID = photoID
	}
	if err := s.db.WithContext(ctx).Create(v).Error; err != nil {
		return nil, fmt.Errorf("create violation: %w", err)
	}
	return v, nil
}

// List 后台列表（分页）
func (s *ViolationService) List(ctx context.Context, page, perPage int, status string) ([]model.Violation, int64, error) {
	q := s.db.WithContext(ctx).Model(&model.Violation{}).Order("id DESC")
	if status != "" {
		q = q.Where("status = ?", status)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []model.Violation
	if err := q.Offset((page - 1) * perPage).Limit(perPage).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

// UpdateStatus 更新状态
func (s *ViolationService) UpdateStatus(ctx context.Context, id uint64, status string) error {
	switch status {
	case model.ViolationStatusUnhandled, model.ViolationStatusHandled, model.ViolationStatusIgnored:
	default:
		return bizerr.BadRequest.WithMessage("invalid status")
	}
	res := s.db.WithContext(ctx).Model(&model.Violation{}).Where("id = ?", id).Updates(map[string]any{
		"status":     status,
		"handled_at": nowUnix(),
	})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return bizerr.ResourceNotFound
	}
	return nil
}
