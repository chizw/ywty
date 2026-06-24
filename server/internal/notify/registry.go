// Package notify 通知驱动注册表（按名切换）
package notify

import "fmt"

// SMSFactory SMS 驱动工厂
type SMSFactory func() SMSer

// MailFactory Mail 驱动工厂
type MailFactory func() Mailer

var (
	smsRegistry  = map[string]SMSFactory{}
	mailRegistry = map[string]MailFactory{}
)

// RegisterSMS 注册 SMS 驱动
func RegisterSMS(name string, f SMSFactory) { smsRegistry[name] = f }

// RegisterMail 注册 Mail 驱动
func RegisterMail(name string, f MailFactory) { mailRegistry[name] = f }

// NewSMSer 构造指定 SMS 驱动
func NewSMSer(name string) (SMSer, error) {
	f, ok := smsRegistry[name]
	if !ok {
		return nil, fmt.Errorf("unsupported sms driver: %s", name)
	}
	return f(), nil
}

// NewMailer 构造指定 Mail 驱动
func NewMailer(name string) (Mailer, error) {
	f, ok := mailRegistry[name]
	if !ok {
		return nil, fmt.Errorf("unsupported mail driver: %s", name)
	}
	return f(), nil
}

// ListSMSDrivers 列出已注册 SMS 驱动
func ListSMSDrivers() []string {
	out := make([]string, 0, len(smsRegistry))
	for k := range smsRegistry {
		out = append(out, k)
	}
	return out
}

// ListMailDrivers 列出已注册 Mail 驱动
func ListMailDrivers() []string {
	out := make([]string, 0, len(mailRegistry))
	for k := range mailRegistry {
		out = append(out, k)
	}
	return out
}

func init() {
	// 默认注册 log 驱动
	RegisterSMS("log", func() SMSer { return NewLogSMSer() })
	RegisterMail("log", func() Mailer { return NewLogMailer() })
	RegisterMail("smtp", func() Mailer { return &SMTPMailer{} })
}
