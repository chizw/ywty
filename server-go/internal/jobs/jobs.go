// Package jobs 队列任务处理器注册（对应 PHP app/Jobs）。
package jobs

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/chizw/ywty/server-go/internal/cache"
	"github.com/chizw/ywty/server-go/internal/config"
	"github.com/chizw/ywty/server-go/internal/mailx"
	"github.com/chizw/ywty/server-go/internal/queue"
	"gorm.io/gorm"
)

// Register 注册全部任务处理器。
func Register(q *queue.Queue, gdb *gorm.DB, c *cache.Cache, cfg *config.Config) {
	q.Register("send_code_mail", func(data []byte) error {
		return sendCodeMail(gdb, c, cfg, data)
	})
	q.Register("send_code_sms", func([]byte) error {
		// 短信驱动在 M5 接入
		return errors.New("短信服务尚未配置")
	})
}

type sendCodeMailData struct {
	GroupID  int64  `json:"group_id"`
	Event    string `json:"event"`
	Email    string `json:"email"`
	SiteName string `json:"site_name"`
}

// sendCodeMail 对齐 SendCodeMailJob → MailService::sendCode：
// 解析组的邮件驱动 → 生成验证码（cache: mail_code:{event}:{email}，TTL 900）→ SMTP 发送。
func sendCodeMail(gdb *gorm.DB, c *cache.Cache, cfg *config.Config, data []byte) error {
	var in sendCodeMailData
	if err := json.Unmarshal(data, &in); err != nil {
		return fmt.Errorf("任务参数解析失败: %w", err)
	}
	smtpCfg, err := mailx.ResolveSMTP(gdb, in.GroupID)
	if err != nil {
		return err
	}
	if smtpCfg.FromAddr == "" {
		smtpCfg.FromAddr = "hello@example.com"
	}
	if smtpCfg.FromName == "" {
		smtpCfg.FromName = in.SiteName
	}
	code := mailx.GenerateCode(c, mailx.CodeKey(in.Event, in.Email))
	body, err := mailx.RenderVerifyCode(in.SiteName, "您的验证码是：", code)
	if err != nil {
		return err
	}
	if err := smtpCfg.Send(in.Email, "验证码", body); err != nil {
		slog.Warn("验证码邮件发送失败", "email", in.Email, "err", err)
		return err
	}
	slog.Info("验证码邮件已发送", "email", in.Email, "event", in.Event)
	return nil
}
