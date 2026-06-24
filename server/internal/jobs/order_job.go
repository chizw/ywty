// Package jobs 订单异步任务
package jobs

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/ywty/server/internal/logger"
	"github.com/ywty/server/internal/model"
	"github.com/ywty/server/internal/queue"
)

// OrderCancelPayload 订单取消任务参数
type OrderCancelPayload struct {
	OrderID string `json:"order_id"`
	Reason  string `json:"reason"`
}

// OrderPaidPayload 订单支付完成任务参数
type OrderPaidPayload struct {
	OrderID string `json:"order_id"`
	TradeNo string `json:"trade_no"`
}

// NewOrderCancelTask 创建订单取消任务
func NewOrderCancelTask(payload OrderCancelPayload) (*asynq.Task, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(queue.TypeOrderCancel, data), nil
}

// NewOrderPaidTask 创建订单完成任务
func NewOrderPaidTask(payload OrderPaidPayload) (*asynq.Task, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(queue.TypeOrderPaid, data), nil
}

// HandleOrderCancel 取消超时未支付订单
func (h *Handlers) HandleOrderCancel(ctx context.Context, t *asynq.Task) error {
	var p OrderCancelPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}

	var order model.Order
	if err := h.DB.WithContext(ctx).Where("trade_no = ? OR out_trade_no = ?", p.OrderID, p.OrderID).First(&order).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}

	// 已支付/已取消则跳过
	if order.Status != model.OrderStatusUnpaid {
		return nil
	}

	now := time.Now().Unix()
	if err := h.DB.WithContext(ctx).Model(&order).Updates(map[string]any{
		"status":       model.OrderStatusCanceled,
		"canceled_at":  now,
	}).Error; err != nil {
		return err
	}

	logger.L.Info("order cancelled",
		zap.String("trade_no", order.TradeNo),
		zap.String("reason", p.Reason),
	)
	return nil
}

// HandleOrderPaid 处理订单完成：发放套餐容量给用户
func (h *Handlers) HandleOrderPaid(ctx context.Context, t *asynq.Task) error {
	var p OrderPaidPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}

	var order model.Order
	if err := h.DB.WithContext(ctx).Where("trade_no = ? OR out_trade_no = ?", p.OrderID, p.OrderID).First(&order).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}

	// 已处理过则跳过
	if order.Status == model.OrderStatusPaid {
		return nil
	}

	// 事务：更新订单 + 创建订阅 + 容量
	err := h.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now().Unix()
		if err := tx.Model(&order).Updates(map[string]any{
			"status":   model.OrderStatusPaid,
			"trade_no": p.TradeNo,
			"paid_at":  now,
		}).Error; err != nil {
			return err
		}

		if order.PlanID == 0 {
			return nil
		}

		var plan model.Plan
		if err := tx.First(&plan, order.PlanID).Error; err != nil {
			return err
		}

		// 套餐时长：从关联的 PlanPrice 取（默认 30 天）
		duration := 30
		var pp model.PlanPrice
		if order.Product != nil {
			if v, ok := order.Product["plan_price_id"]; ok {
				if pid, ok := v.(float64); ok {
					if err := tx.First(&pp, uint64(pid)).Error; err == nil {
						duration = pp.Duration
					}
				}
			}
		}

		// 套餐容量：从关联的 PlanCapacity 取
		var capacity int64
		var pc model.PlanCapacity
		if err := tx.Where("plan_id = ?", order.PlanID).First(&pc).Error; err == nil {
			capacity = pc.Capacity
		}

		// 创建订阅
		sub := model.Subscription{
			UserID:    order.UserID,
			PlanID:    order.PlanID,
			OrderID:   order.ID,
			StartedAt: now,
			ExpireAt:  time.Now().AddDate(0, 0, duration).Unix(),
			Status:    model.SubscriptionStatusActive,
		}
		if err := tx.Create(&sub).Error; err != nil {
			return err
		}

		// 调整用户容量：累加 plan 容量
		if capacity > 0 {
			if err := tx.Model(&model.User{}).
				Where("id = ?", order.UserID).
				Update("storage_quota", gorm.Expr("COALESCE(storage_quota, 0) + ?", capacity)).Error; err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return err
	}

	logger.L.Info("order paid processed",
		zap.String("trade_no", order.TradeNo),
		zap.Uint64("user_id", order.UserID),
	)
	return nil
}
