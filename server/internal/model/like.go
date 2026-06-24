package model

// Like 点赞表
type Like struct {
	Base
	UserID       uint64 `gorm:"not null;index" json:"user_id"` // 用户
	LikeableType string `gorm:"size:64;not null;index:idx_like_polymorphic" json:"likeable_type"`
	LikeableID   uint64 `gorm:"not null;index:idx_like_polymorphic" json:"likeable_id"`
}

// TableName 指定表名
func (Like) TableName() string { return "likes" }
