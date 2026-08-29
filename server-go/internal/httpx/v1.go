package httpx

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/chizw/ywty/server-go/internal/authx"
	"github.com/chizw/ywty/server-go/internal/config"
	"github.com/chizw/ywty/server-go/internal/imageproc"
	"github.com/chizw/ywty/server-go/internal/model"
	"github.com/chizw/ywty/server-go/internal/pagination"
	"github.com/chizw/ywty/server-go/internal/photostore"
	"github.com/chizw/ywty/server-go/internal/support/r"
	"gorm.io/gorm"
)

// v1 旧版（1.x / PicGo 插件）兼容接口。
// 响应信封与 v2 不同：{"status": true|false, "message": "...", "data": {...}}。

type v1Envelope struct {
	Status  bool           `json:"status"`
	Message string         `json:"message"`
	Data    map[string]any `json:"data"`
}

func v1Success(w http.ResponseWriter, message string, data map[string]any) {
	if data == nil {
		data = map[string]any{}
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(v1Envelope{Status: true, Message: message, Data: data})
}

func v1Fail(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(v1Envelope{Status: false, Message: message, Data: map[string]any{}})
}

// GET /api/v1/strategies
func (d *deps) v1Strategies(w http.ResponseWriter, req *http.Request) {
	ctx := authx.From(req)
	if ctx.Group == nil {
		v1Fail(w, "系统未初始化角色组")
		return
	}
	var rows []struct {
		ID   int64
		Name string
	}
	d.gdb.Raw(
		"SELECT s.`id`, s.`name` FROM `storages` s INNER JOIN `group_storage` gs ON gs.storage_id = s.id "+
			"WHERE gs.group_id = ? AND s.deleted_at IS NULL ORDER BY gs.sort ASC", ctx.Group.ID,
	).Scan(&rows)
	list := make([]map[string]any, 0, len(rows))
	for _, r := range rows {
		list = append(list, map[string]any{"id": r.ID, "name": r.Name})
	}
	v1Success(w, "success", map[string]any{"strategies": list})
}

// POST /api/v1/upload
func (d *deps) v1Upload(w http.ResponseWriter, req *http.Request) {
	ctx := authx.From(req)

	if err := req.ParseMultipartForm(32 << 20); err != nil {
		v1Fail(w, "文件资源无效")
		return
	}
	file, header, err := req.FormFile("file")
	if err != nil {
		v1Fail(w, "请选择要上传的文件")
		return
	}
	defer func() { _ = file.Close() }()

	var groupOpts struct {
		AllowFileTypes []string `json:"allow_file_types"`
		MaxUploadSize  int      `json:"max_upload_size"`
	}
	if ctx.Group != nil && ctx.Group.Options != nil {
		_ = jsonUnmarshalInto(string(ctx.Group.Options), &groupOpts)
	}
	ext := ""
	if dot := strings.LastIndex(header.Filename, "."); dot >= 0 {
		ext = imageproc.NormalizeExt(header.Filename[dot+1:])
	}
	allowed := intersectTypes(masterImageTypes, groupOpts.AllowFileTypes)
	if !allowed[ext] {
		v1Fail(w, "不允许的文件类型")
		return
	}
	data, readErr := io.ReadAll(file)
	if readErr != nil {
		v1Fail(w, "文件读取失败")
		return
	}
	sizeKB := float64(len(data)) / 1024
	if groupOpts.MaxUploadSize > 0 && sizeKB > float64(groupOpts.MaxUploadSize) {
		v1Fail(w, "文件超出大小限制")
		return
	}

	// token 临时上传凭证（未登录时）
	user := ctx.User
	if user == nil {
		if token := req.FormValue("token"); token != "" {
			if uid, ok := d.cache.Get("v1_upload_token:" + token); ok {
				id, _ := strconv.ParseInt(uid, 10, 64)
				var u model.User
				if err := d.gdb.First(&u, id).Error; err == nil {
					user = &u
				}
			} else {
				v1Fail(w, "Token 无效或已过期")
				return
			}
		}
	}
	if user == nil && ctx.Group == nil {
		v1Fail(w, "请先登录")
		return
	}

	// 容量
	if user != nil {
		total := photostore.UserTotalCapacityKB(d.gdb, user.ID)
		used := photostore.SumUserUsedKB(d.gdb, user.ID)
		if total > 0 && used+sizeKB > total {
			v1Fail(w, "储存空间不足")
			return
		}
	}

	// 储存：strategy_id 或默认第一个
	strategyID := pathIntForm(req.FormValue("strategy_id"))
	var rec *photostore.StorageRecord
	if strategyID > 0 && ctx.Group != nil {
		rec, _ = photostore.LoadGroupStorage(d.gdb, ctx.Group.ID, strategyID)
	} else if ctx.Group != nil {
		rec, _ = photostore.DefaultGroupStorage(d.gdb, ctx.Group.ID)
	}
	if rec == nil {
		v1Fail(w, "不存在的储存驱动")
		return
	}
	fs, err := rec.Filesystem(d.cfg)
	if err != nil {
		v1Fail(w, err.Error())
		return
	}

	md5Hex, sha1Hex := photostore.HashBytes(data)
	var uid int64
	if user != nil {
		uid = user.ID
	}
	name := header.Filename
	if dot := strings.LastIndex(name, "."); dot >= 0 {
		name = name[:dot]
	}
	pathname := photostore.Pathname(rec.FSOptions().NamingRule, name, ext, md5Hex, sha1Hex, uid, time.Now()) + "." + ext
	if err := fs.AppendOrCreate(pathname, data); err != nil {
		v1Fail(w, "文件写入失败")
		return
	}
	width, height := imageproc.Size(data, ext)
	isPublic := req.FormValue("permission") == "1"
	var userID *int64
	if user != nil {
		userID = &user.ID
	}
	var albumRef *int64
	if aid := pathIntForm(req.FormValue("album_id")); aid > 0 && user != nil {
		albumRef = &aid
	}
	var expiredAt *time.Time
	if secs := pathIntForm(req.FormValue("expired_seconds")); secs > 0 {
		t := time.Now().Add(time.Duration(secs) * time.Second)
		expiredAt = &t
	}
	if ea := req.FormValue("expired_at"); ea != "" {
		for _, layout := range []string{time.RFC3339, "2006-01-02 15:04:05", "2006-01-02"} {
			if t, err := time.Parse(layout, ea); err == nil && t.After(time.Now()) {
				expiredAt = &t
				break
			}
		}
	}
	ip := authx.ClientIP(req)
	photo, err := photostore.Store(d.gdb, photostore.StoreInput{
		UserID: userID, GroupID: ctx.Group.ID, StorageID: rec.ID,
		Filename: header.Filename, Name: name, Pathname: pathname,
		Mimetype: imageproc.MimeTypeByExt(ext), Extension: ext,
		MD5: md5Hex, SHA1: sha1Hex, SizeKB: sizeKB,
		Width: width, Height: height, IsPublic: isPublic,
		IP: ip, ExpiredAt: expiredAt, AlbumID: albumRef,
	})
	if err != nil {
		_ = fs.Delete(pathname)
		v1Fail(w, "上传失败")
		return
	}
	d.afterUpload(rec, photo)
	v1Success(w, "上传成功", v1ImageRow(d.gdb, d.cfg, photo))
}

// GET /api/v1/images
func (d *deps) v1ImagesIndex(w http.ResponseWriter, req *http.Request) {
	u := authx.From(req).User
	order := "`id` DESC"
	switch req.URL.Query().Get("order") {
	case "earliest":
		order = "`id` ASC"
	case "utmost":
		order = "`size` DESC"
	case "least":
		order = "`size` ASC"
	}
	where := "`user_id` = ? AND `deleted_at` IS NULL"
	args := []any{u.ID}
	switch req.URL.Query().Get("permission") {
	case "public":
		where += " AND `is_public` = 1"
	case "private":
		where += " AND `is_public` = 0"
	}
	if q := req.URL.Query().Get("q"); q != "" {
		where += " AND (`name` LIKE ? OR `filename` LIKE ?)"
		args = append(args, "%"+q+"%", "%"+q+"%")
	}
	if aid := pathIntForm(req.URL.Query().Get("album_id")); aid > 0 {
		where += " AND `id` IN (SELECT `photo_id` FROM `album_photo` WHERE `album_id` = ?)"
		args = append(args, aid)
	}

	p := pagination.FromRequest(req)
	var total int64
	d.gdb.Raw("SELECT count(*) FROM `photos` WHERE "+where, args...).Scan(&total)
	var rows []model.Photo
	d.gdb.Raw("SELECT * FROM `photos` WHERE "+where+" ORDER BY "+order+" LIMIT ? OFFSET ?",
		append(append([]any{}, args...), 40, p.Offset())...).Scan(&rows)
	out := make([]map[string]any, 0, len(rows))
	for i := range rows {
		out = append(out, v1ImageRow(d.gdb, d.cfg, &rows[i]))
	}
	r.Success(w, pagination.New(out, total, p))
}

// DELETE /api/v1/images/{key}
func (d *deps) v1ImageDestroy(w http.ResponseWriter, req *http.Request) {
	u := authx.From(req).User
	id := pathInt(req, "key")
	var photo model.Photo
	if err := d.gdb.Where("id = ? AND user_id = ?", id, u.ID).First(&photo).Error; err != nil {
		v1Fail(w, "图片不存在")
		return
	}
	_ = photostore.DeletePhoto(d.gdb, d.cfg, &photo)
	v1Success(w, "删除成功", nil)
}

// POST /api/v1/images/tokens
func (d *deps) v1ImageTokens(w http.ResponseWriter, req *http.Request) {
	u := authx.From(req).User
	var body struct {
		Num     int64 `json:"num"`
		Seconds int64 `json:"seconds"`
	}
	if err := readBody(req, &body); err != nil {
		v1Fail(w, "请求体解析失败")
		return
	}
	if body.Num < 1 || body.Num > 100 {
		v1Fail(w, "数量必须是 1-100 之间的数字")
		return
	}
	if body.Seconds < 1 || body.Seconds > 2626560 {
		v1Fail(w, "到期时间最大 1 个月")
		return
	}
	tokens := make([]map[string]any, 0, body.Num)
	for i := int64(0); i < body.Num; i++ {
		key := randMD5Key(u.ID)
		expiredAt := time.Now().Add(time.Duration(body.Seconds) * time.Second)
		d.cache.Put("v1_upload_token:"+key, itoa(int(u.ID)), int(body.Seconds))
		tokens = append(tokens, map[string]any{
			"token":      key,
			"expired_at": expiredAt.Format("2006-01-02 15:04:05"),
		})
	}
	v1Success(w, "生成成功", map[string]any{"tokens": tokens})
}

// GET /api/v1/albums
func (d *deps) v1AlbumsIndex(w http.ResponseWriter, req *http.Request) {
	u := authx.From(req).User
	order := "`id` DESC"
	switch req.URL.Query().Get("order") {
	case "earliest":
		order = "`id` ASC"
	}
	where := "`user_id` = ? AND `deleted_at` IS NULL"
	args := []any{u.ID}
	switch req.URL.Query().Get("permission") {
	case "public":
		where += " AND `is_public` = 1"
	case "private":
		where += " AND `is_public` = 0"
	}
	if q := req.URL.Query().Get("q"); q != "" {
		where += " AND (`name` LIKE ? OR `intro` LIKE ?)"
		args = append(args, "%"+q+"%", "%"+q+"%")
	}
	p := pagination.FromRequest(req)
	p.PerPage = 40
	var total int64
	d.gdb.Raw("SELECT count(*) FROM `albums` WHERE "+where, args...).Scan(&total)
	var rows []model.Album
	d.gdb.Raw("SELECT * FROM `albums` WHERE "+where+" ORDER BY "+order+" LIMIT ? OFFSET ?",
		append(append([]any{}, args...), p.PerPage, p.Offset())...).Scan(&rows)
	out := make([]map[string]any, 0, len(rows))
	for _, a := range rows {
		var imageNum int64
		d.gdb.Raw("SELECT count(*) FROM `album_photo` ap INNER JOIN `photos` ph ON ph.id = ap.photo_id AND ph.deleted_at IS NULL WHERE ap.album_id = ?", a.ID).Scan(&imageNum)
		out = append(out, map[string]any{"id": a.ID, "name": a.Name, "intro": a.Intro, "image_num": imageNum})
	}
	r.Success(w, pagination.New(out, total, p))
}

// DELETE /api/v1/albums/{id}
func (d *deps) v1AlbumDestroy(w http.ResponseWriter, req *http.Request) {
	u := authx.From(req).User
	id := pathInt(req, "id")
	d.gdb.Model(&model.Album{}).Where("id = ? AND user_id = ?", id, u.ID).Update("deleted_at", time.Now().UTC())
	v1Success(w, "删除成功", nil)
}

// GET /api/v1/profile
func (d *deps) v1Profile(w http.ResponseWriter, req *http.Request) {
	u := authx.From(req).User
	var imageNum, albumNum int64
	d.gdb.Raw("SELECT count(*) FROM `photos` WHERE `user_id` = ? AND `deleted_at` IS NULL", u.ID).Scan(&imageNum)
	d.gdb.Raw("SELECT count(*) FROM `albums` WHERE `user_id` = ? AND `deleted_at` IS NULL", u.ID).Scan(&albumNum)
	v1Success(w, "success", map[string]any{
		"username":      u.Username,
		"name":          u.Name,
		"avatar":        u.Avatar,
		"email":         u.Email,
		"url":           "",
		"capacity":      photostore.UserTotalCapacityKB(d.gdb, u.ID),
		"size":          photostore.SumUserUsedKB(d.gdb, u.ID),
		"image_num":     imageNum,
		"album_num":     albumNum,
		"registered_ip": u.RegisterIP,
	})
}

// v1ImageRow 旧版图片行（含多格式链接与人性化时间）。
func v1ImageRow(gdb *gorm.DB, cfg *config.Config, p *model.Photo) map[string]any {
	publicURL := photostore.PublicURL(gdb, cfg, p)
	thumbnailURL := photostore.ThumbnailURL(gdb, cfg, p)
	humanDate := "刚刚"
	date := ""
	if p.CreatedAt != nil {
		date = p.CreatedAt.Format("2006-01-02 15:04:05")
		humanDate = humanDiff(p.CreatedAt)
	}
	return map[string]any{
		"key":         p.ID,
		"name":        p.Name,
		"pathname":    p.Pathname,
		"origin_name": p.Filename,
		"size":        p.Size,
		"mimetype":    p.Mimetype,
		"extension":   p.Extension,
		"md5":         p.MD5,
		"sha1":        p.SHA1,
		"width":       p.Width,
		"height":      p.Height,
		"human_date":  humanDate,
		"date":        date,
		"links": map[string]any{
			"url":                publicURL,
			"html":               "&lt;img src=\"" + publicURL + "\" alt=\"" + p.Filename + "\" title=\"" + p.Filename + "\" /&gt;",
			"bbcode":             "[img]" + publicURL + "[/img]",
			"markdown":           "![" + p.Filename + "](" + publicURL + ")",
			"markdown_with_link": "[![" + p.Filename + "](" + publicURL + ")](" + publicURL + ")",
			"thumbnail_url":      thumbnailURL,
			"delete_url":         "",
		},
	}
}

// humanDiff 人性化时间差。
func humanDiff(t *time.Time) string {
	d := time.Since(*t)
	switch {
	case d < time.Minute:
		return "刚刚"
	case d < time.Hour:
		return itoa(int(d.Minutes())) + " 分钟前"
	case d < 24*time.Hour:
		return itoa(int(d.Hours())) + " 小时前"
	case d < 30*24*time.Hour:
		return itoa(int(d.Hours()/24)) + " 天前"
	case d < 365*24*time.Hour:
		return itoa(int(d.Hours()/24/30)) + " 个月前"
	default:
		return itoa(int(d.Hours()/24/365)) + " 年前"
	}
}

// randMD5Key 临时上传令牌（md5(random+uid)）。
func randMD5Key(uid int64) string {
	raw := make([]byte, 16)
	_, _ = rand.Read(raw)
	sum := md5.Sum(append(raw, byte(uid), byte(uid>>8)))
	return hex.EncodeToString(sum[:])
}
