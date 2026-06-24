package model

// Group 角色组
type Group struct {
	Base
	Name      string  `gorm:"size:255;not null" json:"name"`              // 名称
	Intro     string  `gorm:"size:2000;not null;default:''" json:"intro"` // 描述
	Options   JSONMap `gorm:"type:json" json:"options,omitempty"`         // 配置
	IsDefault bool    `gorm:"not null;default:false" json:"is_default"`   // 是否为默认组
	IsGuest   bool    `gorm:"not null;default:false" json:"is_guest"`     // 是否为游客组
}

// TableName 指定表名
func (Group) TableName() string { return "groups" }
