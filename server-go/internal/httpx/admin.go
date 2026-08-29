// Package httpx — /api/admin/* 管理接口（替代 Filament 的服务端能力）。
// 认证：会话或令牌 + users.is_admin 校验。
package httpx

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/chizw/ywty/server-go/internal/authx"
	"github.com/chizw/ywty/server-go/internal/model"
	"github.com/chizw/ywty/server-go/internal/orderx"
	"github.com/chizw/ywty/server-go/internal/pagination"
	"github.com/chizw/ywty/server-go/internal/photostore"
	"github.com/chizw/ywty/server-go/internal/setting"
	"github.com/chizw/ywty/server-go/internal/support/r"
	"github.com/chizw/ywty/server-go/internal/validate"
	"golang.org/x/crypto/bcrypt"
)

// adminAuth 管理员鉴权中间件。
func (d *deps) adminAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		u := authx.From(req).User
		if u == nil || !u.IsAdmin {
			r.ErrorWithCode(w, http.StatusForbidden, "您的令牌没有权限访问此API")
			return
		}
		next.ServeHTTP(w, req)
	})
}

// ---------- 登录 ----------

// POST /api/admin/login
func (d *deps) handleAdminLogin(w http.ResponseWriter, req *http.Request) {
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := readBody(req, &body); err != nil {
		r.Error(w, "请求体解析失败")
		return
	}
	var user model.User
	if err := d.gdb.Where("username = ?", strings.ToLower(body.Username)).First(&user).Error; err != nil ||
		bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(body.Password)) != nil {
		r.ErrorData(w, http.StatusUnprocessableEntity, "The given data was invalid.",
			map[string]any{"errors": map[string][]string{"username": {"These credentials do not match our records."}}})
		return
	}
	if !user.IsAdmin {
		r.ErrorWithCode(w, http.StatusForbidden, "您的令牌没有权限访问此API")
		return
	}
	sid, err := authx.CreateSession(d.gdb, user.ID, authx.ClientIP(req), req.UserAgent())
	if err != nil {
		r.Error(w, "登录失败")
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name: authx.SessionCookieName(d.cfg), Value: sid, Path: "/",
		HttpOnly: true, SameSite: http.SameSiteLaxMode, MaxAge: 120 * 60,
	})
	r.Success(w, map[string]any{"username": user.Username, "is_admin": true})
}

// ---------- 仪表盘 ----------

// GET /api/admin/dashboard
func (d *deps) handleAdminDashboard(w http.ResponseWriter, _ *http.Request) {
	count := func(q string, args ...any) int64 {
		var n int64
		d.gdb.Raw(q, args...).Scan(&n)
		return n
	}
	var revenue int64
	d.gdb.Raw("SELECT COALESCE(sum(`amount`), 0) FROM `orders` WHERE `status` = 'paid'").Scan(&revenue)
	r.Success(w, map[string]any{
		"user_count":    count("SELECT count(*) FROM `users` WHERE `deleted_at` IS NULL"),
		"photo_count":   count("SELECT count(*) FROM `photos` WHERE `deleted_at` IS NULL"),
		"album_count":   count("SELECT count(*) FROM `albums` WHERE `deleted_at` IS NULL"),
		"share_count":   count("SELECT count(*) FROM `shares`"),
		"order_count":   count("SELECT count(*) FROM `orders` WHERE `status` = 'paid'"),
		"ticket_open":   count("SELECT count(*) FROM `tickets` WHERE `status` = 'in_progress' AND `deleted_at` IS NULL"),
		"report_open":   count("SELECT count(*) FROM `reports` WHERE `status` = 'unhandled' AND `deleted_at` IS NULL"),
		"photo_size_kb": count("SELECT COALESCE(sum(`size`), 0) FROM `photos` WHERE `deleted_at` IS NULL"),
		"revenue_fen":   revenue,
		"timestamp":     time.Now().Unix(),
	})
}

// ---------- 用户管理 ----------

// GET /api/admin/users
func (d *deps) handleAdminUsers(w http.ResponseWriter, req *http.Request) {
	p := pagination.FromRequest(req)
	where := "u.`deleted_at` IS NULL"
	args := []any{}
	if q := p.Q; q != "" {
		where += " AND (u.`username` LIKE ? OR u.`name` LIKE ? OR u.`email` LIKE ?)"
		args = append(args, "%"+q+"%", "%"+q+"%", "%"+q+"%")
	}
	var total int64
	d.gdb.Raw("SELECT count(*) FROM `users` u WHERE "+where, args...).Scan(&total)
	var rows []struct {
		ID         int64
		Username   string
		Name       string
		Email      *string
		Phone      *string
		IsAdmin    bool
		Status     string
		CreatedAt  *time.Time
		PhotoCount int64
	}
	d.gdb.Raw(
		"SELECT u.`id`, u.`username`, u.`name`, u.`email`, u.`phone`, u.`is_admin`, u.`status`, u.`created_at`, "+
			"(SELECT count(*) FROM `photos` ph WHERE ph.user_id = u.id AND ph.deleted_at IS NULL) AS photo_count "+
			"FROM `users` u WHERE "+where+" ORDER BY u.`id` DESC LIMIT ? OFFSET ?",
		append(append([]any{}, args...), p.PerPage, p.Offset())...,
	).Scan(&rows)
	out := make([]map[string]any, 0, len(rows))
	for _, u := range rows {
		out = append(out, map[string]any{
			"id": u.ID, "username": u.Username, "name": u.Name, "email": u.Email,
			"phone": u.Phone, "is_admin": u.IsAdmin, "status": u.Status,
			"created_at": timePtrJSON(u.CreatedAt), "photo_count": u.PhotoCount,
		})
	}
	r.Success(w, pagination.New(out, total, p))
}

// PUT /api/admin/users/{id}
func (d *deps) handleAdminUserUpdate(w http.ResponseWriter, req *http.Request) {
	id := pathInt(req, "id")
	var body struct {
		Status   *string `json:"status"`
		IsAdmin  *bool   `json:"is_admin"`
		Password *string `json:"password"`
	}
	if err := readBody(req, &body); err != nil {
		r.Error(w, "请求体解析失败")
		return
	}
	updates := map[string]any{"updated_at": time.Now().UTC()}
	if body.Status != nil {
		if !validate.In(*body.Status, "normal", "frozen") {
			r.ErrorData(w, http.StatusUnprocessableEntity, "The given data was invalid.",
				map[string]any{"errors": map[string][]string{"status": {"状态 不存在。"}}})
			return
		}
		updates["status"] = *body.Status
	}
	if body.IsAdmin != nil {
		updates["is_admin"] = *body.IsAdmin
	}
	if body.Password != nil && *body.Password != "" {
		if len(*body.Password) < 8 {
			r.ErrorData(w, http.StatusUnprocessableEntity, "The given data was invalid.",
				map[string]any{"errors": map[string][]string{"password": {"密码 至少需要 8 个字符。"}}})
			return
		}
		hash, _ := bcrypt.GenerateFromPassword([]byte(*body.Password), 12)
		updates["password"] = string(hash)
	}
	if err := d.gdb.Model(&model.User{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		r.Error(w, "更新失败")
		return
	}
	r.Success(w, nil)
}

// DELETE /api/admin/users/{id}
func (d *deps) handleAdminUserDelete(w http.ResponseWriter, req *http.Request) {
	id := pathInt(req, "id")
	d.gdb.Model(&model.User{}).Where("id = ?", id).Update("deleted_at", time.Now().UTC())
	w.WriteHeader(http.StatusNoContent)
}

// ---------- 图片/相册管理 ----------

// GET /api/admin/photos
func (d *deps) handleAdminPhotos(w http.ResponseWriter, req *http.Request) {
	p := pagination.FromRequest(req)
	where := "ph.`deleted_at` IS NULL"
	args := []any{}
	if q := p.Q; q != "" {
		where += " AND (ph.`name` LIKE ? OR ph.`filename` LIKE ? OR ph.`pathname` LIKE ?)"
		args = append(args, "%"+q+"%", "%"+q+"%", "%"+q+"%")
	}
	if s := req.URL.Query().Get("status"); s != "" {
		where += " AND ph.`status` = ?"
		args = append(args, s)
	}
	var total int64
	d.gdb.Raw("SELECT count(*) FROM `photos` ph WHERE "+where, args...).Scan(&total)
	var rows []model.Photo
	d.gdb.Raw("SELECT ph.* FROM `photos` ph WHERE "+where+" ORDER BY ph.`id` DESC LIMIT ? OFFSET ?",
		append(append([]any{}, args...), p.PerPage, p.Offset())...).Scan(&rows)
	out := make([]map[string]any, 0, len(rows))
	for i := range rows {
		row := map[string]any{
			"id": rows[i].ID, "user_id": rows[i].UserID, "name": rows[i].Name,
			"filename": rows[i].Filename, "pathname": rows[i].Pathname,
			"extension": rows[i].Extension, "size": rows[i].Size,
			"width": rows[i].Width, "height": rows[i].Height,
			"is_public": rows[i].IsPublic, "status": rows[i].Status,
			"ip_address": rows[i].IPAddress, "created_at": timePtrJSON(rows[i].CreatedAt),
			"public_url":    photostore.PublicURL(d.gdb, d.cfg, &rows[i]),
			"thumbnail_url": photostore.ThumbnailURL(d.gdb, d.cfg, &rows[i]),
		}
		out = append(out, row)
	}
	r.Success(w, pagination.New(out, total, p))
}

// DELETE /api/admin/photos/{id}
func (d *deps) handleAdminPhotoDelete(w http.ResponseWriter, req *http.Request) {
	id := pathInt(req, "id")
	var photo model.Photo
	if err := d.gdb.First(&photo, id).Error; err != nil {
		r.Error(w, "图片不存在")
		return
	}
	_ = photostore.DeletePhoto(d.gdb, d.cfg, &photo)
	r.Success(w, nil)
}

// PUT /api/admin/photos/{id}/status（审核：normal/violation/pending）
func (d *deps) handleAdminPhotoStatus(w http.ResponseWriter, req *http.Request) {
	id := pathInt(req, "id")
	var body struct {
		Status string `json:"status"`
	}
	if err := readBody(req, &body); err != nil {
		r.Error(w, "请求体解析失败")
		return
	}
	if !validate.In(body.Status, "normal", "violation", "pending") {
		r.ErrorData(w, http.StatusUnprocessableEntity, "The given data was invalid.",
			map[string]any{"errors": map[string][]string{"status": {"状态 不存在。"}}})
		return
	}
	d.gdb.Model(&model.Photo{}).Where("id = ?", id).Updates(map[string]any{
		"status": body.Status, "updated_at": time.Now().UTC(),
	})
	if body.Status == "normal" {
		d.gdb.Model(&model.Violation{}).Where("photo_id = ? AND status = 'unhandled'", id).
			Updates(map[string]any{"status": "handled", "handled_at": time.Now().UTC()})
	}
	r.Success(w, nil)
}

// ---------- 公告/页面 ----------

// GET/POST /api/admin/notices, PUT/DELETE /api/admin/notices/{id}
func (d *deps) handleAdminNotices(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		p := pagination.FromRequest(req)
		var total int64
		d.gdb.Raw("SELECT count(*) FROM `notices` WHERE `deleted_at` IS NULL").Scan(&total)
		var rows []model.Notice
		d.gdb.Raw("SELECT * FROM `notices` WHERE `deleted_at` IS NULL ORDER BY `id` DESC LIMIT ? OFFSET ?",
			p.PerPage, p.Offset()).Scan(&rows)
		out := make([]model.Notice, 0, len(rows))
		out = append(out, rows...)
		r.Success(w, pagination.New(out, total, p))
	case http.MethodPost:
		var body struct {
			Title   string `json:"title"`
			Content string `json:"content"`
			Sort    int32  `json:"sort"`
		}
		if err := readBody(req, &body); err != nil {
			r.Error(w, "请求体解析失败")
			return
		}
		if body.Title == "" {
			r.ErrorData(w, http.StatusUnprocessableEntity, "The given data was invalid.",
				map[string]any{"errors": map[string][]string{"title": {"标题 不能为空。"}}})
			return
		}
		now := time.Now().UTC()
		n := model.Notice{Title: body.Title, Content: &body.Content, Sort: body.Sort, CreatedAt: &now, UpdatedAt: &now}
		if err := d.gdb.Create(&n).Error; err != nil {
			r.Error(w, "创建失败")
			return
		}
		r.Created(w, map[string]any{"id": n.ID})
	}
}

// PUT /api/admin/notices/{id}
func (d *deps) handleAdminNoticeUpdate(w http.ResponseWriter, req *http.Request) {
	id := pathInt(req, "id")
	var body map[string]any
	if err := readBody(req, &body); err != nil {
		r.Error(w, "请求体解析失败")
		return
	}
	updates := map[string]any{"updated_at": time.Now().UTC()}
	for _, k := range []string{"title", "content", "sort"} {
		if v, ok := body[k]; ok {
			updates[k] = v
		}
	}
	d.gdb.Model(&model.Notice{}).Where("id = ?", id).Updates(updates)
	r.Success(w, nil)
}

// DELETE /api/admin/notices/{id}
func (d *deps) handleAdminNoticeDelete(w http.ResponseWriter, req *http.Request) {
	d.gdb.Model(&model.Notice{}).Where("id = ?", pathInt(req, "id")).Update("deleted_at", time.Now().UTC())
	w.WriteHeader(http.StatusNoContent)
}

// GET/POST /api/admin/pages, PUT/DELETE /api/admin/pages/{id}
func (d *deps) handleAdminPages(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		p := pagination.FromRequest(req)
		var total int64
		d.gdb.Raw("SELECT count(*) FROM `pages`").Scan(&total)
		var rows []model.Page
		d.gdb.Raw("SELECT * FROM `pages` ORDER BY `id` DESC LIMIT ? OFFSET ?", p.PerPage, p.Offset()).Scan(&rows)
		r.Success(w, pagination.New(rows, total, p))
	case http.MethodPost:
		var body model.Page
		if err := readBody(req, &body); err != nil {
			r.Error(w, "请求体解析失败")
			return
		}
		now := time.Now().UTC()
		body.CreatedAt, body.UpdatedAt = &now, &now
		if body.Type == "" {
			body.Type = "internal"
		}
		if err := d.gdb.Create(&body).Error; err != nil {
			r.Error(w, "创建失败")
			return
		}
		r.Created(w, map[string]any{"id": body.ID})
	}
}

// PUT /api/admin/pages/{id}
func (d *deps) handleAdminPageUpdate(w http.ResponseWriter, req *http.Request) {
	id := pathInt(req, "id")
	var body map[string]any
	if err := readBody(req, &body); err != nil {
		r.Error(w, "请求体解析失败")
		return
	}
	updates := map[string]any{"updated_at": time.Now().UTC()}
	for _, k := range []string{"name", "title", "content", "slug", "url", "icon", "sort", "is_show", "type", "keywords", "description"} {
		if v, ok := body[k]; ok {
			updates[k] = v
		}
	}
	d.gdb.Model(&model.Page{}).Where("id = ?", id).Updates(updates)
	r.Success(w, nil)
}

// DELETE /api/admin/pages/{id}
func (d *deps) handleAdminPageDelete(w http.ResponseWriter, req *http.Request) {
	d.gdb.Delete(&model.Page{}, pathInt(req, "id"))
	w.WriteHeader(http.StatusNoContent)
}

// ---------- 套餐/优惠码 ----------

// GET/POST /api/admin/plans, PUT/DELETE /api/admin/plans/{id}（含价格阶梯）
func (d *deps) handleAdminPlans(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		var rows []model.Plan
		d.gdb.Where("deleted_at IS NULL").Order("sort ASC, id ASC").Find(&rows)
		out := make([]map[string]any, 0, len(rows))
		for i := range rows {
			var prices []model.PlanPrice
			d.gdb.Where("plan_id = ?", rows[i].ID).Order("price ASC").Find(&prices)
			out = append(out, map[string]any{
				"id": rows[i].ID, "type": rows[i].Type, "name": rows[i].Name,
				"intro": rows[i].Intro, "features": jsonOrNull(rows[i].Features),
				"badge": rows[i].Badge, "sort": rows[i].Sort, "is_up": rows[i].IsUp,
				"prices": prices,
			})
		}
		r.Success(w, map[string]any{"plans": out})
	case http.MethodPost:
		var body struct {
			Type     string            `json:"type"`
			Name     string            `json:"name"`
			Intro    *string           `json:"intro"`
			Features json.RawMessage   `json:"features"`
			Badge    string            `json:"badge"`
			Sort     int32             `json:"sort"`
			IsUp     bool              `json:"is_up"`
			Prices   []model.PlanPrice `json:"prices"`
			GroupID  *int64            `json:"group_id"`
			Capacity *float64          `json:"capacity"`
		}
		if err := readBody(req, &body); err != nil {
			r.Error(w, "请求体解析失败")
			return
		}
		if body.Type == "" {
			body.Type = "vip"
		}
		now := time.Now().UTC()
		plan := model.Plan{
			Type: body.Type, Name: body.Name, Intro: body.Intro,
			Features: featuresJSON(body.Features), Badge: body.Badge,
			Sort: body.Sort, IsUp: body.IsUp, CreatedAt: &now, UpdatedAt: &now,
		}
		if err := d.gdb.Create(&plan).Error; err != nil {
			r.Error(w, "创建失败")
			return
		}
		d.savePlanRelations(plan.ID, body.GroupID, body.Capacity, body.Prices)
		r.Created(w, map[string]any{"id": plan.ID})
	}
}

// PUT /api/admin/plans/{id}
func (d *deps) handleAdminPlanUpdate(w http.ResponseWriter, req *http.Request) {
	id := pathInt(req, "id")
	var body struct {
		Name     *string           `json:"name"`
		Intro    *string           `json:"intro"`
		Features json.RawMessage   `json:"features"`
		Badge    *string           `json:"badge"`
		Sort     *int32            `json:"sort"`
		IsUp     *bool             `json:"is_up"`
		Prices   []model.PlanPrice `json:"prices"`
		GroupID  *int64            `json:"group_id"`
		Capacity *float64          `json:"capacity"`
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
	if body.Badge != nil {
		updates["badge"] = *body.Badge
	}
	if body.Sort != nil {
		updates["sort"] = *body.Sort
	}
	if body.IsUp != nil {
		updates["is_up"] = *body.IsUp
	}
	if len(body.Features) > 0 {
		updates["features"] = string(body.Features)
	}
	d.gdb.Model(&model.Plan{}).Where("id = ?", id).Updates(updates)
	d.savePlanRelations(id, body.GroupID, body.Capacity, body.Prices)
	r.Success(w, nil)
}

func (d *deps) savePlanRelations(planID int64, groupID *int64, capacity *float64, prices []model.PlanPrice) {
	if groupID != nil {
		d.gdb.Exec("DELETE FROM `plan_groups` WHERE `plan_id` = ?", planID)
		if *groupID > 0 {
			d.gdb.Exec("INSERT INTO `plan_groups` (`plan_id`, `group_id`) VALUES (?, ?)", planID, *groupID)
		}
	}
	if capacity != nil {
		d.gdb.Exec("DELETE FROM `plan_capacities` WHERE `plan_id` = ?", planID)
		d.gdb.Exec("INSERT INTO `plan_capacities` (`plan_id`, `capacity`) VALUES (?, ?)", planID, *capacity)
	}
	if prices != nil {
		d.gdb.Where("plan_id = ?", planID).Delete(&model.PlanPrice{})
		now := time.Now().UTC()
		for _, pr := range prices {
			pr.PlanID = planID
			pr.CreatedAt, pr.UpdatedAt = &now, &now
			d.gdb.Create(&pr)
		}
	}
}

func featuresJSON(raw json.RawMessage) (out []byte) {
	if len(raw) == 0 {
		return []byte("null")
	}
	return raw
}

// DELETE /api/admin/plans/{id}
func (d *deps) handleAdminPlanDelete(w http.ResponseWriter, req *http.Request) {
	d.gdb.Model(&model.Plan{}).Where("id = ?", pathInt(req, "id")).Update("deleted_at", time.Now().UTC())
	w.WriteHeader(http.StatusNoContent)
}

// GET/POST /api/admin/coupons, PUT/DELETE /api/admin/coupons/{id}
func (d *deps) handleAdminCoupons(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		p := pagination.FromRequest(req)
		var total int64
		d.gdb.Raw("SELECT count(*) FROM `coupons` WHERE `deleted_at` IS NULL").Scan(&total)
		var rows []model.Coupon
		d.gdb.Raw("SELECT * FROM `coupons` WHERE `deleted_at` IS NULL ORDER BY `id` DESC LIMIT ? OFFSET ?",
			p.PerPage, p.Offset()).Scan(&rows)
		r.Success(w, pagination.New(rows, total, p))
	case http.MethodPost:
		var body model.Coupon
		if err := readBody(req, &body); err != nil {
			r.Error(w, "请求体解析失败")
			return
		}
		if body.Type == "" {
			body.Type = "direct"
		}
		if body.UsageLimit == 0 {
			body.UsageLimit = 1
		}
		if body.Code == "" {
			r.ErrorData(w, http.StatusUnprocessableEntity, "The given data was invalid.",
				map[string]any{"errors": map[string][]string{"code": {"券码 不能为空。"}}})
			return
		}
		now := time.Now().UTC()
		body.CreatedAt, body.UpdatedAt = &now, &now
		if err := d.gdb.Create(&body).Error; err != nil {
			r.Error(w, "券码已存在或创建失败")
			return
		}
		r.Created(w, map[string]any{"id": body.ID})
	}
}

// PUT /api/admin/coupons/{id}
func (d *deps) handleAdminCouponUpdate(w http.ResponseWriter, req *http.Request) {
	id := pathInt(req, "id")
	var body map[string]any
	if err := readBody(req, &body); err != nil {
		r.Error(w, "请求体解析失败")
		return
	}
	updates := map[string]any{"updated_at": time.Now().UTC()}
	for _, k := range []string{"type", "name", "code", "value", "usage_limit", "expired_at"} {
		if v, ok := body[k]; ok {
			updates[k] = v
		}
	}
	d.gdb.Model(&model.Coupon{}).Where("id = ?", id).Updates(updates)
	r.Success(w, nil)
}

// DELETE /api/admin/coupons/{id}
func (d *deps) handleAdminCouponDelete(w http.ResponseWriter, req *http.Request) {
	d.gdb.Model(&model.Coupon{}).Where("id = ?", pathInt(req, "id")).Update("deleted_at", time.Now().UTC())
	w.WriteHeader(http.StatusNoContent)
}

// ---------- 订单/举报/反馈/工单（管理视角） ----------

// GET /api/admin/albums
func (d *deps) handleAdminAlbums(w http.ResponseWriter, req *http.Request) {
	p := pagination.FromRequest(req)
	where := "a.`deleted_at` IS NULL"
	args := []any{}
	if q := p.Q; q != "" {
		where += " AND (a.`name` LIKE ? OR a.`intro` LIKE ?)"
		args = append(args, "%"+q+"%", "%"+q+"%")
	}
	var total int64
	d.gdb.Raw("SELECT count(*) FROM `albums` a WHERE "+where, args...).Scan(&total)
	var rows []struct {
		ID        int64
		UserID    *int64
		Name      string
		Intro     string
		IsPublic  bool
		PhotoCnt  int64
		CreatedAt *time.Time
	}
	d.gdb.Raw(
		"SELECT a.`id`, a.`user_id`, a.`name`, a.`intro`, a.`is_public`, a.`created_at`, "+
			"(SELECT count(*) FROM `album_photo` ap INNER JOIN `photos` ph ON ph.id = ap.photo_id AND ph.deleted_at IS NULL WHERE ap.album_id = a.id) AS photo_cnt "+
			"FROM `albums` a WHERE "+where+" ORDER BY a.`id` DESC LIMIT ? OFFSET ?",
		append(append([]any{}, args...), p.PerPage, p.Offset())...,
	).Scan(&rows)
	out := make([]map[string]any, 0, len(rows))
	for _, r := range rows {
		out = append(out, map[string]any{
			"id": r.ID, "user_id": r.UserID, "name": r.Name, "intro": r.Intro,
			"is_public": r.IsPublic, "photo_count": r.PhotoCnt, "created_at": timePtrJSON(r.CreatedAt),
		})
	}
	r.Success(w, pagination.New(out, total, p))
}

// DELETE /api/admin/albums/{id}
func (d *deps) handleAdminAlbumDelete(w http.ResponseWriter, req *http.Request) {
	d.gdb.Model(&model.Album{}).Where("id = ?", pathInt(req, "id")).Update("deleted_at", time.Now().UTC())
	w.WriteHeader(http.StatusNoContent)
}

// GET /api/admin/shares
func (d *deps) handleAdminShares(w http.ResponseWriter, req *http.Request) {
	p := pagination.FromRequest(req)
	where := "1=1"
	args := []any{}
	if q := p.Q; q != "" {
		where += " AND (`slug` LIKE ? OR `content` LIKE ?)"
		args = append(args, "%"+q+"%", "%"+q+"%")
	}
	var total int64
	d.gdb.Raw("SELECT count(*) FROM `shares` WHERE "+where, args...).Scan(&total)
	var rows []struct {
		ID        int64
		UserID    int64
		Type      string
		Slug      string
		ViewCount int64
		ExpiredAt *time.Time
		CreatedAt *time.Time
	}
	d.gdb.Raw("SELECT `id`, `user_id`, `type`, `slug`, `view_count`, `expired_at`, `created_at` FROM `shares` WHERE "+
		where+" ORDER BY `id` DESC LIMIT ? OFFSET ?",
		append(append([]any{}, args...), p.PerPage, p.Offset())...,
	).Scan(&rows)
	out := make([]map[string]any, 0, len(rows))
	for _, r := range rows {
		out = append(out, map[string]any{
			"id": r.ID, "user_id": r.UserID, "type": r.Type, "slug": r.Slug,
			"view_count": r.ViewCount, "expired_at": timePtrJSON(r.ExpiredAt), "created_at": timePtrJSON(r.CreatedAt),
		})
	}
	r.Success(w, pagination.New(out, total, p))
}

// DELETE /api/admin/shares/{id}
func (d *deps) handleAdminShareDelete(w http.ResponseWriter, req *http.Request) {
	d.gdb.Delete(&model.Share{}, pathInt(req, "id"))
	w.WriteHeader(http.StatusNoContent)
}

// GET /api/admin/violations
func (d *deps) handleAdminViolations(w http.ResponseWriter, req *http.Request) {
	p := pagination.FromRequest(req)
	where := "v.`deleted_at` IS NULL"
	args := []any{}
	if s := req.URL.Query().Get("status"); s != "" {
		where += " AND v.`status` = ?"
		args = append(args, s)
	}
	var total int64
	d.gdb.Raw("SELECT count(*) FROM `violations` v WHERE "+where, args...).Scan(&total)
	var rows []struct {
		ID        int64
		UserID    *int64
		PhotoID   *int64
		Reason    string
		Status    string
		HandledAt *time.Time
		CreatedAt *time.Time
	}
	d.gdb.Raw("SELECT v.* FROM `violations` v WHERE "+where+" ORDER BY v.`id` DESC LIMIT ? OFFSET ?",
		append(append([]any{}, args...), p.PerPage, p.Offset())...,
	).Scan(&rows)
	out := make([]map[string]any, 0, len(rows))
	for _, r := range rows {
		out = append(out, map[string]any{
			"id": r.ID, "user_id": r.UserID, "photo_id": r.PhotoID, "reason": r.Reason,
			"status": r.Status, "handled_at": timePtrJSON(r.HandledAt), "created_at": timePtrJSON(r.CreatedAt),
		})
	}
	r.Success(w, pagination.New(out, total, p))
}

// PUT /api/admin/violations/{id}（标记已处理）
func (d *deps) handleAdminViolationUpdate(w http.ResponseWriter, req *http.Request) {
	id := pathInt(req, "id")
	d.gdb.Model(&model.Violation{}).Where("id = ?", id).Updates(map[string]any{
		"status": "handled", "handled_at": time.Now().UTC(),
	})
	r.Success(w, nil)
}

// GET /api/admin/orders
func (d *deps) handleAdminOrders(w http.ResponseWriter, req *http.Request) {
	p := pagination.FromRequest(req)
	where := "1=1"
	args := []any{}
	if s := req.URL.Query().Get("status"); s != "" {
		where += " AND `status` = ?"
		args = append(args, s)
	}
	if q := p.Q; q != "" {
		where += " AND (`trade_no` LIKE ? OR `out_trade_no` LIKE ?)"
		args = append(args, "%"+q+"%", "%"+q+"%")
	}
	var total int64
	d.gdb.Raw("SELECT count(*) FROM `orders` WHERE "+where, args...).Scan(&total)
	var rows []model.Order
	d.gdb.Raw("SELECT * FROM `orders` WHERE "+where+" ORDER BY `id` DESC LIMIT ? OFFSET ?",
		append(append([]any{}, args...), p.PerPage, p.Offset())...).Scan(&rows)
	out := make([]map[string]any, 0, len(rows))
	for i := range rows {
		out = append(out, orderx.OrderResource(d.gdb, &rows[i]))
	}
	r.Success(w, pagination.New(out, total, p))
}

// GET /api/admin/reports
func (d *deps) handleAdminReports(w http.ResponseWriter, req *http.Request) {
	p := pagination.FromRequest(req)
	var total int64
	d.gdb.Raw("SELECT count(*) FROM `reports` WHERE `deleted_at` IS NULL").Scan(&total)
	var rows []model.Report
	d.gdb.Raw("SELECT * FROM `reports` WHERE `deleted_at` IS NULL ORDER BY `id` DESC LIMIT ? OFFSET ?",
		p.PerPage, p.Offset()).Scan(&rows)
	r.Success(w, pagination.New(rows, total, p))
}

// PUT /api/admin/reports/{id}（标记已处理）
func (d *deps) handleAdminReportUpdate(w http.ResponseWriter, req *http.Request) {
	id := pathInt(req, "id")
	d.gdb.Model(&model.Report{}).Where("id = ?", id).Updates(map[string]any{
		"status": "handled", "handled_at": time.Now().UTC(),
	})
	r.Success(w, nil)
}

// GET /api/admin/feedbacks
func (d *deps) handleAdminFeedbacks(w http.ResponseWriter, req *http.Request) {
	p := pagination.FromRequest(req)
	var total int64
	d.gdb.Raw("SELECT count(*) FROM `feedbacks`").Scan(&total)
	var rows []model.Feedback
	d.gdb.Raw("SELECT * FROM `feedbacks` ORDER BY `id` DESC LIMIT ? OFFSET ?", p.PerPage, p.Offset()).Scan(&rows)
	r.Success(w, pagination.New(rows, total, p))
}

// GET /api/admin/tickets + 回复管理
func (d *deps) handleAdminTickets(w http.ResponseWriter, req *http.Request) {
	p := pagination.FromRequest(req)
	where := "t.`deleted_at` IS NULL"
	args := []any{}
	if s := req.URL.Query().Get("status"); s != "" {
		where += " AND t.`status` = ?"
		args = append(args, s)
	}
	var total int64
	d.gdb.Raw("SELECT count(*) FROM `tickets` t WHERE "+where, args...).Scan(&total)
	var rows []struct {
		ID        int64
		IssueNo   string
		Title     string
		Level     string
		Status    string
		Username  string
		CreatedAt *time.Time
	}
	d.gdb.Raw(
		"SELECT t.`id`, t.`issue_no`, t.`title`, t.`level`, t.`status`, t.`created_at`, u.`username` AS username "+
			"FROM `tickets` t LEFT JOIN `users` u ON u.id = t.user_id "+
			"WHERE "+where+" ORDER BY t.`id` DESC LIMIT ? OFFSET ?",
		append(append([]any{}, args...), p.PerPage, p.Offset())...,
	).Scan(&rows)
	out := make([]map[string]any, 0, len(rows))
	for _, t := range rows {
		out = append(out, map[string]any{
			"id": t.ID, "issue_no": t.IssueNo, "title": t.Title, "level": t.Level,
			"status": t.Status, "username": t.Username, "created_at": timePtrJSON(t.CreatedAt),
		})
	}
	r.Success(w, pagination.New(out, total, p))
}

// GET /api/admin/tickets/{id}/replies
func (d *deps) handleAdminTicketReplies(w http.ResponseWriter, req *http.Request) {
	id := pathInt(req, "id")
	var rows []model.TicketReply
	d.gdb.Where("ticket_id = ? AND deleted_at IS NULL", id).Order("created_at ASC, id ASC").Find(&rows)
	r.Success(w, map[string]any{"data": rows})
}

// POST /api/admin/tickets/{id}/reply（管理员回复）
func (d *deps) handleAdminTicketReply(w http.ResponseWriter, req *http.Request) {
	u := authx.From(req).User
	id := pathInt(req, "id")
	var body struct {
		Content string `json:"content"`
	}
	if err := readBody(req, &body); err != nil || strings.TrimSpace(body.Content) == "" {
		r.Error(w, "回复内容不能为空")
		return
	}
	var t model.Ticket
	if err := d.gdb.First(&t, id).Error; err != nil {
		r.Error(w, "工单不存在")
		return
	}
	now := time.Now().UTC()
	reply := model.TicketReply{TicketID: t.ID, UserID: u.ID, Content: body.Content, IsNotify: true, CreatedAt: &now, UpdatedAt: &now}
	if err := d.gdb.Create(&reply).Error; err != nil {
		r.Error(w, "回复失败")
		return
	}
	r.Created(w, nil)
}

// ---------- 设置/储存/驱动/角色组 ----------

// GET /api/admin/settings
func (d *deps) handleAdminSettings(w http.ResponseWriter, _ *http.Request) {
	var rows []struct {
		Group   string
		Name    string
		Locked  bool
		Payload *string
	}
	d.gdb.Raw("SELECT `group`, `name`, `locked`, `payload` FROM `settings` ORDER BY `group` ASC, `name` ASC").Scan(&rows)
	out := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		var payload any
		if row.Payload != nil {
			_ = json.Unmarshal([]byte(*row.Payload), &payload)
		}
		out = append(out, map[string]any{"group": row.Group, "name": row.Name, "locked": row.Locked, "payload": payload})
	}
	r.Success(w, map[string]any{"settings": out})
}

// PUT /api/admin/settings
func (d *deps) handleAdminSettingsUpdate(w http.ResponseWriter, req *http.Request) {
	var body struct {
		Updates []struct {
			Group   string `json:"group"`
			Name    string `json:"name"`
			Payload any    `json:"payload"`
		} `json:"updates"`
	}
	if err := readBody(req, &body); err != nil {
		r.Error(w, "请求体解析失败")
		return
	}
	for _, u := range body.Updates {
		if err := setting.Set(d.gdb, u.Group, u.Name, u.Payload); err != nil {
			r.Error(w, "保存失败: "+err.Error())
			return
		}
	}
	r.Success(w, nil)
}

// GET/POST /api/admin/storages, PUT/DELETE /api/admin/storages/{id}
func (d *deps) handleAdminStorages(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		var rows []model.Storage
		d.gdb.Where("deleted_at IS NULL").Order("id ASC").Find(&rows)
		r.Success(w, map[string]any{"storages": rows})
	case http.MethodPost:
		var body model.Storage
		if err := readBody(req, &body); err != nil {
			r.Error(w, "请求体解析失败")
			return
		}
		now := time.Now().UTC()
		body.CreatedAt, body.UpdatedAt = &now, &now
		if err := d.gdb.Create(&body).Error; err != nil {
			r.Error(w, "创建失败")
			return
		}
		r.Created(w, map[string]any{"id": body.ID})
	}
}

// PUT /api/admin/storages/{id}
func (d *deps) handleAdminStorageUpdate(w http.ResponseWriter, req *http.Request) {
	id := pathInt(req, "id")
	var body map[string]any
	if err := readBody(req, &body); err != nil {
		r.Error(w, "请求体解析失败")
		return
	}
	updates := map[string]any{"updated_at": time.Now().UTC()}
	for _, k := range []string{"name", "intro", "prefix", "provider", "options"} {
		if v, ok := body[k]; ok {
			updates[k] = v
		}
	}
	d.gdb.Model(&model.Storage{}).Where("id = ?", id).Updates(updates)
	r.Success(w, nil)
}

// DELETE /api/admin/storages/{id}
func (d *deps) handleAdminStorageDelete(w http.ResponseWriter, req *http.Request) {
	d.gdb.Model(&model.Storage{}).Where("id = ?", pathInt(req, "id")).Update("deleted_at", time.Now().UTC())
	w.WriteHeader(http.StatusNoContent)
}

// GET/POST /api/admin/drivers, PUT/DELETE /api/admin/drivers/{id}
func (d *deps) handleAdminDrivers(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		var rows []model.Driver
		d.gdb.Where("deleted_at IS NULL").Order("id ASC").Find(&rows)
		r.Success(w, map[string]any{"drivers": rows})
	case http.MethodPost:
		var body model.Driver
		if err := readBody(req, &body); err != nil {
			r.Error(w, "请求体解析失败")
			return
		}
		now := time.Now().UTC()
		body.CreatedAt, body.UpdatedAt = &now, &now
		if err := d.gdb.Create(&body).Error; err != nil {
			r.Error(w, "创建失败")
			return
		}
		r.Created(w, map[string]any{"id": body.ID})
	}
}

// PUT /api/admin/drivers/{id}
func (d *deps) handleAdminDriverUpdate(w http.ResponseWriter, req *http.Request) {
	id := pathInt(req, "id")
	var body map[string]any
	if err := readBody(req, &body); err != nil {
		r.Error(w, "请求体解析失败")
		return
	}
	updates := map[string]any{"updated_at": time.Now().UTC()}
	for _, k := range []string{"type", "name", "intro", "options"} {
		if v, ok := body[k]; ok {
			updates[k] = v
		}
	}
	d.gdb.Model(&model.Driver{}).Where("id = ?", id).Updates(updates)
	r.Success(w, nil)
}

// DELETE /api/admin/drivers/{id}
func (d *deps) handleAdminDriverDelete(w http.ResponseWriter, req *http.Request) {
	d.gdb.Model(&model.Driver{}).Where("id = ?", pathInt(req, "id")).Update("deleted_at", time.Now().UTC())
	w.WriteHeader(http.StatusNoContent)
}

// GET /api/admin/groups + PUT options
func (d *deps) handleAdminGroups(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		var rows []model.Group
		d.gdb.Where("deleted_at IS NULL").Order("id ASC").Find(&rows)
		r.Success(w, map[string]any{"groups": rows})
	}
}

// PUT /api/admin/groups/{id}
func (d *deps) handleAdminGroupUpdate(w http.ResponseWriter, req *http.Request) {
	id := pathInt(req, "id")
	var body map[string]any
	if err := readBody(req, &body); err != nil {
		r.Error(w, "请求体解析失败")
		return
	}
	updates := map[string]any{"updated_at": time.Now().UTC()}
	for _, k := range []string{"name", "intro", "options", "is_default", "is_guest"} {
		if v, ok := body[k]; ok {
			updates[k] = v
		}
	}
	d.gdb.Model(&model.Group{}).Where("id = ?", id).Updates(updates)
	r.Success(w, nil)
}

// adminIDParam 兼容字符串 ID。
func adminIDParam(req *http.Request, name string) int64 { return pathInt(req, name) }
