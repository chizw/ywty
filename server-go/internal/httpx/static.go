package httpx

import (
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/chizw/ywty/server-go/internal/config"
	"github.com/chizw/ywty/server-go/internal/setting"
	"gorm.io/gorm"
)

// staticHandler 托管前端主题构建产物（web/ npm run build 后输出到 public/）：
//   - 命中文件 → 直接返回
//   - 未命中且期望 HTML（浏览器导航 / SPA 路由）→ 返回 index.html（SPA fallback）
//   - 其余 → 404
//
// index.html 的 <title> 使用 site.title - site.subtitle 动态注入，
// 与原版主题标题注入行为一致。
func staticHandler(cfg *config.Config, gdb *gorm.DB) http.HandlerFunc {
	titleCache := &titleCache{ttl: 10 * time.Second}

	return func(w http.ResponseWriter, req *http.Request) {
		upath := path.Clean("/" + req.URL.Path)
		if upath != "/" {
			fp := filepath.Join(cfg.StaticDir, filepath.FromSlash(upath))
			if info, err := os.Stat(fp); err == nil && !info.IsDir() {
				http.ServeFile(w, req, fp)
				return
			}
		}
		if wantsHTML(req) {
			title, ok := titleCache.get(gdb)
			serveIndex(w, cfg, title, ok)
			return
		}
		http.NotFound(w, req)
	}
}

// serveIndex 输出 index.html；title.ok 为 true 时注入站点标题。
func serveIndex(w http.ResponseWriter, cfg *config.Config, title string, inject bool) {
	raw, err := os.ReadFile(filepath.Join(cfg.StaticDir, "index.html"))
	if err != nil {
		// 主题未构建：给出可直接人工检查的占位页
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("<!doctype html><html><head><meta charset=\"utf-8\"><title>ywty</title></head>" +
			"<body>前端主题未构建：请在 web/ 目录执行 npm run build，或设置 STATIC_DIR 指向构建产物目录。</body></html>"))
		return
	}
	html := string(raw)
	if inject {
		html = replaceTitle(html, title)
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	_, _ = w.Write([]byte(html))
}

// replaceTitle 对齐 blade <title>{{ title }} - {{ subtitle }}</title>。
func replaceTitle(html, title string) string {
	start := strings.Index(html, "<title")
	if start < 0 {
		return html
	}
	end := strings.Index(html[start:], "</title>")
	if end < 0 {
		return html
	}
	end += start + len("</title>")
	return html[:start] + "<title>" + title + "</title>" + html[end:]
}

type titleCache struct {
	mu       sync.Mutex
	value    string
	injected bool
	expireAt time.Time
	ttl      time.Duration
}

func (t *titleCache) get(gdb *gorm.DB) (string, bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if time.Now().Before(t.expireAt) {
		return t.value, t.injected
	}
	t.injected = false
	t.value = "ywty"
	if gdb != nil {
		if title, err := setting.String(gdb, setting.GroupSite, "title"); err == nil && title != "" {
			t.value = title
			if subtitle, err := setting.String(gdb, setting.GroupSite, "subtitle"); err == nil && subtitle != "" {
				t.value = title + " - " + subtitle
			}
			t.injected = true
		}
	}
	t.expireAt = time.Now().Add(t.ttl)
	return t.value, t.injected
}
