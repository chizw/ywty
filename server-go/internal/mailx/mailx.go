// Package mailx 邮件能力：SMTP 发送、验证码生成/校验（cache 键与 PHP 版一致：
// mail_code:{event}:{email}，TTL 900 秒）、邮件模板。
package mailx

import (
	"crypto/rand"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"math/big"
	"net"
	"net/smtp"
	"strings"
	"time"

	"github.com/chizw/ywty/server-go/internal/cache"
	"gorm.io/gorm"
)

// ---------- 验证码 ----------

// GenerateCode 生成 6 位验证码并写入缓存（等价 VerifyCodeService::generateCode，TTL 900）。
func GenerateCode(c *cache.Cache, key string) string {
	n, _ := rand.Int(rand.Reader, big.NewInt(900000))
	code := fmt.Sprintf("%06d", 100000+n.Int64())
	c.Put(key, code, 900)
	return code
}

// VerifyCode 等价 VerifyCodeService::verifyCode（注意 PHP 是松散比较，这里为精确比较）。
func VerifyCode(c *cache.Cache, key, code string) bool {
	stored, ok := c.Get(key)
	return ok && stored == code
}

// ---------- SMTP ----------

type SMTPConfig struct {
	Host       string `json:"host"`
	Port       int    `json:"port"`
	Username   string `json:"username"`
	Password   string `json:"password"`
	Encryption string `json:"encryption"` // tls | ssl | 空串
	Provider   string `json:"provider"`
	FromAddr   string `json:"from_address"`
	FromName   string `json:"from_name"`
}

// ResolveSMTP 从 drivers 表解析当前可用的 SMTP 配置（type='mail' 且 options.provider='smtp'）。
// 对齐 PHP MailService::instance 的 provider 装配，其余 provider（mailgun 等）在 M5 实现。
func ResolveSMTP(gdb *gorm.DB, groupID int64) (*SMTPConfig, error) {
	var raw *string
	err := gdb.Raw(
		"SELECT d.`options` FROM `drivers` d "+
			"INNER JOIN `group_driver` gd ON gd.driver_id = d.id AND gd.type = 'mail' "+
			"WHERE gd.group_id = ? AND d.type = 'mail' AND d.deleted_at IS NULL "+
			"ORDER BY gd.sort ASC LIMIT 1", groupID,
	).Scan(&raw).Error
	if err != nil {
		return nil, err
	}
	if raw == nil || *raw == "" {
		return nil, errors.New("邮件服务尚未初始化")
	}
	var opts SMTPConfig
	if err := json.Unmarshal([]byte(*raw), &opts); err != nil {
		return nil, fmt.Errorf("邮件驱动配置解析失败: %w", err)
	}
	if opts.Provider != "smtp" {
		return nil, errors.New("暂只支持 SMTP 邮件驱动")
	}
	if opts.Host == "" {
		return nil, errors.New("SMTP 配置缺少 host")
	}
	return &opts, nil
}

// Send 发送 HTML 邮件。
func (m *SMTPConfig) Send(to, subject, htmlBody string) error {
	addr := fmt.Sprintf("%s:%d", m.Host, m.Port)
	from := m.FromAddr
	msg := buildMessage(m.FromName, from, to, subject, htmlBody)

	if m.Port == 465 || strings.EqualFold(m.Encryption, "ssl") {
		return sendSSL(addr, m.Username, m.Password, from, to, msg)
	}
	return sendStartTLS(addr, m.Username, m.Password, from, to, msg)
}

func buildMessage(fromName, from, to, subject, htmlBody string) []byte {
	var b strings.Builder
	b.WriteString("From: " + formatAddr(fromName, from) + "\r\n")
	b.WriteString("To: <" + to + ">\r\n")
	b.WriteString("Subject: " + mimeEncode(subject) + "\r\n")
	b.WriteString("MIME-Version: 1.0\r\n")
	b.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	b.WriteString("Date: " + time.Now().Format(time.RFC1123Z) + "\r\n")
	b.WriteString("\r\n")
	b.WriteString(htmlBody)
	return []byte(b.String())
}

func formatAddr(name, addr string) string {
	if name == "" {
		return "<" + addr + ">"
	}
	return mimeEncode(name) + " <" + addr + ">"
}

func mimeEncode(s string) string {
	needs := false
	for _, r := range s {
		if r > 126 || r < 32 {
			needs = true
			break
		}
	}
	if !needs {
		return s
	}
	return "=?UTF-8?B?" + base64Encode(s) + "?="
}

func sendStartTLS(addr, user, pass, from, to string, msg []byte) error {
	conn, err := net.DialTimeout("tcp", addr, 15*time.Second)
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close() }()
	cl, err := smtp.NewClient(conn, addr)
	if err != nil {
		return err
	}
	defer func() { _ = cl.Close() }()
	if err := cl.Hello("localhost"); err != nil {
		return err
	}
	if ok, _ := cl.Extension("STARTTLS"); ok {
		if err := cl.StartTLS(&tls.Config{ServerName: hostOf(addr)}); err != nil {
			return err
		}
	}
	auth := smtp.PlainAuth("", user, pass, hostOf(addr))
	if user != "" {
		if ok, _ := cl.Extension("AUTH"); ok {
			if err := cl.Auth(auth); err != nil {
				return err
			}
		}
	}
	if err := cl.Mail(from); err != nil {
		return err
	}
	if err := cl.Rcpt(to); err != nil {
		return err
	}
	w, err := cl.Data()
	if err != nil {
		return err
	}
	if _, err := w.Write(msg); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}
	return cl.Quit()
}

func sendSSL(addr, user, pass, from, to string, msg []byte) error {
	conn, err := tls.Dial("tcp", addr, &tls.Config{ServerName: hostOf(addr)})
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close() }()
	cl, err := smtp.NewClient(conn, addr)
	if err != nil {
		return err
	}
	defer func() { _ = cl.Close() }()
	if user != "" {
		if err := cl.Auth(smtp.PlainAuth("", user, pass, hostOf(addr))); err != nil {
			return err
		}
	}
	if err := cl.Mail(from); err != nil {
		return err
	}
	if err := cl.Rcpt(to); err != nil {
		return err
	}
	w, err := cl.Data()
	if err != nil {
		return err
	}
	if _, err := w.Write(msg); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}
	return cl.Quit()
}

func hostOf(addr string) string {
	if i := strings.LastIndex(addr, ":"); i > 0 {
		return addr[:i]
	}
	return addr
}

// ---------- 模板 ----------

const verifyCodeTPL = `<!doctype html>
<html><body style="font-family:Arial,'Microsoft YaHei',sans-serif;color:#333">
<div style="max-width:520px;margin:24px auto;padding:32px;border:1px solid #eee;border-radius:12px">
  <h2 style="margin:0 0 12px">{{.SiteName}}</h2>
  <p>{{.Prompt}}</p>
  <p style="font-size:30px;font-weight:bold;letter-spacing:6px;color:#0ea5e9;margin:20px 0">{{.Code}}</p>
  <p style="color:#888;font-size:12px">验证码 15 分钟内有效。若非本人操作，请忽略本邮件。</p>
</div>
</body></html>`

var codeTemplate = template.Must(template.New("code").Parse(verifyCodeTPL))

// RenderVerifyCode 渲染验证码邮件。
func RenderVerifyCode(siteName, prompt, code string) (string, error) {
	var b strings.Builder
	err := codeTemplate.Execute(&b, map[string]string{
		"SiteName": siteName, "Prompt": prompt, "Code": code,
	})
	return b.String(), err
}

// CodeKey 邮件验证码缓存键（对齐 MailService::getCodeKey）。
func CodeKey(event, email string) string {
	return "mail_code:" + event + ":" + email
}

// SendCodeMailJob 发送验证码邮件任务入口（由队列调用）。
func SendCodeMailJob(c *cache.Cache, data json.RawMessage) error {
	var in struct {
		Event    string     `json:"event"`
		Email    string     `json:"email"`
		SiteName string     `json:"site_name"`
		SMTP     SMTPConfig `json:"smtp"`
	}
	if err := json.Unmarshal(data, &in); err != nil {
		return err
	}
	code := GenerateCode(c, CodeKey(in.Event, in.Email))
	prompt := "您的验证码是："
	body, err := RenderVerifyCode(in.SiteName, prompt, code)
	if err != nil {
		return err
	}
	return in.SMTP.Send(in.Email, "验证码", body)
}

func base64Encode(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
}
