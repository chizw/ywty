package model

// 套餐类型
const (
	PlanTypeVip     = "vip"     // VIP
	PlanTypeDefault = "default" // 默认
)

// Plan 计划套餐
type Plan struct {
	Base
	Type     string    `gorm:"size:64;not null;default:'vip'" json:"type"` // 类型
	Name     string    `gorm:"size:255;not null" json:"name"`              // 名称
	Intro    string    `gorm:"type:text" json:"intro,omitempty"`           // 简介
	Features JSONSlice `gorm:"type:json" json:"features,omitempty"`        // 特点
	Badge    string    `gorm:"size:32;not null;default:''" json:"badge"`   // 徽章内容
	Sort     int       `gorm:"not null;default:0" json:"sort"`             // 排序值
	IsUp     bool      `gorm:"not null;default:false" json:"is_up"`        // 是否上架
}

// TableName 指定表名
func (Plan) TableName() string { return "plans" }

// PlanPrice 套餐阶梯价格
type PlanPrice struct {
	Base
	PlanID   uint64 `gorm:"not null;index" json:"plan_id"`      // 计划
	Name     string `gorm:"size:255;not null" json:"name"`      // 名称
	Duration int    `gorm:"not null;default:0" json:"duration"` // 时长(分钟)
	Price    int    `gorm:"not null;default:0" json:"price"`    // 价格(分)
}

// TableName 指定表名
func (PlanPrice) TableName() string { return "plan_prices" }

// PlanGroup 计划可用组（中间表，无时间戳）
type PlanGroup struct {
	ID      uint64 `gorm:"primaryKey;autoIncrement" json:"id"`
	PlanID  uint64 `gorm:"not null;index" json:"plan_id"`  // 计划
	GroupID uint64 `gorm:"index;default:0" json:"group_id"` // 角色组
}

// TableName 指定表名
func (PlanGroup) TableName() string { return "plan_groups" }

// PlanCapacity 计划可用容量
type PlanCapacity struct {
	ID       uint64 `gorm:"primaryKey;autoIncrement" json:"id"`
	PlanID   uint64 `gorm:"not null;index" json:"plan_id"`                         // 计划
	Capacity int64  `gorm:"type:decimal(20,0);not null;default:0" json:"capacity"` // 容量(字节)
}

// TableName 指定表名
func (PlanCapacity) TableName() string { return "plan_capacities" }
