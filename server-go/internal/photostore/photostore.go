// Package photostore 上传管线核心（命名、去重、相册与标签同步）。
package photostore

import (
	"crypto/md5"
	"crypto/rand"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/chizw/ywty/server-go/internal/config"
	"github.com/chizw/ywty/server-go/internal/db/types"
	"github.com/chizw/ywty/server-go/internal/model"
	"github.com/chizw/ywty/server-go/internal/storage"
	"gorm.io/gorm"
)

// StorageRecord storages 行（带 options）。
type StorageRecord struct {
	ID       int64
	Name     string
	Prefix   string
	Provider string
	Options  *string
}

// LoadGroupStorage 按 id 取组内绑定的储存。
func LoadGroupStorage(gdb *gorm.DB, groupID, storageID int64) (*StorageRecord, error) {
	var row struct {
		ID       int64
		Name     string
		Prefix   string
		Provider string
		Options  *string
	}
	err := gdb.Raw(
		"SELECT s.`id`, s.`name`, s.`prefix`, s.`provider`, s.`options` FROM `storages` s "+
			"INNER JOIN `group_storage` gs ON gs.storage_id = s.id AND gs.group_id = ? "+
			"WHERE s.`id` = ? AND s.`deleted_at` IS NULL LIMIT 1", groupID, storageID,
	).Scan(&row).Error
	if err != nil {
		return nil, err
	}
	if row.ID == 0 {
		return nil, nil
	}
	return &StorageRecord{ID: row.ID, Name: row.Name, Prefix: row.Prefix, Provider: row.Provider, Options: row.Options}, nil
}

// DefaultGroupStorage 组内默认（排序最前）的储存。
func DefaultGroupStorage(gdb *gorm.DB, groupID int64) (*StorageRecord, error) {
	var row struct {
		ID       int64
		Name     string
		Prefix   string
		Provider string
		Options  *string
	}
	err := gdb.Raw(
		"SELECT s.`id`, s.`name`, s.`prefix`, s.`provider`, s.`options` FROM `storages` s "+
			"INNER JOIN `group_storage` gs ON gs.storage_id = s.id "+
			"WHERE gs.group_id = ? AND s.`deleted_at` IS NULL ORDER BY gs.`sort` ASC, s.`id` ASC LIMIT 1", groupID,
	).Scan(&row).Error
	if err != nil {
		return nil, err
	}
	if row.ID == 0 {
		return nil, nil
	}
	return &StorageRecord{ID: row.ID, Name: row.Name, Prefix: row.Prefix, Provider: row.Provider, Options: row.Options}, nil
}

// LoadStorageByID 按 id 取储存（不受组约束）。
func LoadStorageByID(gdb *gorm.DB, id int64) (*StorageRecord, error) {
	var row StorageRecord
	err := gdb.Raw("SELECT `id`, `name`, `prefix`, `provider`, `options` FROM `storages` WHERE `id` = ? AND `deleted_at` IS NULL LIMIT 1", id).
		Scan(&row).Error
	if err != nil {
		return nil, err
	}
	if row.ID == 0 {
		return nil, nil
	}
	return &row, nil
}

// FSOptions 解析 options。
func (s *StorageRecord) FSOptions() storage.Options {
	if s.Options == nil {
		return storage.Options{}
	}
	return storage.ParseOptions(*s.Options)
}

// Filesystem 构建适配器。
func (s *StorageRecord) Filesystem(cfg *config.Config) (storage.Filesystem, error) {
	return storage.FilesystemFor(&storage.Storage{
		ID: s.ID, Name: s.Name, Prefix: s.Prefix, Provider: s.Provider, Options: s.FSOptions(),
	}, cfg)
}

// Runtime 转运行时 Storage。
func (s *StorageRecord) Runtime(cfg *config.Config) *storage.Storage {
	return &storage.Storage{ID: s.ID, Name: s.Name, Prefix: s.Prefix, Provider: s.Provider, Options: s.FSOptions()}
}

// namingRe 清理命名规则首尾斜杠。
var namingTrim = regexp.MustCompile(`^/+|/+$`)

// Pathname 按 naming_rule 生成路径（支持 {Ymd}/{md5} 等占位符）。
func Pathname(rule, filename, ext, md5Hex, sha1Hex string, uid int64, now time.Time) string {
	rule = namingTrim.ReplaceAllString(strings.TrimSpace(rule), "")
	uniqid := fmt.Sprintf("%x", now.UnixNano())
	uuidStr := uuidLike()
	replacer := strings.NewReplacer(
		"{Y}", strconv.Itoa(now.Year()),
		"{y}", strconv.Itoa(now.Year())[2:],
		"{m}", fmt.Sprintf("%02d", int(now.Month())),
		"{d}", fmt.Sprintf("%02d", now.Day()),
		"{Ymd}", now.Format("20060102"),
		"{filename}", filename,
		"{ext}", ext,
		"{time}", strconv.FormatInt(now.Unix(), 10),
		"{uniqid}", uniqid,
		"{md5}", md5Hex,
		"{sha1}", sha1Hex,
		"{uuid}", uuidStr,
		"{uid}", strconv.FormatInt(uid, 10),
	)
	return replacer.Replace(rule)
}

// uuidLike 生成 UUID v4 形态字符串。
func uuidLike() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

// StoreInput 入库参数。
type StoreInput struct {
	UserID    *int64
	GroupID   int64
	StorageID int64
	Filename  string // 原始文件名
	Name      string // 去扩展名的别名
	Pathname  string
	Mimetype  string
	Extension string
	MD5       string
	SHA1      string
	SizeKB    float64
	Width     int
	Height    int
	IsPublic  bool
	IP        string
	ExpiredAt *time.Time
	AlbumID   *int64
	Tags      []string
}

// UploadError 业务错误。
type UploadError struct{ Message string }

func (e *UploadError) Error() string { return e.Message }

// Store 入库：按内容指纹去重 + 相册/标签同步。
func Store(gdb *gorm.DB, in StoreInput) (*model.Photo, error) {
	var (
		photo model.Photo
		uid   any
	)
	if in.UserID != nil {
		uid = *in.UserID
	}
	// firstOrCreate 语义
	find := func() (*model.Photo, error) {
		var p model.Photo
		q := "SELECT * FROM `photos` WHERE `storage_id` = ? AND `md5` = ? AND `sha1` = ? AND `pathname` = ? AND `deleted_at` IS NULL"
		args := []any{in.StorageID, in.MD5, in.SHA1, in.Pathname}
		if in.UserID != nil {
			q += " AND `user_id` = ?"
			args = append(args, *in.UserID)
		} else {
			q += " AND `user_id` IS NULL"
		}
		if err := gdb.Raw(q, args...).Scan(&p).Error; err != nil {
			return nil, err
		}
		if p.ID == 0 {
			return nil, nil
		}
		return &p, nil
	}
	existing, err := find()
	if err != nil {
		return nil, err
	}
	if existing != nil {
		photo = *existing
	} else {
		now := time.Now().UTC()
		var ip *string
		if in.IP != "" {
			ip = &in.IP
		}
		photo = model.Photo{
			UserID:    in.UserID,
			GroupID:   &in.GroupID,
			StorageID: &in.StorageID,
			Name:      in.Name,
			Filename:  in.Filename,
			Pathname:  in.Pathname,
			Mimetype:  in.Mimetype,
			Extension: in.Extension,
			MD5:       in.MD5,
			SHA1:      in.SHA1,
			Exif:      types.JSON("{}"),
			Size:      in.SizeKB,
			Width:     int64(in.Width),
			Height:    int64(in.Height),
			IsPublic:  in.IsPublic,
			Status:    "normal",
			IPAddress: ip,
			ExpiredAt: in.ExpiredAt,
			CreatedAt: &now,
			UpdatedAt: &now,
		}
		_ = uid
		if err := gdb.Create(&photo).Error; err != nil {
			return nil, err
		}
	}

	// 相册
	if in.AlbumID != nil && *in.AlbumID > 0 {
		insertIgnore(gdb, "INSERT INTO `album_photo` (`album_id`, `photo_id`, `sort`) VALUES (?, ?, 0)", *in.AlbumID, photo.ID)
	}

	// 标签
	for _, name := range in.Tags {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		var tagID int64
		gdb.Raw("SELECT `id` FROM `tags` WHERE `name` = ? LIMIT 1", name).Scan(&tagID)
		if tagID == 0 {
			res := gdb.Exec("INSERT INTO `tags` (`name`, `created_at`, `updated_at`) VALUES (?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)", name)
			if res.Error != nil {
				continue
			}
			gdb.Raw("SELECT `id` FROM `tags` WHERE `name` = ? LIMIT 1", name).Scan(&tagID)
		}
		var uidVal any
		if in.UserID != nil {
			uidVal = *in.UserID
		}
		insertIgnore(gdb, "INSERT INTO `taggables` (`tag_id`, `user_id`, `taggable_type`, `taggable_id`) VALUES (?, ?, 'App\\Models\\Photo', ?)", tagID, uidVal, photo.ID)
	}
	return &photo, nil
}

// SumUserUsedKB 用户已用容量（KB）。
func SumUserUsedKB(gdb *gorm.DB, userID int64) float64 {
	var used float64
	gdb.Raw("SELECT COALESCE(sum(`size`), 0) FROM `photos` WHERE `user_id` = ? AND `deleted_at` IS NULL", userID).Scan(&used)
	return used
}

// UserTotalCapacityKB 用户总容量（KB）。
func UserTotalCapacityKB(gdb *gorm.DB, userID int64) float64 {
	var total float64
	gdb.Raw(
		"SELECT COALESCE(sum(`capacity`), 0) FROM `user_capacities` WHERE `user_id` = ? AND `deleted_at` IS NULL "+
			"AND (`expired_at` > CURRENT_TIMESTAMP OR `expired_at` IS NULL)", userID,
	).Scan(&total)
	return total
}

// PublicURL 图片公开 URL。
func PublicURL(gdb *gorm.DB, cfg *config.Config, photo *model.Photo) string {
	s, err := LoadStorageByID(gdb, deref(photo.StorageID))
	if err != nil || s == nil {
		return strings.TrimRight(cfg.AppURL, "/") + "/" + strings.Trim(photo.Pathname, "/")
	}
	return storage.PublicURL(s.Runtime(cfg), cfg, photo.Pathname)
}

// ThumbnailURL 缩略图 URL（public 磁盘 thumbnails/{pathname}，对齐 thumbnail_url accessor）。
func ThumbnailURL(gdb *gorm.DB, cfg *config.Config, photo *model.Photo) string {
	tp := "thumbnails/" + photo.Pathname
	// public 磁盘根 = storage/app/public
	if _, err := diskStat("storage/app/public", tp); err == nil {
		return strings.TrimRight(cfg.AppURL, "/") + "/storage/" + tp
	}
	return PublicURL(gdb, cfg, photo)
}

func diskStat(root, path string) (os.FileInfo, error) {
	return os.Stat(filepath.Join(root, filepath.FromSlash(path)))
}

// ResourceUrls 复制链接集合。
func ResourceUrls(gdb *gorm.DB, cfg *config.Config, photo *model.Photo) map[string]any {
	u := PublicURL(gdb, cfg, photo)
	tu := ThumbnailURL(gdb, cfg, photo)
	return map[string]any{
		"url":                u,
		"html":               fmt.Sprintf(`<img src="%s" alt="%s" title="%s" />`, u, photo.Name, photo.Name),
		"bbcode":             "[img]" + u + "[/img]",
		"markdown":           fmt.Sprintf("![%s](%s)", photo.Name, u),
		"markdown_with_link": fmt.Sprintf("[![%s](%s)](%s)", photo.Name, u, u),
		"thumbnail_url":      tu,
	}
}

// DeletePhoto 删除图片（软删 + 物理文件与缩略图，对齐 deleted 事件）。
func DeletePhoto(gdb *gorm.DB, cfg *config.Config, photo *model.Photo) error {
	if s, err := LoadStorageByID(gdb, deref(photo.StorageID)); err == nil && s != nil {
		if fs, ferr := s.Filesystem(cfg); ferr == nil {
			_ = fs.Delete(photo.Pathname)
		}
	}
	_ = gdb.Exec("DELETE FROM `photos` WHERE `id` = ?", photo.ID).Error
	_ = gdb.Exec("DELETE FROM `album_photo` WHERE `photo_id` = ?", photo.ID).Error
	_ = gdb.Exec("DELETE FROM `taggables` WHERE `taggable_type` = 'App\\Models\\Photo' AND `taggable_id` = ?", photo.ID).Error
	// 缩略图（public 磁盘）
	_ = gdb.Exec("DELETE FROM `photos` WHERE `id` = ?", photo.ID).Error
	tp := filepath.Join("storage/app/public", filepath.FromSlash("thumbnails/"+photo.Pathname))
	_ = os.Remove(tp)
	return nil
}

// HashBytes 计算文件 md5/sha1。
func HashBytes(data []byte) (string, string) {
	m := md5.Sum(data)
	s := sha1.Sum(data)
	return hex.EncodeToString(m[:]), hex.EncodeToString(s[:])
}

func deref(p *int64) int64 {
	if p == nil {
		return 0
	}
	return *p
}

var ErrNotFound = errors.New("photostore: 未找到")

// jsonUnmarshal 引用包装。
func jsonUnmarshal(s string, v any) error { return json.Unmarshal([]byte(s), v) }

// InsertIgnore 幂等插入（方言分支）。
func InsertIgnore(gdb *gorm.DB, q string, args ...any) { insertIgnore(gdb, q, args...) }

// insertIgnore 幂等插入（方言分支）。
func insertIgnore(gdb *gorm.DB, q string, args ...any) {
	if gdb.Dialector.Name() == "mysql" {
		gdb.Exec(strings.Replace(q, "INSERT INTO", "INSERT IGNORE INTO", 1), args...)
		return
	}
	gdb.Exec(strings.Replace(q, "INSERT INTO", "INSERT OR IGNORE INTO", 1), args...)
}
