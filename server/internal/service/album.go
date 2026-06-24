// Package service 相册服务
package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	bizerr "github.com/ywty/server/internal/errors"
	"github.com/ywty/server/internal/model"
)

// AlbumService 相册服务
type AlbumService struct {
	db *gorm.DB
}

func NewAlbumService(db *gorm.DB) *AlbumService { return &AlbumService{db: db} }

// AlbumDTO 相册输出
type AlbumDTO struct {
	ID         uint64    `json:"id"`
	Name       string    `json:"name"`
	Intro      string    `json:"intro"`
	IsPublic   bool      `json:"is_public"`
	PhotoCount int64     `json:"photo_count"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// CreateAlbumReq 新建相册
type CreateAlbumReq struct {
	Name     string `json:"name" binding:"required,max=255"`
	Intro    string `json:"intro"`
	IsPublic bool   `json:"is_public"`
}

// Create 新建
func (s *AlbumService) Create(ctx context.Context, userID uint64, req CreateAlbumReq) (*AlbumDTO, error) {
	uid := userID
	a := &model.Album{
		UserID:   &uid,
		Name:     req.Name,
		Intro:    req.Intro,
		IsPublic: req.IsPublic,
	}
	if err := s.db.WithContext(ctx).Create(a).Error; err != nil {
		return nil, fmt.Errorf("create album: %w", err)
	}
	return s.Get(ctx, a.ID, userID)
}

// List 列表
func (s *AlbumService) List(ctx context.Context, userID uint64) ([]*AlbumDTO, error) {
	var rows []model.Album
	if err := s.db.WithContext(ctx).Where("user_id = ?", userID).Order("id DESC").Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]*AlbumDTO, 0, len(rows))
	for i := range rows {
		out = append(out, toAlbumDTO(&rows[i], s.photoCount(ctx, rows[i].ID)))
	}
	return out, nil
}

// Get 单个
func (s *AlbumService) Get(ctx context.Context, id, userID uint64) (*AlbumDTO, error) {
	var a model.Album
	if err := s.db.WithContext(ctx).Where("id = ? AND user_id = ?", id, userID).First(&a).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, bizerr.ResourceNotFound
		}
		return nil, err
	}
	return toAlbumDTO(&a, s.photoCount(ctx, id)), nil
}

// UpdateAlbumReq 更新
type UpdateAlbumReq struct {
	Name     *string `json:"name"`
	Intro    *string `json:"intro"`
	IsPublic *bool   `json:"is_public"`
}

// Update 更新
func (s *AlbumService) Update(ctx context.Context, id, userID uint64, req UpdateAlbumReq) (*AlbumDTO, error) {
	var a model.Album
	if err := s.db.WithContext(ctx).Where("id = ? AND user_id = ?", id, userID).First(&a).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, bizerr.ResourceNotFound
		}
		return nil, err
	}
	updates := map[string]any{}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Intro != nil {
		updates["intro"] = *req.Intro
	}
	if req.IsPublic != nil {
		updates["is_public"] = *req.IsPublic
	}
	if len(updates) > 0 {
		if err := s.db.WithContext(ctx).Model(&a).Updates(updates).Error; err != nil {
			return nil, err
		}
	}
	return s.Get(ctx, id, userID)
}

// Delete 删除（解绑照片但不删照片）
func (s *AlbumService) Delete(ctx context.Context, id, userID uint64) error {
	res := s.db.WithContext(ctx).Where("id = ? AND user_id = ?", id, userID).Delete(&model.Album{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return bizerr.ResourceNotFound
	}
	_ = s.db.WithContext(ctx).Where("album_id = ?", id).Delete(&model.AlbumPhoto{}).Error
	return nil
}

// AddPhoto 绑定照片到相册
func (s *AlbumService) AddPhoto(ctx context.Context, albumID, photoID, userID uint64) error {
	// 校验相册归属
	var a model.Album
	if err := s.db.WithContext(ctx).Where("id = ? AND user_id = ?", albumID, userID).First(&a).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return bizerr.ResourceNotFound
		}
		return err
	}
	// 校验照片归属
	var p model.Photo
	if err := s.db.WithContext(ctx).Where("id = ? AND user_id = ?", photoID, userID).First(&p).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return bizerr.PhotoNotFound
		}
		return err
	}
	rel := &model.AlbumPhoto{AlbumID: albumID, PhotoID: photoID}
	return s.db.WithContext(ctx).FirstOrCreate(rel, model.AlbumPhoto{AlbumID: albumID, PhotoID: photoID}).Error
}

// RemovePhoto 从相册解绑
func (s *AlbumService) RemovePhoto(ctx context.Context, albumID, photoID, userID uint64) error {
	if err := s.db.WithContext(ctx).Where("id = ? AND user_id = ?", albumID, userID).First(&model.Album{}).Error; err != nil {
		return bizerr.ResourceNotFound
	}
	return s.db.WithContext(ctx).Where("album_id = ? AND photo_id = ?", albumID, photoID).Delete(&model.AlbumPhoto{}).Error
}

// ListPhotos 列相册内的照片
func (s *AlbumService) ListPhotos(ctx context.Context, albumID, userID uint64) ([]model.Photo, error) {
	if err := s.db.WithContext(ctx).Where("id = ? AND user_id = ?", albumID, userID).First(&model.Album{}).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, bizerr.ResourceNotFound
		}
		return nil, err
	}
	var rows []model.Photo
	if err := s.db.WithContext(ctx).
		Table("photos").
		Joins("JOIN album_photo ap ON ap.photo_id = photos.id").
		Where("ap.album_id = ? AND photos.deleted_at IS NULL", albumID).
		Order("ap.sort ASC, ap.id DESC").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (s *AlbumService) photoCount(ctx context.Context, albumID uint64) int64 {
	var c int64
	_ = s.db.WithContext(ctx).Model(&model.AlbumPhoto{}).Where("album_id = ?", albumID).Count(&c).Error
	return c
}

func toAlbumDTO(a *model.Album, count int64) *AlbumDTO {
	return &AlbumDTO{
		ID:         a.ID,
		Name:       a.Name,
		Intro:      a.Intro,
		IsPublic:   a.IsPublic,
		PhotoCount: count,
		CreatedAt:  a.CreatedAt,
		UpdatedAt:  a.UpdatedAt,
	}
}
