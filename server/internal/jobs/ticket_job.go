// Package jobs 工单通知任务
package jobs

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/ywty/server/internal/logger"
	"github.com/ywty/server/internal/model"
	"github.com/ywty/server/internal/queue"
)

// TicketNotifyPayload 工单通知任务参数
type TicketNotifyPayload struct {
	TicketID uint64 `json:"ticket_id"`
	ReplyID  uint64 `json:"reply_id,omitempty"` // 回复 ID（可选）
	IsAdmin  bool   `json:"is_admin"`           // 是否管理员回复
}

// NewTicketNotifyTask 创建工单通知任务
func NewTicketNotifyTask(payload TicketNotifyPayload) (*asynq.Task, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(queue.TypeTicketNotify, data), nil
}

// HandleTicketNotify 处理工单通知
func (h *Handlers) HandleTicketNotify(ctx context.Context, t *asynq.Task) error {
	var p TicketNotifyPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}

	var ticket model.Ticket
	if err := h.DB.WithContext(ctx).First(&ticket, p.TicketID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil
		}
		return err
	}

	// 获取用户信息
	var user model.User
	if err := h.DB.WithContext(ctx).First(&user, ticket.UserID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil
		}
		return err
	}

	// 获取回复内容（如果有）
	var reply model.TicketReply
	var replyContent string
	if p.ReplyID > 0 {
		if err := h.DB.WithContext(ctx).First(&reply, p.ReplyID).Error; err == nil {
			replyContent = reply.Content
		}
	}

	// 构建通知内容（当前仅记录日志，后续可扩展为用户通知表）
	var title, content string
	if p.IsAdmin {
		title = fmt.Sprintf("工单 #%s 有新回复", ticket.IssueNo)
		content = fmt.Sprintf("管理员回复了工单「%s」", ticket.Title)
		if replyContent != "" {
			content += fmt.Sprintf("，回复：%s", replyContent)
		}
	} else {
		title = fmt.Sprintf("新工单 #%s", ticket.IssueNo)
		content = fmt.Sprintf("用户 %s 提交了新工单「%s」", user.Username, ticket.Title)
	}

	logger.L.Info("ticket notification",
		zap.String("title", title),
		zap.String("content", content),
		zap.Uint64("user_id", ticket.UserID),
	)

	logger.L.Info("ticket notify sent",
		zap.Uint64("ticket_id", p.TicketID),
		zap.Bool("is_admin", p.IsAdmin),
	)
	return nil
}