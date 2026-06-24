package model

// UserGroup 用户角色组关联
type UserGroup struct {
	Base
	UserID    uint64  `gorm:"not null;index" json:"user_id"`              // 用户
	GroupID   uint64  `gorm:"not null;index" json:"group_id"`             // 角色组
	OrderID   *uint64 `gorm:"index" json:"order_id,omitempty"`           // 来源订单
	From      string  `gorm:"size:32;not null;default:'system'" json:"from"` // 来源
	ExpiredAt *int64  `gorm:"index" json:"expired_at,omitempty"`         // 到期时间
}

// TableName 指定表名
func (UserGroup) TableName() string { return "user_groups" }

// UserCapacity 用户容量
type UserCapacity struct {
	Base
	UserID    uint64  `gorm:"not null;index" json:"user_id"`              // 用户
	OrderID   *uint64 `gorm:"index" json:"order_id,omitempty"`           // 来源订单
	Capacity  float64 `gorm:"type:decimal(20,0);not null;default:0;index" json:"capacity"` // 容量(kb)
	From      string  `gorm:"size:32;not null;default:'system'" json:"from"` // 来源
	ExpiredAt *int64  `gorm:"index" json:"expired_at,omitempty"`         // 到期时间
}

// TableName 指定表名
func (UserCapacity) TableName() string { return "user_capacities" }
