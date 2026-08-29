package imageproc

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"os"
	"strconv"
	"strings"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

// ParseHexColor 解析 #rgb / #rrggbb。
func ParseHexColor(s string) (color.Color, error) {
	s = strings.TrimPrefix(strings.TrimSpace(s), "#")
	switch len(s) {
	case 3:
		var r, g, b int
		if _, err := fmt.Sscanf(s, "%1x%1x%1x", &r, &g, &b); err != nil {
			return nil, err
		}
		return color.RGBA{R: uint8(r * 17), G: uint8(g * 17), B: uint8(b * 17), A: 255}, nil
	case 6:
		v, err := strconv.ParseUint(s, 16, 32)
		if err != nil {
			return nil, err
		}
		return color.RGBA{R: uint8(v >> 16), G: uint8(v >> 8), B: uint8(v), A: 255}, nil
	}
	return color.Black, nil
}

// RenderText 把文字渲染为带透明的 RGBA 图片（近似 PHP 水印文本图层）。
func RenderText(text, fontPath string, sizePx float64, colorHex string) (image.Image, error) {
	if text == "" {
		return nil, errors.New("imageproc: 空水印文本")
	}
	ttf, err := os.ReadFile(fontPath)
	if err != nil {
		return nil, fmt.Errorf("imageproc: 字体文件读取失败: %w", err)
	}
	f, err := opentype.Parse(ttf)
	if err != nil {
		return nil, fmt.Errorf("imageproc: 字体解析失败: %w", err)
	}
	size := sizePx
	if size <= 0 {
		size = 24
	}
	face, err := opentype.NewFace(f, &opentype.FaceOptions{Size: size, DPI: 72, Hinting: font.HintingFull})
	if err != nil {
		return nil, err
	}

	// 测量
	w := font.MeasureString(face, text).Ceil()
	metrics := face.Metrics()
	h := metrics.Ascent.Ceil() + metrics.Descent.Ceil()
	if w <= 0 || h <= 0 {
		return nil, errors.New("imageproc: 水印文本尺寸异常")
	}

	dst := image.NewRGBA(image.Rect(0, 0, w+2, h+2))
	col, err := ParseHexColor(colorHex)
	if err != nil {
		col = color.Black
	}
	d := &font.Drawer{
		Dst:  dst,
		Src:  image.NewUniform(col),
		Face: face,
		Dot:  fixed.P(1, metrics.Ascent.Ceil()),
	}
	d.DrawString(text)
	return dst, nil
}

// PlaceWatermark 把水印图层叠加到底图（对齐 PhotoHandleService::placeWatermark + image->place）。
// ratio：水印占底图短边百分比；position：9 宫格；opacity：0-100。
func PlaceWatermark(base, mark image.Image, ratio int, tiled bool, position string, offsetX, offsetY, opacity int) image.Image {
	b := base.Bounds()
	bw, bh := b.Dx(), b.Dy()

	// 缩放水印到 ratio%
	if ratio > 0 && ratio < 100 {
		mw := bw * ratio / 100
		mh := bh * ratio / 100
		mb := mark.Bounds()
		scale := 1.0
		if mw > 0 && mb.Dx() > mw {
			scale = float64(mw) / float64(mb.Dx())
		}
		if mh > 0 && float64(mb.Dy())*scale > float64(mh) {
			scale = float64(mh) / float64(mb.Dy())
		}
		if scale < 1 {
			mark = Resize(mark, int(float64(mb.Dx())*scale), int(float64(mb.Dy())*scale))
		}
	}
	mark = applyOpacity(mark, opacity)
	mb := mark.Bounds()
	mw, mh := mb.Dx(), mb.Dy()

	dst := image.NewRGBA(b)
	draw.Draw(dst, b, base, b.Min, draw.Src)

	place := func(x, y int) {
		r := image.Rect(x, y, x+mw, y+mh).Add(b.Min)
		draw.Draw(dst, r.Intersect(b), mark, mb.Min, draw.Over)
	}

	if tiled {
		for x := 0; x < bw; x += mw + offsetX {
			for y := 0; y < bh; y += mh + offsetY {
				place(x, y)
			}
		}
		return dst
	}

	switch position {
	case "center":
		place((bw-mw)/2, (bh-mh)/2)
	case "top":
		place((bw-mw)/2, offsetY)
	case "left":
		place(offsetX, (bh-mh)/2)
	case "right":
		place(bw-mw-offsetX, (bh-mh)/2)
	case "bottom":
		place((bw-mw)/2, bh-mh-offsetY)
	case "top-left", "":
		place(offsetX, offsetY)
	case "top-right":
		place(bw-mw-offsetX, offsetY)
	case "bottom-left":
		place(offsetX, bh-mh-offsetY)
	case "bottom-right":
		place(bw-mw-offsetX, bh-mh-offsetY)
	default:
		place(offsetX, offsetY)
	}
	return dst
}

// applyOpacity 将图片整体透明度调整为 opacity%（0-100）。
func applyOpacity(img image.Image, opacity int) image.Image {
	if opacity >= 100 || opacity < 0 {
		if opacity >= 100 {
			return img
		}
	}
	b := img.Bounds()
	dst := image.NewNRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	factor := 1.0
	if opacity < 100 {
		factor = float64(opacity) / 100
	}
	for y := 0; y < b.Dy(); y++ {
		for x := 0; x < b.Dx(); x++ {
			r, g, bb, a := img.At(b.Min.X+x, b.Min.Y+y).RGBA()
			na := uint8(float64(a>>8) * factor)
			dst.SetNRGBA(x, y, color.NRGBA{R: uint8(r >> 8), G: uint8(g >> 8), B: uint8(bb >> 8), A: na})
		}
	}
	return dst
}
