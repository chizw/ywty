// Package service 意见反馈
package service

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	bizerr "github.com/ywty/server/internal/errors"
	"github.com/ywty/server/internal/model"
)

// CreateFeedbackReq 创建反馈请求
type CreateFeedbackReq struct {
	Type    string `json:"type" binding:"required,oneof=general bug suggest"`
	Title   string `json:"title" binding:"required"`
	Name    string `json:"name" binding:"required"`
	Email   string `json:"email" binding:"required,email"`
	Content string `json:"content" binding:"required"`
}

// FeedbackService 反馈服务
type FeedbackService struct {
	db *gorm.DB
}

// NewFeedbackService 创建反馈服务
func NewFeedbackService(db *gorm.DB) *FeedbackService {
	return &FeedbackService{db: db}
}

// Create 创建反馈（公开）
func (s *FeedbackService) Create(ctx context.Context, ip string, req CreateFeedbackReq) (*model.Feedback, error) {
	f := &model.Feedback{
		Type:      req.Type,
		Title:     req.Title,
		Name:      req.Name,
		Email:     req.Email,
		Content:   req.Content,
		IPAddress: ip,
	}
	if err := s.db.WithContext(ctx).Create(f).Error; err != nil {
		return nil, fmt.Errorf("create feedback: %w", err)
	}
	return f, nil
}

// List 后台列表（分页）
func (s *FeedbackService) List(ctx context.Context, page, perPage int) ([]model.Feedback, int64, error) {
	q := s.db.WithContext(ctx).Model(&model.Feedback{}).Order("id DESC")
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []model.Feedback
	if err := q.Offset((page - 1) * perPage).Limit(perPage).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

// Delete 删除反馈
func (s *FeedbackService) Delete(ctx context.Context, id uint64) error {
	res := s.db.WithContext(ctx).Delete(&model.Feedback{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return bizerr.ResourceNotFound
	}
	return nil
}
