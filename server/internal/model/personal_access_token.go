package model

// PersonalAccessToken 用户 API Token（兼容 Laravel Sanctum 字段命名）
type PersonalAccessToken struct {
	Base
	UserID     uint64    `gorm:"not null;index" json:"-"`               // 所属用户
	Name       string    `gorm:"size:255;not null" json:"name"`         // Token 名称
	Token      string    `gorm:"size:80;not null;uniqueIndex" json:"-"` // 哈希后的 Token
	Abilities  JSONSlice `gorm:"type:json" json:"abilities,omitempty"`  // 权限列表
	LastUsedAt *int64    `json:"last_used_at,omitempty"`                // 最后使用时间
	ExpiresAt  *int64    `gorm:"index" json:"expires_at,omitempty"`     // 过期时间
}

// TableName 指定表名
func (PersonalAccessToken) TableName() string { return "personal_access_tokens" }
