// Package service 分享服务
package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	bizerr "github.com/ywty/server/internal/errors"
	"github.com/ywty/server/internal/model"
)

// ShareService 分享
type ShareService struct {
	db *gorm.DB
}

func NewShareService(db *gorm.DB) *ShareService { return &ShareService{db: db} }

// CreateShareReq 新建分享
type CreateShareReq struct {
	Type          string   `json:"type" binding:"required,oneof=photo album"`
	IDs           []uint64 `json:"ids" binding:"required,min=1"`
	Password      string   `json:"password"`
	ExpireMinutes int      `json:"expire_minutes"` // 0 = 永不过期
}

// Create 创建
func (s *ShareService) Create(ctx context.Context, userID uint64, req CreateShareReq) (*model.Share, error) {
	slug, err := randomSlug(10)
	if err != nil {
		return nil, fmt.Errorf("gen slug: %w", err)
	}
	sh := &model.Share{
		UserID:  userID,
		Type:    req.Type,
		Slug:    slug,
		Content: req.Password,
	}
	if req.ExpireMinutes > 0 {
		exp := time.Now().Add(time.Duration(req.ExpireMinutes) * time.Minute).Unix()
		sh.ExpiredAt = &exp
	}
	if err := s.db.WithContext(ctx).Create(sh).Error; err != nil {
		return nil, err
	}
	// 绑定资源
	for _, id := range req.IDs {
		rel := &model.Shareable{
			ShareID:       sh.ID,
			ShareableType: req.Type,
			ShareableID:   id,
		}
		if err := s.db.WithContext(ctx).Create(rel).Error; err != nil {
			return nil, err
		}
	}
	return sh, nil
}

// List 列出我的分享
func (s *ShareService) List(ctx context.Context, userID uint64) ([]*model.Share, error) {
	var rows []model.Share
	if err := s.db.WithContext(ctx).Where("user_id = ?", userID).Order("id DESC").Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]*model.Share, 0, len(rows))
	for i := range rows {
		out = append(out, &rows[i])
	}
	return out, nil
}

// UpdateShareReq 更新分享
type UpdateShareReq struct {
	Password      *string `json:"password"`       // nil=不改, ""=清除, 非空=设置
	ExpireMinutes *int    `json:"expire_minutes"` // nil=不改, >0=设置过期, 0=取消过期
}

// Update 更新分享（仅所有者）
func (s *ShareService) Update(ctx context.Context, userID, id uint64, req UpdateShareReq) (*model.Share, error) {
	var sh model.Share
	if err := s.db.WithContext(ctx).Where("id = ? AND user_id = ?", id, userID).First(&sh).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, bizerr.ResourceNotFound
		}
		return nil, err
	}

	// Password: nil 不改，空串清除，非空设置（密码存于 Content 字段）
	if req.Password != nil {
		if err := s.db.WithContext(ctx).Model(&sh).Update("content", *req.Password).Error; err != nil {
			return nil, err
		}
		sh.Content = *req.Password
	}

	// ExpireMinutes: nil 不改，>0 设置过期，0 取消过期
	if req.ExpireMinutes != nil {
		if *req.ExpireMinutes > 0 {
			exp := time.Now().Add(time.Duration(*req.ExpireMinutes) * time.Minute).Unix()
			if err := s.db.WithContext(ctx).Model(&sh).Update("expired_at", exp).Error; err != nil {
				return nil, err
			}
			sh.ExpiredAt = &exp
		} else {
			if err := s.db.WithContext(ctx).Model(&sh).Update("expired_at", nil).Error; err != nil {
				return nil, err
			}
			sh.ExpiredAt = nil
		}
	}

	return &sh, nil
}

// Delete 删除
func (s *ShareService) Delete(ctx context.Context, userID, id uint64) error {
	res := s.db.WithContext(ctx).Where("id = ? AND user_id = ?", id, userID).Delete(&model.Share{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return bizerr.ResourceNotFound
	}
	_ = s.db.WithContext(ctx).Where("share_id = ?", id).Delete(&model.Shareable{}).Error
	return nil
}

// GetBySlug 通过 slug 获取（公开）
func (s *ShareService) GetBySlug(ctx context.Context, slug string) (*model.Share, []map[string]any, error) {
	var sh model.Share
	if err := s.db.WithContext(ctx).Where("slug = ?", slug).First(&sh).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, bizerr.ResourceNotFound
		}
		return nil, nil, err
	}
	if sh.ExpiredAt != nil && *sh.ExpiredAt < time.Now().Unix() {
		return nil, nil, bizerr.ResourceNotFound.WithMessage("share expired")
	}
	// 递增浏览
	_ = s.db.WithContext(ctx).Model(&sh).Update("view_count", gorm.Expr("view_count + 1")).Error

	// 取出关联资源
	relType := sh.Type
	var shareableIDs []uint64
	if err := s.db.WithContext(ctx).Model(&model.Shareable{}).
		Where("share_id = ? AND shareable_type = ?", sh.ID, relType).
		Pluck("shareable_id", &shareableIDs).Error; err != nil {
		return nil, nil, err
	}
	items := make([]map[string]any, 0, len(shareableIDs))
	if relType == model.ShareTypePhoto {
		var photos []model.Photo
		if err := s.db.WithContext(ctx).Where("id IN ?", shareableIDs).Find(&photos).Error; err == nil {
			for i := range photos {
				items = append(items, map[string]any{
					"id": photos[i].ID, "name": photos[i].Name,
					"pathname": photos[i].Pathname, "mimetype": photos[i].Mimetype,
					"size": photos[i].Size, "is_public": photos[i].IsPublic,
				})
			}
		}
	} else {
		var albums []model.Album
		if err := s.db.WithContext(ctx).Where("id IN ?", shareableIDs).Find(&albums).Error; err == nil {
			for i := range albums {
				items = append(items, map[string]any{
					"id": albums[i].ID, "name": albums[i].Name, "intro": albums[i].Intro,
					"is_public": albums[i].IsPublic,
				})
			}
		}
	}
	return &sh, items, nil
}

func randomSlug(n int) (string, error) {
	if n < 6 {
		n = 6
	}
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}
