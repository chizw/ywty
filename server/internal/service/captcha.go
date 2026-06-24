// Package service 图形验证码服务
package service

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math/big"
	"sync"
	"time"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

const (
	captchaLength = 4
	captchaWidth  = 120
	captchaHeight = 40
	captchaExpire = 5 * time.Minute
)

// captchaChars 去掉易混淆字符（0/O/1/I/L）
const captchaChars = "ABCDEFGHJKMNPQRSTUVWXYZ23456789"

// captchaEntry 内存中的验证码条目
type captchaEntry struct {
	code      string
	expiresAt time.Time
}

// CaptchaService 图形验证码服务（内存存储，带过期）
type CaptchaService struct {
	store sync.Map
}

// NewCaptchaService 创建验证码服务并启动过期清理
func NewCaptchaService() *CaptchaService {
	s := &CaptchaService{}
	go s.cleanup()
	return s
}

// Generate 生成一个新的图形验证码
// 返回: id（用于后续校验）、code（明文验证码，仅供内部使用）、imageBase64（data URI）
func (s *CaptchaService) Generate() (id, code, imageBase64 string) {
	code = randomCaptchaCode(captchaLength)
	id = randomCaptchaID(16)
	s.store.Store(id, captchaEntry{code: code, expiresAt: time.Now().Add(captchaExpire)})

	img := drawCaptchaImage(code)
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return id, code, ""
	}
	imageBase64 = "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes())
	return id, code, imageBase64
}

// Verify 校验验证码：成功后立即删除（一次性）
func (s *CaptchaService) Verify(id, code string) bool {
	if id == "" || code == "" {
		return false
	}
	v, ok := s.store.LoadAndDelete(id)
	if !ok {
		return false
	}
	entry := v.(captchaEntry)
	if time.Now().After(entry.expiresAt) {
		return false
	}
	return entry.code == code
}

// cleanup 定期清理过期条目
func (s *CaptchaService) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		now := time.Now()
		s.store.Range(func(k, v any) bool {
			if now.After(v.(captchaEntry).expiresAt) {
				s.store.Delete(k)
			}
			return true
		})
	}
}

// randomCaptchaCode 生成指定长度的字母数字验证码
func randomCaptchaCode(n int) string {
	b := make([]byte, n)
	max := big.NewInt(int64(len(captchaChars)))
	for i := range b {
		idx, _ := rand.Int(rand.Reader, max)
		b[i] = captchaChars[idx.Int64()]
	}
	return string(b)
}

// randomCaptchaID 生成十六进制随机 ID
func randomCaptchaID(n int) string {
	const hex = "0123456789abcdef"
	b := make([]byte, n)
	max := big.NewInt(int64(len(hex)))
	for i := range b {
		idx, _ := rand.Int(rand.Reader, max)
		b[i] = hex[idx.Int64()]
	}
	return string(b)
}

// drawCaptchaImage 用标准库绘制验证码图片
func drawCaptchaImage(code string) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, captchaWidth, captchaHeight))
	// 浅色背景
	draw.Draw(img, img.Bounds(), &image.Uniform{color.RGBA{R: 238, G: 246, B: 255, A: 255}}, image.Point{}, draw.Src)

	// 干扰线
	for i := 0; i < 6; i++ {
		x0 := randInt(captchaWidth)
		y0 := randInt(captchaHeight)
		x1 := randInt(captchaWidth)
		y1 := randInt(captchaHeight)
		drawLine(img, x0, y0, x1, y1, randomNoiseColor())
	}

	// 干扰点
	for i := 0; i < 80; i++ {
		img.Set(randInt(captchaWidth), randInt(captchaHeight), randomNoiseColor())
	}

	// 绘制字符
	face := basicfont.Face7x13
	charW := face.Advance // basicfont.Face.Advance 为每字前进宽度（int）
	if charW <= 0 {
		charW = 7
	}
	totalW := charW * len(code)
	startX := (captchaWidth - totalW) / 2
	for i := 0; i < len(code); i++ {
		d := &font.Drawer{
			Dst:  img,
			Src:  &image.Uniform{randomCharColor()},
			Face: face,
			Dot: fixed.Point26_6{
				X: fixed.I(startX + i*charW),
				Y: fixed.I(captchaHeight/2 + 6),
			},
		}
		d.DrawString(string(code[i]))
	}
	return img
}

func drawLine(img *image.RGBA, x0, y0, x1, y1 int, c color.Color) {
	// Bresenham 直线算法
	dx := abs(x1 - x0)
	dy := -abs(y1 - y0)
	sx := -1
	if x0 < x1 {
		sx = 1
	}
	sy := -1
	if y0 < y1 {
		sy = 1
	}
	err := dx + dy
	for {
		if x0 >= 0 && x0 < captchaWidth && y0 >= 0 && y0 < captchaHeight {
			img.Set(x0, y0, c)
		}
		if x0 == x1 && y0 == y1 {
			return
		}
		e2 := 2 * err
		if e2 >= dy {
			err += dy
			x0 += sx
		}
		if e2 <= dx {
			err += dx
			y0 += sy
		}
	}
}

func randomNoiseColor() color.RGBA {
	return color.RGBA{
		R: uint8(randInt(180)),
		G: uint8(randInt(180)),
		B: uint8(randInt(180)),
		A: 200,
	}
}

func randomCharColor() color.RGBA {
	return color.RGBA{
		R: uint8(20 + randInt(80)),
		G: uint8(20 + randInt(80)),
		B: uint8(80 + randInt(120)),
		A: 255,
	}
}

func randInt(max int) int {
	if max <= 0 {
		return 0
	}
	n, _ := rand.Int(rand.Reader, big.NewInt(int64(max)))
	return int(n.Int64())
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
