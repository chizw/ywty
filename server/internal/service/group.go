// Package service 群组管理服务
package service

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	bizerr "github.com/ywty/server/internal/errors"
	"github.com/ywty/server/internal/model"
)

// GroupService 群组管理服务
type GroupService struct {
	db *gorm.DB
}

// NewGroupService 创建群组管理服务
func NewGroupService(db *gorm.DB) *GroupService {
	return &GroupService{db: db}
}

// CreateGroupRequest 创建群组请求
type CreateGroupRequest struct {
	Name      string        `json:"name" binding:"required"`
	Intro     string        `json:"intro"`
	Options   model.JSONMap `json:"options"`
	IsDefault bool          `json:"is_default"`
	IsGuest   bool          `json:"is_guest"`
}

// UpdateGroupRequest 更新群组请求
type UpdateGroupRequest struct {
	Name      string        `json:"name"`
	Intro     string        `json:"intro"`
	Options   model.JSONMap `json:"options"`
	IsDefault bool          `json:"is_default"`
	IsGuest   bool          `json:"is_guest"`
}

// Create 创建群组
func (s *GroupService) Create(ctx context.Context, req CreateGroupRequest) (*model.Group, error) {
	// 检查名称是否已存在
	var count int64
	if err := s.db.WithContext(ctx).Model(&model.Group{}).Where("name = ?", req.Name).Count(&count).Error; err != nil {
		return nil, fmt.Errorf("count group: %w", err)
	}
	if count > 0 {
		return nil, bizerr.AlreadyExists.WithMessage("群组名称已存在")
	}

	group := &model.Group{
		Name:      req.Name,
		Intro:     req.Intro,
		Options:   req.Options,
		IsDefault: req.IsDefault,
		IsGuest:   req.IsGuest,
	}

	if err := s.db.WithContext(ctx).Create(group).Error; err != nil {
		return nil, fmt.Errorf("create group: %w", err)
	}
	return group, nil
}

// List 列出群组
func (s *GroupService) List(ctx context.Context, page, pageSize int) ([]model.Group, int64, error) {
	var total int64
	if err := s.db.WithContext(ctx).Model(&model.Group{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count groups: %w", err)
	}

	var groups []model.Group
	offset := (page - 1) * pageSize
	if err := s.db.WithContext(ctx).Order("id DESC").Offset(offset).Limit(pageSize).Find(&groups).Error; err != nil {
		return nil, 0, fmt.Errorf("list groups: %w", err)
	}
	return groups, total, nil
}

// Get 获取群组详情
func (s *GroupService) Get(ctx context.Context, id uint64) (*model.Group, error) {
	var group model.Group
	if err := s.db.WithContext(ctx).First(&group, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, bizerr.NotFound
		}
		return nil, fmt.Errorf("get group: %w", err)
	}
	return &group, nil
}

// Update 更新群组
func (s *GroupService) Update(ctx context.Context, id uint64, req UpdateGroupRequest) (*model.Group, error) {
	var group model.Group
	if err := s.db.WithContext(ctx).First(&group, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, bizerr.NotFound
		}
		return nil, fmt.Errorf("get group: %w", err)
	}

	// 如果修改名称，检查是否重复
	if req.Name != "" && req.Name != group.Name {
		var count int64
		if err := s.db.WithContext(ctx).Model(&model.Group{}).Where("name = ? AND id != ?", req.Name, id).Count(&count).Error; err != nil {
			return nil, fmt.Errorf("count group: %w", err)
		}
		if count > 0 {
			return nil, bizerr.AlreadyExists.WithMessage("群组名称已存在")
		}
		group.Name = req.Name
	}

	if req.Intro != "" {
		group.Intro = req.Intro
	}
	if req.Options != nil {
		group.Options = req.Options
	}
	group.IsDefault = req.IsDefault
	group.IsGuest = req.IsGuest

	if err := s.db.WithContext(ctx).Save(&group).Error; err != nil {
		return nil, fmt.Errorf("update group: %w", err)
	}
	return &group, nil
}

// Delete 删除群组
func (s *GroupService) Delete(ctx context.Context, id uint64) error {
	// 检查是否有用户属于该群组
	var count int64
	if err := s.db.WithContext(ctx).Model(&model.UserGroup{}).Where("group_id = ?", id).Count(&count).Error; err != nil {
		return fmt.Errorf("count user groups: %w", err)
	}
	if count > 0 {
		return bizerr.BadRequest.WithMessage("该群组下还有用户，无法删除")
	}

	result := s.db.WithContext(ctx).Delete(&model.Group{}, id)
	if result.Error != nil {
		return fmt.Errorf("delete group: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return bizerr.NotFound
	}
	return nil
}
