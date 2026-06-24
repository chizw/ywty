// Package service 点赞/举报
package service

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	bizerr "github.com/ywty/server/internal/errors"
	"github.com/ywty/server/internal/model"
)

// LikeService 点赞
type LikeService struct {
	db *gorm.DB
}

func NewLikeService(db *gorm.DB) *LikeService { return &LikeService{db: db} }

// Like 点赞（已点则取消）
func (s *LikeService) Like(ctx context.Context, userID uint64, targetType string, targetID uint64) (bool, error) {
	var l model.Like
	err := s.db.WithContext(ctx).Where("user_id = ? AND likeable_type = ? AND likeable_id = ?",
		userID, targetType, targetID).First(&l).Error
	if err == nil {
		// 已存在 -> 取消
		if err := s.db.WithContext(ctx).Delete(&l).Error; err != nil {
			return false, err
		}
		return false, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return false, err
	}
	if err := s.db.WithContext(ctx).Create(&model.Like{
		UserID:       userID,
		LikeableType: targetType,
		LikeableID:   targetID,
	}).Error; err != nil {
		return false, err
	}
	return true, nil
}

// Count 数量
func (s *LikeService) Count(ctx context.Context, targetType string, targetID uint64) int64 {
	var c int64
	_ = s.db.WithContext(ctx).Model(&model.Like{}).Where("likeable_type = ? AND likeable_id = ?", targetType, targetID).Count(&c).Error
	return c
}

// Liked 是否已赞
func (s *LikeService) Liked(ctx context.Context, userID uint64, targetType string, targetID uint64) bool {
	var c int64
	_ = s.db.WithContext(ctx).Model(&model.Like{}).Where("user_id = ? AND likeable_type = ? AND likeable_id = ?",
		userID, targetType, targetID).Count(&c).Error
	return c > 0
}

// ReportService 举报
type ReportService struct {
	db *gorm.DB
}

func NewReportService(db *gorm.DB) *ReportService { return &ReportService{db: db} }

// CreateReportReq 举报参数
type CreateReportReq struct {
	TargetType string `json:"target_type" binding:"required"`
	TargetID   uint64 `json:"target_id" binding:"required"`
	Content    string `json:"content"`
}

// Create 提交举报
func (s *ReportService) Create(ctx context.Context, userID uint64, ip string, req CreateReportReq) (*model.Report, error) {
	r := &model.Report{
		ReportUserID:   &userID,
		ReportableType: req.TargetType,
		ReportableID:   req.TargetID,
		Content:        req.Content,
		Status:         model.ReportStatusUnhandled,
		IPAddress:      ip,
	}
	if err := s.db.WithContext(ctx).Create(r).Error; err != nil {
		return nil, fmt.Errorf("create report: %w", err)
	}
	return r, nil
}

// ListAll 列出（管理员）
func (s *ReportService) ListAll(ctx context.Context, status string) ([]*model.Report, error) {
	q := s.db.WithContext(ctx).Model(&model.Report{}).Order("id DESC")
	if status != "" {
		q = q.Where("status = ?", status)
	}
	var rows []model.Report
	if err := q.Limit(200).Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]*model.Report, 0, len(rows))
	for i := range rows {
		out = append(out, &rows[i])
	}
	return out, nil
}

// AdminList 后台分页列表
func (s *ReportService) AdminList(ctx context.Context, page, perPage int, status string) ([]model.Report, int64, error) {
	q := s.db.WithContext(ctx).Model(&model.Report{}).Order("id DESC")
	if status != "" {
		q = q.Where("status = ?", status)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []model.Report
	if err := q.Offset((page - 1) * perPage).Limit(perPage).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

// UpdateStatus 更新状态
func (s *ReportService) UpdateStatus(ctx context.Context, id uint64, status string) error {
	switch status {
	case model.ReportStatusHandled, model.ReportStatusIgnored, model.ReportStatusUnhandled:
	default:
		return bizerr.BadRequest.WithMessage("invalid status")
	}
	res := s.db.WithContext(ctx).Model(&model.Report{}).Where("id = ?", id).Updates(map[string]any{
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

func nowUnix() int64 { return timeNow().Unix() }
