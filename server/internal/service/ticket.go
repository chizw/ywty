// Package service 工单服务
package service

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"gorm.io/gorm"

	bizerr "github.com/ywty/server/internal/errors"
	"github.com/ywty/server/internal/model"
)

// TicketService 工单服务
type TicketService struct {
	db *gorm.DB
}

func NewTicketService(db *gorm.DB) *TicketService { return &TicketService{db: db} }

// CreateTicketReq 创建工单请求
type CreateTicketReq struct {
	Title   string `json:"title" binding:"required"`
	Level   string `json:"level" binding:"required,oneof=low medium high urgent"`
	Type    string `json:"type" binding:"required,oneof=bug feature complaint other"`
	Content string `json:"content" binding:"required"`
}

// Create 创建工单
func (s *TicketService) Create(ctx context.Context, userID uint64, req CreateTicketReq) (*model.Ticket, error) {
	issueNo, err := s.generateIssueNo(ctx)
	if err != nil {
		return nil, err
	}

	ticket := &model.Ticket{
		UserID:  userID,
		IssueNo: issueNo,
		Title:   req.Title,
		Type:    req.Type,
		Level:   req.Level,
		Status:  model.TicketStatusInProgress,
	}

	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(ticket).Error; err != nil {
			return fmt.Errorf("create ticket: %w", err)
		}
		reply := &model.TicketReply{
			TicketID: ticket.ID,
			UserID:   userID,
			Content:  req.Content,
			IsNotify: true,
		}
		if err := tx.Create(reply).Error; err != nil {
			return fmt.Errorf("create ticket reply: %w", err)
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return ticket, nil
}

// List 用户列出自己的工单
func (s *TicketService) List(ctx context.Context, userID uint64, page, perPage int) ([]model.Ticket, int64, error) {
	page, perPage = normalizePage(page, perPage)
	var total int64
	q := s.db.WithContext(ctx).Model(&model.Ticket{}).Where("user_id = ?", userID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []model.Ticket
	if err := q.Order("id DESC").Offset((page - 1) * perPage).Limit(perPage).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

// Get 用户获取工单详情（含 replies）
func (s *TicketService) Get(ctx context.Context, userID, ticketID uint64) (*model.Ticket, []model.TicketReply, error) {
	var ticket model.Ticket
	if err := s.db.WithContext(ctx).Where("id = ? AND user_id = ?", ticketID, userID).First(&ticket).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, bizerr.ResourceNotFound
		}
		return nil, nil, err
	}
	var replies []model.TicketReply
	if err := s.db.WithContext(ctx).Where("ticket_id = ?", ticketID).Order("id ASC").Find(&replies).Error; err != nil {
		return nil, nil, err
	}
	return &ticket, replies, nil
}

// Reply 用户回复工单
func (s *TicketService) Reply(ctx context.Context, userID, ticketID uint64, content string) (*model.TicketReply, error) {
	var ticket model.Ticket
	if err := s.db.WithContext(ctx).Where("id = ? AND user_id = ?", ticketID, userID).First(&ticket).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, bizerr.ResourceNotFound
		}
		return nil, err
	}
	if ticket.Status == model.TicketStatusClosed {
		return nil, bizerr.BadRequest.WithMessage("工单已关闭")
	}
	reply := &model.TicketReply{
		TicketID: ticketID,
		UserID:   userID,
		Content:  content,
		IsNotify: true,
	}
	if err := s.db.WithContext(ctx).Create(reply).Error; err != nil {
		return nil, fmt.Errorf("create reply: %w", err)
	}
	return reply, nil
}

// Close 用户关闭自己的工单
func (s *TicketService) Close(ctx context.Context, userID, ticketID uint64) error {
	res := s.db.WithContext(ctx).Model(&model.Ticket{}).
		Where("id = ? AND user_id = ? AND status != ?", ticketID, userID, model.TicketStatusClosed).
		Updates(map[string]any{"status": model.TicketStatusClosed})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return bizerr.ResourceNotFound
	}
	return nil
}

// AdminList 后台列表
func (s *TicketService) AdminList(ctx context.Context, page, perPage int, status, level string) ([]model.Ticket, int64, error) {
	page, perPage = normalizePage(page, perPage)
	q := s.db.WithContext(ctx).Model(&model.Ticket{})
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if level != "" {
		q = q.Where("level = ?", level)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []model.Ticket
	if err := q.Order("id DESC").Offset((page - 1) * perPage).Limit(perPage).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

// AdminGet 后台详情
func (s *TicketService) AdminGet(ctx context.Context, ticketID uint64) (*model.Ticket, []model.TicketReply, error) {
	var ticket model.Ticket
	if err := s.db.WithContext(ctx).First(&ticket, ticketID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, bizerr.ResourceNotFound
		}
		return nil, nil, err
	}
	var replies []model.TicketReply
	if err := s.db.WithContext(ctx).Where("ticket_id = ?", ticketID).Order("id ASC").Find(&replies).Error; err != nil {
		return nil, nil, err
	}
	return &ticket, replies, nil
}

// AdminReply 后台回复
func (s *TicketService) AdminReply(ctx context.Context, ticketID uint64, content string) (*model.TicketReply, error) {
	var ticket model.Ticket
	if err := s.db.WithContext(ctx).First(&ticket, ticketID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, bizerr.ResourceNotFound
		}
		return nil, err
	}
	reply := &model.TicketReply{
		TicketID: ticketID,
		UserID:   0,
		Content:  content,
		IsNotify: true,
	}
	if err := s.db.WithContext(ctx).Create(reply).Error; err != nil {
		return nil, fmt.Errorf("create reply: %w", err)
	}
	if ticket.Status == model.TicketStatusClosed {
		_ = s.db.WithContext(ctx).Model(&ticket).Update("status", model.TicketStatusInProgress).Error
	}
	return reply, nil
}

// AdminUpdateStatus 后台更新状态
func (s *TicketService) AdminUpdateStatus(ctx context.Context, ticketID uint64, status string) error {
	switch status {
	case model.TicketStatusInProgress, model.TicketStatusResolved, model.TicketStatusClosed:
	default:
		return bizerr.BadRequest.WithMessage("invalid status")
	}
	res := s.db.WithContext(ctx).Model(&model.Ticket{}).Where("id = ?", ticketID).Update("status", status)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return bizerr.ResourceNotFound
	}
	return nil
}

// generateIssueNo 格式 TICKET-YYYYMMDD-XXXX（4位随机）
func (s *TicketService) generateIssueNo(ctx context.Context) (string, error) {
	for i := 0; i < 10; i++ {
		prefix := time.Now().Format("TICKET-20060102-")
		no := prefix + fmt.Sprintf("%04d", rand.Intn(10000))
		var cnt int64
		if err := s.db.WithContext(ctx).Model(&model.Ticket{}).Where("issue_no = ?", no).Count(&cnt).Error; err != nil {
			return "", err
		}
		if cnt == 0 {
			return no, nil
		}
	}
	return "", bizerr.Conflict.WithMessage("无法生成唯一工单编号")
}

func normalizePage(page, perPage int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if perPage <= 0 || perPage > 100 {
		perPage = 20
	}
	return page, perPage
}
