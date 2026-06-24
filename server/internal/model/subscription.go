package model

// 订阅状态
const (
	SubscriptionStatusActive   = "active"   // 生效中
	SubscriptionStatusExpired  = "expired"  // 已过期
	SubscriptionStatusCanceled = "canceled" // 已取消
)

// Subscription 用户订阅
type Subscription struct {
	Base
	UserID    uint64 `gorm:"not null;index" json:"user_id"`
	PlanID    uint64 `gorm:"not null;index" json:"plan_id"`
	OrderID   uint64 `gorm:"not null;index" json:"order_id"`
	StartedAt int64  `gorm:"not null" json:"started_at"`
	ExpireAt  int64  `gorm:"not null;index" json:"expire_at"`
	Status    string `gorm:"size:32;not null;default:'active';index" json:"status"`
}

// TableName 指定表名
func (Subscription) TableName() string { return "subscriptions" }
