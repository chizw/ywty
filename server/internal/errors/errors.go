// Package errors 定义统一的业务错误码
package errors

import (
	"errors"
	"fmt"
	"net/http"
)

// 业务错误类型
type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	HTTP    int    `json:"-"`
	cause   error
}

func (e *Error) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("[%d] %s: %v", e.Code, e.Message, e.cause)
	}
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

func (e *Error) Unwrap() error { return e.cause }

func (e *Error) WithCause(err error) *Error {
	clone := *e
	clone.cause = err
	return &clone
}

func (e *Error) WithMessage(msg string) *Error {
	clone := *e
	clone.Message = msg
	return &clone
}

func New(code int, message string, httpStatus int) *Error {
	return &Error{Code: code, Message: message, HTTP: httpStatus}
}

// 通用错误码（与 HTTP 状态码保持一致区间）
var (
	// 2xx 成功（业务码用 0）
	OK = New(0, "ok", http.StatusOK)

	// 4xx 客户端错误
	BadRequest       = New(40000, "请求参数错误", http.StatusBadRequest)
	Unauthorized     = New(40100, "未登录或登录已过期", http.StatusUnauthorized)
	TokenInvalid     = New(40101, "Token 无效", http.StatusUnauthorized)
	TokenExpired     = New(40102, "Token 已过期", http.StatusUnauthorized)
	Forbidden        = New(40300, "无权访问", http.StatusForbidden)
	NotFound         = New(40400, "资源不存在", http.StatusNotFound)
	MethodNotAllowed = New(40500, "方法不被允许", http.StatusMethodNotAllowed)
	Conflict         = New(40900, "资源冲突", http.StatusConflict)
	Validation       = New(42200, "数据校验失败", http.StatusUnprocessableEntity)
	TooManyRequests  = New(42900, "请求过于频繁", http.StatusTooManyRequests)

	// 5xx 服务端错误
	Internal       = New(50000, "服务器内部错误", http.StatusInternalServerError)
	NotImplemented = New(50100, "功能未实现", http.StatusNotImplemented)
	ServiceUnavail = New(50300, "服务暂不可用", http.StatusServiceUnavailable)

	// 业务错误码 1xxxx
	UserNotFound      = New(10001, "用户不存在", http.StatusNotFound)
	UserExists        = New(10002, "用户已存在", http.StatusConflict)
	UserDisabled      = New(10003, "用户已被禁用", http.StatusForbidden)
	PasswordIncorrect = New(10004, "密码错误", http.StatusUnauthorized)
	CaptchaIncorrect  = New(10005, "验证码错误", http.StatusUnauthorized)
	CodeIncorrect     = New(10006, "验证码错误或已过期", http.StatusUnauthorized)

	// 存储 / 文件
	StorageNotFound   = New(20001, "存储驱动不存在", http.StatusNotFound)
	StorageUploadFail = New(20002, "文件上传失败", http.StatusInternalServerError)
	FileTooLarge      = New(20003, "文件超过大小限制", http.StatusRequestEntityTooLarge)
	FileTypeInvalid   = New(20004, "文件类型不被允许", http.StatusUnsupportedMediaType)

	// 照片
	PhotoNotFound  = New(30001, "照片不存在", http.StatusNotFound)
	CapacityExceed = New(30002, "存储容量已用尽", http.StatusForbidden)

	// 支付
	OrderNotFound         = New(40001, "订单不存在", http.StatusNotFound)
	OrderPaid             = New(40002, "订单已支付", http.StatusConflict)
	OrderCanceled         = New(40003, "订单已取消", http.StatusConflict)
	PaymentFailed         = New(40004, "支付失败", http.StatusInternalServerError)
	PaymentChannelInvalid = New(40005, "支付渠道无效", http.StatusBadRequest)

	// 验证码
	CodeSendTooFrequent = New(50001, "请求过于频繁，请稍后再试", http.StatusTooManyRequests)
	CodeInvalid         = New(50002, "验证码错误", http.StatusUnauthorized)
	CodeExpired         = New(50003, "验证码已过期", http.StatusUnauthorized)
	CodeUsed            = New(50004, "验证码已使用", http.StatusUnauthorized)
	CodeChannelInvalid  = New(50005, "验证码渠道无效", http.StatusBadRequest)
	MailSendFailed      = New(50006, "邮件发送失败", http.StatusInternalServerError)
	SMSSendFailed       = New(50007, "短信发送失败", http.StatusInternalServerError)

	// 账号
	PhoneBound  = New(10101, "手机号已被绑定", http.StatusConflict)
	EmailBound  = New(10102, "邮箱已被绑定", http.StatusConflict)
	OldPwdWrong = New(10103, "原密码错误", http.StatusUnauthorized)

	// Token
	TokenNotFound = New(60001, "Token 不存在", http.StatusNotFound)
	TokenRevoked  = New(60002, "Token 已被吊销", http.StatusUnauthorized)

	// 业务通用
	ResourceNotFound = New(70001, "资源不存在", http.StatusNotFound)
	AlreadyExists    = New(70002, "资源已存在", http.StatusConflict)
)

// As 检查是否为业务错误
func As(err error) (*Error, bool) {
	var be *Error
	if errors.As(err, &be) {
		return be, true
	}
	return nil, false
}

// Wrap 包装任意 error 为业务错误
func Wrap(err error, biz *Error) *Error {
	if err == nil {
		return nil
	}
	return biz.WithCause(err)
}
