package model

// 优惠券类型
const (
	CouponTypeDirect   = "direct"   // 直接抵扣
	CouponTypeDiscount = "discount" // 折扣
)

// Coupon 优惠券
type Coupon struct {
	Base
	Type       string  `gorm:"size:32;not null;default:'direct'" json:"type"`     // 折扣类型
	Name       string  `gorm:"size:32;not null;default:''" json:"name"`           // 名称
	Code       string  `gorm:"size:64;not null;uniqueIndex" json:"code"`          // 券码
	Value      float64 `gorm:"type:decimal(8,2);not null;default:0" json:"value"` // 金额或折扣率
	UsageLimit uint    `gorm:"not null;default:1" json:"usage_limit"`             // 可使用次数
	ExpiredAt  *int64  `gorm:"index" json:"expired_at,omitempty"`                 // 到期时间
}

// TableName 指定表名
func (Coupon) TableName() string { return "coupons" }
