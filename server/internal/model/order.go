package model

// 订单类型
const (
	OrderTypePlan   = "plan"   // 套餐
	OrderTypeCharge = "charge" // 充值
)

// 订单状态
const (
	OrderStatusUnpaid   = "unpaid"   // 待支付
	OrderStatusPaid     = "paid"     // 已支付
	OrderStatusCanceled = "canceled" // 已取消
	OrderStatusRefunded = "refunded" // 已退款
)

// Order 订单
type Order struct {
	Base
	PlanID       *uint64 `gorm:"index" json:"plan_id,omitempty"`                        // 计划
	UserID       *uint64 `gorm:"index" json:"user_id,omitempty"`                        // 用户
	CouponID     *uint64 `gorm:"index" json:"coupon_id,omitempty"`                      // 优惠券
	TradeNo      string  `gorm:"size:64;not null;uniqueIndex" json:"trade_no"`          // 系统订单号
	OutTradeNo   string  `gorm:"size:64;not null;uniqueIndex" json:"out_trade_no"`      // 支付订单号
	Type         string  `gorm:"size:32;not null;default:'plan'" json:"type"`           // 类型
	Amount       uint    `gorm:"not null;default:0" json:"amount"`                      // 实际付款金额(分)
	DeductAmount uint    `gorm:"not null;default:0" json:"deduct_amount"`               // 抵扣金额(分)
	Snapshot     JSONMap `gorm:"type:json" json:"snapshot,omitempty"`                   // 产品快照
	Product      JSONMap `gorm:"type:json" json:"product,omitempty"`                    // 购买产品数据
	PayMethod    string  `gorm:"size:32;not null;default:''" json:"pay_method"`         // 支付方式
	Status       string  `gorm:"size:32;not null;default:'unpaid';index" json:"status"` // 状态
	PaidAt       *int64  `json:"paid_at,omitempty"`                                     // 支付时间
	CanceledAt   *int64  `json:"canceled_at,omitempty"`                                 // 取消时间
}

// TableName 指定表名
func (Order) TableName() string { return "orders" }
