package model

// 分享类型
const (
	ShareTypeAlbum = "album" // 相册分享
	ShareTypePhoto = "photo" // 图片分享
)

// Share 分享
type Share struct {
	Base
	UserID    uint64 `gorm:"not null;index" json:"user_id"`                // 用户
	Type      string `gorm:"size:32;not null;default:'album'" json:"type"` // 分享类型
	Slug      string `gorm:"size:64;not null;uniqueIndex" json:"slug"`     // url slug
	Content   string `gorm:"type:text" json:"content,omitempty"`           // 分享内容
	Password  string `gorm:"size:128;not null;default:''" json:"password"` // 密码
	ViewCount uint64 `gorm:"not null;default:0" json:"view_count"`         // 浏览量
	ExpiredAt *int64 `gorm:"index" json:"expired_at,omitempty"`            // 到期时间
}

// TableName 指定表名
func (Share) TableName() string { return "shares" }

// Shareable 分享内容多态表
type Shareable struct {
	Base
	ShareID       uint64 `gorm:"not null;index" json:"share_id"` // 分享
	ShareableType string `gorm:"size:64;not null;index:idx_share_polymorphic" json:"shareable_type"`
	ShareableID   uint64 `gorm:"not null;index:idx_share_polymorphic" json:"shareable_id"`
}

// TableName 指定表名
func (Shareable) TableName() string { return "shareables" }
