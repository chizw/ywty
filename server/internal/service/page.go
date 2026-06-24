// Package service 单页服务
package service

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	bizerr "github.com/ywty/server/internal/errors"
	"github.com/ywty/server/internal/model"
)

// PageService 单页服务
type PageService struct {
	db *gorm.DB
}

func NewPageService(db *gorm.DB) *PageService { return &PageService{db: db} }

// CreatePageReq 创建单页请求
type CreatePageReq struct {
	Type        string `json:"type" binding:"required,oneof=internal external"`
	Name        string `json:"name" binding:"required"`
	Icon        string `json:"icon"`
	Title       string `json:"title"`
	Content     string `json:"content"`
	Keywords    string `json:"keywords"`
	Description string `json:"description"`
	Slug        string `json:"slug" binding:"required"`
	URL         string `json:"url"`
	Sort        int    `json:"sort"`
	IsShow      bool   `json:"is_show"`
}

// UpdatePageReq 更新单页请求
type UpdatePageReq struct {
	Type        *string `json:"type"`
	Name        *string `json:"name"`
	Icon        *string `json:"icon"`
	Title       *string `json:"title"`
	Content     *string `json:"content"`
	Keywords    *string `json:"keywords"`
	Description *string `json:"description"`
	Slug        *string `json:"slug"`
	URL         *string `json:"url"`
	Sort        *int    `json:"sort"`
	IsShow      *bool   `json:"is_show"`
}

// GetBySlug 通过 slug 查公开单页（is_show=true）
func (s *PageService) GetBySlug(ctx context.Context, slug string) (*model.Page, error) {
	var p model.Page
	if err := s.db.WithContext(ctx).Where("slug = ? AND is_show = ?", slug, true).First(&p).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, bizerr.ResourceNotFound
		}
		return nil, err
	}
	_ = s.db.WithContext(ctx).Model(&p).UpdateColumn("view_count", gorm.Expr("view_count + 1")).Error
	p.ViewCount++
	return &p, nil
}

// ListPublic 公开列表
func (s *PageService) ListPublic(ctx context.Context) ([]model.Page, error) {
	var rows []model.Page
	if err := s.db.WithContext(ctx).Where("is_show = ?", true).Order("sort DESC, id DESC").Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

// AdminList 后台列表
func (s *PageService) AdminList(ctx context.Context, page, perPage int) ([]model.Page, int64, error) {
	page, perPage = normalizePage(page, perPage)
	var total int64
	q := s.db.WithContext(ctx).Model(&model.Page{})
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []model.Page
	if err := q.Order("sort DESC, id DESC").Offset((page - 1) * perPage).Limit(perPage).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

// AdminGet 后台详情
func (s *PageService) AdminGet(ctx context.Context, id uint64) (*model.Page, error) {
	var p model.Page
	if err := s.db.WithContext(ctx).First(&p, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, bizerr.ResourceNotFound
		}
		return nil, err
	}
	return &p, nil
}

// AdminCreate 后台创建
func (s *PageService) AdminCreate(ctx context.Context, req CreatePageReq) (*model.Page, error) {
	p := &model.Page{
		Type:        req.Type,
		Name:        req.Name,
		Icon:        req.Icon,
		Title:       req.Title,
		Content:     req.Content,
		Keywords:    req.Keywords,
		Description: req.Description,
		Slug:        req.Slug,
		URL:         req.URL,
		Sort:        req.Sort,
		IsShow:      req.IsShow,
	}
	if err := s.db.WithContext(ctx).Create(p).Error; err != nil {
		return nil, fmt.Errorf("create page: %w", err)
	}
	return p, nil
}

// AdminUpdate 后台更新
func (s *PageService) AdminUpdate(ctx context.Context, id uint64, req UpdatePageReq) (*model.Page, error) {
	var p model.Page
	if err := s.db.WithContext(ctx).First(&p, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, bizerr.ResourceNotFound
		}
		return nil, err
	}
	updates := map[string]any{}
	if req.Type != nil {
		updates["type"] = *req.Type
	}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Icon != nil {
		updates["icon"] = *req.Icon
	}
	if req.Title != nil {
		updates["title"] = *req.Title
	}
	if req.Content != nil {
		updates["content"] = *req.Content
	}
	if req.Keywords != nil {
		updates["keywords"] = *req.Keywords
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Slug != nil {
		updates["slug"] = *req.Slug
	}
	if req.URL != nil {
		updates["url"] = *req.URL
	}
	if req.Sort != nil {
		updates["sort"] = *req.Sort
	}
	if req.IsShow != nil {
		updates["is_show"] = *req.IsShow
	}
	if len(updates) > 0 {
		if err := s.db.WithContext(ctx).Model(&p).Updates(updates).Error; err != nil {
			return nil, err
		}
	}
	return &p, nil
}

// AdminDelete 后台删除
func (s *PageService) AdminDelete(ctx context.Context, id uint64) error {
	res := s.db.WithContext(ctx).Delete(&model.Page{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return bizerr.ResourceNotFound
	}
	return nil
}
