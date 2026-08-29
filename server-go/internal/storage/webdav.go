package storage

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/chizw/ywty/server-go/internal/config"
	"github.com/studio-b12/gowebdav"
)

// WebDAV 适配器（base_uri/username/password/auth_type + root 前缀）。
type WebDAV struct {
	client *gowebdav.Client
	root   string
}

// NewWebDAV 构建 WebDAV 适配器。
func NewWebDAV(rawOptions string, cfg *config.Config) (Filesystem, error) {
	var o struct {
		BaseURI  string `json:"base_uri"`
		Username string `json:"username"`
		Password string `json:"password"`
		AuthType string `json:"auth_type"`
		Root     string `json:"root"`
	}
	if rawOptions != "" {
		_ = json.Unmarshal([]byte(rawOptions), &o)
	}
	_ = cfg
	if o.BaseURI == "" {
		return nil, errors.New("storage: WebDAV 配置缺少 base_uri")
	}
	client := gowebdav.NewClient(o.BaseURI, o.Username, o.Password)
	switch strings.ToLower(o.AuthType) {
	case "digest":
		client.SetTransport(nil) // gowebdav 自动协商
	case "", "basic":
		// NewClient 默认 basic
	}
	root := strings.Trim(o.Root, "/")
	return &WebDAV{client: client, root: root}, nil
}

func (w *WebDAV) key(path string) string {
	if w.root == "" {
		return path
	}
	return w.root + "/" + strings.TrimPrefix(path, "/")
}

func (w *WebDAV) Write(path string, data []byte) error {
	return w.client.Write(w.key(path), data, 0o644)
}

func (w *WebDAV) AppendOrCreate(path string, data []byte) error {
	if w.Exists(path) {
		return nil
	}
	return w.Write(path, data)
}

func (w *WebDAV) Exists(path string) bool {
	_, err := w.client.Stat(w.key(path))
	return err == nil
}

func (w *WebDAV) Read(path string) ([]byte, error) {
	return w.client.Read(w.key(path))
}

func (w *WebDAV) Delete(path string) error {
	return w.client.Remove(w.key(path))
}

func (w *WebDAV) Move(from, to string) error {
	return w.client.Rename(w.key(from), w.key(to), false)
}
