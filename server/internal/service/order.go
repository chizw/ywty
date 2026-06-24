// Package service 订单与支付服务
package service

import (
	"context"
	"crypto/rand"
	"fmt"
	"math"
	"strings"
	"time"

	"gorm.io/gorm"

	bizerr "github.com/ywty/server/internal/errors"
	"github.com/ywty/server/internal/model"
	"github.com/ywty/server/internal/payment"
)

// OrderService 订单服务
type OrderService struct {
	db      *gorm.DB
	baseURL string
}

// NewOrderService 创建订单服务
func NewOrderService(db *gorm.DB, baseURL string) *OrderService {
	return &OrderService{db: db, baseURL: strings.TrimRight(baseURL, "/")}
}

// CreateOrderReq 创建订单请求
type CreateOrderReq struct {
	PlanID     uint64 `json:"plan_id" binding:"required"`
	PriceID    uint64 `json:"price_id" binding:"required"`
	CouponCode string `json:"coupon_code"`
	PayMethod  string `json:"pay_method" binding:"required"`
}

// CreatePlanOrder 创建套餐订单
func (s *OrderService) CreatePlanOrder(ctx context.Context, userID uint64, req CreateOrderReq) (*model.Order, error) {
	var plan model.Plan
	if err := s.db.WithContext(ctx).First(&plan, req.PlanID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, bizerr.ResourceNotFound.WithMessage("套餐不存在")
		}
		return nil, err
	}

	var price model.PlanPrice
	if err := s.db.WithContext(ctx).Where("id = ? AND plan_id = ?", req.PriceID, req.PlanID).First(&price).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, bizerr.ResourceNotFound.WithMessage("套餐价格不存在")
		}
		return nil, err
	}

	amount := uint(price.Price)
	var couponID *uint64
	if req.CouponCode != "" {
		var c model.Coupon
		if err := s.db.WithContext(ctx).Where("code = ?", req.CouponCode).First(&c).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil, bizerr.ResourceNotFound.WithMessage("优惠券不存在")
			}
			return nil, err
		}
		now := time.Now().Unix()
		if c.ExpiredAt != nil && *c.ExpiredAt < now {
			return nil, bizerr.BadRequest.WithMessage("优惠券已过期")
		}
		if c.UsageLimit == 0 {
			return nil, bizerr.BadRequest.WithMessage("优惠券已用完")
		}
		couponID = &c.ID
		switch c.Type {
		case model.CouponTypeDirect:
			amount = uint(math.Max(0, float64(amount)-math.Round(c.Value*100)))
		case model.CouponTypeDiscount:
			amount = uint(math.Max(0, math.Round(float64(amount)*c.Value)))
		}
	}

	tradeNo := s.GenerateTradeNo(userID)
	expiredAt := time.Now().Add(30 * time.Minute).Unix()

	order := &model.Order{
		PlanID:       &req.PlanID,
		UserID:       &userID,
		CouponID:     couponID,
		TradeNo:      tradeNo,
		OutTradeNo:   tradeNo,
		Type:         model.OrderTypePlan,
		Amount:       amount,
		DeductAmount: uint(price.Price) - amount,
		Snapshot: model.JSONMap{
			"plan": map[string]any{
				"id":   plan.ID,
				"name": plan.Name,
				"type": plan.Type,
			},
			"price": map[string]any{
				"id":       price.ID,
				"name":     price.Name,
				"duration": price.Duration,
				"price":    price.Price,
			},
			"expired_at": expiredAt,
		},
		Product: model.JSONMap{
			"plan_id":  req.PlanID,
			"price_id": req.PriceID,
		},
		PayMethod: req.PayMethod,
		Status:    model.OrderStatusUnpaid,
	}

	if err := s.db.WithContext(ctx).Create(order).Error; err != nil {
		return nil, fmt.Errorf("create order: %w", err)
	}
	return order, nil
}

// List 列出我的订单
func (s *OrderService) List(ctx context.Context, userID uint64, page, perPage int) ([]model.Order, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}
	var rows []model.Order
	err := s.db.WithContext(ctx).Where("user_id = ?", userID).Order("id DESC").
		Offset((page - 1) * perPage).Limit(perPage).Find(&rows).Error
	return rows, err
}

// Get 获取订单详情
func (s *OrderService) Get(ctx context.Context, userID, orderID uint64) (*model.Order, error) {
	var o model.Order
	if err := s.db.WithContext(ctx).Where("id = ? AND user_id = ?", orderID, userID).First(&o).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, bizerr.OrderNotFound
		}
		return nil, err
	}
	return &o, nil
}

// Cancel 取消未支付订单
func (s *OrderService) Cancel(ctx context.Context, userID, orderID uint64) error {
	var o model.Order
	if err := s.db.WithContext(ctx).Where("id = ? AND user_id = ?", orderID, userID).First(&o).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return bizerr.OrderNotFound
		}
		return err
	}
	if o.Status != model.OrderStatusUnpaid {
		return bizerr.OrderCanceled.WithMessage("订单状态不允许取消")
	}
	now := time.Now().Unix()
	return s.db.WithContext(ctx).Model(&o).Updates(map[string]any{
		"status":      model.OrderStatusCanceled,
		"canceled_at": now,
	}).Error
}

// Pay 发起支付
func (s *OrderService) Pay(ctx context.Context, userID, orderID uint64, clientIP string) (*payment.CreateResult, error) {
	var o model.Order
	if err := s.db.WithContext(ctx).Where("id = ? AND user_id = ?", orderID, userID).First(&o).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, bizerr.OrderNotFound
		}
		return nil, err
	}
	if o.Status != model.OrderStatusUnpaid {
		return nil, bizerr.OrderPaid.WithMessage("订单不可支付")
	}

	driver, err := payment.Get(o.PayMethod, nil)
	if err != nil {
		return nil, bizerr.PaymentChannelInvalid
	}

	cfg := map[string]string{
		"return_url": s.baseURL + "/orders/result",
		"notify_url": s.baseURL + "/api/v1/orders/notify",
		"client_ip":  clientIP,
	}
	if planSnap, ok := o.Snapshot["plan"].(map[string]any); ok {
		if name, ok := planSnap["name"].(string); ok && name != "" {
			cfg["subject"] = name
		}
	}

	return driver.CreatePayment(ctx, o, cfg)
}

// HandlePaid 处理支付成功
func (s *OrderService) HandlePaid(ctx context.Context, tradeNo, outTradeNo, payMethod string) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var o model.Order
		if err := tx.Where("trade_no = ?", tradeNo).First(&o).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return bizerr.OrderNotFound
			}
			return err
		}
		if o.Status == model.OrderStatusPaid {
			return nil
		}
		if o.Status != model.OrderStatusUnpaid {
			return bizerr.OrderCanceled.WithMessage("订单状态不允许支付")
		}

		now := time.Now().Unix()
		if err := tx.Model(&o).Updates(map[string]any{
			"status":       model.OrderStatusPaid,
			"paid_at":      now,
			"pay_method":   payMethod,
			"out_trade_no": outTradeNo,
		}).Error; err != nil {
			return err
		}

		if o.PlanID == nil || o.UserID == nil {
			return nil
		}

		var capacities []model.PlanCapacity
		if err := tx.Where("plan_id = ?", *o.PlanID).Find(&capacities).Error; err != nil {
			return err
		}
		for _, pc := range capacities {
			uc := model.UserCapacity{
				UserID:   *o.UserID,
				OrderID:  &o.ID,
				Capacity: pc.Capacity,
				From:     model.CapacityFromOrder,
			}
			if err := tx.Create(&uc).Error; err != nil {
				return err
			}
		}

		var groups []model.PlanGroup
		if err := tx.Where("plan_id = ?", *o.PlanID).Find(&groups).Error; err != nil {
			return err
		}
		for _, pg := range groups {
			if pg.GroupID != nil {
				ug := model.UserGroup{
					UserID:  *o.UserID,
					GroupID: *pg.GroupID,
					OrderID: &o.ID,
					From:    model.GroupFromOrder,
				}
				if err := tx.Create(&ug).Error; err != nil {
					return err
				}
			}
		}

		if o.CouponID != nil {
			if err := tx.Model(&model.Coupon{}).Where("id = ?", *o.CouponID).
				Update("usage_limit", gorm.Expr("usage_limit - ?", 1)).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

// GenerateTradeNo 生成订单号
func (s *OrderService) GenerateTradeNo(userID uint64) string {
	return fmt.Sprintf("%s%d%s", time.Now().Format("20060102150405"), userID, randomString(6))
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	for i := range b {
		b[i] = letters[int(b[i])%len(letters)]
	}
	return string(b)
}
