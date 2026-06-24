package model

// Notice 系统公告
type Notice struct {
	Base
	Title     string `gorm:"size:255;not null" json:"title"`         // 标题
	Content   string `gorm:"type:longtext" json:"content,omitempty"` // 内容
	ViewCount uint64 `gorm:"not null;default:0" json:"view_count"`   // 阅读量
	Sort      int    `gorm:"not null;default:0" json:"sort"`         // 排序值
}

// TableName 指定表名
func (Notice) TableName() string { return "notices" }
