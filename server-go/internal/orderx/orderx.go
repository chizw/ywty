// Package orderx 订单/套餐/优惠码业务。
package orderx

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/chizw/ywty/server-go/internal/config"
	"github.com/chizw/ywty/server-go/internal/db/types"
	"github.com/chizw/ywty/server-go/internal/model"
	"gorm.io/gorm"
)

// ---------- 金额/券码计算 ----------

// PreviewData preview 结果（对齐 OrderService::preview 返回数组）。
type PreviewData struct {
	PlanID       int64
	CouponID     *int64
	Snapshot     types.JSON
	Product      types.JSON
	Amount       int64
	DeductAmount int64
	Duration     int32
	PlanType     string
	GroupID      *int64 // plan_groups.group_id（vip 用）
	Capacity     float64
}

var (
	ErrPriceNotFound = errors.New("不存在该价格方案")
	ErrCouponInvalid = errors.New("优惠券不存在或已到期")
	ErrOrderNotFound = errors.New("订单不存在")
)

// Preview 计算价格方案 + 优惠码后的订单数据。
func Preview(gdb *gorm.DB, priceID int64, couponCode *string) (*PreviewData, error) {
	var price struct {
		ID       int64
		PlanID   int64
		Name     string
		Duration int32
		Price    int64
	}
	if err := gdb.Raw(
		"SELECT p.`id`, p.`plan_id`, p.`name`, p.`duration`, p.`price` FROM `plan_prices` p "+
			"INNER JOIN `plans` pl ON pl.id = p.plan_id AND pl.deleted_at IS NULL "+
			"WHERE p.`id` = ? LIMIT 1", priceID,
	).Scan(&price).Error; err != nil {
		return nil, err
	}
	if price.ID == 0 {
		return nil, ErrPriceNotFound
	}

	// plan 快照
	var plan model.Plan
	if err := gdb.First(&plan, price.PlanID).Error; err != nil {
		return nil, ErrPriceNotFound
	}
	planJSON, _ := json.Marshal(planResource(gdb, &plan))

	product, _ := json.Marshal(map[string]any{
		"id": price.ID, "plan_id": price.PlanID, "name": price.Name,
		"duration": price.Duration, "price": price.Price,
	})

	out := &PreviewData{
		PlanID:   price.PlanID,
		Snapshot: types.JSON(planJSON),
		Product:  types.JSON(product),
		Amount:   price.Price,
		Duration: price.Duration,
		PlanType: plan.Type,
	}

	// 优惠码
	if couponCode != nil && *couponCode != "" {
		var coupon model.Coupon
		err := gdb.Where("code = ? AND expired_at > ?", *couponCode, time.Now()).First(&coupon).Error
		if err != nil {
			return nil, ErrCouponInvalid
		}
		var used int64
		gdb.Raw("SELECT count(*) FROM `orders` WHERE `coupon_id` = ? AND `status` <> 'cancelled'", coupon.ID).Scan(&used)
		if used >= coupon.UsageLimit {
			return nil, ErrCouponInvalid
		}
		amount := couponAmountOf(&coupon, price.Price)
		out.CouponID = &coupon.ID
		out.DeductAmount = price.Price - amount
		if out.DeductAmount < 0 {
			out.DeductAmount = 0
		}
		out.Amount = amount
		if out.Amount < 0 {
			out.Amount = 0
		}
	}

	// plan 关联的组/容量
	var planGroup struct {
		GroupID *int64
	}
	gdb.Raw("SELECT `group_id` FROM `plan_groups` WHERE `plan_id` = ? LIMIT 1", price.PlanID).Scan(&planGroup)
	out.GroupID = planGroup.GroupID

	var planCap struct {
		Capacity float64
	}
	gdb.Raw("SELECT `capacity` FROM `plan_capacities` WHERE `plan_id` = ? LIMIT 1", price.PlanID).Scan(&planCap)
	out.Capacity = planCap.Capacity

	return out, nil
}

// couponAmountOf 优惠码抵扣后的金额（direct 直减 / percent 乘系数）。
func couponAmountOf(c *model.Coupon, price int64) int64 {
	switch c.Type {
	case "percent":
		return int64(float64(price) * c.Value)
	default: // direct
		return price - int64(c.Value)
	}
}

// planResource 快照结构（UserOrderResource 中 snapshot 只取 name/intro/features/badge，
// 但 OrderService::preview 的 snapshot 是整个 plan；这里保留整个 plan 的字段）。
func planResource(gdb *gorm.DB, plan *model.Plan) map[string]any {
	return map[string]any{
		"id": plan.ID, "type": plan.Type, "name": plan.Name,
		"intro": plan.Intro, "features": jsonOrNull(plan.Features),
		"badge": plan.Badge, "sort": plan.Sort, "is_up": plan.IsUp,
	}
}

func jsonOrNull(j types.JSON) any {
	if len(j) == 0 || string(j) == "null" {
		return nil
	}
	var v any
	if json.Unmarshal(j, &v) != nil {
		return nil
	}
	return v
}

// ---------- 订单号 ----------

// GenerateTradeNo 系统订单号：YmdHis + 5 位随机（对齐 Order::generateTradeNo）。
func GenerateTradeNo(gdb *gorm.DB) string {
	for {
		tradeNo := time.Now().Format("20060102150405") + padLeft(strconv.Itoa(1+rand.Intn(99999)), 5, '0')
		var n int64
		gdb.Raw("SELECT count(*) FROM `orders` WHERE `trade_no` = ?", tradeNo).Scan(&n)
		if n == 0 {
			return tradeNo
		}
	}
}

// GenerateOutTradeNo 支付订单号（更长的随机段）。
func GenerateOutTradeNo(gdb *gorm.DB) string {
	for {
		no := time.Now().Format("20060102150405") + padLeft(strconv.Itoa(1+rand.Intn(99999999)), 8, '0')
		var n int64
		gdb.Raw("SELECT count(*) FROM `orders` WHERE `out_trade_no` = ?", no).Scan(&n)
		if n == 0 {
			return no
		}
	}
}

func padLeft(s string, n int, c byte) string {
	for len(s) < n {
		s = string(c) + s
	}
	return s
}

// ---------- 创建/完成/取消 ----------

// Create 创建订单；金额为 0 时直接完成（对齐 OrderService::create）。
func Create(gdb *gorm.DB, cfg *config.Config, userID int64, priceID int64, couponCode *string) (*model.Order, bool, error) {
	pv, err := Preview(gdb, priceID, couponCode)
	if err != nil {
		return nil, false, err
	}
	now := time.Now().UTC()
	order := model.Order{
		PlanID: &pv.PlanID, UserID: &userID, CouponID: pv.CouponID,
		TradeNo: GenerateTradeNo(gdb), OutTradeNo: GenerateOutTradeNo(gdb),
		Type: "plan", Amount: pv.Amount, DeductAmount: pv.DeductAmount,
		Snapshot: pv.Snapshot, Product: pv.Product,
		Status: "unpaid", CreatedAt: &now, UpdatedAt: &now,
	}
	if err := gdb.Create(&order).Error; err != nil {
		return nil, false, err
	}
	if order.Amount <= 0 {
		if err := Complete(gdb, cfg, &order, ""); err != nil {
			return nil, false, err
		}
		return &order, true, nil
	}
	return &order, false, nil
}

// Complete 支付完成：发放套餐/容量（对齐 OrderService::complete）。
func Complete(gdb *gorm.DB, cfg *config.Config, order *model.Order, payMethod string) error {
	if order.PlanID == nil || *order.PlanID == 0 {
		return errors.New("订单缺少套餐信息")
	}
	var plan model.Plan
	if err := gdb.First(&plan, *order.PlanID).Error; err != nil {
		return err
	}
	product := map[string]any{}
	_ = json.Unmarshal(order.Product, &product)
	durationMin := 0
	if d, ok := product["duration"].(float64); ok {
		durationMin = int(d)
	}
	expiredAt := time.Now().Add(time.Duration(durationMin) * time.Minute)

	if plan.Type == "vip" {
		var planGroup struct {
			GroupID *int64
		}
		gdb.Raw("SELECT `group_id` FROM `plan_groups` WHERE `plan_id` = ? LIMIT 1", plan.ID).Scan(&planGroup)
		if planGroup.GroupID != nil && *planGroup.GroupID > 0 {
			oid := order.ID
			ug := model.UserGroup{
				UserID: *order.UserID, GroupID: *planGroup.GroupID, OrderID: &oid,
				From: "subscribe", ExpiredAt: &expiredAt,
			}
			if err := gdb.Create(&ug).Error; err != nil {
				return fmt.Errorf("发放用户组失败: %w", err)
			}
		}
	}
	if plan.Type == "storage" {
		var planCap struct {
			Capacity float64
		}
		gdb.Raw("SELECT `capacity` FROM `plan_capacities` WHERE `plan_id` = ? LIMIT 1", plan.ID).Scan(&planCap)
		oid := order.ID
		uc := model.UserCapacity{
			UserID: *order.UserID, OrderID: &oid, Capacity: planCap.Capacity,
			From: "subscribe", ExpiredAt: &expiredAt,
		}
		if err := gdb.Create(&uc).Error; err != nil {
			return fmt.Errorf("发放容量失败: %w", err)
		}
	}

	now := time.Now().UTC()
	return gdb.Model(&model.Order{}).Where("id = ?", order.ID).Updates(map[string]any{
		"paid_at": now, "status": "paid", "pay_method": payMethod, "updated_at": now,
	}).Error
}

// Cancel 取消未支付订单。
func Cancel(gdb *gorm.DB, order *model.Order) error {
	now := time.Now().UTC()
	return gdb.Model(&model.Order{}).Where("id = ?", order.ID).Updates(map[string]any{
		"canceled_at": now, "status": "cancelled", "updated_at": now,
	}).Error
}

// CancelIfUnpaid 延时取消任务：仅取消仍未支付的订单。
func CancelIfUnpaid(gdb *gorm.DB, orderID int64) error {
	var order model.Order
	if err := gdb.First(&order, orderID).Error; err != nil {
		return nil
	}
	if order.Status != "unpaid" {
		return nil
	}
	return Cancel(gdb, &order)
}

// GetByTradeNo 按系统订单号取用户订单。
func GetByTradeNo(gdb *gorm.DB, userID int64, tradeNo string) (*model.Order, error) {
	var order model.Order
	err := gdb.Where("user_id = ? AND trade_no = ?", userID, tradeNo).First(&order).Error
	if err != nil {
		return nil, ErrOrderNotFound
	}
	return &order, nil
}

// OrderResource 对齐 UserOrderResource 可见字段。
func OrderResource(gdb *gorm.DB, o *model.Order) map[string]any {
	// snapshot/product 只保留部分字段
	snapshot := map[string]any{}
	_ = json.Unmarshal(o.Snapshot, &snapshot)
	snapOut := map[string]any{
		"name": snapshot["name"], "intro": snapshot["intro"],
		"features": snapshot["features"], "badge": snapshot["badge"],
	}
	product := map[string]any{}
	_ = json.Unmarshal(o.Product, &product)
	prodOut := map[string]any{
		"name": product["name"], "duration": product["duration"], "price": product["price"],
	}

	var coupon any
	if o.CouponID != nil && *o.CouponID > 0 {
		var c model.Coupon
		if err := gdb.First(&c, *o.CouponID).Error; err == nil {
			coupon = map[string]any{"name": c.Name, "code": c.Code}
		}
	}

	return map[string]any{
		"trade_no": o.TradeNo, "coupon": coupon, "amount": o.Amount,
		"snapshot": snapOut, "product": prodOut, "pay_method": o.PayMethod,
		"deduct_amount": o.DeductAmount, "status": o.Status,
		"paid_at": timePtr(o.PaidAt), "canceled_at": timePtr(o.CanceledAt),
		"created_at": timePtr(o.CreatedAt),
	}
}

func timePtr(t *time.Time) any {
	if t == nil {
		return nil
	}
	return t.Format(time.RFC3339)
}
