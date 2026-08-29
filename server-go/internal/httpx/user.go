package httpx

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/chizw/ywty/server-go/internal/authx"
	"github.com/chizw/ywty/server-go/internal/mailx"
	"github.com/chizw/ywty/server-go/internal/model"
	"github.com/chizw/ywty/server-go/internal/pagination"
	"github.com/chizw/ywty/server-go/internal/support/r"
	"github.com/chizw/ywty/server-go/internal/validate"
	"github.com/go-chi/chi/v5"
)

// ---------- GET /api/v2/user/profile ----------

func (d *deps) handleProfile(w http.ResponseWriter, req *http.Request) {
	u := authx.From(req).User

	counts := func(table, col string) int64 {
		var n int64
		d.gdb.Raw("SELECT count(*) FROM `"+table+"` WHERE `"+col+"` = ? AND `deleted_at` IS NULL", u.ID).Scan(&n)
		return n
	}
	var usedStorage float64
	d.gdb.Raw("SELECT COALESCE(sum(`size`), 0) FROM `photos` WHERE `user_id` = ? AND `deleted_at` IS NULL", u.ID).Scan(&usedStorage)
	var totalStorage float64
	d.gdb.Raw(
		"SELECT COALESCE(sum(`capacity`), 0) FROM `user_capacities` WHERE `user_id` = ? AND `deleted_at` IS NULL "+
			"AND (`expired_at` > CURRENT_TIMESTAMP OR `expired_at` IS NULL)", u.ID,
	).Scan(&totalStorage)

	avatarURL := u.Avatar
	if avatarURL != "" && !strings.HasPrefix(avatarURL, "http") {
		avatarURL = d.cfg.AppURL + "/storage/" + strings.TrimPrefix(avatarURL, "/")
	}

	r.Success(w, map[string]any{
		"id":                u.ID,
		"avatar_url":        avatarURL,
		"name":              u.Name,
		"username":          u.Username,
		"email":             u.Email,
		"phone":             u.Phone,
		"tagline":           u.Tagline,
		"bio":               u.Bio,
		"url":               u.URL,
		"location":          u.Location,
		"company":           u.Company,
		"company_title":     u.CompanyTitle,
		"interests":         jsonOrNull(u.Interests),
		"socials":           jsonOrNull(u.Socials),
		"options":           jsonOrNull(u.Options),
		"is_admin":          u.IsAdmin,
		"country_code":      u.CountryCode,
		"login_ip":          u.LoginIP,
		"email_verified_at": timePtrJSON(u.EmailVerifiedAt),
		"phone_verified_at": timePtrJSON(u.PhoneVerifiedAt),
		"created_at":        timePtrJSON(u.CreatedAt),

		"group_count":    counts("user_groups", "user_id"),
		"capacity_count": counts("user_capacities", "user_id"),
		"order_count":    counts("orders", "user_id"),
		"share_count":    counts("shares", "user_id"),
		"ticket_count":   counts("tickets", "user_id"),
		"photo_count":    counts("photos", "user_id"),
		"album_count":    counts("albums", "user_id"),
		"used_storage":   round2(usedStorage),
		"total_storage":  round2(totalStorage),
	})
}

func timePtrJSON(t *time.Time) any {
	if t == nil {
		return nil
	}
	return t.Format(time.RFC3339)
}

func round2(f float64) float64 {
	return float64(int64(f*100+0.5)) / 100
}

// ---------- POST /api/v2/user/profile ----------

func (d *deps) handleUpdateProfile(w http.ResponseWriter, req *http.Request) {
	u := authx.From(req).User
	var body map[string]any
	if err := readBody(req, &body); err != nil {
		r.Error(w, "请求体解析失败")
		return
	}

	v := validate.New()
	updates := map[string]any{}

	if s, ok := body["username"].(string); ok && s != "" {
		if len(s) > 20 || !d.checkUnique("users", "username", s, u.ID) {
			v.Add("username", "用户名", "格式不正确或已被占用。")
		} else {
			updates["username"] = s
		}
	}
	strField := func(key, attr string, max int) {
		if s, ok := body[key].(string); ok {
			if len(s) > max {
				v.Add(key, attr, "过长。")
				return
			}
			updates[key] = s
		}
	}
	strField("name", "昵称", 20)
	strField("tagline", "个性签名", 60)
	strField("bio", "个人简介", 200)
	strField("company", "公司", 80)
	strField("company_title", "职位", 60)
	strField("location", "位置", 60)
	if s, ok := body["url"].(string); ok && s != "" {
		if !validate.URL(s) || len(s) > 200 {
			v.Add("url", "个人网站", "必须是合法的 URL。")
		} else {
			updates["url"] = s
		}
	}
	arrField := func(key, attr string) {
		if raw, ok := body[key]; ok {
			list, err := json.Marshal(raw)
			if err != nil {
				v.Add(key, attr, "格式不正确。")
				return
			}
			var arr []any
			_ = json.Unmarshal(list, &arr)
			if len(arr) > 6 {
				v.Add(key, attr, "最多 6 项。")
				return
			}
			updates[key] = string(list)
		}
	}
	arrField("interests", "兴趣爱好")
	arrField("socials", "社交账号")

	// 头像文件上传在 M2 图片管线中实现；此处忽略 avatar 文件字段

	if v.Fail() {
		v.Respond(w)
		return
	}
	if len(updates) > 0 {
		updates["updated_at"] = time.Now().UTC()
		if err := d.gdb.Model(&model.User{}).Where("id = ?", u.ID).Updates(updates).Error; err != nil {
			r.Error(w, "fail")
			return
		}
	}
	r.Success(w, nil)
}

// ---------- POST /api/v2/user/setting ----------

func (d *deps) handleUpdateSetting(w http.ResponseWriter, req *http.Request) {
	u := authx.From(req).User
	var body map[string]any
	if err := readBody(req, &body); err != nil {
		r.Error(w, "请求体解析失败")
		return
	}
	allowed := map[string]bool{
		"language": true, "show_original_photos": true, "encode_copied_url": true,
		"auto_upload_after_select": true, "upload_button_action": true, "default_storage_id": true,
	}
	merged := map[string]any{}
	if len(u.Options) > 0 {
		_ = json.Unmarshal(u.Options, &merged)
	}
	patched := 0
	for k, val := range body {
		if allowed[k] {
			merged[k] = val
			patched++
		}
	}
	if patched > 0 {
		opts, _ := json.Marshal(merged)
		if err := d.gdb.Model(&model.User{}).Where("id = ?", u.ID).
			Updates(map[string]any{"options": string(opts), "updated_at": time.Now().UTC()}).Error; err != nil {
			r.Error(w, "fail")
			return
		}
	}
	r.Success(w, nil)
}

// ---------- POST /api/v2/user/bind_email ----------

func (d *deps) handleBindEmail(w http.ResponseWriter, req *http.Request) {
	u := authx.From(req).User
	var body struct {
		Email string `json:"email"`
		Code  string `json:"code"`
	}
	if err := readBody(req, &body); err != nil {
		r.Error(w, "请求体解析失败")
		return
	}
	body.Email = strings.ToLower(strings.TrimSpace(body.Email))
	v := validate.New()
	if !validate.Required(body.Email) || !validate.Email(body.Email) || !d.checkUnique("users", "email", body.Email, u.ID) {
		v.Add("email", "邮箱", "必须是合法且未被占用的邮箱。")
	}
	if !mailx.VerifyCode(d.cache, mailx.CodeKey("bind", body.Email), body.Code) {
		v.Add("code", "验证码", "Invalid verification code.")
	}
	if v.Fail() {
		v.Respond(w)
		return
	}
	now := time.Now().UTC()
	if err := d.gdb.Model(&model.User{}).Where("id = ?", u.ID).
		Updates(map[string]any{"email": body.Email, "email_verified_at": now, "updated_at": now}).Error; err != nil {
		r.Error(w, "绑定失败，请稍后重试")
		return
	}
	r.Created(w, nil)
}

// ---------- POST /api/v2/user/bind_phone ----------

func (d *deps) handleBindPhone(w http.ResponseWriter, req *http.Request) {
	u := authx.From(req).User
	var body struct {
		Phone       string `json:"phone"`
		CountryCode string `json:"country_code"`
		Code        string `json:"code"`
	}
	if err := readBody(req, &body); err != nil {
		r.Error(w, "请求体解析失败")
		return
	}
	if body.CountryCode == "" {
		body.CountryCode = "cn"
	}
	v := validate.New()
	if !validate.Required(body.Phone) || !d.checkUnique("users", "phone", body.Phone, u.ID) {
		v.Add("phone", "手机号", "必须是合法且未被占用的手机号。")
	}
	if !smsVerify(d.cache, "bind", body.Phone, body.Code) {
		v.Add("code", "验证码", "Invalid verification code.")
	}
	if v.Fail() {
		v.Respond(w)
		return
	}
	now := time.Now().UTC()
	cc := body.CountryCode
	if err := d.gdb.Model(&model.User{}).Where("id = ?", u.ID).
		Updates(map[string]any{"phone": body.Phone, "country_code": cc, "phone_verified_at": now, "updated_at": now}).Error; err != nil {
		r.Error(w, "绑定失败，请稍后重试")
		return
	}
	r.Created(w, nil)
}

// ---------- GET /api/v2/user/groups ----------

func (d *deps) handleUserGroups(w http.ResponseWriter, req *http.Request) {
	u := authx.From(req).User
	p := pagination.FromRequest(req)

	var flat []struct {
		ID         int64
		From       string
		ExpiredAt  *time.Time
		CreatedAt  *time.Time
		GroupID    int64
		GroupName  string
		GroupIntro string
		IsDefault  bool
		IsGuest    bool
		Options    *string
	}
	var total int64
	d.gdb.Raw(
		"SELECT count(*) FROM `user_groups` ug INNER JOIN `groups` g ON g.id = ug.group_id AND g.deleted_at IS NULL "+
			"WHERE ug.user_id = ? AND ug.deleted_at IS NULL AND (ug.expired_at > CURRENT_TIMESTAMP OR ug.expired_at IS NULL)", u.ID,
	).Scan(&total)
	d.gdb.Raw(
		"SELECT ug.`id`, ug.`from`, ug.`expired_at`, ug.`created_at`, "+
			"g.`id` AS group_id, g.`name` AS group_name, g.`intro` AS group_intro, "+
			"g.`is_default` AS is_default, g.`is_guest` AS is_guest, g.`options` AS options "+
			"FROM `user_groups` ug INNER JOIN `groups` g ON g.id = ug.group_id AND g.deleted_at IS NULL "+
			"WHERE ug.user_id = ? AND ug.deleted_at IS NULL AND (ug.expired_at > CURRENT_TIMESTAMP OR ug.expired_at IS NULL) "+
			"ORDER BY ug.created_at DESC LIMIT ? OFFSET ?", u.ID, p.PerPage, p.Offset(),
	).Scan(&flat)

	out := make([]map[string]any, 0, len(flat))
	for _, r := range flat {
		var options any
		if r.Options != nil {
			options = jsonOrNullStr(*r.Options)
		}
		out = append(out, map[string]any{
			"id": r.ID, "from": r.From,
			"expired_at": timePtrJSON(r.ExpiredAt), "created_at": timePtrJSON(r.CreatedAt),
			"group": map[string]any{
				"id": r.GroupID, "name": r.GroupName, "intro": r.GroupIntro,
				"is_default": r.IsDefault, "is_guest": r.IsGuest, "options": options,
			},
		})
	}
	r.Success(w, pagination.New(out, total, p))
}

// DELETE /api/v2/user/groups/{id}
func (d *deps) handleUserGroupDestroy(w http.ResponseWriter, req *http.Request) {
	u := authx.From(req).User
	id := pathInt(req, "id")
	res := d.gdb.Exec(
		"UPDATE `user_groups` SET `deleted_at` = CURRENT_TIMESTAMP WHERE `id` = ? AND `user_id` = ? AND `deleted_at` IS NULL "+
			"AND (`expired_at` > CURRENT_TIMESTAMP OR `expired_at` IS NULL)", id, u.ID,
	)
	if res.Error != nil || res.RowsAffected == 0 {
		r.Error(w, "记录不存在")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ---------- GET /api/v2/user/capacities ----------

func (d *deps) handleUserCapacities(w http.ResponseWriter, req *http.Request) {
	u := authx.From(req).User
	p := pagination.FromRequest(req)

	type row struct {
		ID        int64      `json:"id"`
		From      string     `json:"from"`
		Capacity  float64    `json:"capacity"`
		ExpiredAt *time.Time `json:"expired_at"`
		CreatedAt *time.Time `json:"created_at"`
	}
	var total int64
	d.gdb.Raw(
		"SELECT count(*) FROM `user_capacities` WHERE `user_id` = ? AND `deleted_at` IS NULL "+
			"AND (`expired_at` > CURRENT_TIMESTAMP OR `expired_at` IS NULL)", u.ID,
	).Scan(&total)
	var rows []row
	d.gdb.Raw(
		"SELECT `id`, `from`, `capacity`, `expired_at`, `created_at` FROM `user_capacities` "+
			"WHERE `user_id` = ? AND `deleted_at` IS NULL AND (`expired_at` > CURRENT_TIMESTAMP OR `expired_at` IS NULL) "+
			"ORDER BY `created_at` DESC LIMIT ? OFFSET ?", u.ID, p.PerPage, p.Offset(),
	).Scan(&rows)
	r.Success(w, pagination.New(rows, total, p))
}

// DELETE /api/v2/user/capacities/{id}
func (d *deps) handleUserCapacityDestroy(w http.ResponseWriter, req *http.Request) {
	u := authx.From(req).User
	id := pathInt(req, "id")
	res := d.gdb.Exec(
		"UPDATE `user_capacities` SET `deleted_at` = CURRENT_TIMESTAMP WHERE `id` = ? AND `user_id` = ? AND `deleted_at` IS NULL "+
			"AND (`expired_at` > CURRENT_TIMESTAMP OR `expired_at` IS NULL)", id, u.ID,
	)
	if res.Error != nil || res.RowsAffected == 0 {
		r.Error(w, "记录不存在")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ---------- GET /api/v2/user/tokens ----------

func (d *deps) handleTokensIndex(w http.ResponseWriter, req *http.Request) {
	u := authx.From(req).User
	p := pagination.FromRequest(req)

	order := "`created_at` DESC" // 默认 latest
	if q := p.Q; q != "" {
		for _, part := range strings.Fields(q) {
			if strings.HasPrefix(part, "sort:") {
				switch part {
				case "sort:oldest":
					order = "`created_at` ASC"
				case "sort:latest":
					order = "`created_at` DESC"
				case "sort:last_used_at:ascend":
					order = "`last_used_at` ASC"
				case "sort:last_used_at:descend":
					order = "`last_used_at` DESC"
				case "sort:expires_at:ascend":
					order = "`expires_at` ASC"
				case "sort:expires_at:descend":
					order = "`expires_at` DESC"
				case "sort:created_at:ascend":
					order = "`created_at` ASC"
				case "sort:created_at:descend":
					order = "`created_at` DESC"
				}
			}
		}
	}

	var flat []struct {
		ID         int64
		Name       string
		LastUsedAt *time.Time
		ExpiresAt  *time.Time
		CreatedAt  *time.Time
		Abilities  *string
	}
	var total int64
	d.gdb.Raw(
		"SELECT count(*) FROM `personal_access_tokens` WHERE `tokenable_type` = ? AND `tokenable_id` = ?",
		authx.TokenableType, u.ID,
	).Scan(&total)

	like := "%" + p.Q + "%"
	d.gdb.Raw(
		"SELECT `id`, `name`, `last_used_at`, `expires_at`, `created_at`, `abilities` FROM `personal_access_tokens` "+
			"WHERE `tokenable_type` = ? AND `tokenable_id` = ? AND `name` LIKE ? "+
			"ORDER BY "+order+" LIMIT ? OFFSET ?", authx.TokenableType, u.ID, like, p.PerPage, p.Offset(),
	).Scan(&flat)

	out := make([]map[string]any, 0, len(flat))
	for _, t := range flat {
		abilities := json.RawMessage("null")
		if t.Abilities != nil && *t.Abilities != "" {
			abilities = json.RawMessage(*t.Abilities)
		}
		out = append(out, map[string]any{
			"id": t.ID, "name": t.Name,
			"last_used_at": timePtrJSON(t.LastUsedAt), "expires_at": timePtrJSON(t.ExpiresAt),
			"created_at": timePtrJSON(t.CreatedAt), "abilities": abilities,
		})
	}
	r.Success(w, pagination.New(out, total, p))
}

// POST /api/v2/user/tokens
func (d *deps) handleTokensStore(w http.ResponseWriter, req *http.Request) {
	u := authx.From(req).User
	var body struct {
		Name      string   `json:"name"`
		ExpiresAt *string  `json:"expires_at"`
		Abilities []string `json:"abilities"`
	}
	if err := readBody(req, &body); err != nil {
		r.Error(w, "请求体解析失败")
		return
	}
	v := validate.New()
	if !validate.Required(body.Name) || len(body.Name) > 40 {
		v.Add("name", "名称", "不能为空且最长 40 字符。")
	}
	for _, a := range body.Abilities {
		if a == "*" {
			continue
		}
		if !isValidAbility(a) {
			v.Add("abilities", "权限项", "包含无效权限 "+a+"。")
		}
	}
	if v.Fail() {
		v.Respond(w)
		return
	}

	var expiresAt *time.Time
	if body.ExpiresAt != nil && *body.ExpiresAt != "" {
		for _, layout := range []string{time.RFC3339, "2006-01-02 15:04:05", "2006-01-02"} {
			if t, err := time.Parse(layout, *body.ExpiresAt); err == nil {
				expiresAt = &t
				break
			}
		}
	}
	abilities := body.Abilities
	if len(abilities) == 0 {
		abilities = []string{"*"}
	}
	plain, err := authx.CreateToken(d.gdb, u.ID, body.Name, abilities, expiresAt)
	if err != nil {
		r.Error(w, "创建令牌失败")
		return
	}
	var exp any
	if expiresAt != nil {
		exp = expiresAt.Format(time.RFC3339)
	}
	r.Success(w, map[string]any{
		"name": body.Name, "token": plain, "expires_at": exp, "abilities": abilities,
	})
}

func isValidAbility(a string) bool {
	for _, p := range []string{
		"user:profile:read", "user:profile:write", "user:token:read", "user:token:write",
		"user:group:read", "user:group:write", "user:capacity:read", "user:capacity:write",
		"user:album:read", "user:album:write", "user:photo:read", "user:photo:write",
		"user:share:read", "user:share:write", "user:ticket:read", "user:ticket:write",
		"user:order:read", "user:order:write", "oauth:read", "oauth:write",
		"explore:read", "explore:write", "upload:write", "basic",
	} {
		if a == p {
			return true
		}
	}
	return false
}

// DELETE /api/v2/user/tokens/{id}
func (d *deps) handleTokensDestroy(w http.ResponseWriter, req *http.Request) {
	u := authx.From(req).User
	id := pathInt(req, "id")
	res := d.gdb.Where("id = ? AND tokenable_id = ?", id, u.ID).Delete(&model.PersonalAccessToken{})
	if res.Error != nil || res.RowsAffected == 0 {
		r.Error(w, "记录不存在")
		return
	}
	r.Success(w, nil)
}

// GET /api/v2/user/tokens/permissions
func (d *deps) handleTokensPermissions(w http.ResponseWriter, req *http.Request) {
	tok := authx.From(req).Token
	if tok == nil {
		// 会话认证（网页登录态）没有当前令牌
		r.Success(w, nil)
		return
	}
	var exp any
	if tok.ExpiresAt != nil {
		exp = tok.ExpiresAt.Format(time.RFC3339)
	}
	var lastUsed any
	if tok.LastUsedAt != nil {
		lastUsed = tok.LastUsedAt.Format(time.RFC3339)
	}
	r.Success(w, map[string]any{
		"token_name":   tok.Name,
		"abilities":    tok.Abilities,
		"last_used_at": lastUsed,
		"expires_at":   exp,
	})
}

// ---------- 工具 ----------

func pathInt(req *http.Request, name string) int64 {
	var n int64
	_, _ = fmt.Sscanf(chi.URLParam(req, name), "%d", &n)
	return n
}

func jsonOrNullStr(s string) any {
	if s == "" || s == "null" {
		return nil
	}
	var v any
	if json.Unmarshal([]byte(s), &v) != nil {
		return nil
	}
	return v
}
