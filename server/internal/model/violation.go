package model

// 违规状态
const (
	ViolationStatusUnhandled = "unhandled" // 未处理
	ViolationStatusHandled   = "handled"   // 已处理
	ViolationStatusIgnored   = "ignored"   // 已忽略
)

// Violation 图片违规记录
type Violation struct {
	Base
	UserID    *uint64 `gorm:"index" json:"user_id,omitempty"`                           // 用户
	PhotoID   *uint64 `gorm:"index" json:"photo_id,omitempty"`                          // 图片
	Reason    string  `gorm:"size:255;not null;default:''" json:"reason"`               // 违规原因
	Status    string  `gorm:"size:32;not null;default:'unhandled';index" json:"status"` // 状态
	HandledAt *int64  `json:"handled_at,omitempty"`                                     // 处理时间
}

// TableName 指定表名
func (Violation) TableName() string { return "violations" }
