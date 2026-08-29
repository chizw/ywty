package httpx

import (
	"net/http"
	"strings"
	"time"

	"github.com/chizw/ywty/server-go/internal/authx"
	"github.com/chizw/ywty/server-go/internal/model"
	"github.com/chizw/ywty/server-go/internal/pagination"
	"github.com/chizw/ywty/server-go/internal/photostore"
	"github.com/chizw/ywty/server-go/internal/support/r"
	"github.com/chizw/ywty/server-go/internal/validate"
)

// photoRow 列表/详情序列化（对齐 UserPhotoResource）。
func (d *deps) photoRow(gdb photoRowDeps, p *model.Photo) map[string]any {
	out := map[string]any{
		"id": p.ID, "name": p.Name, "intro": p.Intro,
		"filename": p.Filename, "pathname": p.Pathname,
		"mimetype": p.Mimetype, "extension": p.Extension,
		"md5": p.MD5, "sha1": p.SHA1,
		"width": p.Width, "height": p.Height,
		"ip_address": p.IPAddress, "is_public": p.IsPublic,
		"expired_at": timePtrJSON(p.ExpiredAt), "created_at": timePtrJSON(p.CreatedAt),
		"public_url":    photostore.PublicURL(d.gdb, d.cfg, p),
		"thumbnail_url": photostore.ThumbnailURL(d.gdb, d.cfg, p),
	}
	// group/storage/albums/tags 关联
	var group struct {
		ID    int64
		Name  string
		Intro string
	}
	d.gdb.Raw("SELECT `id`, `name`, `intro` FROM `groups` WHERE `id` = ? LIMIT 1", derefI64(p.GroupID)).Scan(&group)
	out["group"] = map[string]any{"id": group.ID, "name": group.Name, "intro": group.Intro}

	var stor struct {
		ID       int64
		Name     string
		Intro    string
		Provider string
	}
	d.gdb.Raw("SELECT `id`, `name`, `intro`, `provider` FROM `storages` WHERE `id` = ? LIMIT 1", derefI64(p.StorageID)).Scan(&stor)
	out["storage"] = map[string]any{"id": stor.ID, "name": stor.Name, "intro": stor.Intro, "provider": stor.Provider}

	albums := []map[string]any{}
	var arows []struct {
		ID    int64
		Name  string
		Intro string
	}
	d.gdb.Raw(
		"SELECT a.`id`, a.`name`, a.`intro` FROM `albums` a INNER JOIN `album_photo` ap ON ap.album_id = a.id "+
			"WHERE ap.photo_id = ? AND a.deleted_at IS NULL", p.ID,
	).Scan(&arows)
	for _, a := range arows {
		albums = append(albums, map[string]any{"id": a.ID, "name": a.Name, "intro": a.Intro})
	}
	out["albums"] = albums

	tags := []map[string]any{}
	var trows []struct {
		ID   int64
		Name string
	}
	d.gdb.Raw(
		"SELECT t.`id`, t.`name` FROM `tags` t INNER JOIN `taggables` tt ON tt.tag_id = t.id "+
			"WHERE tt.taggable_type = 'photo' AND tt.taggable_id = ?", p.ID,
	).Scan(&trows)
	for _, t := range trows {
		tags = append(tags, map[string]any{"id": t.ID, "name": t.Name})
	}
	out["tags"] = tags
	return out
}

type photoRowDeps struct{}

func derefI64(p *int64) int64 {
	if p == nil {
		return 0
	}
	return *p
}

// ---------- GET /api/v2/user/photos ----------

func (d *deps) handleUserPhotos(w http.ResponseWriter, req *http.Request) {
	u := authx.From(req).User
	p := pagination.FromRequest(req)
	qp := req.URL.Query()

	where := "ph.`user_id` = ? AND ph.`deleted_at` IS NULL AND ph.`status` = 'normal'"
	args := []any{u.ID}
	if gid := pathIntForm(qp.Get("group_id")); gid > 0 {
		where += " AND ph.`group_id` = ?"
		args = append(args, gid)
	}
	if sid := pathIntForm(qp.Get("storage_id")); sid > 0 {
		where += " AND ph.`storage_id` = ?"
		args = append(args, sid)
	}
	if aid := pathIntForm(qp.Get("album_id")); aid > 0 {
		where += " AND EXISTS (SELECT 1 FROM `album_photo` ap WHERE ap.photo_id = ph.id AND ap.album_id = ?)"
		args = append(args, aid)
	}
	if q := p.Q; q != "" {
		for _, part := range strings.Fields(q) {
			if strings.Contains(part, ":") {
				continue // sort 等指令已在 M1 实现；表达式过滤此处跳过
			}
			where += " AND ph.`name` LIKE ?"
			args = append(args, "%"+part+"%")
		}
	}

	var total int64
	d.gdb.Raw("SELECT count(*) FROM `photos` ph WHERE "+where, args...).Scan(&total)
	var rows []model.Photo
	d.gdb.Raw("SELECT ph.* FROM `photos` ph WHERE "+where+" ORDER BY ph.`created_at` DESC, ph.`id` DESC LIMIT ? OFFSET ?",
		append(append([]any{}, args...), p.PerPage, p.Offset())...).Scan(&rows)

	out := make([]map[string]any, 0, len(rows))
	for i := range rows {
		out = append(out, d.photoRow(photoRowDeps{}, &rows[i]))
	}
	r.Success(w, pagination.New(out, total, p))
}

// GET /api/v2/user/photos/{id}
func (d *deps) handleUserPhotoShow(w http.ResponseWriter, req *http.Request) {
	u := authx.From(req).User
	id := pathInt(req, "id")
	var photo model.Photo
	if err := d.gdb.Where("id = ? AND user_id = ?", id, u.ID).First(&photo).Error; err != nil {
		r.Error(w, "图片不存在")
		return
	}
	r.Success(w, d.photoRow(photoRowDeps{}, &photo))
}

// PUT /api/v2/photos/update（批量更新）
func (d *deps) handlePhotosUpdate(w http.ResponseWriter, req *http.Request) {
	u := authx.From(req).User
	var body struct {
		IDs      []int64 `json:"ids"`
		Name     *string `json:"name"`
		Intro    *string `json:"intro"`
		IsPublic *bool   `json:"is_public"`
		Albums   []int64 `json:"albums"`
	}
	if err := readBody(req, &body); err != nil {
		r.Error(w, "请求体解析失败")
		return
	}
	if len(body.IDs) == 0 {
		r.ErrorData(w, http.StatusUnprocessableEntity, "The given data was invalid.",
			map[string]any{"errors": map[string][]string{"ids": {"ids 不能为空。"}}})
		return
	}
	updates := map[string]any{"updated_at": time.Now().UTC()}
	if body.Name != nil {
		updates["name"] = *body.Name
	}
	if body.Intro != nil {
		updates["intro"] = *body.Intro
	}
	if body.IsPublic != nil {
		updates["is_public"] = *body.IsPublic
	}
	d.gdb.Model(&model.Photo{}).Where("user_id = ? AND id IN ?", u.ID, body.IDs).Updates(updates)
	if body.Albums != nil {
		for _, id := range body.IDs {
			d.gdb.Exec("DELETE FROM `album_photo` WHERE `photo_id` = ?", id)
			for _, aid := range body.Albums {
				photostore.InsertIgnore(d.gdb, "INSERT INTO `album_photo` (`album_id`, `photo_id`, `sort`) VALUES (?, ?, 0)", aid, id)
			}
		}
	}
	r.Success(w, nil)
}

// DELETE /api/v2/photos（批量删除）
func (d *deps) handlePhotosDestroy(w http.ResponseWriter, req *http.Request) {
	u := authx.From(req).User
	var body struct {
		IDs []int64 `json:"ids"`
	}
	if err := readBody(req, &body); err != nil {
		r.Error(w, "请求体解析失败")
		return
	}
	if len(body.IDs) == 0 {
		r.ErrorData(w, http.StatusUnprocessableEntity, "The given data was invalid.",
			map[string]any{"errors": map[string][]string{"ids": {"ids 不能为空。"}}})
		return
	}
	var photos []model.Photo
	d.gdb.Where("user_id = ? AND id IN ?", u.ID, body.IDs).Find(&photos)
	for i := range photos {
		_ = photostore.DeletePhoto(d.gdb, d.cfg, &photos[i])
	}
	r.Success(w, nil)
}

// POST /api/v2/user/photos/{id}/tags
func (d *deps) handlePhotoTagsAttach(w http.ResponseWriter, req *http.Request) {
	d.mutatePhotoTags(w, req, true)
}

// DELETE /api/v2/user/photos/{id}/tags
func (d *deps) handlePhotoTagsRemove(w http.ResponseWriter, req *http.Request) {
	d.mutatePhotoTags(w, req, false)
}

func (d *deps) mutatePhotoTags(w http.ResponseWriter, req *http.Request, attach bool) {
	u := authx.From(req).User
	id := pathInt(req, "id")
	var body struct {
		Tags []string `json:"tags"`
	}
	if err := readBody(req, &body); err != nil {
		r.Error(w, "请求体解析失败")
		return
	}
	var photo model.Photo
	if err := d.gdb.Where("id = ? AND user_id = ?", id, u.ID).First(&photo).Error; err != nil {
		r.Error(w, "图片不存在")
		return
	}
	for _, name := range body.Tags {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		if !attach {
			d.gdb.Exec("DELETE FROM `taggables` WHERE `taggable_type` = 'photo' AND `taggable_id` = ? "+
				"AND `tag_id` IN (SELECT `id` FROM `tags` WHERE `name` = ?)", photo.ID, name)
			continue
		}
		var tagID int64
		d.gdb.Raw("SELECT `id` FROM `tags` WHERE `name` = ? LIMIT 1", name).Scan(&tagID)
		if tagID == 0 {
			res := d.gdb.Exec("INSERT INTO `tags` (`name`, `created_at`, `updated_at`) VALUES (?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)", name)
			if res.Error != nil {
				continue
			}
			d.gdb.Raw("SELECT `id` FROM `tags` WHERE `name` = ? LIMIT 1", name).Scan(&tagID)
		}
		d.gdb.Exec("INSERT OR IGNORE INTO `taggables` (`tag_id`, `user_id`, `taggable_type`, `taggable_id`) VALUES (?, ?, 'photo', ?)",
			tagID, u.ID, photo.ID)
	}
	r.Success(w, nil)
}

// ---------- GET /api/v2/user/albums ----------

func (d *deps) handleUserAlbums(w http.ResponseWriter, req *http.Request) {
	u := authx.From(req).User
	p := pagination.FromRequest(req)

	where := "a.`user_id` = ? AND a.`deleted_at` IS NULL"
	args := []any{u.ID}
	if q := p.Q; q != "" {
		where += " AND (a.`name` LIKE ? OR a.`intro` LIKE ?)"
		args = append(args, "%"+q+"%", "%"+q+"%")
	}
	var total int64
	d.gdb.Raw("SELECT count(*) FROM `albums` a WHERE "+where, args...).Scan(&total)
	var rows []model.Album
	d.gdb.Raw("SELECT a.* FROM `albums` a WHERE "+where+" ORDER BY a.`created_at` DESC, a.`id` DESC LIMIT ? OFFSET ?",
		append(append([]any{}, args...), p.PerPage, p.Offset())...).Scan(&rows)

	out := make([]map[string]any, 0, len(rows))
	for _, a := range rows {
		out = append(out, d.albumRow(&a))
	}
	r.Success(w, pagination.New(out, total, p))
}

// albumRow 序列化（对齐 UserAlbumResource：tags + covers + photo_count）。
func (d *deps) albumRow(a *model.Album) map[string]any {
	var photoCount int64
	d.gdb.Raw("SELECT count(*) FROM `album_photo` ap INNER JOIN `photos` ph ON ph.id = ap.photo_id AND ph.deleted_at IS NULL "+
		"WHERE ap.album_id = ?", a.ID).Scan(&photoCount)

	covers := []any{}
	var paths []string
	d.gdb.Raw(
		"SELECT ph.`pathname` FROM `photos` ph INNER JOIN `album_photo` ap ON ap.photo_id = ph.id "+
			"WHERE ap.album_id = ? AND ph.deleted_at IS NULL ORDER BY ap.sort ASC, ph.id DESC LIMIT 3", a.ID,
	).Scan(&paths)
	for _, pn := range paths {
		p := &model.Photo{Pathname: pn}
		covers = append(covers, photostore.ThumbnailURL(d.gdb, d.cfg, p))
	}

	tags := []map[string]any{}
	var trows []struct {
		ID   int64
		Name string
	}
	d.gdb.Raw(
		"SELECT t.`id`, t.`name` FROM `tags` t INNER JOIN `taggables` tt ON tt.tag_id = t.id "+
			"WHERE tt.taggable_type = 'album' AND tt.taggable_id = ?", a.ID,
	).Scan(&trows)
	for _, t := range trows {
		tags = append(tags, map[string]any{"id": t.ID, "name": t.Name})
	}

	return map[string]any{
		"id": a.ID, "name": a.Name, "intro": a.Intro,
		"tags": tags, "photo_count": photoCount, "covers": covers,
		"is_public": a.IsPublic, "created_at": timePtrJSON(a.CreatedAt),
	}
}

// POST /api/v2/user/albums
func (d *deps) handleAlbumsStore(w http.ResponseWriter, req *http.Request) {
	u := authx.From(req).User
	var body struct {
		Name     string   `json:"name"`
		Intro    string   `json:"intro"`
		IsPublic *bool    `json:"is_public"`
		Tags     []string `json:"tags"`
	}
	if err := readBody(req, &body); err != nil {
		r.Error(w, "请求体解析失败")
		return
	}
	v := validate.New()
	if !validate.Required(body.Name) || len(body.Name) > 255 {
		v.Add("name", "名称", "不能为空。")
	}
	if v.Fail() {
		v.Respond(w)
		return
	}
	now := time.Now().UTC()
	album := model.Album{
		UserID: &u.ID, Name: body.Name, Intro: body.Intro,
		IsPublic:  body.IsPublic != nil && *body.IsPublic,
		CreatedAt: &now, UpdatedAt: &now,
	}
	if err := d.gdb.Create(&album).Error; err != nil {
		r.Error(w, "创建失败")
		return
	}
	for _, name := range body.Tags {
		var tagID int64
		d.gdb.Raw("SELECT `id` FROM `tags` WHERE `name` = ? LIMIT 1", name).Scan(&tagID)
		if tagID == 0 {
			d.gdb.Exec("INSERT INTO `tags` (`name`, `created_at`, `updated_at`) VALUES (?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)", name)
			d.gdb.Raw("SELECT `id` FROM `tags` WHERE `name` = ? LIMIT 1", name).Scan(&tagID)
		}
		d.gdb.Exec("INSERT OR IGNORE INTO `taggables` (`tag_id`, `user_id`, `taggable_type`, `taggable_id`) VALUES (?, ?, 'album', ?)", tagID, u.ID, album.ID)
	}
	r.Created(w, d.albumRow(&album))
}

// GET /api/v2/user/albums/{id}
func (d *deps) handleUserAlbumShow(w http.ResponseWriter, req *http.Request) {
	u := authx.From(req).User
	id := pathInt(req, "id")
	var album model.Album
	if err := d.gdb.Where("id = ? AND user_id = ?", id, u.ID).First(&album).Error; err != nil {
		r.Error(w, "相册不存在")
		return
	}
	r.Success(w, d.albumRow(&album))
}

// PUT /api/v2/user/albums/{id}
func (d *deps) handleUserAlbumUpdate(w http.ResponseWriter, req *http.Request) {
	u := authx.From(req).User
	id := pathInt(req, "id")
	var album model.Album
	if err := d.gdb.Where("id = ? AND user_id = ?", id, u.ID).First(&album).Error; err != nil {
		r.Error(w, "相册不存在")
		return
	}
	var body struct {
		Name     *string `json:"name"`
		Intro    *string `json:"intro"`
		IsPublic *bool   `json:"is_public"`
	}
	if err := readBody(req, &body); err != nil {
		r.Error(w, "请求体解析失败")
		return
	}
	updates := map[string]any{"updated_at": time.Now().UTC()}
	if body.Name != nil {
		updates["name"] = *body.Name
	}
	if body.Intro != nil {
		updates["intro"] = *body.Intro
	}
	if body.IsPublic != nil {
		updates["is_public"] = *body.IsPublic
	}
	if err := d.gdb.Model(&model.Album{}).Where("id = ?", album.ID).Updates(updates).Error; err != nil {
		r.Error(w, "更新失败")
		return
	}
	_ = d.gdb.First(&album, album.ID).Error
	r.Success(w, d.albumRow(&album))
}

// DELETE /api/v2/user/albums/{id}
func (d *deps) handleUserAlbumDestroy(w http.ResponseWriter, req *http.Request) {
	u := authx.From(req).User
	id := pathInt(req, "id")
	res := d.gdb.Model(&model.Album{}).Where("id = ? AND user_id = ?", id, u.ID).
		Update("deleted_at", time.Now().UTC())
	if res.Error != nil || res.RowsAffected == 0 {
		r.Error(w, "相册不存在")
		return
	}
	r.Success(w, nil)
}

// POST /api/v2/user/albums/{id}/photos（添加图片）
func (d *deps) handleAlbumAddPhotos(w http.ResponseWriter, req *http.Request) {
	d.mutateAlbumPhotos(w, req, true)
}

// DELETE /api/v2/user/albums/{id}/photos
func (d *deps) handleAlbumRemovePhotos(w http.ResponseWriter, req *http.Request) {
	d.mutateAlbumPhotos(w, req, false)
}

func (d *deps) mutateAlbumPhotos(w http.ResponseWriter, req *http.Request, add bool) {
	u := authx.From(req).User
	id := pathInt(req, "id")
	var body struct {
		PhotoIDs []int64 `json:"photo_ids" `
	}
	if err := readBody(req, &body); err != nil {
		// 兼容 ids 字段名
		var alt struct {
			IDs []int64 `json:"ids"`
		}
		if err2 := readBody(req, &alt); err2 != nil || len(alt.IDs) == 0 {
			r.Error(w, "请求体解析失败")
			return
		}
		body.PhotoIDs = alt.IDs
	}
	var album model.Album
	if err := d.gdb.Where("id = ? AND user_id = ?", id, u.ID).First(&album).Error; err != nil {
		r.Error(w, "相册不存在")
		return
	}
	for _, pid := range body.PhotoIDs {
		var n int64
		d.gdb.Raw("SELECT count(*) FROM `photos` WHERE `id` = ? AND `user_id` = ? AND `deleted_at` IS NULL", pid, u.ID).Scan(&n)
		if n == 0 {
			continue
		}
		if add {
			d.gdb.Exec("INSERT OR IGNORE INTO `album_photo` (`album_id`, `photo_id`, `sort`) VALUES (?, ?, 0)", album.ID, pid)
		} else {
			d.gdb.Exec("DELETE FROM `album_photo` WHERE `album_id` = ? AND `photo_id` = ?", album.ID, pid)
		}
	}
	r.Success(w, nil)
}

// POST /api/v2/user/albums/{id}/tags
func (d *deps) handleAlbumTagsAttach(w http.ResponseWriter, req *http.Request) {
	d.mutateAlbumTags(w, req, true)
}

// DELETE /api/v2/user/albums/{id}/tags
func (d *deps) handleAlbumTagsRemove(w http.ResponseWriter, req *http.Request) {
	d.mutateAlbumTags(w, req, false)
}

func (d *deps) mutateAlbumTags(w http.ResponseWriter, req *http.Request, attach bool) {
	u := authx.From(req).User
	id := pathInt(req, "id")
	var body struct {
		Tags []string `json:"tags"`
	}
	if err := readBody(req, &body); err != nil {
		r.Error(w, "请求体解析失败")
		return
	}
	var album model.Album
	if err := d.gdb.Where("id = ? AND user_id = ?", id, u.ID).First(&album).Error; err != nil {
		r.Error(w, "相册不存在")
		return
	}
	for _, name := range body.Tags {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		if !attach {
			d.gdb.Exec("DELETE FROM `taggables` WHERE `taggable_type` = 'album' AND `taggable_id` = ? "+
				"AND `tag_id` IN (SELECT `id` FROM `tags` WHERE `name` = ?)", album.ID, name)
			continue
		}
		var tagID int64
		d.gdb.Raw("SELECT `id` FROM `tags` WHERE `name` = ? LIMIT 1", name).Scan(&tagID)
		if tagID == 0 {
			d.gdb.Exec("INSERT INTO `tags` (`name`, `created_at`, `updated_at`) VALUES (?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)", name)
			d.gdb.Raw("SELECT `id` FROM `tags` WHERE `name` = ? LIMIT 1", name).Scan(&tagID)
		}
		d.gdb.Exec("INSERT OR IGNORE INTO `taggables` (`tag_id`, `user_id`, `taggable_type`, `taggable_id`) VALUES (?, ?, 'album', ?)", tagID, u.ID, album.ID)
	}
	r.Success(w, nil)
}
