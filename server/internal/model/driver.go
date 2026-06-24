package model

// 驱动类型
const (
	DriverTypeStorage = "storage" // 存储
	DriverTypeProcess = "process" // 处理
	DriverTypeMail    = "mail"    // 邮件
	DriverTypeSMS     = "sms"     // 短信
	DriverTypePayment = "payment" // 支付
	DriverTypeOAuth   = "oauth"   // 三方登录
	DriverTypeScan    = "scan"    // 内容审核
)

// Driver 驱动表（多态：存储/处理/邮件/短信/支付/三方登录/审核）
type Driver struct {
	Base
	Type    string  `gorm:"size:64;not null;index" json:"type"`         // 驱动类型
	Name    string  `gorm:"size:255;not null" json:"name"`              // 名称
	Intro   string  `gorm:"size:2000;not null;default:''" json:"intro"` // 简介
	Options JSONMap `gorm:"type:json" json:"options,omitempty"`         // 配置
}

// TableName 指定表名
func (Driver) TableName() string { return "drivers" }
