// Package storage 移植 app/Contracts/StorageAbstract + app/Drivers/Storage：
// 按 storages.options 装配文件系统。M2 实现 Local，其余驱动 M5 接入。
package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/chizw/ywty/server-go/internal/config"
)

// Filesystem 储存适配器接口（flysystem 语义的最小子集）。
type Filesystem interface {
	Write(path string, data []byte) error
	AppendOrCreate(path string, data []byte) error // 不存在则写入
	Exists(path string) bool
	Read(path string) ([]byte, error)
	Delete(path string) error
	Move(from, to string) error
}

// Options storages.options JSON 的结构化形式。
type Options struct {
	Raw               string `json:"-"` // 原始 JSON（透传给非本地驱动）
	PublicURL         string `json:"public_url"`
	NamingRule        string `json:"naming_rule"`
	GenerateThumbnail *bool  `json:"generate_thumbnail"`
	ThumbnailMaxSize  int    `json:"thumbnail_max_size"`
	ThumbnailQuality  int    `json:"thumbnail_quality"`
	Root              string `json:"root"`
}

// ParseOptions 解析 options JSON（容忍空值）。
func ParseOptions(raw string) Options {
	var o Options
	if raw != "" {
		_ = json.Unmarshal([]byte(raw), &o)
	}
	o.Raw = raw
	if o.NamingRule == "" {
		o.NamingRule = "{Ymd}/{md5}"
	}
	if o.ThumbnailMaxSize == 0 {
		o.ThumbnailMaxSize = 800
	}
	if o.ThumbnailQuality == 0 {
		o.ThumbnailQuality = 90
	}
	return o
}

// Storage 一条 storages 表记录的运行时形态。
type Storage struct {
	ID       int64
	Name     string
	Prefix   string
	Provider string
	Options  Options
}

// FilesystemFor 按提供者构建适配器。
// s3/oss/cos 共用 S3 兼容实现（OSS/COS 需配置其 S3 兼容端点）。
func FilesystemFor(s *Storage, cfg *config.Config) (Filesystem, error) {
	switch s.Provider {
	case "local":
		return NewLocal(s.Options.Root, cfg), nil
	case "s3", "oss", "cos":
		return NewS3FromRaw(s.Options.Raw)
	case "webdav":
		return NewWebDAV(s.Options.Raw, cfg)
	default:
		return nil, fmt.Errorf("暂不支持的储存提供者: %s（七牛/又拍/FTP/SFTP 在后续版本支持）", s.Provider)
	}
}

// Local 本地储存适配器（对齐 LocalStorage::getAdapter：root 不存在时回退 public 磁盘根）。
type Local struct {
	root string
}

// NewLocal 构建本地适配器。
func NewLocal(root string, cfg *config.Config) *Local {
	if root == "" || root == "/" || root == "." || !dirExists(root) {
		// 对齐 PHP：回退到 public 磁盘根目录
		if cfg != nil && cfg.UploadsDir != "" && dirExists(cfg.UploadsDir) {
			root = cfg.UploadsDir
		} else if root == "" || root == "/" || root == "." {
			root = "storage/app/public"
		}
	}
	return &Local{root: root}
}

func dirExists(p string) bool {
	info, err := os.Stat(p)
	return err == nil && info.IsDir()
}

// Resolve 安全拼路径（防目录穿越）。
func (l *Local) Resolve(path string) (string, error) {
	clean := filepath.Clean("/" + strings.TrimPrefix(filepath.ToSlash(path), "/"))
	full := filepath.Join(l.root, filepath.FromSlash(clean))
	rel, err := filepath.Rel(l.root, full)
	if err != nil || strings.HasPrefix(rel, "..") {
		return "", errors.New("storage: 非法路径")
	}
	return full, nil
}

func (l *Local) Write(path string, data []byte) error {
	full, err := l.Resolve(path)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		return err
	}
	return os.WriteFile(full, data, 0o644)
}

func (l *Local) AppendOrCreate(path string, data []byte) error {
	if l.Exists(path) {
		return nil
	}
	return l.Write(path, data)
}

func (l *Local) Exists(path string) bool {
	full, err := l.Resolve(path)
	if err != nil {
		return false
	}
	_, err = os.Stat(full)
	return err == nil
}

func (l *Local) Read(path string) ([]byte, error) {
	full, err := l.Resolve(path)
	if err != nil {
		return nil, err
	}
	return os.ReadFile(full)
}

func (l *Local) Delete(path string) error {
	full, err := l.Resolve(path)
	if err != nil {
		return nil
	}
	return os.Remove(full)
}

func (l *Local) Move(from, to string) error {
	src, err := l.Resolve(from)
	if err != nil {
		return err
	}
	dst, err := l.Resolve(to)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	return os.Rename(src, dst)
}

// PublicURL 组装公开访问 URL：public_url/{prefix}/{path}（对齐 Photo::publicUrl）。
func PublicURL(s *Storage, cfg *config.Config, path string) string {
	base := s.Options.PublicURL
	if base == "" && cfg != nil {
		base = cfg.AppURL
	}
	prefix := strings.Trim(s.Prefix, "/")
	p := strings.Trim(path, "/")
	if prefix != "" {
		return strings.TrimRight(base, "/") + "/" + prefix + "/" + p
	}
	return strings.TrimRight(base, "/") + "/" + p
}
