// Package jobs 异步任务实现（被 worker 进程消费）
package jobs

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"path/filepath"
	"strings"
	"time"

	"github.com/disintegration/imaging"
	"github.com/hibiken/asynq"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/ywty/server/internal/logger"
	"github.com/ywty/server/internal/model"
	"github.com/ywty/server/internal/queue"
	"github.com/ywty/server/internal/storage"
)

// ThumbnailPayload 缩略图任务参数
type ThumbnailPayload struct {
	PhotoID  uint64 `json:"photo_id"`
	Width    int    `json:"width"`
	Height   int    `json:"height"`
	Strategy string `json:"strategy"` // fit | fill | stretch
}

// ProcessPayload 图片处理任务（resize / watermark）
type ProcessPayload struct {
	PhotoID  uint64 `json:"photo_id"`
	MaxWidth int    `json:"max_width"`
	Quality  int    `json:"quality"`
}

// AutoDeletePayload 自动删除任务
type AutoDeletePayload struct {
	BeforeTime int64 `json:"before_time"` // 删除此时间之前 + soft_deleted 的图片
}

// Handlers 任务处理器集合
type Handlers struct {
	DB     *gorm.DB
	Driver storage.Driver
}

// Register 注册所有任务到 mux
func (h *Handlers) Register(mux *asynq.ServeMux) {
	mux.HandleFunc(queue.TypeThumbnail, h.HandleThumbnail)
	mux.HandleFunc(queue.TypeProcess, h.HandleProcess)
	mux.HandleFunc(queue.TypeAutoDelete, h.HandleAutoDelete)
	mux.HandleFunc(queue.TypeTicketNotify, h.HandleTicketNotify)
	mux.HandleFunc(queue.TypeOrderCancel, h.HandleOrderCancel)
	mux.HandleFunc(queue.TypeOrderPaid, h.HandleOrderPaid)
}

// HandleThumbnail 生成缩略图
func (h *Handlers) HandleThumbnail(ctx context.Context, t *asynq.Task) error {
	var p ThumbnailPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}
	if p.Width == 0 {
		p.Width = 320
	}
	if p.Height == 0 {
		p.Height = 320
	}
	if p.Strategy == "" {
		p.Strategy = "fit"
	}

	var photo model.Photo
	if err := h.DB.First(&photo, p.PhotoID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil // 照片已不存在，跳过
		}
		return err
	}

	// 仅当 driver 为本地存储时才做本地缩略图
	if h.Driver == nil || h.Driver.Name() != storage.DriverNameLocal {
		logger.L.Info("thumbnail skipped: driver not local",
			zap.Uint64("photo_id", p.PhotoID),
		)
		return nil
	}

	rd, err := h.Driver.Get(ctx, photo.Pathname)
	if err != nil {
		return fmt.Errorf("read source: %w", err)
	}
	defer rd.Close()

	img, _, err := image.Decode(rd)
	if err != nil {
		return fmt.Errorf("decode image: %w", err)
	}

	var thumb *image.NRGBA
	switch p.Strategy {
	case "fill":
		thumb = imaging.Fill(img, p.Width, p.Height, imaging.Center, imaging.Lanczos)
	case "stretch":
		thumb = imaging.Resize(img, p.Width, p.Height, imaging.Lanczos)
	default:
		thumb = imaging.Fit(img, p.Width, p.Height, imaging.Lanczos)
	}

	buf := &bytes.Buffer{}
	if err := imaging.Encode(buf, thumb, imaging.JPEG, imaging.JPEGQuality(85)); err != nil {
		return fmt.Errorf("encode jpeg: %w", err)
	}

	dir := filepath.Dir(photo.Pathname)
	base := strings.TrimSuffix(filepath.Base(photo.Pathname), filepath.Ext(photo.Pathname))
	thumbKey := filepath.ToSlash(filepath.Join(dir, ".thumbs", base+".jpg"))

	if _, err := h.Driver.Put(ctx, thumbKey, buf, int64(buf.Len()), "image/jpeg"); err != nil {
		return fmt.Errorf("save thumb: %w", err)
	}

	if err := h.DB.Model(&photo).Update("thumb_path", thumbKey).Error; err != nil {
		logger.L.Warn("save thumb_path failed", zap.Uint64("photo_id", p.PhotoID), zap.Error(err))
	}
	logger.L.Info("thumbnail generated",
		zap.Uint64("photo_id", p.PhotoID),
		zap.String("key", thumbKey),
	)
	return nil
}

// HandleProcess 图片处理（resize + 压缩）
func (h *Handlers) HandleProcess(ctx context.Context, t *asynq.Task) error {
	var p ProcessPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}
	if p.MaxWidth == 0 {
		p.MaxWidth = 1920
	}
	if p.Quality == 0 {
		p.Quality = 90
	}

	var photo model.Photo
	if err := h.DB.First(&photo, p.PhotoID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}
	if h.Driver == nil || h.Driver.Name() != storage.DriverNameLocal {
		return nil
	}
	rd, err := h.Driver.Get(ctx, photo.Pathname)
	if err != nil {
		return fmt.Errorf("read: %w", err)
	}
	defer rd.Close()
	img, _, err := image.Decode(rd)
	if err != nil {
		return fmt.Errorf("decode: %w", err)
	}
	b := img.Bounds()
	if b.Dx() <= p.MaxWidth {
		logger.L.Info("process skipped: smaller than max", zap.Uint64("photo_id", p.PhotoID))
		return nil
	}
	resized := imaging.Resize(img, p.MaxWidth, 0, imaging.Lanczos)
	buf := &bytes.Buffer{}
	if err := imaging.Encode(buf, resized, imaging.JPEG, imaging.JPEGQuality(p.Quality)); err != nil {
		return fmt.Errorf("encode: %w", err)
	}
	if _, err := h.Driver.Put(ctx, photo.Pathname, buf, int64(buf.Len()), "image/jpeg"); err != nil {
		return fmt.Errorf("put: %w", err)
	}
	newSize := float64(buf.Len()) / 1024
	if err := h.DB.Model(&photo).Update("size", newSize).Error; err != nil {
		return err
	}
	logger.L.Info("photo processed",
		zap.Uint64("photo_id", p.PhotoID),
		zap.String("new_path", photo.Pathname),
	)
	return nil
}

// HandleAutoDelete 清理已软删超过 N 天的图片（含存储文件）
func (h *Handlers) HandleAutoDelete(ctx context.Context, t *asynq.Task) error {
	var p AutoDeletePayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}
	if p.BeforeTime == 0 {
		p.BeforeTime = time.Now().AddDate(0, 0, -30).Unix() // 默认 30 天前
	}

	var photos []model.Photo
	if err := h.DB.Unscoped().
		Where("deleted_at IS NOT NULL AND deleted_at < ?", p.BeforeTime).
		Limit(500).
		Find(&photos).Error; err != nil {
		return err
	}
	if len(photos) == 0 {
		return nil
	}
	ids := make([]uint64, 0, len(photos))
	for i := range photos {
		ids = append(ids, photos[i].ID)
		if h.Driver != nil {
			_ = h.Driver.Delete(ctx, photos[i].Pathname)
			if photos[i].ThumbPath != "" {
				_ = h.Driver.Delete(ctx, photos[i].ThumbPath)
			}
		}
	}
	if err := h.DB.Unscoped().Where("id IN ?", ids).Delete(&model.Photo{}).Error; err != nil {
		return err
	}
	logger.L.Info("auto delete done", zap.Int("count", len(ids)))
	return nil
}
