// Package service 容量服务：统计用户已用 / 总配额 / 是否超额
package service

import (
	"context"
	"errors"
	"math"
	"time"

	"gorm.io/gorm"

	"github.com/ywty/server/internal/model"
)

// 默认配额（开发期）：单图 10MB / 总容量 100MB
const (
	defaultMaxImageSize int64 = 10 * 1024 * 1024
	defaultCapacity     int64 = 100 * 1024 * 1024
)

// CapacityService 容量服务
type CapacityService struct {
	db *gorm.DB
}

func NewCapacityService(db *gorm.DB) *CapacityService { return &CapacityService{db: db} }

// Info 用户容量信息
type CapacityInfo struct {
	UserID    uint64 `json:"user_id"`
	Used      int64  `json:"used"`         // 已用字节
	Capacity  int64  `json:"capacity"`     // 配额
	MaxImage  int64  `json:"max_image"`    // 单图上限
	UsedPct   int    `json:"used_percent"` // 已用百分比 0-100
	Remain    int64  `json:"remain"`       // 剩余
	Unlimited bool   `json:"unlimited"`    // 是否无限
}

// GetUserCapacity 取用户的容量配置
func (s *CapacityService) GetUserCapacity(ctx context.Context, userID uint64) (*CapacityInfo, error) {
	// 1) 查用户
	var u model.User
	if err := s.db.WithContext(ctx).First(&u, userID).Error; err != nil {
		return nil, err
	}
	// 2) 找用户所在 group
	var ug model.UserGroup
	maxImage := defaultMaxImageSize
	capacity := defaultCapacity
	unlimited := false

	hasUG := errors.Is(s.db.WithContext(ctx).Where("user_id = ?", userID).First(&ug).Error, gorm.ErrRecordNotFound) == false
	if hasUG {
		var g model.Group
		if err := s.db.WithContext(ctx).First(&g, ug.GroupID).Error; err == nil {
			if v, ok := g.Options["max_image_size"]; ok {
				if n, ok := v.(float64); ok && n > 0 {
					maxImage = int64(n)
				}
			}
			if v, ok := g.Options["capacity"]; ok {
				if n, ok := v.(float64); ok && n > 0 {
					capacity = int64(n)
				} else if n == -1 {
					unlimited = true
				}
			}
		}
	}
	// 管理员无限
	if u.IsAdmin {
		unlimited = true
	}
	// 3) 计算已用
	var used int64
	_ = s.db.WithContext(ctx).Model(&model.Photo{}).Where("user_id = ?", userID).
		Select("COALESCE(SUM(size), 0)").Scan(&used).Error
	// 兼容 float64 字段：使用 KB
	if used > 0 && used < 1e9 {
		used = int64(float64(used) * 1024) // 数据库里是 KB，转字节
	}

	remain := capacity - used
	if remain < 0 {
		remain = 0
	}
	pct := 0
	if capacity > 0 && !unlimited {
		pct = int(math.Min(100, float64(used)*100/float64(capacity)))
	}
	return &CapacityInfo{
		UserID:    userID,
		Used:      used,
		Capacity:  capacity,
		MaxImage:  maxImage,
		UsedPct:   pct,
		Remain:    remain,
		Unlimited: unlimited,
	}, nil
}

// CheckUpload 检查是否能上传 size 字节
func (s *CapacityService) CheckUpload(ctx context.Context, userID uint64, sizeBytes int64) error {
	info, err := s.GetUserCapacity(ctx, userID)
	if err != nil {
		return err
	}
	if info.Unlimited {
		return nil
	}
	if sizeBytes > info.MaxImage {
		return errors.New("image too large")
	}
	if sizeBytes > info.Remain {
		return errors.New("capacity exhausted")
	}
	return nil
}

// Recalculate 重新统计（供 admin 手动触发）
func (s *CapacityService) Recalculate(ctx context.Context, userID uint64) (*CapacityInfo, error) {
	// 直接复用 GetUserCapacity（已实时统计）
	return s.GetUserCapacity(ctx, userID)
}

// Quota 配额同步任务（每日执行）
func (s *CapacityService) Quota(_ context.Context) error {
	// 简化：实际可在 Worker 中调用以记录历史快照
	_ = time.Now()
	return nil
}
