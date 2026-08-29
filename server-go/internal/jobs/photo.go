package jobs

import (
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/chizw/ywty/server-go/internal/config"
	"github.com/chizw/ywty/server-go/internal/imageproc"
	"github.com/chizw/ywty/server-go/internal/model"
	"github.com/chizw/ywty/server-go/internal/photostore"
	"github.com/chizw/ywty/server-go/internal/storage"
	"gorm.io/gorm"
)

func jsonUnmarshal(data []byte, v any) error { return json.Unmarshal(data, v) }

// RunHandlePhoto 图片处理（对齐 HandlePhotoJob → PhotoHandleService::format）。
// optionsJSON 为 handle 驱动的 options JSON（operations/output）。
func RunHandlePhoto(gdb *gorm.DB, cfg *config.Config, photoID int64, optionsJSON string) error {
	photo, fs, _, err := loadPhotoFS(gdb, cfg, photoID)
	if err != nil {
		return err
	}
	raw, err := fs.Read(photo.Pathname)
	if err != nil {
		return err
	}
	img, err := imageproc.Decode(raw, photo.Extension)
	if err != nil {
		return err
	}

	var opts struct {
		Operations map[string][]struct {
			Type string         `json:"type"`
			Data map[string]any `json:"data"`
		} `json:"operations"`
		Output struct {
			Format  string `json:"format"`
			Quality int    `json:"quality"`
		} `json:"output"`
	}
	if err := jsonUnmarshal([]byte(optionsJSON), &opts); err != nil {
		return fmt.Errorf("处理驱动配置解析失败: %w", err)
	}

	for _, ops := range opts.Operations {
		for _, op := range ops {
			switch op.Type {
			case "scaleDown", "scale", "fit":
				img = imageproc.Resize(img, intAny(op.Data["width"]), intAny(op.Data["height"]))
			case "crop":
				img = crop(img, intAny(op.Data["x"]), intAny(op.Data["y"]),
					intAny(op.Data["width"]), intAny(op.Data["height"]))
			case "rotate":
				if angle, ok := op.Data["angle"].(float64); ok {
					img = imageproc.Rotate(img, int(angle))
				}
			case "blur", "brightness", "contrast", "greyscale", "gamma":
				// 滤镜类操作 M5 完善，暂跳过
			case "text":
				if mark, err := textWatermarkFromOptions(op.Data); err == nil {
					img = placeWatermarkFromOptions(img, mark, op.Data)
				}
			case "watermark":
				if mark, err := imageWatermarkFromOptions(op.Data); err == nil {
					img = placeWatermarkFromOptions(img, mark, op.Data)
				}
			}
		}
	}

	outExt := opts.Output.Format
	if outExt == "" || outExt == "auto" {
		outExt = photo.Extension
	}
	data, finalExt, err := imageproc.Encode(img, outExt, opts.Output.Quality)
	if err != nil {
		return err
	}
	newPathname := photo.Pathname
	if finalExt != photo.Extension {
		if i := strings.LastIndex(photo.Pathname, photo.Extension); i >= 0 {
			newPathname = photo.Pathname[:i] + finalExt + photo.Pathname[i+len(photo.Extension):]
		}
	}
	if newPathname != photo.Pathname {
		_ = fs.Delete(photo.Pathname)
	}
	if err := fs.Write(newPathname, data); err != nil {
		return err
	}
	w, h := imageproc.Size(data, finalExt)
	md5Hex, sha1Hex := photostore.HashBytes(data)
	return gdb.Model(&model.Photo{}).Where("id = ?", photo.ID).Updates(map[string]any{
		"pathname":   newPathname,
		"extension":  finalExt,
		"mimetype":   imageproc.MimeTypeByExt(finalExt),
		"md5":        md5Hex,
		"sha1":       sha1Hex,
		"size":       float64(len(data)) / 1024,
		"width":      w,
		"height":     h,
		"updated_at": time.Now().UTC(),
	}).Error
}

// generateThumbnail 缩略图（对齐 GeneratePhotoThumbnailJob）。
func generateThumbnail(gdb *gorm.DB, cfg *config.Config, data []byte) error {
	var in struct {
		PhotoID int64 `json:"photo_id"`
	}
	if err := jsonUnmarshal(data, &in); err != nil {
		return err
	}
	photo, fs, rec, err := loadPhotoFS(gdb, cfg, in.PhotoID)
	if err != nil {
		return err
	}
	raw, err := fs.Read(photo.Pathname)
	if err != nil {
		return err
	}
	out, _, err := imageproc.Thumbnail(raw, photo.Extension,
		rec.FSOptions().ThumbnailMaxSize, rec.FSOptions().ThumbnailQuality)
	if err != nil {
		return err
	}
	return writePublicFile("thumbnails/"+photo.Pathname, out)
}

// autoDeletePhoto 到期图片删除（对齐 AutoDeletePhotoJob）。
func autoDeletePhoto(gdb *gorm.DB, cfg *config.Config, data []byte) error {
	var in struct {
		PhotoID int64 `json:"photo_id"`
	}
	if err := jsonUnmarshal(data, &in); err != nil {
		return err
	}
	var photo model.Photo
	if err := gdb.First(&photo, in.PhotoID).Error; err != nil {
		return nil // 已删除
	}
	if photo.ExpiredAt == nil || photo.ExpiredAt.After(time.Now()) {
		return nil // 未到期或已取消过期
	}
	return photostore.DeletePhoto(gdb, cfg, &photo)
}

// ---------- 水印 ----------

// textWatermarkFromOptions 文本水印（对齐 PhotoHandleService text 分支的常用参数）。
func textWatermarkFromOptions(data map[string]any) (image.Image, error) {
	text, _ := data["text"].(string)
	font, _ := data["font"].(map[string]any)
	fontPath := ""
	fontSize := 24.0
	colorHex := "#000000"
	if font != nil {
		if fn, ok := font["filename"].(map[string]any); ok {
			fontPath, _ = fn["value"].(string)
		}
		if fs, ok := font["size"].(map[string]any); ok {
			if v, ok := fs["value"].(float64); ok {
				fontSize = v
			}
		} else if fs, ok := font["size"].(float64); ok {
			fontSize = fs
		}
		if fc, ok := font["color"].(map[string]any); ok {
			if v, ok := fc["value"].(string); ok {
				colorHex = v
			}
		} else if fc, ok := font["color"].(string); ok {
			colorHex = fc
		}
	}
	if fontPath == "" {
		return nil, errors.New("文本水印缺少字体文件")
	}
	return imageproc.RenderText(text, filepath.Join("storage/app", filepath.FromSlash(fontPath)), fontSize, colorHex)
}

// imageWatermarkFromOptions 图片水印（element 为 local 磁盘相对路径）。
func imageWatermarkFromOptions(data map[string]any) (image.Image, error) {
	element, _ := data["element"].(string)
	if element == "" {
		return nil, errors.New("图片水印缺少素材")
	}
	raw, err := os.ReadFile(filepath.Join("storage/app", filepath.FromSlash(element)))
	if err != nil {
		return nil, err
	}
	ext := ""
	if i := strings.LastIndex(element, "."); i >= 0 {
		ext = element[i+1:]
	}
	return imageproc.Decode(raw, ext)
}

// placeWatermarkFromOptions 叠加水印（ratio/position/offset/tiled/opacity）。
func placeWatermarkFromOptions(base, mark image.Image, data map[string]any) image.Image {
	ratio := intAny(data["ratio"])
	if ratio == 0 {
		ratio = 10
	}
	tiled, _ := data["is_tiled"].(bool)
	position, _ := data["position"].(string)
	offsetX := intAny(data["offset_x"])
	offsetY := intAny(data["offset_y"])
	opacity := 100
	if v, ok := data["opacity"].(float64); ok {
		opacity = int(v)
	}
	return imageproc.PlaceWatermark(base, mark, ratio, tiled, position, offsetX, offsetY, opacity)
}

// ---------- 内部工具 ----------

func loadPhotoFS(gdb *gorm.DB, cfg *config.Config, photoID int64) (*model.Photo, storage.Filesystem, *photostore.StorageRecord, error) {
	var photo model.Photo
	if err := gdb.First(&photo, photoID).Error; err != nil {
		return nil, nil, nil, err
	}
	rec, err := photostore.LoadStorageByID(gdb, deref(photo.StorageID))
	if err != nil || rec == nil {
		return nil, nil, nil, errors.New("图片的储存策略不存在")
	}
	fs, err := rec.Filesystem(cfg)
	if err != nil {
		return nil, nil, nil, err
	}
	return &photo, fs, rec, nil
}

func deref(p *int64) int64 {
	if p == nil {
		return 0
	}
	return *p
}

func intAny(v any) int {
	switch x := v.(type) {
	case float64:
		return int(x)
	case int:
		return x
	case string:
		var n int
		_, _ = fmt.Sscanf(x, "%d", &n)
		return n
	}
	return 0
}

func crop(img image.Image, x, y, w, h int) image.Image {
	b := img.Bounds()
	if w <= 0 {
		w = b.Dx()
	}
	if h <= 0 {
		h = b.Dy()
	}
	r := image.Rect(0, 0, w, h).Add(image.Pt(b.Min.X+x, b.Min.Y+y))
	return imageproc.CropTo(img, r)
}

// writePublicFile 写入 public 磁盘（storage/app/public，对齐 Storage::disk('public')）。
func writePublicFile(rel string, data []byte) error {
	full := filepath.Join("storage/app/public", filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		return err
	}
	return os.WriteFile(full, data, 0o644)
}
