// Package service 验证码服务
package service

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	bizerr "github.com/ywty/server/internal/errors"
	"github.com/ywty/server/internal/logger"
	"github.com/ywty/server/internal/model"
	"github.com/ywty/server/internal/notify"
)

// VerifyCodeService 验证码服务
type VerifyCodeService struct {
	db *gorm.DB
}

// NewVerifyCodeService 创建验证码服务
func NewVerifyCodeService(db *gorm.DB) *VerifyCodeService {
	return &VerifyCodeService{db: db}
}

// Send 发送验证码
// channel: email | sms
// account: 邮箱或手机号
// event:   register/reset_password/...
func (s *VerifyCodeService) Send(ctx context.Context, channel, account, event, ip string) error {
	account = strings.TrimSpace(account)
	if account == "" {
		return bizerr.BadRequest.WithMessage("account is required")
	}
	if channel != model.VerifyChannelEmail && channel != model.VerifyChannelSMS {
		return bizerr.CodeChannelInvalid
	}

	// 1 分钟内同一账号同事件最多 1 次
	now := time.Now().Unix()
	oneMinAgo := now - 60
	var recent int64
	if err := s.db.WithContext(ctx).Model(&model.VerifyCode{}).
		Where("account = ? AND event = ? AND channel = ? AND created_at > ?",
			account, event, channel, time.Unix(oneMinAgo, 0)).
		Count(&recent).Error; err != nil {
		return fmt.Errorf("count verify codes: %w", err)
	}
	if recent > 0 {
		return bizerr.CodeSendTooFrequent
	}

	// 24h 内同一事件，同一账号限制 10 次
	dayAgo := time.Unix(now-86400, 0)
	var dayCount int64
	if err := s.db.WithContext(ctx).Model(&model.VerifyCode{}).
		Where("account = ? AND event = ? AND channel = ? AND created_at > ?",
			account, event, channel, dayAgo).
		Count(&dayCount).Error; err != nil {
		return fmt.Errorf("count day codes: %w", err)
	}
	if dayCount >= 10 {
		return bizerr.CodeSendTooFrequent
	}

	// 生成 6 位数字
	code := generateCode(6)
	exp := now + 600 // 10 分钟
	rec := &model.VerifyCode{
		Channel:   channel,
		Account:   account,
		Event:     event,
		Code:      code,
		IPAddress: ip,
		ExpiredAt: exp,
	}
	if err := s.db.WithContext(ctx).Create(rec).Error; err != nil {
		return fmt.Errorf("save verify code: %w", err)
	}

	// 发送
	switch channel {
	case model.VerifyChannelEmail:
		m := notify.Mail{
			To:      []string{account},
			Subject: "ywty 验证码",
			Text:    fmt.Sprintf("您的验证码是：%s ，10 分钟内有效，请勿泄露给他人。", code),
		}
		if err := notify.GetMailer().Send(ctx, m); err != nil {
			logger.L.Warn("send mail failed", zap.Error(err))
			return bizerr.MailSendFailed
		}
	case model.VerifyChannelSMS:
		s := notify.SMS{To: account, Body: fmt.Sprintf("【ywty】您的验证码：%s，10 分钟内有效。", code)}
		if err := notify.GetSMSer().Send(ctx, s); err != nil {
			logger.L.Warn("send sms failed", zap.Error(err))
			return bizerr.SMSSendFailed
		}
	}
	return nil
}

// Verify 校验并消费验证码。成功会标记 used_at。
func (s *VerifyCodeService) Verify(ctx context.Context, channel, account, event, code string) error {
	account = strings.TrimSpace(account)
	code = strings.TrimSpace(code)
	if account == "" || code == "" {
		return bizerr.CodeInvalid
	}
	now := time.Now().Unix()
	var rec model.VerifyCode
	err := s.db.WithContext(ctx).
		Where("account = ? AND event = ? AND channel = ? AND used_at IS NULL",
			account, event, channel).
		Order("id DESC").
		First(&rec).Error
	if err != nil {
		return bizerr.CodeInvalid
	}
	if rec.ExpiredAt < now {
		return bizerr.CodeExpired
	}
	if rec.Code != code {
		return bizerr.CodeInvalid
	}
	if err := s.db.WithContext(ctx).Model(&rec).Update("used_at", now).Error; err != nil {
		return fmt.Errorf("mark used: %w", err)
	}
	return nil
}

// generateCode 生成 n 位数字验证码
func generateCode(n int) string {
	if n < 4 {
		n = 4
	}
	if n > 10 {
		n = 10
	}
	max := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(n)), nil)
	num, err := rand.Int(rand.Reader, max)
	if err != nil {
		return fmt.Sprintf("%0*d", n, time.Now().UnixNano()%int64(max.Int64()))
	}
	return fmt.Sprintf("%0*d", n, num.Int64())
}
