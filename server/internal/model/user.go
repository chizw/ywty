package model

// User 用户表
// 与原 Laravel users 表结构保持一致
type User struct {
	Base
	Avatar          string    `gorm:"size:255;not null;default:''" json:"avatar"`            // 头像
	Name            string    `gorm:"size:255;not null;default:''" json:"name"`              // 姓名
	Username        string    `gorm:"size:255;not null;uniqueIndex" json:"username"`         // 用户名
	Phone           *string   `gorm:"size:64;uniqueIndex" json:"phone,omitempty"`            // 手机号
	Email           string    `gorm:"size:255;not null;uniqueIndex" json:"email"`            // 邮箱
	Password        string    `gorm:"size:255;not null" json:"-"`                            // 密码（不序列化）
	Location        string    `gorm:"size:64;not null;default:''" json:"location"`           // 所在地
	URL             string    `gorm:"size:255;not null;default:''" json:"url"`               // 个人网站
	Company         string    `gorm:"size:128;not null;default:''" json:"company"`           // 所在公司
	CompanyTitle    string    `gorm:"size:128;not null;default:''" json:"company_title"`     // 工作职位
	Tagline         string    `gorm:"size:255;not null;default:''" json:"tagline"`           // 个性签名
	Bio             string    `gorm:"size:255;not null;default:''" json:"bio"`               // 个人简介
	Interests       JSONSlice `gorm:"type:json" json:"interests,omitempty"`                  // 兴趣标签
	Socials         JSONSlice `gorm:"type:json" json:"socials,omitempty"`                    // 社交账号
	PhoneVerifiedAt *int64    `json:"phone_verified_at,omitempty"`                           // 手机验证时间（unix 秒）
	EmailVerifiedAt *int64    `json:"email_verified_at,omitempty"`                           // 邮箱验证时间
	RememberToken   string    `gorm:"size:100" json:"-"`                                     // 记住令牌
	IsAdmin         bool      `gorm:"not null;default:false;index" json:"is_admin"`          // 是否为管理员
	Options         JSONMap   `gorm:"type:json" json:"options,omitempty"`                    // 配置
	LoginIP         string    `gorm:"size:45" json:"login_ip,omitempty"`                     // 最后登录 IP
	RegisterIP      string    `gorm:"size:45" json:"register_ip,omitempty"`                  // 注册 IP
	CountryCode     string    `gorm:"size:32" json:"country_code,omitempty"`                 // 国家
	Status          string    `gorm:"size:64;not null;default:'normal';index" json:"status"` // 状态
}

// TableName 指定表名
func (User) TableName() string { return "users" }
