// Package notify 邮件 / 短信驱动抽象
package notify

import "context"

// Mail 邮件内容
type Mail struct {
	To      []string // 收件人
	Subject string   // 主题
	Text    string   // 纯文本正文
	HTML    string   // HTML 正文
}

// Mailer 邮件驱动接口
type Mailer interface {
	Name() string
	Send(ctx context.Context, m Mail) error
}

// SMS 短信内容
type SMS struct {
	To   string // 手机号
	Body string // 内容
	Sign string // 签名（可选）
}

// SMSer 短信驱动接口
type SMSer interface {
	Name() string
	Send(ctx context.Context, s SMS) error
}
