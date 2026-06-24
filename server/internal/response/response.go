// Package response 提供统一的 API 响应结构
package response

import (
	"github.com/gin-gonic/gin"

	bizerr "github.com/ywty/server/internal/errors"
)

// PageMeta 兼容老版本 meta 分页结构
type PageMeta struct {
	CurrentPage int   `json:"current_page"`
	PerPage     int   `json:"per_page"`
	Total       int64 `json:"total"`
	LastPage    int   `json:"last_page"`
}

type Body struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Meta    *PageMeta   `json:"meta,omitempty"`
}

const (
	HeaderTraceID = "X-Trace-Id"
)

// JSON 写入普通 JSON 响应
func JSON(c *gin.Context, status int, data interface{}) {
	c.JSON(status, Body{Code: 0, Message: "ok", Data: data})
}

// Page 写入分页响应
func Page(c *gin.Context, data interface{}, meta PageMeta) {
	c.JSON(200, Body{Code: 0, Message: "ok", Data: data, Meta: &meta})
}

// Success 200 成功
func Success(c *gin.Context, data interface{}) {
	JSON(c, 200, data)
}

// Fail 业务错误
func Fail(c *gin.Context, err error) {
	if be, ok := bizerr.As(err); ok {
		c.AbortWithStatusJSON(be.HTTP, Body{Code: be.Code, Message: be.Message})
		return
	}
	// 非业务错误：内部错误只对外返回通用文案
	_ = err
	c.AbortWithStatusJSON(500, Body{Code: 50000, Message: "服务器内部错误"})
}

// FailCode 便捷方法：传入业务错误对象
func FailCode(c *gin.Context, be *bizerr.Error) {
	if be == nil {
		c.AbortWithStatusJSON(500, Body{Code: 50000, Message: "internal error"})
		return
	}
	c.AbortWithStatusJSON(be.HTTP, Body{Code: be.Code, Message: be.Message})
}

// FailMsg 便捷方法：直接传业务码 + 文案 + HTTP
func FailMsg(c *gin.Context, httpStatus, code int, message string) {
	c.AbortWithStatusJSON(httpStatus, Body{Code: code, Message: message})
}
