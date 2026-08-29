// Package imageproc 纯 Go 图像处理管线（解码/编码/缩略图/尺寸），
// 可选外挂 vips CLI 处理 heic/avif 等高级格式（保持 CGO_ENABLED=0 静态编译）。
package imageproc

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	xbmp "golang.org/x/image/bmp"
	xwebp "golang.org/x/image/webp"
)

// NormalizeExt 规范化扩展名。
func NormalizeExt(ext string) string {
	return strings.ToLower(strings.TrimPrefix(ext, "."))
}

// SupportedDecode 无 vips 时可解码的格式。
func SupportedDecode() map[string]bool {
	return map[string]bool{"jpg": true, "jpeg": true, "png": true, "gif": true, "webp": true, "bmp": true}
}

// SupportedEncode 无 vips 时可编码的格式。
func SupportedEncode() map[string]bool {
	return map[string]bool{"jpg": true, "jpeg": true, "png": true, "gif": true}
}

var vipsAvailable *bool

// HasVips 检测系统 vips CLI（首次调用后缓存）。
func HasVips() bool {
	if vipsAvailable != nil {
		return *vipsAvailable
	}
	path, err := exec.LookPath("vips")
	ok := err == nil && path != ""
	vipsAvailable = &ok
	return ok
}

// Decode 解码图片；heic/avif 等在 vips 可用时经 vips 转 png 再解。
func Decode(data []byte, ext string) (image.Image, error) {
	ext = NormalizeExt(ext)
	r := bytes.NewReader(data)
	switch ext {
	case "jpg", "jpeg":
		return jpeg.Decode(r)
	case "png":
		return png.Decode(r)
	case "gif":
		return gif.Decode(r)
	case "webp":
		return xwebp.Decode(r)
	case "bmp":
		return xbmp.Decode(r)
	}
	if HasVips() {
		out, err := vipsConvert(data, ext, "png")
		if err == nil {
			return png.Decode(bytes.NewReader(out))
		}
	}
	return nil, fmt.Errorf("imageproc: 不支持的图片格式 %q", ext)
}

// Encode 按目标扩展名编码（quality 对 jpeg/webp 有效）。
// 无 vips 时 webp/avif/heic 等目标格式回退为 jpeg/png。
func Encode(img image.Image, ext string, quality int) ([]byte, string, error) {
	ext = NormalizeExt(ext)
	var buf bytes.Buffer
	var err error
	switch ext {
	case "png":
		err = png.Encode(&buf, img)
	case "gif":
		err = gif.Encode(&buf, img, nil)
	case "webp", "avif", "heic", "heif", "tif", "tiff", "jp2", "j2k", "jp2k", "jpf", "jpm", "jpg2", "j2c", "jpc", "jpx":
		if HasVips() && ext == "webp" {
			// vips 需要 png 中转
			tmp, _ := pngEncodeToBytes(img)
			out, cerr := vipsConvert(tmp, "png", ext)
			if cerr == nil {
				return out, ext, nil
			}
		}
		// 回退：jpeg（有透明通道时 png）
		if hasAlpha(img) {
			err = png.Encode(&buf, img)
			ext = "png"
		} else {
			err = jpeg.Encode(&buf, img, jpegOpts(quality))
			ext = "jpg"
		}
	default: // jpg/jpeg 及未知
		err = jpeg.Encode(&buf, img, jpegOpts(quality))
		ext = "jpg"
	}
	if err != nil {
		return nil, "", err
	}
	return buf.Bytes(), ext, nil
}

func jpegOpts(quality int) *jpeg.Options {
	if quality <= 0 || quality > 100 {
		quality = 90
	}
	return &jpeg.Options{Quality: quality}
}

func pngEncodeToBytes(img image.Image) ([]byte, error) {
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// hasAlpha 粗判是否含透明通道（决定回退编码格式）。
func hasAlpha(img image.Image) bool {
	switch m := img.(type) {
	case *image.NRGBA:
		for i := 3; i < len(m.Pix); i += 4 {
			if m.Pix[i] < 255 {
				return true
			}
		}
	case *image.RGBA:
		for i := 3; i < len(m.Pix); i += 4 {
			if m.Pix[i] < 255 {
				return true
			}
		}
	case interface {
		image.Image
		At(x, y int) color.Color
	}:
		b := m.Bounds()
		for y := b.Min.Y; y < b.Max.Y; y++ {
			for x := b.Min.X; x < b.Max.X; x++ {
				if hasAlphaColor(m.At(x, y)) {
					return true
				}
			}
		}
	}
	return false
}

// Thumbnail 生成缩略图：等比缩到 max 以内（只缩不放），按原格式编码。
// 对齐 PhotoService::generateThumbnail（scale + AutoEncoder）。
func Thumbnail(data []byte, ext string, maxSize, quality int) ([]byte, string, error) {
	img, err := Decode(data, ext)
	if err != nil {
		return nil, "", err
	}
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	if w <= 0 || h <= 0 {
		return nil, "", errors.New("imageproc: 无效图片尺寸")
	}
	if w > maxSize && h > maxSize {
		scale := float64(maxSize) / float64(w)
		if sh := float64(maxSize) / float64(h); sh < scale {
			scale = sh
		}
		img = Resize(img, int(float64(w)*scale), int(float64(h)*scale))
	}
	// AutoEncoder 语义：jpg→jpg，png→png（保透明），gif→gif
	ext = NormalizeExt(ext)
	switch ext {
	case "png":
		out, err := pngEncodeToBytes(img)
		return out, "png", err
	case "gif":
		var buf bytes.Buffer
		if err := gif.Encode(&buf, img, nil); err != nil {
			return nil, "", err
		}
		return buf.Bytes(), "gif", nil
	default:
		var buf bytes.Buffer
		if err := jpeg.Encode(&buf, img, jpegOpts(quality)); err != nil {
			return nil, "", err
		}
		return buf.Bytes(), "jpg", nil
	}
}

// vipsConvert 用 vips CLI 做格式转换（inExt/outExt 为扩展名标识）。
func vipsConvert(data []byte, inExt, outExt string) ([]byte, error) {
	if !HasVips() {
		return nil, errors.New("imageproc: vips 不可用")
	}
	dir, err := os.MkdirTemp("", "ywty-vips-")
	if err != nil {
		return nil, err
	}
	defer func() { _ = os.RemoveAll(dir) }()
	in := filepath.Join(dir, "in."+NormalizeExt(inExt))
	out := filepath.Join(dir, "out."+NormalizeExt(outExt))
	if err := os.WriteFile(in, data, 0o600); err != nil {
		return nil, err
	}
	cmd := exec.Command("vips", "copy", in, out)
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("imageproc: vips 转换失败: %w", err)
	}
	return os.ReadFile(out)
}

// Size 解码取宽高。
func Size(data []byte, ext string) (int, int) {
	img, err := Decode(data, ext)
	if err != nil {
		return 0, 0
	}
	b := img.Bounds()
	return b.Dx(), b.Dy()
}

// MimeTypeByExt 扩展名 → MIME（与 Intervention MediaType 对齐的常用值）。
func MimeTypeByExt(ext string) string {
	switch NormalizeExt(ext) {
	case "jpg", "jpeg":
		return "image/jpeg"
	case "png":
		return "image/png"
	case "gif":
		return "image/gif"
	case "webp":
		return "image/webp"
	case "avif":
		return "image/avif"
	case "bmp":
		return "image/bmp"
	case "tif", "tiff":
		return "image/tiff"
	case "heic":
		return "image/heic"
	case "heif":
		return "image/heif"
	case "jp2", "j2k", "jp2k", "jpf", "jpm", "jpg2", "j2c", "jpc", "jpx":
		return "image/jp2"
	default:
		return "application/octet-stream"
	}
}
