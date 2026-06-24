// Package process 图片处理驱动抽象（缩略图 / 压缩 / 水印 / 转换）
package process

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"strings"

	"github.com/disintegration/imaging"
)

// Job 处理任务
// Action 支持：thumbnail | resize | fit | fill | compress | watermark | convert
// Params 常用键：
//   - "width", "height"  目标尺寸
//   - "quality"          JPEG/WebP 质量 1-100（默认 85）
//   - "format"           输出格式 jpeg | png | webp（默认 jpeg）
//   - "watermark_text"   文字水印
//   - "raw"              []byte 形式源数据（可选，优先于 Source）
type Job struct {
	Action string         // 操作类型
	Params map[string]any // 任务参数
	Source image.Image    // 源图（内存，优先于 Params["raw"]）
}

// Result 处理结果
type Result struct {
	Data   []byte // 输出图片字节
	Width  int
	Height int
	Format string // jpeg | png | webp
}

// Driver 图片处理驱动
type Driver interface {
	Name() string
	Process(ctx context.Context, j Job) (*Result, error)
}

// Factory 工厂
type Factory func(cfg map[string]string) (Driver, error)

var registry = map[string]Factory{}

// Register 注册
func Register(name string, f Factory) { registry[name] = f }

// Get 构造
func Get(name string, cfg map[string]string) (Driver, error) {
	f, ok := registry[name]
	if !ok {
		return nil, errors.New("unsupported process driver: " + name)
	}
	return f(cfg)
}

// Drivers 列表
func Drivers() []string {
	out := make([]string, 0, len(registry))
	for k := range registry {
		out = append(out, k)
	}
	return out
}

// ============================================================
// LocalDriver 使用 imaging 库进行本地图片处理
// ============================================================

// LocalDriver 本地图片处理驱动（基于 disintegration/imaging）
type LocalDriver struct {
	DefaultQuality int
}

func NewLocalDriver(cfg map[string]string) (Driver, error) {
	d := &LocalDriver{DefaultQuality: 85}
	if cfg != nil {
		if v := cfg["quality"]; v != "" {
			var q int
			fmt.Sscanf(v, "%d", &q)
			if q > 0 && q <= 100 {
				d.DefaultQuality = q
			}
		}
	}
	return d, nil
}

func (l *LocalDriver) Name() string { return "local" }

// Process 本地处理入口
func (l *LocalDriver) Process(ctx context.Context, j Job) (*Result, error) {
	img, err := l.resolveSource(j)
	if err != nil {
		return nil, err
	}
	if img == nil {
		return nil, errors.New("process: no source image provided")
	}

	format := l.paramStr(j.Params, "format", "jpeg")
	quality := l.paramInt(j.Params, "quality", l.DefaultQuality)
	if quality <= 0 {
		quality = 85
	}

	switch strings.ToLower(j.Action) {
	case "", "thumbnail", "resize":
		w := l.paramInt(j.Params, "width", 0)
		h := l.paramInt(j.Params, "height", 0)
		if w == 0 && h == 0 {
			w = 320
		}
		img = imaging.Resize(img, w, h, imaging.Lanczos)
	case "fit":
		w := l.paramInt(j.Params, "width", 320)
		h := l.paramInt(j.Params, "height", 320)
		img = imaging.Fit(img, w, h, imaging.Lanczos)
	case "fill":
		w := l.paramInt(j.Params, "width", 320)
		h := l.paramInt(j.Params, "height", 320)
		img = imaging.Fill(img, w, h, imaging.Center, imaging.Lanczos)
	case "compress":
		// 仅做质量压缩（直接走下面编码逻辑）
	case "convert":
		// 仅做格式转换
	default:
		return nil, fmt.Errorf("process: unsupported action %q", j.Action)
	}

	bounds := img.Bounds()
	out := &Result{
		Width:  bounds.Dx(),
		Height: bounds.Dy(),
		Format: strings.ToLower(format),
	}
	buf := &bytes.Buffer{}
	switch out.Format {
	case "png":
		enc := &png.Encoder{CompressionLevel: png.DefaultCompression}
		if err := enc.Encode(buf, img); err != nil {
			return nil, err
		}
	default:
		out.Format = "jpeg"
		if err := jpeg.Encode(buf, img, &jpeg.Options{Quality: quality}); err != nil {
			return nil, err
		}
	}
	out.Data = buf.Bytes()
	_ = ctx
	return out, nil
}

// resolveSource 解析输入：raw bytes 优先，其次 image.Image
func (l *LocalDriver) resolveSource(j Job) (image.Image, error) {
	if j.Source != nil {
		return j.Source, nil
	}
	if raw, ok := j.Params["raw"].([]byte); ok && len(raw) > 0 {
		img, _, err := image.Decode(bytes.NewReader(raw))
		if err != nil {
			return nil, fmt.Errorf("process: decode source: %w", err)
		}
		return img, nil
	}
	if raw, ok := j.Params["raw"].(string); ok && raw != "" {
		b, err := base64.StdEncoding.DecodeString(raw)
		if err != nil {
			return nil, fmt.Errorf("process: decode b64 source: %w", err)
		}
		img, _, err := image.Decode(bytes.NewReader(b))
		if err != nil {
			return nil, fmt.Errorf("process: decode source: %w", err)
		}
		return img, nil
	}
	return nil, nil
}

func (l *LocalDriver) paramStr(m map[string]any, key, def string) string {
	if m == nil {
		return def
	}
	if v, ok := m[key].(string); ok && v != "" {
		return v
	}
	return def
}

func (l *LocalDriver) paramInt(m map[string]any, key string, def int) int {
	if m == nil {
		return def
	}
	switch v := m[key].(type) {
	case int:
		return v
	case int32:
		return int(v)
	case int64:
		return int(v)
	case float32:
		return int(v)
	case float64:
		return int(v)
	}
	return def
}

// ============================================================
// CustomHTTPDriver 通过 HTTP 委派给外部图片处理服务
// ============================================================

// CustomHTTPDriver 自定义 HTTP 图片处理驱动
// 请求协议：
//
//	POST {URL}
//	Headers:
//	  Content-Type: application/json
//	  Authorization: Bearer {Token} （可选）
//	Body:
//	  {
//	    "action": "thumbnail|resize|...",
//	    "format": "jpeg|png|webp",
//	    "quality": 85,
//	    "width": 320,
//	    "height": 320,
//	    "image_b64": "data:image/jpeg;base64,...",
//	    "params": {...}
//	  }
//	响应：
//	  {
//	    "format": "jpeg",
//	    "width": 320,
//	    "height": 240,
//	    "image_b64": "...."
//	  }
type CustomHTTPDriver struct {
	URL    string
	Token  string
	Method string // 默认 POST
}

func NewCustomHTTPDriver(cfg map[string]string) (Driver, error) {
	if cfg == nil || cfg["url"] == "" {
		return nil, errors.New("custom_http driver: url required")
	}
	method := cfg["method"]
	if method == "" {
		method = http.MethodPost
	}
	return &CustomHTTPDriver{
		URL:    cfg["url"],
		Token:  cfg["token"],
		Method: method,
	}, nil
}

func (c *CustomHTTPDriver) Name() string { return "custom_http" }

func (c *CustomHTTPDriver) Process(ctx context.Context, j Job) (*Result, error) {
	// 组装 payload
	raw, err := c.encodeSource(j)
	if err != nil {
		return nil, err
	}
	body := map[string]any{
		"action":    j.Action,
		"params":    j.Params,
		"image_b64": base64.StdEncoding.EncodeToString(raw),
	}
	if v, ok := j.Params["format"].(string); ok {
		body["format"] = v
	}
	if v, ok := j.Params["quality"]; ok {
		body["quality"] = v
	}
	if v, ok := j.Params["width"]; ok {
		body["width"] = v
	}
	if v, ok := j.Params["height"]; ok {
		body["height"] = v
	}

	rawJSON, _ := json.Marshal(body)
	req, err := http.NewRequestWithContext(ctx, c.Method, c.URL, bytes.NewReader(rawJSON))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("custom_http: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("custom_http: status=%d body=%s", resp.StatusCode, string(b))
	}
	var out struct {
		Format   string `json:"format"`
		Width    int    `json:"width"`
		Height   int    `json:"height"`
		ImageB64 string `json:"image_b64"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("custom_http: decode response: %w", err)
	}
	if out.ImageB64 == "" {
		return nil, errors.New("custom_http: empty image_b64 in response")
	}
	// 允许 data URL 前缀
	if idx := strings.Index(out.ImageB64, ","); strings.HasPrefix(out.ImageB64, "data:") && idx > 0 {
		out.ImageB64 = out.ImageB64[idx+1:]
	}
	data, err := base64.StdEncoding.DecodeString(out.ImageB64)
	if err != nil {
		return nil, fmt.Errorf("custom_http: decode image_b64: %w", err)
	}
	return &Result{
		Data:   data,
		Width:  out.Width,
		Height: out.Height,
		Format: strings.ToLower(out.Format),
	}, nil
}

func (c *CustomHTTPDriver) encodeSource(j Job) ([]byte, error) {
	if raw, ok := j.Params["raw"].([]byte); ok && len(raw) > 0 {
		return raw, nil
	}
	if s, ok := j.Params["raw"].(string); ok && s != "" {
		// data URL
		if idx := strings.Index(s, ","); strings.HasPrefix(s, "data:") && idx > 0 {
			s = s[idx+1:]
		}
		return base64.StdEncoding.DecodeString(s)
	}
	if j.Source != nil {
		buf := &bytes.Buffer{}
		if err := jpeg.Encode(buf, j.Source, &jpeg.Options{Quality: 90}); err != nil {
			return nil, err
		}
		return buf.Bytes(), nil
	}
	return nil, errors.New("custom_http: no source image provided")
}

func init() {
	Register("local", NewLocalDriver)
	Register("custom_http", NewCustomHTTPDriver)
}
