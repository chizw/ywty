// Package service 公告服务
package service

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	bizerr "github.com/ywty/server/internal/errors"
	"github.com/ywty/server/internal/model"
)

// NoticeService 公告服务
type NoticeService struct {
	db *gorm.DB
}

func NewNoticeService(db *gorm.DB) *NoticeService { return &NoticeService{db: db} }

// CreateNoticeReq 创建公告请求
type CreateNoticeReq struct {
	Title   string `json:"title" binding:"required"`
	Content string `json:"content"`
	Sort    int    `json:"sort"`
}

// UpdateNoticeReq 更新公告请求
type UpdateNoticeReq struct {
	Title   *string `json:"title"`
	Content *string `json:"content"`
	Sort    *int    `json:"sort"`
}

// ListPublic 公开列表（按 sort 降序 / id 降序）
func (s *NoticeService) ListPublic(ctx context.Context, page, perPage int) ([]model.Notice, int64, error) {
	page, perPage = normalizePage(page, perPage)
	var total int64
	q := s.db.WithContext(ctx).Model(&model.Notice{})
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []model.Notice
	if err := q.Order("sort DESC, id DESC").Offset((page - 1) * perPage).Limit(perPage).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

// GetPublic 公开详情，同时 view_count + 1
func (s *NoticeService) GetPublic(ctx context.Context, id uint64) (*model.Notice, error) {
	var n model.Notice
	if err := s.db.WithContext(ctx).First(&n, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, bizerr.ResourceNotFound
		}
		return nil, err
	}
	_ = s.db.WithContext(ctx).Model(&n).UpdateColumn("view_count", gorm.Expr("view_count + 1")).Error
	n.ViewCount++
	return &n, nil
}

// AdminList 后台列表
func (s *NoticeService) AdminList(ctx context.Context, page, perPage int) ([]model.Notice, int64, error) {
	page, perPage = normalizePage(page, perPage)
	var total int64
	q := s.db.WithContext(ctx).Model(&model.Notice{})
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []model.Notice
	if err := q.Order("sort DESC, id DESC").Offset((page - 1) * perPage).Limit(perPage).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

// AdminCreate 后台创建
func (s *NoticeService) AdminCreate(ctx context.Context, req CreateNoticeReq) (*model.Notice, error) {
	n := &model.Notice{
		Title:   req.Title,
		Content: req.Content,
		Sort:    req.Sort,
	}
	if err := s.db.WithContext(ctx).Create(n).Error; err != nil {
		return nil, fmt.Errorf("create notice: %w", err)
	}
	return n, nil
}

// AdminUpdate 后台更新
func (s *NoticeService) AdminUpdate(ctx context.Context, id uint64, req UpdateNoticeReq) (*model.Notice, error) {
	var n model.Notice
	if err := s.db.WithContext(ctx).First(&n, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, bizerr.ResourceNotFound
		}
		return nil, err
	}
	updates := map[string]any{}
	if req.Title != nil {
		updates["title"] = *req.Title
	}
	if req.Content != nil {
		updates["content"] = *req.Content
	}
	if req.Sort != nil {
		updates["sort"] = *req.Sort
	}
	if len(updates) > 0 {
		if err := s.db.WithContext(ctx).Model(&n).Updates(updates).Error; err != nil {
			return nil, err
		}
	}
	return &n, nil
}

// AdminDelete 后台删除
func (s *NoticeService) AdminDelete(ctx context.Context, id uint64) error {
	res := s.db.WithContext(ctx).Delete(&model.Notice{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return bizerr.ResourceNotFound
	}
	return nil
}
