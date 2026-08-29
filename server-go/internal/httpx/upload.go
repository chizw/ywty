package httpx

import (
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/chizw/ywty/server-go/internal/authx"
	"github.com/chizw/ywty/server-go/internal/imageproc"
	"github.com/chizw/ywty/server-go/internal/jobs"
	"github.com/chizw/ywty/server-go/internal/model"
	"github.com/chizw/ywty/server-go/internal/photostore"
	"github.com/chizw/ywty/server-go/internal/setting"
	"github.com/chizw/ywty/server-go/internal/support/r"
	"github.com/chizw/ywty/server-go/internal/validate"
)

// masterImageTypes 全部可处理的图片扩展名（原版格式清单）。
var masterImageTypes = []string{
	"jpg", "jpeg", "webp", "avif", "bmp", "gif", "png", "tif", "tiff",
	"jp2", "j2k", "jp2k", "jpf", "jpm", "jpg2", "j2c", "jpc", "jpx", "heic", "heif",
}

// uploadVerify 对齐 UploadVerify 中间件。
func (d *deps) uploadVerify(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		ctx := authx.From(req)
		emailVerify, _ := setting.Bool(d.gdb, setting.GroupApp, "user_email_verify")
		phoneVerify, _ := setting.Bool(d.gdb, setting.GroupApp, "user_phone_verify")
		if ctx.User != nil {
			if emailVerify && ctx.User.EmailVerifiedAt == nil {
				r.Error(w, "请先验证邮箱")
				return
			}
			if phoneVerify && ctx.User.PhoneVerifiedAt == nil {
				r.Error(w, "请先绑定手机号")
				return
			}
		} else {
			guestUpload, _ := setting.Bool(d.gdb, setting.GroupApp, "guest_upload")
			if !guestUpload {
				r.Error(w, "系统暂不支持游客上传")
				return
			}
		}
		next.ServeHTTP(w, req)
	})
}

// uploadFrequencyLimit 对齐 UploadFrequencyLimit 中间件（组级频率限制，0 不限制）。
func (d *deps) uploadFrequencyLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		ctx := authx.From(req)
		if ctx.Group == nil {
			r.Error(w, "系统未初始化角色组")
			return
		}
		var opts struct {
			LimitPerMinute int `json:"limit_per_minute"`
			LimitPerHour   int `json:"limit_per_hour"`
			LimitPerDay    int `json:"limit_per_day"`
			LimitPerWeek   int `json:"limit_per_week"`
			LimitPerMonth  int `json:"limit_per_month"`
		}
		if ctx.Group.Options != nil {
			_ = jsonUnmarshalInto(string(ctx.Group.Options), &opts)
		}
		scopes := []struct {
			opt   int
			since string
			name  string
		}{
			{opts.LimitPerMinute, "-1 minute", "分钟"},
			{opts.LimitPerHour, "-1 hour", "小时"},
			{opts.LimitPerDay, "-1 day", "天"},
			{opts.LimitPerWeek, "-7 day", "周"},
			{opts.LimitPerMonth, "-1 month", "个月"},
		}
		sinceExpr := map[string]string{
			"-1 minute": "datetime('now', '-1 minute')",
			"-1 hour":   "datetime('now', '-1 hour')",
			"-1 day":    "datetime('now', '-1 day')",
			"-7 day":    "datetime('now', '-7 day')",
			"-1 month":  "datetime('now', '-1 month')",
		}
		for _, sc := range scopes {
			if sc.opt == 0 {
				continue
			}
			where, arg := d.freqScope(req, sc.since, sinceExpr)
			var count int64
			d.gdb.Raw("SELECT count(*) FROM `photos` WHERE `deleted_at` IS NULL AND "+where, arg).Scan(&count)
			if count >= int64(sc.opt) {
				r.ErrorWithCode(w, http.StatusTooManyRequests,
					"一"+sc.name+"内你只能上传 "+itoa(sc.opt)+" 张图片")
				return
			}
		}
		next.ServeHTTP(w, req)
	})
}

func (d *deps) freqScope(req *http.Request, since string, sinceExpr map[string]string) (string, any) {
	ctx := authx.From(req)
	if ctx.User != nil {
		if d.gdb.Dialector.Name() == "mysql" {
			return "`user_id` = ? AND `created_at` >= DATE_SUB(NOW(), INTERVAL ? SECOND)", userSinceSeconds(since)
		}
		return "`user_id` = ? AND `created_at` >= " + sinceExpr[since], ctx.User.ID
	}
	ip := authx.ClientIP(req)
	if d.gdb.Dialector.Name() == "mysql" {
		return "`ip_address` = ? AND `created_at` >= DATE_SUB(NOW(), INTERVAL ? SECOND)", userSinceSeconds(since)
	}
	return "`ip_address` = ? AND `created_at` >= " + sinceExpr[since], ip
}

func userSinceSeconds(since string) int {
	switch since {
	case "-1 minute":
		return 60
	case "-1 hour":
		return 3600
	case "-1 day":
		return 86400
	case "-7 day":
		return 7 * 86400
	case "-1 month":
		return 30 * 86400
	}
	return 60
}

// handleUpload POST /api/v2/upload（HomeController::upload）。
func (d *deps) handleUpload(w http.ResponseWriter, req *http.Request) {
	ctx := authx.From(req)
	if ctx.Group == nil {
		r.Error(w, "系统未初始化角色组")
		return
	}

	if err := req.ParseMultipartForm(32 << 20); err != nil {
		r.Error(w, "文件资源无效")
		return
	}
	file, header, err := req.FormFile("file")
	if err != nil {
		r.ErrorData(w, http.StatusUnprocessableEntity, "The given data was invalid.",
			map[string]any{"errors": map[string][]string{"file": {"文件 不能为空。"}}})
		return
	}
	defer func() { _ = file.Close() }()

	// 组配置
	var groupOpts struct {
		AllowFileTypes    []string `json:"allow_file_types"`
		MaxUploadSize     int      `json:"max_upload_size"`
		FileExpireSeconds int      `json:"file_expire_seconds"`
	}
	if ctx.Group.Options != nil {
		_ = jsonUnmarshalInto(string(ctx.Group.Options), &groupOpts)
	}
	if groupOpts.MaxUploadSize == 0 {
		groupOpts.MaxUploadSize = 5120
	}

	v := validate.New()
	// 文件类型/大小
	ext := ""
	if dot := strings.LastIndex(header.Filename, "."); dot >= 0 {
		ext = imageproc.NormalizeExt(header.Filename[dot+1:])
	}
	allowed := intersectTypes(masterImageTypes, groupOpts.AllowFileTypes)
	if !allowed[ext] {
		v.Add("file", "文件", "不允许的文件类型。")
	}
	data, readErr := io.ReadAll(file)
	if readErr != nil {
		r.Error(w, "文件读取失败")
		return
	}
	sizeKB := float64(len(data)) / 1024
	if sizeKB > float64(groupOpts.MaxUploadSize) {
		v.Add("file", "文件", "不能超过 "+itoa(groupOpts.MaxUploadSize)+"KB。")
	}
	// 容量校验
	if ctx.User != nil {
		total := photostore.UserTotalCapacityKB(d.gdb, ctx.User.ID)
		used := photostore.SumUserUsedKB(d.gdb, ctx.User.ID)
		if used+sizeKB > total && total > 0 {
			v.Add("file", "文件", "储存空间不足")
		}
	}
	// storage_id
	storageID := pathIntForm(req.FormValue("storage_id"))
	if storageID == 0 {
		v.Add("storage_id", "储存", "不能为空。")
	} else {
		s, _ := photostore.LoadGroupStorage(d.gdb, ctx.Group.ID, storageID)
		if s == nil {
			v.Add("storage_id", "储存", "不存在的储存驱动")
		}
	}
	albumID := pathIntForm(req.FormValue("album_id"))
	if albumID > 0 && ctx.User != nil {
		var n int64
		d.gdb.Raw("SELECT count(*) FROM `albums` WHERE `id` = ? AND `user_id` = ? AND `deleted_at` IS NULL", albumID, ctx.User.ID).Scan(&n)
		if n == 0 {
			v.Add("album_id", "相册", "相册不存在")
		}
	}
	isPublic := req.FormValue("is_public") == "1" || req.FormValue("is_public") == "true"
	tags := splitTags(req.FormValue("tags"))
	var expiredAt *time.Time
	if ea := req.FormValue("expired_at"); ea != "" {
		for _, layout := range []string{time.RFC3339, "2006-01-02 15:04:05", "2006-01-02"} {
			if t, err := time.Parse(layout, ea); err == nil && t.After(time.Now()) {
				expiredAt = &t
				break
			}
		}
	}
	if v.Fail() {
		v.Respond(w)
		return
	}

	rec, _ := photostore.LoadGroupStorage(d.gdb, ctx.Group.ID, storageID)
	fs, err := rec.Filesystem(d.cfg)
	if err != nil {
		r.Error(w, err.Error())
		return
	}

	// 命名 + 写入
	md5Hex, sha1Hex := photostore.HashBytes(data)
	var uid int64
	if ctx.User != nil {
		uid = ctx.User.ID
	}
	name := header.Filename
	if dot := strings.LastIndex(name, "."); dot >= 0 {
		name = name[:dot]
	}
	pathname := photostore.Pathname(rec.FSOptions().NamingRule, name, ext, md5Hex, sha1Hex, uid, time.Now()) + "." + ext
	if err := fs.AppendOrCreate(pathname, data); err != nil {
		r.Error(w, "文件写入失败: "+err.Error())
		return
	}

	width, height := imageproc.Size(data, ext)
	if expiredAt != nil && groupOpts.FileExpireSeconds > 0 {
		t := time.Now().Add(time.Duration(groupOpts.FileExpireSeconds) * time.Second)
		expiredAt = &t
	}

	var userID *int64
	if ctx.User != nil {
		userID = &ctx.User.ID
	} else {
		isPublic = false
	}
	var albumRef *int64
	if ctx.User != nil && albumID > 0 {
		albumRef = &albumID
	}
	ip := authx.ClientIP(req)
	photo, err := photostore.Store(d.gdb, photostore.StoreInput{
		UserID: userID, GroupID: ctx.Group.ID, StorageID: rec.ID,
		Filename: header.Filename, Name: name, Pathname: pathname,
		Mimetype: imageproc.MimeTypeByExt(ext), Extension: ext,
		MD5: md5Hex, SHA1: sha1Hex, SizeKB: sizeKB,
		Width: width, Height: height, IsPublic: isPublic,
		IP: ip, ExpiredAt: expiredAt, AlbumID: albumRef, Tags: tags,
	})
	if err != nil {
		_ = fs.Delete(pathname)
		r.Error(w, "上传失败: "+err.Error())
		return
	}

	// 同步处理驱动（is_sync）+ 队列链
	d.afterUpload(rec, photo)

	r.Success(w, map[string]any{
		"id": photo.ID, "name": photo.Name, "filename": photo.Filename,
		"pathname": photo.Pathname, "mimetype": photo.Mimetype, "extension": photo.Extension,
		"md5": photo.MD5, "sha1": photo.SHA1,
		"width": photo.Width, "height": photo.Height,
		"ip_address": photo.IPAddress,
		"public_url": photostore.PublicURL(d.gdb, d.cfg, photo),
	})
}

// afterUpload 同步处理 + 派发任务链（对齐 UploadFinished → PhotoUploadComplete）。
func (d *deps) afterUpload(rec *photostore.StorageRecord, photo *model.Photo) {
	// 储存绑定的 handle/scan 驱动（storage_driver type）
	var syncHandle bool
	var driverOptions []struct {
		Type    string
		Options *string
	}
	d.gdb.Raw(
		"SELECT dr.`type`, dr.`options` FROM `drivers` dr "+
			"INNER JOIN `storage_driver` sd ON sd.driver_id = dr.id AND dr.deleted_at IS NULL "+
			"WHERE sd.storage_id = ? AND dr.`type` IN ('handle','scan')", rec.ID,
	).Scan(&driverOptions)
	for _, dr := range driverOptions {
		if dr.Options == nil {
			continue
		}
		var opts map[string]any
		if jsonUnmarshalInto(*dr.Options, &opts) != nil {
			continue
		}
		if isSync, _ := opts["is_sync"].(bool); isSync && dr.Type == "handle" {
			syncHandle = true
			_ = jobs.RunHandlePhoto(d.gdb, d.cfg, photo.ID, *dr.Options)
		}
	}

	// 任务链：缩略图（按储存配置）、auto_delete
	generateThumbnail := true
	if gt := rec.FSOptions().GenerateThumbnail; gt != nil {
		generateThumbnail = *gt
	}
	if generateThumbnail {
		_ = d.queue.Dispatch("generate_thumbnail", map[string]any{"photo_id": photo.ID})
	}
	if photo.ExpiredAt != nil {
		_ = d.queue.DispatchAt("auto_delete_photo", map[string]any{"photo_id": photo.ID},
			photo.ExpiredAt.Unix())
	}
	_ = syncHandle
}

func intersectTypes(master, allowed []string) map[string]bool {
	set := map[string]bool{}
	for _, a := range allowed {
		for _, m := range master {
			if a == m {
				set[m] = true
			}
		}
	}
	return set
}

func pathIntForm(s string) int64 {
	n, _ := strconv.ParseInt(strings.TrimSpace(s), 10, 64)
	return n
}

func splitTags(s string) []string {
	if s == "" {
		return nil
	}
	// 支持逗号分隔或 JSON 数组
	if strings.HasPrefix(s, "[") {
		var arr []string
		if jsonUnmarshalInto(s, &arr) == nil {
			return arr
		}
	}
	return strings.Split(s, ",")
}
