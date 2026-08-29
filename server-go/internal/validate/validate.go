// Package validate 提供请求校验错误收集（原版格式）：
// 失败时以 {"message": "The given data was invalid.", "data": {"errors": {字段: [消息]}}} 422 输出。
package validate

import (
	"net/http"
	"regexp"

	"github.com/chizw/ywty/server-go/internal/support/r"
)

type Errors map[string][]string

type V struct {
	errs Errors
}

func New() *V { return &V{errs: Errors{}} }

func (v *V) Add(field, attr, rule string) {
	v.errs[field] = append(v.errs[field], attr+" "+rule)
}

func (v *V) Fail() bool { return len(v.errs) > 0 }

// Errors 返回收集结果。
func (v *V) Errors() Errors { return v.errs }

// Respond 输出 422 envelope。
func (v *V) Respond(w http.ResponseWriter) {
	r.ErrorData(w, http.StatusUnprocessableEntity, "The given data was invalid.",
		map[string]any{"errors": v.errs})
}

var (
	emailPattern = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)
	urlPattern   = regexp.MustCompile(`^https?://[^\s]+$`)
	alphaDash    = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
)

func Required(s string) bool      { return s != "" }
func Email(s string) bool         { return emailPattern.MatchString(s) }
func URL(s string) bool           { return urlPattern.MatchString(s) }
func AlphaDash(s string) bool     { return alphaDash.MatchString(s) }
func MaxLen(s string, n int) bool { return len(s) <= n }
func MinLen(s string, n int) bool { return len(s) >= n }

// In 判断值是否在枚举内。
func In(s string, values ...string) bool {
	for _, v := range values {
		if s == v {
			return true
		}
	}
	return false
}
