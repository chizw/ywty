package model

// OAuth 三方授权
type OAuth struct {
	Base
	DriverID uint64  `gorm:"not null;index" json:"driver_id"`              // 三方授权驱动ID
	UserID   uint64  `gorm:"not null;index" json:"user_id"`                // 用户ID
	OpenID   string  `gorm:"size:255;not null" json:"openid"`              // 三方授权ID
	Avatar   string  `gorm:"size:512;not null;default:''" json:"avatar"`   // 三方授权头像
	Email    string  `gorm:"size:255;not null;default:''" json:"email"`    // 三方授权邮箱
	Name     string  `gorm:"size:255;not null;default:''" json:"name"`     // 三方授权名称
	Nickname string  `gorm:"size:255;not null;default:''" json:"nickname"` // 三方授权昵称
	Raw      JSONMap `gorm:"type:json" json:"raw,omitempty"`               // 三方授权原始信息
}

// TableName 指定表名
func (OAuth) TableName() string { return "oauth" }
