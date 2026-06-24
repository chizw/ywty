package model

// 验证码事件
const (
	VerifyEventRegister      = "register"       // 注册
	VerifyEventLogin         = "login"          // 登录
	VerifyEventResetPassword = "reset_password" // 重置密码
	VerifyEventChangeEmail   = "change_email"   // 更换邮箱（新邮箱验证）
	VerifyEventChangePhone   = "change_phone"   // 更换手机（新手机验证）
	VerifyEventBindEmail     = "bind_email"     // 绑定邮箱
	VerifyEventBindPhone     = "bind_phone"     // 绑定手机
)

// 验证码渠道
const (
	VerifyChannelEmail = "email"
	VerifyChannelSMS   = "sms"
)

// VerifyCode 验证码
type VerifyCode struct {
	Base
	Channel   string `gorm:"size:16;not null;index" json:"channel"`  // email | sms
	Account   string `gorm:"size:128;not null;index" json:"account"` // 邮箱 / 手机号
	Event     string `gorm:"size:32;not null;index" json:"event"`    // 事件
	Code      string `gorm:"size:16;not null" json:"-"`              // 验证码（不外泄）
	IPAddress string `gorm:"size:45" json:"ip_address,omitempty"`    // 请求 IP
	UsedAt    *int64 `json:"used_at,omitempty"`                      // 使用时间
	ExpiredAt int64  `gorm:"not null;index" json:"expired_at"`       // 过期时间（unix 秒）
}

// TableName 指定表名
func (VerifyCode) TableName() string { return "verify_codes" }
