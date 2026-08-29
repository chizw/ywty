// Package httpx 组装 HTTP 路由、中间件与静态资源托管。
package httpx

import (
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/chizw/ywty/server-go/internal/config"
	"github.com/chizw/ywty/server-go/internal/support/r"
	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

// New 构建应用路由。
func New(cfg *config.Config, gdb *gorm.DB) http.Handler {
	mux := chi.NewRouter()
	mux.Use(recoverMiddleware)
	mux.Use(accessLog)

	mux.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = w.Write([]byte("ok"))
	})

	mux.Mount("/api", apiRouter(cfg, gdb))
	mux.Handle("/*", staticHandler(cfg, gdb))

	return mux
}

// recoverMiddleware 捕获 panic，按统一 envelope 返回 500。
func recoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				slog.Error("panic", "path", req.URL.Path, "recover", rec)
				r.ErrorWithCode(w, http.StatusInternalServerError, "Server Error")
			}
		}()
		next.ServeHTTP(w, req)
	})
}

// accessLog 极简访问日志。
func accessLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		start := time.Now()
		sw := &statusWriter{ResponseWriter: w, code: http.StatusOK}
		next.ServeHTTP(sw, req)
		if req.URL.Path == "/healthz" {
			return // 健康检查刷屏，跳过
		}
		slog.Info("http",
			"method", req.Method,
			"path", req.URL.Path,
			"status", sw.code,
			"cost", time.Since(start).Round(time.Millisecond).String(),
		)
	})
}

type statusWriter struct {
	http.ResponseWriter
	code int
}

func (w *statusWriter) WriteHeader(code int) {
	w.code = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *statusWriter) Flush() {
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// wantsHTML 判断请求是否期望 HTML（浏览器导航 / SPA 路由）。
func wantsHTML(req *http.Request) bool {
	if req.Method != http.MethodGet && req.Method != http.MethodHead {
		return false
	}
	accept := req.Header.Get("Accept")
	return accept == "" || strings.Contains(accept, "text/html") || strings.Contains(accept, "*/*")
}
