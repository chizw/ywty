// Package service 标签服务
package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"

	bizerr "github.com/ywty/server/internal/errors"
	"github.com/ywty/server/internal/model"
)

// TagService 标签
type TagService struct {
	db *gorm.DB
}

func NewTagService(db *gorm.DB) *TagService { return &TagService{db: db} }

// Create 新建标签（同名复用）
func (s *TagService) Create(ctx context.Context, name string) (*model.Tag, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, bizerr.BadRequest.WithMessage("name is required")
	}
	var t model.Tag
	if err := s.db.WithContext(ctx).Where("name = ?", name).First(&t).Error; err == nil {
		return &t, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	t = model.Tag{Name: name}
	if err := s.db.WithContext(ctx).Create(&t).Error; err != nil {
		return nil, fmt.Errorf("create tag: %w", err)
	}
	return &t, nil
}

// List 列出标签
func (s *TagService) List(ctx context.Context) ([]*model.Tag, error) {
	var rows []model.Tag
	if err := s.db.WithContext(ctx).Order("id DESC").Limit(200).Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]*model.Tag, 0, len(rows))
	for i := range rows {
		out = append(out, &rows[i])
	}
	return out, nil
}

// Delete 删除标签
func (s *TagService) Delete(ctx context.Context, id uint64) error {
	res := s.db.WithContext(ctx).Delete(&model.Tag{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return bizerr.ResourceNotFound
	}
	_ = s.db.WithContext(ctx).Where("tag_id = ?", id).Delete(&model.Taggable{}).Error
	return nil
}

// Attach 绑定标签到资源
func (s *TagService) Attach(ctx context.Context, userID, tagID uint64, targetType string, targetID uint64) error {
	rel := &model.Taggable{
		TagID:        &tagID,
		UserID:       &userID,
		TaggableType: targetType,
		TaggableID:   targetID,
	}
	return s.db.WithContext(ctx).Where("tag_id = ? AND taggable_type = ? AND taggable_id = ?", tagID, targetType, targetID).
		FirstOrCreate(rel).Error
}

// Detach 解绑
func (s *TagService) Detach(ctx context.Context, userID, tagID uint64, targetType string, targetID uint64) error {
	return s.db.WithContext(ctx).
		Where("tag_id = ? AND user_id = ? AND taggable_type = ? AND taggable_id = ?", tagID, userID, targetType, targetID).
		Delete(&model.Taggable{}).Error
}

// AttachByNames 一次性按名称绑定（不存在则自动创建）
func (s *TagService) AttachByNames(ctx context.Context, userID uint64, names []string, targetType string, targetID uint64) ([]*model.Tag, error) {
	out := make([]*model.Tag, 0, len(names))
	seen := map[string]bool{}
	for _, n := range names {
		n = strings.TrimSpace(n)
		if n == "" || seen[n] {
			continue
		}
		seen[n] = true
		tag, err := s.Create(ctx, n)
		if err != nil {
			return nil, err
		}
		if err := s.Attach(ctx, userID, tag.ID, targetType, targetID); err != nil {
			return nil, err
		}
		out = append(out, tag)
	}
	return out, nil
}

// ListForTarget 查询某资源绑定的标签
func (s *TagService) ListForTarget(ctx context.Context, targetType string, targetID uint64) ([]*model.Tag, error) {
	var rows []model.Tag
	err := s.db.WithContext(ctx).
		Table("tags").
		Joins("JOIN taggables t ON t.tag_id = tags.id").
		Where("t.taggable_type = ? AND t.taggable_id = ?", targetType, targetID).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]*model.Tag, 0, len(rows))
	for i := range rows {
		out = append(out, &rows[i])
	}
	return out, nil
}
