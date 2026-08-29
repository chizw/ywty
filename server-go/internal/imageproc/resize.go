package imageproc

import (
	"image"
	"image/color"

	"golang.org/x/image/draw"
)

// Resize 高质量等比缩放（CatmullRom，接近 Imagick 输出质量）。
func Resize(img image.Image, w, h int) image.Image {
	if w <= 0 || h <= 0 {
		return img
	}
	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	draw.CatmullRom.Scale(dst, dst.Bounds(), img, img.Bounds(), draw.Over, nil)
	return dst
}

// Rotate 旋转 90/180/270 度（EXIF Orientation 用）。
func Rotate(img image.Image, angle int) image.Image {
	switch ((angle % 360) + 360) % 360 {
	case 90:
		b := img.Bounds()
		dst := image.NewRGBA(image.Rect(0, 0, b.Dy(), b.Dx()))
		for y := 0; y < b.Dy(); y++ {
			for x := 0; x < b.Dx(); x++ {
				dst.Set(b.Dy()-1-y, x, img.At(x, y))
			}
		}
		return dst
	case 180:
		b := img.Bounds()
		dst := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
		for y := 0; y < b.Dy(); y++ {
			for x := 0; x < b.Dx(); x++ {
				dst.Set(b.Dx()-1-x, b.Dy()-1-y, img.At(x, y))
			}
		}
		return dst
	case 270:
		b := img.Bounds()
		dst := image.NewRGBA(image.Rect(0, 0, b.Dy(), b.Dx()))
		for y := 0; y < b.Dy(); y++ {
			for x := 0; x < b.Dx(); x++ {
				dst.Set(y, b.Dx()-1-x, img.At(x, y))
			}
		}
		return dst
	}
	return img
}

// FlipH 水平翻转（EXIF orientation 2/4/5/6/7/8 需要时组合）。
func FlipH(img image.Image) image.Image {
	b := img.Bounds()
	dst := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	for y := 0; y < b.Dy(); y++ {
		for x := 0; x < b.Dx(); x++ {
			dst.Set(b.Dx()-1-x, y, img.At(x, y))
		}
	}
	return dst
}

// AutoOrient 按 EXIF orientation 校正方向。
// orientation 值来自 ExifOrientation。
func AutoOrient(img image.Image, orientation int) image.Image {
	switch orientation {
	case 2:
		return FlipH(img)
	case 3:
		return Rotate(img, 180)
	case 4:
		return Rotate(FlipH(img), 180)
	case 5:
		return Rotate(FlipH(img), 90)
	case 6:
		return Rotate(img, 90)
	case 7:
		return Rotate(FlipH(img), 270)
	case 8:
		return Rotate(img, 270)
	}
	return img
}

// hasAlphaColor 辅助：判断颜色是否非不透明。
func hasAlphaColor(c color.Color) bool {
	_, _, _, a := c.RGBA()
	return a < 0xffff
}

// CropTo 裁剪到指定矩形（越界部分丢弃）。
func CropTo(img image.Image, r image.Rectangle) image.Image {
	dst := image.NewRGBA(image.Rect(0, 0, r.Dx(), r.Dy()))
	draw.Draw(dst, dst.Bounds(), img, r.Min, draw.Src)
	return dst
}
