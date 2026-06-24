package model

// 页面类型
const (
	PageTypeInternal = "internal" // 内部
	PageTypeExternal = "external" // 外部跳转
)

// Page 页面表
type Page struct {
	Base
	Type        string `gorm:"size:32;not null;default:'internal'" json:"type"` // 类型
	Name        string `gorm:"size:255;not null" json:"name"`                   // 名称
	Icon        string `gorm:"size:64;not null;default:''" json:"icon"`         // 图标
	Title       string `gorm:"size:255;not null;default:''" json:"title"`       // 标题
	Content     string `gorm:"type:longtext" json:"content,omitempty"`          // 网页内容
	Keywords    string `gorm:"type:text" json:"keywords,omitempty"`             // 关键字
	Description string `gorm:"type:text" json:"description,omitempty"`          // 描述
	Slug        string `gorm:"size:255;not null;default:''" json:"slug"`        // url slug
	URL         string `gorm:"size:255;not null;default:''" json:"url"`         // 跳转 url
	ViewCount   uint64 `gorm:"not null;default:0" json:"view_count"`            // 浏览量
	Sort        int    `gorm:"not null;default:0" json:"sort"`                  // 排序值
	IsShow      bool   `gorm:"not null;default:false" json:"is_show"`           // 是否显示
}

// TableName 指定表名
func (Page) TableName() string { return "pages" }
