// Package service PhotoService 图片服务
package service

import (
	"context"
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/hibiken/asynq"
	"gorm.io/gorm"

	bizerr "github.com/ywty/server/internal/errors"
	"github.com/ywty/server/internal/jobs"
	"github.com/ywty/server/internal/model"
	"github.com/ywty/server/internal/queue"
	"github.com/ywty/server/internal/storage"
)

// PhotoService 图片服务
type PhotoService struct {
	db       *gorm.DB
	driver   storage.Driver
	rootURL  string
	imageDir string
	maxSize  int64
	queue    *queue.Client
}

// NewPhotoService 创建图片服务
func NewPhotoService(db *gorm.DB, driver storage.Driver, rootURL string) *PhotoService {
	return &PhotoService{
		db:       db,
		driver:   driver,
		rootURL:  strings.TrimRight(rootURL, "/"),
		imageDir: "photos",
		maxSize:  20 * 1024 * 1024, // 20MB
	}
}

// SetQueue 注入队列客户端（上传后可异步缩略图）
func (s *PhotoService) SetQueue(c *queue.Client) {
	s.queue = c
}

// UploadResult 上传结果
type UploadResult struct {
	Photo    *model.Photo `json:"photo"`
	URL      string       `json:"url"`
	Markdown string       `json:"markdown"`
	HTML     string       `json:"html"`
}

// UploadOptions 上传参数
type UploadOptions struct {
	Name     string // 自定义名
	Intro    string // 介绍
	IsPublic bool   // 是否公开
}

// Upload 上传文件
func (s *PhotoService) Upload(ctx context.Context, userID uint64, filename string, reader io.Reader, size int64, opts ...UploadOptions) (*UploadResult, error) {
	if size > s.maxSize {
		return nil, bizerr.FileTooLarge
	}
	if size <= 0 {
		return nil, bizerr.BadRequest.WithMessage("empty file")
	}
	var opt UploadOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	// 读出文件内容用于算 hash + 探测 mime
	buf, err := io.ReadAll(io.LimitReader(reader, s.maxSize+1))
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	md5sum := md5.Sum(buf)
	sha1sum := sha1.Sum(buf)
	md5Hex := hex.EncodeToString(md5sum[:])
	sha1Hex := hex.EncodeToString(sha1sum[:])

	mime := detectMIME(buf)
	ext := extensionFromMIME(mime, filename)

	// 路径：yyyy/MM/dd/<hash>.<ext>
	now := time.Now()
	key := fmt.Sprintf("%s/%s/%s/%s/%s.%s",
		s.imageDir,
		now.Format("2006"),
		now.Format("01"),
		now.Format("02"),
		md5Hex,
		ext,
	)

	// 写存储
	_, err = s.driver.Put(ctx, key, strings.NewReader(string(buf)), int64(len(buf)), mime)
	if err != nil {
		return nil, bizerr.StorageUploadFail.WithCause(err)
	}

	// 名称策略：opts.Name > 原文件名
	name := strings.TrimSpace(opt.Name)
	if name == "" {
		name = strings.TrimSuffix(filepath.Base(filename), filepath.Ext(filename))
	}

	// 入库
	photo := &model.Photo{
		UserID:    &userID,
		Name:      name,
		Intro:     opt.Intro,
		Filename:  filepath.Base(filename),
		Pathname:  key,
		Mimetype:  mime,
		Extension: ext,
		MD5:       md5Hex,
		SHA1:      sha1Hex,
		Size:      float64(len(buf)) / 1024, // 保留为 kb
		IsPublic:  opt.IsPublic,
		Status:    model.PhotoStatusNormal,
	}
	if err := s.db.WithContext(ctx).Create(photo).Error; err != nil {
		_ = s.driver.Delete(ctx, key)
		return nil, fmt.Errorf("save photo: %w", err)
	}

	url := s.driver.PublicURL(key)

	// 异步入队缩略图生成
	if s.queue != nil {
		payload, _ := json.Marshal(jobs.ThumbnailPayload{
			PhotoID:  photo.ID,
			Width:    320,
			Height:   320,
			Strategy: "fit",
		})
		_ = s.queue.EnqueueLow(asynq.NewTask(queue.TypeThumbnail, payload))
	}

	return &UploadResult{
		Photo:    photo,
		URL:      url,
		Markdown: fmt.Sprintf("![%s](%s)", photo.Name, url),
		HTML:     fmt.Sprintf(`<img src="%s" alt="%s" title="%s" />`, url, photo.Name, photo.Name),
	}, nil
}

// List 列出我的图片
func (s *PhotoService) List(ctx context.Context, userID uint64, page, perPage int) ([]model.Photo, int64, error) {
	if page <= 0 {
		page = 1
	}
	if perPage <= 0 || perPage > 100 {
		perPage = 20
	}
	var photos []model.Photo
	var total int64
	q := s.db.WithContext(ctx).Model(&model.Photo{}).Where("user_id = ?", userID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Order("id DESC").Offset((page - 1) * perPage).Limit(perPage).Find(&photos).Error; err != nil {
		return nil, 0, err
	}
	return photos, total, nil
}

// Get 获取单张图片
func (s *PhotoService) Get(ctx context.Context, id uint64, userID uint64) (*model.Photo, error) {
	var p model.Photo
	if err := s.db.WithContext(ctx).First(&p, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, bizerr.PhotoNotFound
		}
		return nil, err
	}
	if p.UserID == nil || *p.UserID != userID {
		return nil, bizerr.Forbidden
	}
	return &p, nil
}

// Delete 删除图片
func (s *PhotoService) Delete(ctx context.Context, id uint64, userID uint64) error {
	var p model.Photo
	if err := s.db.WithContext(ctx).First(&p, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return bizerr.PhotoNotFound
		}
		return err
	}
	if p.UserID == nil || *p.UserID != userID {
		return bizerr.Forbidden
	}
	// 软删 + 删除存储文件
	if err := s.db.WithContext(ctx).Delete(&p).Error; err != nil {
		return err
	}
	_ = s.driver.Delete(ctx, p.Pathname)
	// 解绑相册
	_ = s.db.WithContext(ctx).Where("photo_id = ?", p.ID).Delete(&model.AlbumPhoto{}).Error
	return nil
}

// BatchDelete 批量删除
func (s *PhotoService) BatchDelete(ctx context.Context, ids []uint64, userID uint64) (int, error) {
	if len(ids) == 0 {
		return 0, bizerr.BadRequest.WithMessage("ids required")
	}
	if len(ids) > 100 {
		return 0, bizerr.BadRequest.WithMessage("too many ids")
	}
	// 找出归属
	var rows []model.Photo
	if err := s.db.WithContext(ctx).Where("id IN ? AND user_id = ?", ids, userID).Find(&rows).Error; err != nil {
		return 0, err
	}
	if len(rows) == 0 {
		return 0, nil
	}
	// 删除数据库
	if err := s.db.WithContext(ctx).Where("id IN ?", ids).Delete(&model.Photo{}).Error; err != nil {
		return 0, err
	}
	// 删除存储
	for _, p := range rows {
		_ = s.driver.Delete(ctx, p.Pathname)
	}
	_ = s.db.WithContext(ctx).Where("photo_id IN ?", ids).Delete(&model.AlbumPhoto{}).Error
	return len(rows), nil
}

// MoveToAlbum 移入相册（albumID=0 表示移出）
func (s *PhotoService) MoveToAlbum(ctx context.Context, photoID, albumID, userID uint64) error {
	var p model.Photo
	if err := s.db.WithContext(ctx).Where("id = ? AND user_id = ?", photoID, userID).First(&p).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return bizerr.PhotoNotFound
		}
		return err
	}
	// 删旧
	if err := s.db.WithContext(ctx).Where("photo_id = ?", photoID).Delete(&model.AlbumPhoto{}).Error; err != nil {
		return err
	}
	if albumID == 0 {
		return nil
	}
	// 校验相册归属
	var a model.Album
	if err := s.db.WithContext(ctx).Where("id = ? AND user_id = ?", albumID, userID).First(&a).Error; err != nil {
		return bizerr.ResourceNotFound
	}
	return s.db.WithContext(ctx).Create(&model.AlbumPhoto{AlbumID: albumID, PhotoID: photoID}).Error
}

// Copy 复制（在新路径下创建另一份记录 + 物理文件）
func (s *PhotoService) Copy(ctx context.Context, photoID, userID uint64) (*model.Photo, error) {
	var src model.Photo
	if err := s.db.WithContext(ctx).Where("id = ? AND user_id = ?", photoID, userID).First(&src).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, bizerr.PhotoNotFound
		}
		return nil, err
	}
	// 读源文件
	rd, err := s.driver.Get(ctx, src.Pathname)
	if err != nil {
		return nil, fmt.Errorf("read source: %w", err)
	}
	defer rd.Close()

	// 重新生成 key
	now := time.Now()
	key := fmt.Sprintf("%s/%s/%s/%s/%s-%d.%s",
		s.imageDir,
		now.Format("2006"), now.Format("01"), now.Format("02"),
		src.MD5, time.Now().UnixNano()%1_000_000,
		strings.TrimPrefix(src.Extension, "."),
	)
	if _, err := s.driver.Put(ctx, key, rd, int64(src.Size*1024), src.Mimetype); err != nil {
		return nil, bizerr.StorageUploadFail.WithCause(err)
	}
	dup := &model.Photo{
		UserID:    src.UserID,
		Name:      src.Name + " (副本)",
		Intro:     src.Intro,
		Filename:  src.Filename,
		Pathname:  key,
		Mimetype:  src.Mimetype,
		Extension: src.Extension,
		MD5:       src.MD5,
		SHA1:      src.SHA1,
		Size:      src.Size,
		IsPublic:  src.IsPublic,
		Status:    model.PhotoStatusNormal,
	}
	if err := s.db.WithContext(ctx).Create(dup).Error; err != nil {
		_ = s.driver.Delete(ctx, key)
		return nil, err
	}
	return dup, nil
}

// UpdatePhotoReq 更新图片元数据
type UpdatePhotoReq struct {
	Name     *string `json:"name"`
	Intro    *string `json:"intro"`
	IsPublic *bool   `json:"is_public"`
}

// Update 更新
func (s *PhotoService) Update(ctx context.Context, id, userID uint64, req UpdatePhotoReq) (*model.Photo, error) {
	var p model.Photo
	if err := s.db.WithContext(ctx).Where("id = ? AND user_id = ?", id, userID).First(&p).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, bizerr.PhotoNotFound
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
		if err := s.db.WithContext(ctx).Model(&p).Updates(updates).Error; err != nil {
			return nil, err
		}
	}
	return &p, nil
}

// ServeRedirect 通过 ID 重定向到实际文件 URL
func (s *PhotoService) ServeRedirect(ctx context.Context, id uint64) (string, *model.Photo, error) {
	var p model.Photo
	if err := s.db.WithContext(ctx).First(&p, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", nil, bizerr.PhotoNotFound
		}
		return "", nil, err
	}
	if !p.IsPublic {
		// 非公开图片：必须有 user_id 上下文校验
		return "", nil, bizerr.Forbidden
	}
	return s.driver.PublicURL(p.Pathname), &p, nil
}

// 简单 mime 探测
func detectMIME(b []byte) string {
	if len(b) >= 3 && b[0] == 0xFF && b[1] == 0xD8 && b[2] == 0xFF {
		return "image/jpeg"
	}
	if len(b) >= 8 && string(b[:8]) == "\x89PNG\r\n\x1a\n" {
		return "image/png"
	}
	if len(b) >= 6 && (string(b[:6]) == "GIF87a" || string(b[:6]) == "GIF89a") {
		return "image/gif"
	}
	if len(b) >= 4 && string(b[:4]) == "RIFF" && len(b) >= 12 && string(b[8:12]) == "WEBP" {
		return "image/webp"
	}
	return "application/octet-stream"
}

func extensionFromMIME(mime, filename string) string {
	if ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(filename)), "."); ext != "" {
		return ext
	}
	switch mime {
	case "image/jpeg":
		return "jpg"
	case "image/png":
		return "png"
	case "image/gif":
		return "gif"
	case "image/webp":
		return "webp"
	}
	return "bin"
}
