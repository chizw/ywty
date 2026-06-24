// Package service 优惠券服务
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

// CouponService 优惠券服务
type CouponService struct {
	db *gorm.DB
}

// NewCouponService 创建优惠券服务
func NewCouponService(db *gorm.DB) *CouponService {
	return &CouponService{db: db}
}

// AdminCouponReq 优惠券创建/更新请求
type AdminCouponReq struct {
	Type       string  `json:"type" binding:"required,oneof=direct discount"`
	Name       string  `json:"name" binding:"required"`
	Code       string  `json:"code" binding:"required"`
	Value      float64 `json:"value" binding:"required"`
	UsageLimit uint    `json:"usage_limit"`
	ExpiredAt  *int64  `json:"expired_at"`
}

// List 分页列表
func (s *CouponService) List(ctx context.Context, page, perPage int) ([]model.Coupon, int64, error) {
	page, perPage = normalizePage(page, perPage)

	var total int64
	q := s.db.WithContext(ctx).Model(&model.Coupon{})
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []model.Coupon
	if err := q.Order("id DESC").Offset((page - 1) * perPage).Limit(perPage).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

// Get 根据 ID 查询
func (s *CouponService) Get(ctx context.Context, id uint64) (*model.Coupon, error) {
	var coupon model.Coupon
	if err := s.db.WithContext(ctx).First(&coupon, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, bizerr.ResourceNotFound
		}
		return nil, err
	}
	return &coupon, nil
}

// Create 创建优惠券
func (s *CouponService) Create(ctx context.Context, req AdminCouponReq) (*model.Coupon, error) {
	coupon := &model.Coupon{
		Type:       req.Type,
		Name:       req.Name,
		Code:       req.Code,
		Value:      req.Value,
		UsageLimit: req.UsageLimit,
		ExpiredAt:  req.ExpiredAt,
	}
	if err := s.db.WithContext(ctx).Create(coupon).Error; err != nil {
		if isDuplicateKeyError(err) {
			return nil, bizerr.AlreadyExists.WithMessage("优惠券码已存在")
		}
		return nil, fmt.Errorf("create coupon: %w", err)
	}
	return coupon, nil
}

// Update 更新优惠券
func (s *CouponService) Update(ctx context.Context, id uint64, req AdminCouponReq) (*model.Coupon, error) {
	coupon, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Code != coupon.Code {
		var cnt int64
		if err := s.db.WithContext(ctx).Model(&model.Coupon{}).
			Where("code = ? AND id != ?", req.Code, id).Count(&cnt).Error; err != nil {
			return nil, err
		}
		if cnt > 0 {
			return nil, bizerr.AlreadyExists.WithMessage("优惠券码已存在")
		}
	}

	updates := map[string]any{
		"type":        req.Type,
		"name":        req.Name,
		"code":        req.Code,
		"value":       req.Value,
		"usage_limit": req.UsageLimit,
		"expired_at":  req.ExpiredAt,
	}
	if err := s.db.WithContext(ctx).Model(&model.Coupon{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		if isDuplicateKeyError(err) {
			return nil, bizerr.AlreadyExists.WithMessage("优惠券码已存在")
		}
		return nil, err
	}

	return s.Get(ctx, id)
}

// Delete 删除优惠券
func (s *CouponService) Delete(ctx context.Context, id uint64) error {
	res := s.db.WithContext(ctx).Delete(&model.Coupon{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return bizerr.ResourceNotFound
	}
	return nil
}

// Validate 校验券码并计算抵扣金额（分）
// direct 类型抵扣 value（分）；discount 类型抵扣 amount * (1 - value/100)
// 校验通过后会将 usage_limit 减一
func (s *CouponService) Validate(ctx context.Context, code string, amount uint) (*model.Coupon, uint, error) {
	var coupon model.Coupon
	if err := s.db.WithContext(ctx).Where("code = ?", code).First(&coupon).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, 0, bizerr.ResourceNotFound.WithMessage("优惠券不存在")
		}
		return nil, 0, err
	}

	if coupon.ExpiredAt != nil && time.Now().Unix() > *coupon.ExpiredAt {
		return nil, 0, bizerr.BadRequest.WithMessage("优惠券已过期")
	}
	if coupon.UsageLimit == 0 {
		return nil, 0, bizerr.BadRequest.WithMessage("优惠券使用次数已用完")
	}

	var discount uint
	switch coupon.Type {
	case model.CouponTypeDirect:
		discount = uint(coupon.Value)
	case model.CouponTypeDiscount:
		discount = uint(float64(amount) * (100.0 - coupon.Value) / 100.0)
	default:
		return nil, 0, bizerr.BadRequest.WithMessage("优惠券类型无效")
	}
	if discount > amount {
		discount = amount
	}

	if err := s.db.WithContext(ctx).Model(&model.Coupon{}).
		Where("id = ?", coupon.ID).
		UpdateColumn("usage_limit", gorm.Expr("usage_limit - ?", 1)).Error; err != nil {
		return nil, 0, err
	}
	if coupon.UsageLimit > 0 {
		coupon.UsageLimit--
	}

	return &coupon, discount, nil
}

// isDuplicateKeyError 判断是否为唯一键冲突错误
func isDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, gorm.ErrDuplicatedKey)
}
