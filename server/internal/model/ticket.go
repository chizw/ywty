package model

// 工单级别
const (
	TicketLevelLow    = "low"    // 低
	TicketLevelMedium = "medium" // 中
	TicketLevelHigh   = "high"   // 高
	TicketLevelUrgent = "urgent" // 紧急
)

// 工单状态
const (
	TicketStatusInProgress = "in_progress" // 进行中
	TicketStatusResolved   = "resolved"    // 已解决
	TicketStatusClosed     = "closed"      // 已关闭
)

// 工单类型
const (
	TicketTypeBug       = "bug"       // Bug
	TicketTypeFeature   = "feature"   // 功能
	TicketTypeComplaint = "complaint" // 投诉
	TicketTypeOther     = "other"     // 其他
)

// Ticket 工单
type Ticket struct {
	Base
	UserID  uint64 `gorm:"not null;index" json:"user_id"`                              // 用户
	IssueNo string `gorm:"size:64;not null;uniqueIndex" json:"issue_no"`               // 工单编号
	Title   string `gorm:"size:255;not null" json:"title"`                             // 标题
	Type    string `gorm:"size:32;not null;default:'other'" json:"type"`               // 类型
	Level   string `gorm:"size:32;not null;default:'low'" json:"level"`                // 级别
	Status  string `gorm:"size:32;not null;default:'in_progress';index" json:"status"` // 状态
}

// TableName 指定表名
func (Ticket) TableName() string { return "tickets" }

// TicketReply 工单回复
type TicketReply struct {
	Base
	TicketID uint64 `gorm:"not null;index" json:"ticket_id"`        // 工单
	UserID   uint64 `gorm:"not null;index" json:"user_id"`          // 用户
	Content  string `gorm:"type:longtext;not null" json:"content"`  // 内容
	IsNotify bool   `gorm:"not null;default:true" json:"is_notify"` // 是否需要接收通知
	ReadAt   *int64 `json:"read_at,omitempty"`                      // 已读时间
}

// TableName 指定表名
func (TicketReply) TableName() string { return "ticket_replies" }
