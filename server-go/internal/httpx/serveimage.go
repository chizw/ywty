package httpx

import (
	"bytes"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/chizw/ywty/server-go/internal/config"
	"github.com/chizw/ywty/server-go/internal/imageproc"
	"github.com/chizw/ywty/server-go/internal/storage"
	"gorm.io/gorm"
)

// imageExtPattern 图片扩展名正则（决定哪些路径走储存直出）。
var imageExtPattern = regexp.MustCompile(`(?i)\.(jpg|jpeg|webp|avif|bmp|gif|png|tif|tiff|jp2|j2k|jp2k|jpf|jpm|jpg2|j2c|jpc|jpx|heic|heif)$`)

// combinedStatic 图片直出 + 静态资源 + 管理后台 + SPA fallback
// （对齐原版 nginx try_files → 程序处理的流转）。
func combinedStatic(cfg *config.Config, gdb *gorm.DB) http.HandlerFunc {
	inner := staticHandler(cfg, gdb)
	adminHandler := adminStaticHandler(cfg)
	return func(w http.ResponseWriter, req *http.Request) {
		if imageExtPattern.MatchString(req.URL.Path) &&
			serveStorageImage(cfg, gdb, w, req) {
			return
		}
		// /storage/* → public 磁盘（缩略图、后台素材）
		if p, ok := strings.CutPrefix(req.URL.Path, "/storage/"); ok && !strings.Contains(p, "..") {
			full := filepath.Join("storage/app/public", filepath.FromSlash(path.Clean("/"+p)))
			if info, err := os.Stat(full); err == nil && !info.IsDir() {
				http.ServeFile(w, req, full)
				return
			}
		}
		// /admin/* → 管理后台 SPA
		if _, ok := strings.CutPrefix(req.URL.Path, "/admin"); ok {
			adminHandler(w, req)
			return
		}
		inner.ServeHTTP(w, req)
	}
}

// adminStaticHandler 管理后台静态托管（SPA fallback 到 admin/index.html）。
func adminStaticHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		upath := path.Clean("/" + req.URL.Path)
		if rel, ok := strings.CutPrefix(upath, "/admin/"); ok && rel != "" {
			fp := filepath.Join(cfg.AdminStaticDir, filepath.FromSlash(rel))
			if info, err := os.Stat(fp); err == nil && !info.IsDir() {
				http.ServeFile(w, req, fp)
				return
			}
		}
		index := filepath.Join(cfg.AdminStaticDir, "index.html")
		raw, err := os.ReadFile(index)
		if err != nil {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("<!doctype html><html><head><meta charset=\"utf-8\"><title>admin</title></head>" +
				"<body>管理后台未构建：请在 admin/ 目录执行 npm install && npm run build。</body></html>"))
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-cache")
		_, _ = w.Write(raw)
	}
}

// serveStorageImage 按前缀定位储存并输出本地图片；返回 false 表示未命中（流转到静态/SPA）。
func serveStorageImage(cfg *config.Config, gdb *gorm.DB, w http.ResponseWriter, req *http.Request) bool {
	rel := strings.TrimPrefix(req.URL.Path, "/")
	prefix, rest, _ := strings.Cut(rel, "/")
	if prefix == "" || rest == "" {
		return false
	}
	var row struct {
		ID       int64
		Provider string
		Options  *string
	}
	gdb.Raw("SELECT `id`, `provider`, `options` FROM `storages` WHERE `prefix` = ? AND `deleted_at` IS NULL LIMIT 1", prefix).
		Scan(&row)
	if row.ID == 0 || row.Provider != "local" {
		return false
	}
	opts := storage.ParseOptions(derefStr(row.Options))
	fs := storage.NewLocal(opts.Root, cfg)
	data, err := fs.Read(rest)
	if err != nil {
		// 本地无此文件：尝试原图转码输出（on-the-fly 简化实现——仅当带处理参数时）
		return false
	}

	// glide 风格参数（w/h/q）：按需缩放输出
	q := req.URL.Query()
	wParam, hParam := q.Get("w"), q.Get("h")
	if (wParam != "" || hParam != "") && imageproc.SupportedDecode()[imageproc.NormalizeExt(path.Ext(rest))] {
		if img, err := imageproc.Decode(data, path.Ext(rest)); err == nil {
			b := img.Bounds()
			tw := atoiOr(b.Dx(), wParam)
			th := atoiOr(scaleH(b.Dy(), b.Dx(), wParam), hParam)
			img = imageproc.Resize(img, tw, th)
			var buf bytes.Buffer
			outExt := imageproc.NormalizeExt(path.Ext(rest))
			quality := atoiOr(90, q.Get("q"))
			if data2, ext2, err := imageproc.Encode(img, outExt, quality); err == nil {
				buf.Write(data2)
				outExt = ext2
				w.Header().Set("Content-Type", imageproc.MimeTypeByExt(outExt))
				w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
				_, _ = w.Write(buf.Bytes())
				return true
			}
		}
	}

	w.Header().Set("Content-Type", imageproc.MimeTypeByExt(imageproc.NormalizeExt(path.Ext(rest))))
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	_, _ = w.Write(data)
	return true
}

func scaleH(h, w int, newW string) int {
	n := atoiOr(0, newW)
	if n <= 0 || w == 0 {
		return h
	}
	return h * n / w
}

func atoiOr(def int, s string) int {
	if s == "" {
		return def
	}
	n := 0
	for _, r := range s {
		if r < '0' || r > '9' {
			return def
		}
		n = n*10 + int(r-'0')
	}
	if n <= 0 {
		return def
	}
	return n
}

func derefStr(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}
