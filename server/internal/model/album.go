package model

// Album 相册
type Album struct {
	Base
	UserID   uint64 `gorm:"index" json:"user_id"`                            // 用户
	Name     string `gorm:"size:255;not null;default:''" json:"name"`        // 名称
	Intro    string `gorm:"size:2000;not null;default:''" json:"intro"`      // 介绍
	IsPublic bool   `gorm:"not null;default:false;index" json:"is_public"`   // 是否公开
}

// TableName 指定表名
func (Album) TableName() string { return "albums" }

// AlbumPhoto 相册与图片中间表
type AlbumPhoto struct {
	AlbumID uint64 `gorm:"primaryKey;index" json:"album_id"` // 相册
	PhotoID uint64 `gorm:"primaryKey;index" json:"photo_id"` // 图片
	Sort    int    `gorm:"not null;default:0" json:"sort"`   // 排序值
}

// TableName 指定表名
func (AlbumPhoto) TableName() string { return "album_photo" }
