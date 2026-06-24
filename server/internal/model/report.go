package model

// 举报状态
const (
	ReportStatusUnhandled = "unhandled" // 未处理
	ReportStatusHandled   = "handled"   // 已处理
	ReportStatusIgnored   = "ignored"   // 已忽略
)

// Report 举报记录
type Report struct {
	Base
	ReportUserID   uint64 `gorm:"index;default:0" json:"report_user_id"` // 被举报用户
	ReportableType string `gorm:"size:64;not null;index:idx_report_polymorphic" json:"reportable_type"`
	ReportableID   uint64 `gorm:"not null;index:idx_report_polymorphic" json:"reportable_id"`
	Content        string `gorm:"size:255" json:"content,omitempty"`                        // 原因
	Status         string `gorm:"size:32;not null;default:'unhandled';index" json:"status"` // 状态
	HandledAt      int64  `gorm:"default:0" json:"handled_at"`                              // 处理时间
	IPAddress      string `gorm:"size:45" json:"ip_address,omitempty"`                      // IP 地址
}

// TableName 指定表名
func (Report) TableName() string { return "reports" }
