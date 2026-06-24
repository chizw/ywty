// Package notify 通知驱动全局单例
package notify

import (
	"fmt"
	"sync"
)

var (
	mailerOnce sync.Once
	mailer     Mailer
	smserOnce  sync.Once
	smser      SMSer
)

// SetMailer 注入邮件驱动
func SetMailer(m Mailer) { mailer = m }

// GetMailer 获取邮件驱动（未注入时使用 log 驱动）
func GetMailer() Mailer {
	mailerOnce.Do(func() {
		if mailer == nil {
			mailer = NewLogMailer()
		}
	})
	return mailer
}

// SetSMSer 注入短信驱动
func SetSMSer(s SMSer) { smser = s }

// GetSMSer 获取短信驱动（未注入时使用 log 驱动）
func GetSMSer() SMSer {
	smserOnce.Do(func() {
		if smser == nil {
			smser = NewLogSMSer()
		}
	})
	return smser
}

// 编译期接口检查
var (
	_ Mailer = (*LogMailer)(nil)
	_ SMSer  = (*LogSMSer)(nil)
	_ Mailer = (*SMTPMailer)(nil)
	_        = fmt.Sprintf
)
