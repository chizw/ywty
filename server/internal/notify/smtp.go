// Package notify smtp mailer
package notify

import (
	"context"
	"fmt"
	"net/smtp"

	"go.uber.org/zap"

	"github.com/ywty/server/internal/logger"
)

// SMTPMailer 通过 SMTP 发送邮件
type SMTPMailer struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
	FromName string
}

func NewSMTPMailer(host string, port int, user, pass, from, fromName string) *SMTPMailer {
	return &SMTPMailer{Host: host, Port: port, Username: user, Password: pass, From: from, FromName: fromName}
}

func (s *SMTPMailer) Name() string { return "smtp" }

func (s *SMTPMailer) Send(_ context.Context, m Mail) error {
	auth := smtp.PlainAuth("", s.Username, s.Password, s.Host)
	addr := fmt.Sprintf("%s:%d", s.Host, s.Port)

	header := make(map[string]string)
	header["From"] = fmt.Sprintf("%s <%s>", s.FromName, s.From)
	header["To"] = m.To[0]
	header["Subject"] = m.Subject
	header["MIME-Version"] = "1.0"
	header["Content-Type"] = "text/plain; charset=UTF-8"

	body := m.Text
	msg := ""
	for k, v := range header {
		msg += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	msg += "\r\n" + body

	if err := smtp.SendMail(addr, auth, s.From, m.To, []byte(msg)); err != nil {
		logger.L.Error("smtp send failed", zap.Error(err))
		return err
	}
	return nil
}
