package model

// 反馈类型
const (
	FeedbackTypeGeneral = "general" // 一般
	FeedbackTypeBug     = "bug"     // Bug
	FeedbackTypeSuggest = "suggest" // 建议
)

// Feedback 意见与反馈
type Feedback struct {
	Base
	Type      string `gorm:"size:32;not null;default:'general'" json:"type"` // 类型
	Title     string `gorm:"size:64;not null" json:"title"`                  // 标题
	Name      string `gorm:"size:64;not null" json:"name"`                   // 姓名
	Email     string `gorm:"size:128;not null" json:"email"`                 // email
	Content   string `gorm:"type:longtext;not null" json:"content"`          // 内容
	IPAddress string `gorm:"size:45" json:"ip_address,omitempty"`            // IP 地址
}

// TableName 指定表名
func (Feedback) TableName() string { return "feedbacks" }
