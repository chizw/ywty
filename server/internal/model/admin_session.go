package model

// AdminSession 后台 session
type AdminSession struct {
	Base
	UserID    uint64 `gorm:"not null;index" json:"user_id"`          // 用户
	Token     string `gorm:"size:128;not null;uniqueIndex" json:"-"` // session 标识（哈希）
	IPAddress string `gorm:"size:45" json:"ip_address,omitempty"`    // 登录 IP
	UserAgent string `gorm:"size:512" json:"user_agent,omitempty"`   // UA
	ExpiresAt int64  `gorm:"not null;index" json:"expires_at"`       // 过期时间
}

// TableName 指定表名
func (AdminSession) TableName() string { return "admin_sessions" }
