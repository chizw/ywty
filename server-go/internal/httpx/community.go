package httpx

import (
	"crypto/rand"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/chizw/ywty/server-go/internal/authx"
	"github.com/chizw/ywty/server-go/internal/config"
	"github.com/chizw/ywty/server-go/internal/model"
	"github.com/chizw/ywty/server-go/internal/pagination"
	"github.com/chizw/ywty/server-go/internal/photostore"
	"github.com/chizw/ywty/server-go/internal/setting"
	"github.com/chizw/ywty/server-go/internal/support/r"
	"github.com/chizw/ywty/server-go/internal/validate"
	"github.com/go-chi/chi/v5"
)

// 多态类型值与原版数据库一致（完整类名）。
const (
	morphPhoto = "App\\Models\\Photo"
	morphAlbum = "App\\Models\\Album"
	morphShare = "App\\Models\\Share"
	morphUser  = "App\\Models\\User"
)

// ---------- 用户分享（auth） ----------

// GET /api/v2/user/shares
func (d *deps) handleUserShares(w http.ResponseWriter, req *http.Request) {
	u := authx.From(req).User
	p := pagination.FromRequest(req)

	order := "`created_at` DESC"
	where := "s.`user_id` = ?"
	args := []any{u.ID}
	if q := p.Q; q != "" {
		for _, part := range strings.Fields(q) {
			if strings.HasPrefix(part, "sort:") {
				switch part {
				case "sort:view_count:ascend":
					order = "`view_count` ASC"
				case "sort:view_count:descend":
					order = "`view_count` DESC"
				case "sort:expired_at:ascend":
					order = "`expired_at` ASC"
				case "sort:expired_at:descend":
					order = "`expired_at` DESC"
				case "sort:created_at:ascend":
					order = "`created_at` ASC"
				case "sort:created_at:descend":
					order = "`created_at` DESC"
				}
				continue
			}
			where += " AND (s.`slug` LIKE ? OR s.`content` LIKE ?)"
			args = append(args, "%"+part+"%", "%"+part+"%")
		}
	}

	var total int64
	d.gdb.Raw("SELECT count(*) FROM `shares` s WHERE "+where, args...).Scan(&total)
	var rows []struct {
		ID        int64
		Type      string
		Slug      string
		Content   *string
		ViewCount int64
		LikeCount int64
		Password  string
		ExpiredAt *time.Time
		CreatedAt *time.Time
	}
	d.gdb.Raw(
		"SELECT s.*, (SELECT count(*) FROM `likes` l WHERE l.likeable_type = ? AND l.likeable_id = s.id) AS like_count "+
			"FROM `shares` s WHERE "+where+" ORDER BY "+order+" LIMIT ? OFFSET ?",
		append(append([]any{morphShare}, args...), p.PerPage, p.Offset())...,
	).Scan(&rows)

	out := make([]map[string]any, 0, len(rows))
	for _, s := range rows {
		out = append(out, map[string]any{
			"id": s.ID, "type": s.Type, "slug": s.Slug, "content": s.Content,
			"view_count": s.ViewCount, "like_count": s.LikeCount, "password": s.Password,
			"expired_at": timePtrJSON(s.ExpiredAt), "created_at": timePtrJSON(s.CreatedAt),
		})
	}
	r.Success(w, pagination.New(out, total, p))
}

// POST /api/v2/user/shares
func (d *deps) handleUserSharesStore(w http.ResponseWriter, req *http.Request) {
	u := authx.From(req).User
	var body struct {
		Type      string  `json:"type"`
		Content   string  `json:"content"`
		Password  string  `json:"password"`
		ExpiredAt *string `json:"expired_at"`
		IDs       []int64 `json:"ids"`
	}
	if err := readBody(req, &body); err != nil {
		r.Error(w, "请求体解析失败")
		return
	}
	v := validate.New()
	if !validate.In(body.Type, "album", "photo") {
		v.Add("type", "类型", "不存在。")
	}
	if body.Type == "album" && len(body.IDs) > 1 {
		v.Add("ids", "分享内容", "相册分享只能选择一个相册。")
	}
	if len(body.IDs) == 0 {
		v.Add("ids", "分享内容", "不能为空。")
	}
	if v.Fail() {
		v.Respond(w)
		return
	}

	slug := newShareSlug()
	now := time.Now().UTC()
	var expiredAt *time.Time
	if body.ExpiredAt != nil && *body.ExpiredAt != "" {
		for _, layout := range []string{time.RFC3339, "2006-01-02 15:04:05", "2006-01-02"} {
			if t, err := time.Parse(layout, *body.ExpiredAt); err == nil && t.After(now) {
				expiredAt = &t
				break
			}
		}
	}
	share := model.Share{
		UserID: u.ID, Type: body.Type, Slug: slug,
		Content:   strPtrOrNil(body.Content),
		Password:  body.Password,
		ExpiredAt: expiredAt,
		CreatedAt: &now, UpdatedAt: &now,
	}
	if err := d.gdb.Create(&share).Error; err != nil {
		r.Error(w, "创建分享失败")
		return
	}

	morph := morphPhoto
	table := "photos"
	if body.Type == "album" {
		morph = morphAlbum
		table = "albums"
	}
	var ids []int64
	d.gdb.Raw("SELECT `id` FROM `"+table+"` WHERE `user_id` = ? AND `id` IN ? AND `deleted_at` IS NULL", u.ID, body.IDs).Scan(&ids)
	if len(ids) == 0 {
		_ = d.gdb.Exec("DELETE FROM `shares` WHERE `id` = ?", share.ID).Error
		msg := "没有找到图片"
		if body.Type == "album" {
			msg = "没有找到相册"
		}
		r.Error(w, msg)
		return
	}
	for _, id := range ids {
		photostore.InsertIgnore(d.gdb,
			"INSERT INTO `shareables` (`share_id`, `shareable_type`, `shareable_id`) VALUES (?, '"+morph+"', ?)",
			share.ID, id)
	}
	r.Created(w, map[string]any{"id": share.ID, "slug": share.Slug})
}

// GET /api/v2/user/shares/{id}
func (d *deps) handleUserShareShow(w http.ResponseWriter, req *http.Request) {
	u := authx.From(req).User
	id := pathInt(req, "id")
	var share model.Share
	if err := d.gdb.Where("id = ? AND user_id = ?", id, u.ID).First(&share).Error; err != nil {
		r.Error(w, "分享不存在")
		return
	}
	var likeCount int64
	d.gdb.Raw("SELECT count(*) FROM `likes` WHERE `likeable_type` = ? AND `likeable_id` = ?", morphShare, share.ID).Scan(&likeCount)
	r.Success(w, map[string]any{
		"id": share.ID, "type": share.Type, "slug": share.Slug, "content": share.Content,
		"view_count": share.ViewCount, "like_count": likeCount, "password": share.Password,
		"expired_at": timePtrJSON(share.ExpiredAt), "created_at": timePtrJSON(share.CreatedAt),
	})
}

// PUT /api/v2/user/shares/{id}
func (d *deps) handleUserShareUpdate(w http.ResponseWriter, req *http.Request) {
	u := authx.From(req).User
	id := pathInt(req, "id")
	var share model.Share
	if err := d.gdb.Where("id = ? AND user_id = ?", id, u.ID).First(&share).Error; err != nil {
		r.Error(w, "分享不存在")
		return
	}
	var body struct {
		Content   *string `json:"content"`
		Password  *string `json:"password"`
		ExpiredAt *string `json:"expired_at"`
	}
	if err := readBody(req, &body); err != nil {
		r.Error(w, "请求体解析失败")
		return
	}
	updates := map[string]any{"updated_at": time.Now().UTC()}
	if body.Content != nil {
		updates["content"] = *body.Content
	}
	if body.Password != nil {
		updates["password"] = *body.Password
	}
	if body.ExpiredAt != nil {
		var t *time.Time
		for _, layout := range []string{time.RFC3339, "2006-01-02 15:04:05", "2006-01-02"} {
			if parsed, err := time.Parse(layout, *body.ExpiredAt); err == nil {
				t = &parsed
				break
			}
		}
		if t != nil {
			updates["expired_at"] = *t
		}
	}
	if err := d.gdb.Model(&model.Share{}).Where("id = ?", share.ID).Updates(updates).Error; err != nil {
		r.Error(w, "更新失败")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// DELETE /api/v2/user/shares
func (d *deps) handleUserSharesDestroy(w http.ResponseWriter, req *http.Request) {
	u := authx.From(req).User
	var body struct {
		IDs []int64 `json:"ids"`
	}
	if err := readBody(req, &body); err != nil || len(body.IDs) == 0 {
		r.Error(w, "请求体解析失败")
		return
	}
	d.gdb.Where("user_id = ? AND id IN ?", u.ID, body.IDs).Delete(&model.Share{})
	w.WriteHeader(http.StatusNoContent)
}

// ---------- 公共分享 ----------

func (d *deps) loadShareBySlug(slug string) (*model.Share, int64) {
	var share model.Share
	if err := d.gdb.Where("slug = ?", slug).First(&share).Error; err != nil {
		return nil, 0
	}
	var likeCount int64
	d.gdb.Raw("SELECT count(*) FROM `likes` WHERE `likeable_type` = ? AND `likeable_id` = ?", morphShare, share.ID).Scan(&likeCount)
	return &share, likeCount
}

// GET /api/v2/shares/{slug}
func (d *deps) handleShareShow(w http.ResponseWriter, req *http.Request) {
	share, likeCount := d.loadShareBySlug(chi.URLParam(req, "slug"))
	if share == nil {
		r.ErrorWithCode(w, http.StatusNotFound, "Not Found")
		return
	}
	if share.ExpiredAt != nil && share.ExpiredAt.Before(time.Now()) {
		r.ErrorWithCode(w, http.StatusNotFound, "分享已过期")
		return
	}
	if !d.verifySharePassword(share, req.URL.Query().Get("password")) {
		r.Success(w, map[string]any{"is_valid": false})
		return
	}
	d.gdb.Exec("UPDATE `shares` SET `view_count` = `view_count` + 1 WHERE `id` = ?", share.ID)

	var user struct {
		ID       int64
		Username string
		Name     string
		IsAdmin  bool
		Avatar   string
	}
	d.gdb.Raw("SELECT `id`, `username`, `name`, `is_admin`, `avatar` FROM `users` WHERE `id` = ?", share.UserID).Scan(&user)

	var album map[string]any
	if share.Type == "album" {
		var arow struct {
			ID    int64
			Name  string
			Intro string
		}
		d.gdb.Raw(
			"SELECT a.`id`, a.`name`, a.`intro` FROM `albums` a INNER JOIN `shareables` sb ON sb.shareable_id = a.id AND sb.shareable_type = ? "+
				"WHERE sb.share_id = ? AND a.deleted_at IS NULL LIMIT 1",
			morphAlbum, share.ID,
		).Scan(&arow)
		if arow.ID > 0 {
			album = map[string]any{"id": arow.ID, "name": arow.Name, "intro": arow.Intro}
		}
	}

	r.Success(w, map[string]any{
		"id": share.ID, "type": share.Type, "slug": share.Slug, "content": share.Content,
		"album": album,
		"user": map[string]any{
			"id": user.ID, "avatar_url": avatarURL(d.cfg, user.Avatar),
			"username": user.Username, "name": user.Name, "is_admin": user.IsAdmin,
		},
		"is_valid":   true,
		"view_count": share.ViewCount + 1,
		"like_count": likeCount,
		"is_liked":   d.userLikedBy(authx.From(req).User, morphShare, share.ID),
		"expired_at": timePtrJSON(share.ExpiredAt),
		"created_at": timePtrJSON(share.CreatedAt),
	})
}

// GET /api/v2/shares/{slug}/photos
func (d *deps) handleSharePhotos(w http.ResponseWriter, req *http.Request) {
	share, _ := d.loadShareBySlug(chi.URLParam(req, "slug"))
	if share == nil {
		r.ErrorWithCode(w, http.StatusNotFound, "Not Found")
		return
	}
	if share.ExpiredAt != nil && share.ExpiredAt.Before(time.Now()) {
		r.ErrorWithCode(w, http.StatusNotFound, "分享已过期")
		return
	}
	if !d.verifySharePassword(share, req.URL.Query().Get("password")) {
		r.Success(w, map[string]any{"is_valid": false})
		return
	}
	p := pagination.FromRequest(req)
	p.PerPage = 40
	rows, total := d.sharePhotosQuery(share, p)
	out := make([]map[string]any, 0, len(rows))
	for i := range rows {
		out = append(out, d.explorePhotoRow(authx.From(req).User, &rows[i]))
	}
	page := pagination.New(out, total, p)
	// 对齐原版响应形状（附加数据平级）：
	// {data, links, meta, is_valid} 平级
	r.Success(w, map[string]any{
		"data": page.Data, "links": page.Links, "meta": page.Meta, "is_valid": true,
	})
}

func (d *deps) sharePhotosQuery(share *model.Share, p pagination.Params) ([]model.Photo, int64) {
	morph := morphPhoto
	join, args := "", []any{}
	if share.Type == "album" {
		morph = morphAlbum
		join = "INNER JOIN `shareables` sb0 ON sb0.shareable_type = ? AND sb0.share_id = ? " +
			"INNER JOIN `album_photo` ap ON ap.album_id = sb0.shareable_id AND ap.photo_id = ph.id "
		args = append(args, morph, share.ID)
	} else {
		join = "INNER JOIN `shareables` sb0 ON sb0.shareable_type = ? AND sb0.shareable_id = ph.id AND sb0.share_id = ? "
		args = append(args, morph, share.ID)
	}
	where := join + "WHERE ph.`deleted_at` IS NULL AND EXISTS (SELECT 1 FROM `users` uu WHERE uu.id = ph.user_id AND uu.deleted_at IS NULL) " +
		"AND ph.`status` = 'normal'"
	var total int64
	d.gdb.Raw("SELECT count(*) FROM `photos` ph "+where, args...).Scan(&total)
	var rows []model.Photo
	d.gdb.Raw("SELECT ph.* FROM `photos` ph "+where+" ORDER BY ph.`created_at` DESC LIMIT ? OFFSET ?",
		append(append([]any{}, args...), p.PerPage, p.Offset())...).Scan(&rows)
	return rows, total
}

func (d *deps) verifySharePassword(share *model.Share, password string) bool {
	if share.Password != "" && password != share.Password {
		return false
	}
	return true
}

// POST /api/v2/shares/{slug}/report
func (d *deps) handleShareReport(w http.ResponseWriter, req *http.Request) {
	share, _ := d.loadShareBySlug(chi.URLParam(req, "slug"))
	if share == nil {
		r.ErrorWithCode(w, http.StatusNotFound, "Not Found")
		return
	}
	owner := share.UserID
	d.createReport(w, req, morphShare, share.ID, &owner)
}

// POST/DELETE /api/v2/shares/{slug}/like|unlike
func (d *deps) handleShareLike(w http.ResponseWriter, req *http.Request) {
	share, _ := d.loadShareBySlug(chi.URLParam(req, "slug"))
	if share == nil {
		r.ErrorWithCode(w, http.StatusNotFound, "Not Found")
		return
	}
	d.handleLike(morphShare, share.ID)(w, req)
}

func (d *deps) handleShareUnlike(w http.ResponseWriter, req *http.Request) {
	share, _ := d.loadShareBySlug(chi.URLParam(req, "slug"))
	if share == nil {
		r.ErrorWithCode(w, http.StatusNotFound, "Not Found")
		return
	}
	d.handleUnlike(morphShare, share.ID)(w, req)
}

// ---------- 点赞/举报公共实现 ----------

func (d *deps) handleLike(morph string, id int64) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		u := authx.From(req).User
		if u == nil {
			r.ErrorWithCode(w, http.StatusUnauthorized, "Unauthenticated.")
			return
		}
		var n int64
		d.gdb.Raw("SELECT count(*) FROM `likes` WHERE `likeable_type` = ? AND `likeable_id` = ? AND `user_id` = ?", morph, id, u.ID).Scan(&n)
		if n == 0 {
			now := time.Now().UTC()
			_ = d.gdb.Create(&model.Like{UserID: u.ID, LikeableType: morph, LikeableID: id, CreatedAt: &now, UpdatedAt: &now}).Error
		}
		r.Created(w, nil)
	}
}

func (d *deps) handleUnlike(morph string, id int64) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		u := authx.From(req).User
		if u == nil {
			r.ErrorWithCode(w, http.StatusUnauthorized, "Unauthenticated.")
			return
		}
		res := d.gdb.Exec("DELETE FROM `likes` WHERE `likeable_type` = ? AND `likeable_id` = ? AND `user_id` = ?", morph, id, u.ID)
		if res.Error != nil || res.RowsAffected == 0 {
			r.ErrorWithCode(w, http.StatusNotFound, "Not Found")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func (d *deps) userLikedBy(u *model.User, morph string, id int64) bool {
	if u == nil {
		return false
	}
	var n int64
	d.gdb.Raw("SELECT count(*) FROM `likes` WHERE `likeable_type` = ? AND `likeable_id` = ? AND `user_id` = ?", morph, id, u.ID).Scan(&n)
	return n > 0
}

// createReport 创建举报（content 必填）。
func (d *deps) createReport(w http.ResponseWriter, req *http.Request, morph string, id int64, ownerID *int64) {
	var body struct {
		Content string `json:"content"`
	}
	if err := readBody(req, &body); err != nil {
		r.Error(w, "请求体解析失败")
		return
	}
	v := validate.New()
	if !validate.Required(body.Content) || len(body.Content) > 2000 {
		v.Add("content", "举报内容", "不能为空。")
	}
	if v.Fail() {
		v.Respond(w)
		return
	}
	ip := authx.ClientIP(req)
	rep := model.Report{
		ReportUserID: ownerID, ReportableType: morph, ReportableID: id,
		Content: &body.Content, Status: "unhandled", IPAddress: &ip,
	}
	if err := d.gdb.Create(&rep).Error; err != nil {
		r.Error(w, "提交失败")
		return
	}
	r.Created(w, nil)
}

// ---------- 探索（广场） ----------

// explorePhotoRow 对齐 ExplorePhotoResource。
func (d *deps) explorePhotoRow(currentUser *model.User, p *model.Photo) map[string]any {
	tags := []map[string]any{}
	var trows []struct {
		ID   int64
		Name string
	}
	d.gdb.Raw(
		"SELECT t.`id`, t.`name` FROM `tags` t INNER JOIN `taggables` tt ON tt.tag_id = t.id "+
			"WHERE tt.taggable_type = ? AND tt.taggable_id = ?", morphPhoto, p.ID,
	).Scan(&trows)
	for _, t := range trows {
		tags = append(tags, map[string]any{"id": t.ID, "name": t.Name})
	}
	var user struct {
		ID       int64
		Username string
		Name     string
		IsAdmin  bool
		Avatar   string
	}
	d.gdb.Raw("SELECT `id`, `username`, `name`, `is_admin`, `avatar` FROM `users` WHERE `id` = ?", derefI64(p.UserID)).Scan(&user)
	return map[string]any{
		"id": p.ID, "tags": tags,
		"user": map[string]any{
			"id": user.ID, "avatar_url": avatarURL(d.cfg, user.Avatar),
			"username": user.Username, "name": user.Name, "is_admin": user.IsAdmin,
		},
		"name": p.Name, "intro": p.Intro, "extension": p.Extension,
		"width": p.Width, "height": p.Height,
		"thumbnail_url": photostore.ThumbnailURL(d.gdb, d.cfg, p),
		"public_url":    photostore.PublicURL(d.gdb, d.cfg, p),
		"is_liked":      d.userLikedBy(currentUser, morphPhoto, p.ID),
		"size":          p.Size,
	}
}

// exploreAlbumRow 对齐 ExploreAlbumResource。
func (d *deps) exploreAlbumRow(currentUser *model.User, a *model.Album) map[string]any {
	var photoCount int64
	d.gdb.Raw("SELECT count(*) FROM `album_photo` ap INNER JOIN `photos` ph ON ph.id = ap.photo_id AND ph.deleted_at IS NULL WHERE ap.album_id = ?", a.ID).Scan(&photoCount)

	var paths []string
	d.gdb.Raw(
		"SELECT ph.`pathname` FROM `photos` ph INNER JOIN `album_photo` ap ON ap.photo_id = ph.id "+
			"WHERE ap.album_id = ? AND ph.deleted_at IS NULL ORDER BY ph.created_at DESC LIMIT 3", a.ID,
	).Scan(&paths)
	covers := make([]any, 0, len(paths))
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
			"WHERE tt.taggable_type = ? AND tt.taggable_id = ?", morphAlbum, a.ID,
	).Scan(&trows)
	for _, t := range trows {
		tags = append(tags, map[string]any{"id": t.ID, "name": t.Name})
	}
	var user struct {
		ID       int64
		Username string
		Name     string
		IsAdmin  bool
		Avatar   string
	}
	d.gdb.Raw("SELECT `id`, `username`, `name`, `is_admin`, `avatar` FROM `users` WHERE `id` = ?", derefI64(a.UserID)).Scan(&user)
	return map[string]any{
		"id": a.ID, "name": a.Name, "intro": a.Intro, "tags": tags,
		"user": map[string]any{
			"id": user.ID, "avatar_url": avatarURL(d.cfg, user.Avatar),
			"username": user.Username, "name": user.Name, "is_admin": user.IsAdmin,
		},
		"photo_count": photoCount, "is_liked": d.userLikedBy(currentUser, morphAlbum, a.ID),
		"covers": covers, "created_at": timePtrJSON(a.CreatedAt),
	}
}

// explore photos: GET /api/v2/explore/photos
func (d *deps) handleExplorePhotos(w http.ResponseWriter, req *http.Request) {
	current := authx.From(req).User
	p := pagination.FromRequest(req)
	p.PerPage = clampPerPage(req.URL.Query().Get("per_page"), 40)
	where := "ph.`deleted_at` IS NULL AND ph.`is_public` = 1 AND ph.`status` = 'normal' " +
		"AND EXISTS (SELECT 1 FROM `users` uu WHERE uu.id = ph.user_id AND uu.deleted_at IS NULL)"
	args := []any{}
	if q := req.URL.Query().Get("q"); q != "" {
		where += " AND (ph.`name` LIKE ? OR ph.`intro` LIKE ?)"
		args = append(args, "%"+q+"%", "%"+q+"%")
	}
	var total int64
	d.gdb.Raw("SELECT count(*) FROM `photos` ph WHERE "+where, args...).Scan(&total)
	var rows []model.Photo
	d.gdb.Raw("SELECT ph.* FROM `photos` ph WHERE "+where+" ORDER BY ph.`created_at` DESC LIMIT ? OFFSET ?",
		append(append([]any{}, args...), p.PerPage, p.Offset())...).Scan(&rows)
	out := make([]map[string]any, 0, len(rows))
	for i := range rows {
		out = append(out, d.explorePhotoRow(current, &rows[i]))
	}
	r.Success(w, pagination.New(out, total, p))
}

func (d *deps) loadExplorePhoto(id int64) *model.Photo {
	var p model.Photo
	err := d.gdb.Raw(
		"SELECT ph.* FROM `photos` ph WHERE ph.`id` = ? AND ph.`deleted_at` IS NULL AND ph.`is_public` = 1 AND ph.`status` = 'normal' "+
			"AND EXISTS (SELECT 1 FROM `users` uu WHERE uu.id = ph.user_id AND uu.deleted_at IS NULL) LIMIT 1", id,
	).Scan(&p).Error
	if err != nil || p.ID == 0 {
		return nil
	}
	return &p
}

func (d *deps) handleExplorePhotoShow(w http.ResponseWriter, req *http.Request) {
	p := d.loadExplorePhoto(pathInt(req, "id"))
	if p == nil {
		r.ErrorWithCode(w, http.StatusNotFound, "Not Found")
		return
	}
	r.Success(w, d.explorePhotoRow(authx.From(req).User, p))
}

func (d *deps) handleExplorePhotoReport(w http.ResponseWriter, req *http.Request) {
	p := d.loadExplorePhoto(pathInt(req, "id"))
	if p == nil {
		r.ErrorWithCode(w, http.StatusNotFound, "Not Found")
		return
	}
	d.createReport(w, req, morphPhoto, p.ID, p.UserID)
}

func (d *deps) handleExplorePhotoLike(w http.ResponseWriter, req *http.Request) {
	p := d.loadExplorePhoto(pathInt(req, "id"))
	if p == nil {
		r.ErrorWithCode(w, http.StatusNotFound, "Not Found")
		return
	}
	d.handleLike(morphPhoto, p.ID)(w, req)
}

func (d *deps) handleExplorePhotoUnlike(w http.ResponseWriter, req *http.Request) {
	p := d.loadExplorePhoto(pathInt(req, "id"))
	if p == nil {
		r.ErrorWithCode(w, http.StatusNotFound, "Not Found")
		return
	}
	d.handleUnlike(morphPhoto, p.ID)(w, req)
}

// explore users
func (d *deps) loadExploreUser(username string) *model.User {
	var u model.User
	err := d.gdb.Where("status = 'normal' AND username = ? AND deleted_at IS NULL", username).First(&u).Error
	if err != nil {
		return nil
	}
	return &u
}

func (d *deps) handleExploreUserProfile(w http.ResponseWriter, req *http.Request) {
	u := d.loadExploreUser(chi.URLParam(req, "username"))
	if u == nil {
		r.ErrorWithCode(w, http.StatusNotFound, "Not Found")
		return
	}
	var photoCount, albumCount, likedPhotoCount, likedAlbumCount int64
	d.gdb.Raw("SELECT count(*) FROM `photos` WHERE `user_id` = ? AND `is_public` = 1 AND `deleted_at` IS NULL", u.ID).Scan(&photoCount)
	d.gdb.Raw("SELECT count(*) FROM `albums` WHERE `user_id` = ? AND `is_public` = 1 AND `deleted_at` IS NULL", u.ID).Scan(&albumCount)
	d.gdb.Raw("SELECT count(*) FROM `photos` WHERE `user_id` = ? AND `is_public` = 1 AND `deleted_at` IS NULL "+
		"AND EXISTS (SELECT 1 FROM `likes` l WHERE l.likeable_type = ? AND l.likeable_id = photos.id)", u.ID, morphPhoto).Scan(&likedPhotoCount)
	d.gdb.Raw("SELECT count(*) FROM `albums` WHERE `user_id` = ? AND `is_public` = 1 AND `deleted_at` IS NULL "+
		"AND EXISTS (SELECT 1 FROM `likes` l WHERE l.likeable_type = ? AND l.likeable_id = albums.id)", u.ID, morphAlbum).Scan(&likedAlbumCount)

	r.Success(w, map[string]any{
		"id": u.ID, "avatar_url": avatarURL(d.cfg, u.Avatar),
		"name": u.Name, "username": u.Username, "location": u.Location,
		"bio": u.Bio, "interests": jsonOrNull(u.Interests), "socials": jsonOrNull(u.Socials),
		"is_admin": u.IsAdmin, "created_at": timePtrJSON(u.CreatedAt),
		"photo_count": photoCount, "album_count": albumCount,
		"liked_photo_count": likedPhotoCount, "liked_album_count": likedAlbumCount,
	})
}

func (d *deps) handleExploreUserPhotos(w http.ResponseWriter, req *http.Request) {
	u := d.loadExploreUser(chi.URLParam(req, "username"))
	if u == nil {
		r.ErrorWithCode(w, http.StatusNotFound, "Not Found")
		return
	}
	p := pagination.FromRequest(req)
	p.PerPage = clampPerPage(req.URL.Query().Get("per_page"), 20)
	where := "ph.`user_id` = ? AND ph.`deleted_at` IS NULL AND ph.`is_public` = 1 AND ph.`status` = 'normal'"
	args := []any{u.ID}
	if q := req.URL.Query().Get("q"); q != "" {
		where += " AND (ph.`name` LIKE ? OR ph.`intro` LIKE ?)"
		args = append(args, "%"+q+"%", "%"+q+"%")
	}
	var total int64
	d.gdb.Raw("SELECT count(*) FROM `photos` ph WHERE "+where, args...).Scan(&total)
	var rows []model.Photo
	d.gdb.Raw("SELECT ph.* FROM `photos` ph WHERE "+where+" ORDER BY ph.`created_at` DESC LIMIT ? OFFSET ?",
		append(append([]any{}, args...), p.PerPage, p.Offset())...).Scan(&rows)
	out := make([]map[string]any, 0, len(rows))
	for i := range rows {
		out = append(out, d.explorePhotoRow(authx.From(req).User, &rows[i]))
	}
	r.Success(w, pagination.New(out, total, p))
}

func (d *deps) handleExploreUserAlbums(w http.ResponseWriter, req *http.Request) {
	u := d.loadExploreUser(chi.URLParam(req, "username"))
	if u == nil {
		r.ErrorWithCode(w, http.StatusNotFound, "Not Found")
		return
	}
	p := pagination.FromRequest(req)
	p.PerPage = clampPerPage(req.URL.Query().Get("per_page"), 20)
	where := "a.`user_id` = ? AND a.`deleted_at` IS NULL AND a.`is_public` = 1"
	args := []any{u.ID}
	if q := req.URL.Query().Get("q"); q != "" {
		where += " AND (a.`name` LIKE ? OR a.`intro` LIKE ?)"
		args = append(args, "%"+q+"%", "%"+q+"%")
	}
	var total int64
	d.gdb.Raw("SELECT count(*) FROM `albums` a WHERE "+where, args...).Scan(&total)
	var rows []model.Album
	d.gdb.Raw("SELECT a.* FROM `albums` a WHERE "+where+" ORDER BY a.`created_at` DESC LIMIT ? OFFSET ?",
		append(append([]any{}, args...), p.PerPage, p.Offset())...).Scan(&rows)
	out := make([]map[string]any, 0, len(rows))
	for i := range rows {
		out = append(out, d.exploreAlbumRow(authx.From(req).User, &rows[i]))
	}
	r.Success(w, pagination.New(out, total, p))
}

func (d *deps) handleExploreUserReport(w http.ResponseWriter, req *http.Request) {
	u := d.loadExploreUser(chi.URLParam(req, "username"))
	if u == nil {
		r.ErrorWithCode(w, http.StatusNotFound, "Not Found")
		return
	}
	owner := u.ID
	d.createReport(w, req, morphUser, u.ID, &owner)
}

// explore albums
func (d *deps) handleExploreAlbums(w http.ResponseWriter, req *http.Request) {
	current := authx.From(req).User
	p := pagination.FromRequest(req)
	p.PerPage = clampPerPage(req.URL.Query().Get("per_page"), 20)
	where := "a.`deleted_at` IS NULL AND a.`is_public` = 1 AND EXISTS (SELECT 1 FROM `users` uu WHERE uu.id = a.user_id AND uu.deleted_at IS NULL)"
	args := []any{}
	if q := req.URL.Query().Get("q"); q != "" {
		where += " AND (a.`name` LIKE ? OR a.`intro` LIKE ?)"
		args = append(args, "%"+q+"%", "%"+q+"%")
	}
	var total int64
	d.gdb.Raw("SELECT count(*) FROM `albums` a WHERE "+where, args...).Scan(&total)
	var rows []model.Album
	d.gdb.Raw("SELECT a.* FROM `albums` a WHERE "+where+" ORDER BY a.`created_at` DESC LIMIT ? OFFSET ?",
		append(append([]any{}, args...), p.PerPage, p.Offset())...).Scan(&rows)
	out := make([]map[string]any, 0, len(rows))
	for i := range rows {
		out = append(out, d.exploreAlbumRow(current, &rows[i]))
	}
	r.Success(w, pagination.New(out, total, p))
}

func (d *deps) loadExploreAlbum(id int64) *model.Album {
	var a model.Album
	err := d.gdb.Raw(
		"SELECT a.* FROM `albums` a WHERE a.`id` = ? AND a.`deleted_at` IS NULL AND a.`is_public` = 1 "+
			"AND EXISTS (SELECT 1 FROM `users` uu WHERE uu.id = a.user_id AND uu.deleted_at IS NULL) LIMIT 1", id,
	).Scan(&a).Error
	if err != nil || a.ID == 0 {
		return nil
	}
	return &a
}

func (d *deps) handleExploreAlbumShow(w http.ResponseWriter, req *http.Request) {
	a := d.loadExploreAlbum(pathInt(req, "id"))
	if a == nil {
		r.ErrorWithCode(w, http.StatusNotFound, "Not Found")
		return
	}
	r.Success(w, d.exploreAlbumRow(authx.From(req).User, a))
}

func (d *deps) handleExploreAlbumPhotos(w http.ResponseWriter, req *http.Request) {
	a := d.loadExploreAlbum(pathInt(req, "id"))
	if a == nil {
		r.ErrorWithCode(w, http.StatusNotFound, "Not Found")
		return
	}
	current := authx.From(req).User
	p := pagination.FromRequest(req)
	p.PerPage = clampPerPage(req.URL.Query().Get("per_page"), 40)
	where := "INNER JOIN `album_photo` ap ON ap.photo_id = ph.id AND ap.album_id = " + fmt.Sprintf("%d", a.ID) +
		" WHERE ph.`deleted_at` IS NULL AND ph.`is_public` = 1 AND ph.`status` = 'normal'"
	var total int64
	d.gdb.Raw("SELECT count(*) FROM `photos` ph " + where).Scan(&total)
	var rows []model.Photo
	d.gdb.Raw("SELECT ph.* FROM `photos` ph "+where+" ORDER BY ph.`created_at` DESC LIMIT ? OFFSET ?", p.PerPage, p.Offset()).Scan(&rows)
	out := make([]map[string]any, 0, len(rows))
	for i := range rows {
		out = append(out, d.explorePhotoRow(current, &rows[i]))
	}
	r.Success(w, pagination.New(out, total, p))
}

func (d *deps) handleExploreAlbumReport(w http.ResponseWriter, req *http.Request) {
	a := d.loadExploreAlbum(pathInt(req, "id"))
	if a == nil {
		r.ErrorWithCode(w, http.StatusNotFound, "Not Found")
		return
	}
	owner := derefI64(a.UserID)
	d.createReport(w, req, morphAlbum, a.ID, &owner)
}

func (d *deps) handleExploreAlbumLike(w http.ResponseWriter, req *http.Request) {
	a := d.loadExploreAlbum(pathInt(req, "id"))
	if a == nil {
		r.ErrorWithCode(w, http.StatusNotFound, "Not Found")
		return
	}
	d.handleLike(morphAlbum, a.ID)(w, req)
}

func (d *deps) handleExploreAlbumUnlike(w http.ResponseWriter, req *http.Request) {
	a := d.loadExploreAlbum(pathInt(req, "id"))
	if a == nil {
		r.ErrorWithCode(w, http.StatusNotFound, "Not Found")
		return
	}
	d.handleUnlike(morphAlbum, a.ID)(w, req)
}

// ---------- 公告与页面 ----------

// GET /api/v2/notices
func (d *deps) handleNoticesIndex(w http.ResponseWriter, req *http.Request) {
	p := pagination.FromRequest(req)
	var total int64
	d.gdb.Raw("SELECT count(*) FROM `notices` WHERE `deleted_at` IS NULL").Scan(&total)
	var rows []model.Notice
	d.gdb.Raw("SELECT * FROM `notices` WHERE `deleted_at` IS NULL ORDER BY `created_at` DESC LIMIT ? OFFSET ?",
		p.PerPage, p.Offset()).Scan(&rows)
	out := make([]map[string]any, 0, len(rows))
	for _, n := range rows {
		out = append(out, map[string]any{
			"id": n.ID, "title": n.Title, "content": excerpt(n.Content, 200),
			"created_at": timePtrJSON(n.CreatedAt),
		})
	}
	r.Success(w, pagination.New(out, total, p))
}

// GET /api/v2/notices/{id}
func (d *deps) handleNoticeShow(w http.ResponseWriter, req *http.Request) {
	id := pathInt(req, "id")
	var n model.Notice
	if err := d.gdb.First(&n, id).Error; err != nil {
		r.ErrorWithCode(w, http.StatusNotFound, "Not Found")
		return
	}
	d.gdb.Exec("UPDATE `notices` SET `view_count` = `view_count` + 1 WHERE `id` = ?", n.ID)
	r.Success(w, map[string]any{
		"id": n.ID, "title": n.Title, "content": n.Content, "created_at": timePtrJSON(n.CreatedAt),
	})
}

// GET /api/v2/pages
func (d *deps) handlePagesIndex(w http.ResponseWriter, req *http.Request) {
	p := pagination.FromRequest(req)
	var total int64
	d.gdb.Raw("SELECT count(*) FROM `pages` WHERE `is_show` = 1").Scan(&total)
	var rows []model.Page
	d.gdb.Raw("SELECT * FROM `pages` WHERE `is_show` = 1 ORDER BY `sort` ASC, `id` ASC LIMIT ? OFFSET ?",
		p.PerPage, p.Offset()).Scan(&rows)
	out := make([]map[string]any, 0, len(rows))
	for _, pg := range rows {
		out = append(out, map[string]any{
			"id": pg.ID, "type": pg.Type, "icon": pg.Icon, "name": pg.Name,
			"title": pg.Title, "slug": pg.Slug, "url": pg.URL, "view_count": pg.ViewCount,
		})
	}
	r.Success(w, pagination.New(out, total, p))
}

// GET /api/v2/pages/{slug}
func (d *deps) handlePageShow(w http.ResponseWriter, req *http.Request) {
	slug := chi.URLParam(req, "slug")
	var pg model.Page
	if err := d.gdb.Where("is_show = 1 AND slug = ?", slug).First(&pg).Error; err != nil {
		r.ErrorWithCode(w, http.StatusNotFound, "Not Found")
		return
	}
	d.gdb.Exec("UPDATE `pages` SET `view_count` = `view_count` + 1 WHERE `id` = ?", pg.ID)
	r.Success(w, map[string]any{
		"id": pg.ID, "type": pg.Type, "icon": pg.Icon, "name": pg.Name,
		"title": pg.Title, "slug": pg.Slug, "url": pg.URL, "view_count": pg.ViewCount + 1,
		"content": pg.Content, "created_at": timePtrJSON(pg.CreatedAt),
	})
}

// ---------- 工具 ----------

func newShareSlug() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x%x%x%x%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

func strPtrOrNil(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func avatarURL(cfg *config.Config, avatar string) string {
	if avatar == "" {
		return ""
	}
	if strings.HasPrefix(avatar, "http://") || strings.HasPrefix(avatar, "https://") {
		return avatar
	}
	return strings.TrimRight(cfg.AppURL, "/") + "/storage/" + strings.TrimPrefix(avatar, "/")
}

func excerpt(s *string, n int) string {
	if s == nil {
		return ""
	}
	runes := []rune(*s)
	if len(runes) <= n {
		return *s
	}
	return string(runes[:n]) + "..."
}

func clampPerPage(raw string, def int) int {
	if raw == "" {
		return def
	}
	n := atoiOr(def, raw)
	if n < 1 {
		return def
	}
	return n
}

// requireExplore 广场开关中间件（CheckExploreEnabled）。
func (d *deps) requireExplore(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if !d.exploreEnabled() {
			r.ErrorWithCode(w, http.StatusNotFound, "Gallery feature is disabled")
			return
		}
		next.ServeHTTP(w, req)
	})
}

func (d *deps) exploreEnabled() bool {
	v, _ := setting.Bool(d.gdb, setting.GroupApp, "enable_explore")
	return v
}
