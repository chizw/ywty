// Package notify log mailer：把邮件直接打到日志（开发环境默认）
package notify

import (
	"context"
	"strings"

	"go.uber.org/zap"

	"github.com/ywty/server/internal/logger"
)

// LogMailer 把邮件内容输出到日志
type LogMailer struct{}

func NewLogMailer() *LogMailer { return &LogMailer{} }

func (l *LogMailer) Name() string { return "log" }

func (l *LogMailer) Send(_ context.Context, m Mail) error {
	logger.L.Info("mail send",
		zap.Strings("to", m.To),
		zap.String("subject", m.Subject),
		zap.String("body", strings.TrimSpace(m.Text)),
	)
	return nil
}
