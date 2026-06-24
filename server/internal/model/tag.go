package model

// Tag 标签
type Tag struct {
	Base
	Name string `gorm:"size:255;not null" json:"name"` // 名称
}

// TableName 指定表名
func (Tag) TableName() string { return "tags" }

// Taggable 标签多态关联表
type Taggable struct {
	Base
	TagID        *uint64 `gorm:"index" json:"tag_id,omitempty"`  // 标签
	UserID       *uint64 `gorm:"index" json:"user_id,omitempty"` // 用户
	TaggableType string  `gorm:"size:64;not null;index:idx_tag_polymorphic" json:"taggable_type"`
	TaggableID   uint64  `gorm:"not null;index:idx_tag_polymorphic" json:"taggable_id"`
}

// TableName 指定表名
func (Taggable) TableName() string { return "taggables" }
