// Package notify log sms：把短信内容输出到日志（开发环境默认）
package notify

import (
	"context"

	"go.uber.org/zap"

	"github.com/ywty/server/internal/logger"
)

// LogSMSer 把短信内容输出到日志
type LogSMSer struct{}

func NewLogSMSer() *LogSMSer { return &LogSMSer{} }

func (l *LogSMSer) Name() string { return "log" }

func (l *LogSMSer) Send(_ context.Context, s SMS) error {
	logger.L.Info("sms send",
		zap.String("to", s.To),
		zap.String("body", s.Body),
	)
	return nil
}
