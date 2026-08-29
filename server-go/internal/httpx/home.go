package httpx

import (
	"encoding/json"
	"net/http"
	"time"

	"strings"

	"github.com/chizw/ywty/server-go/internal/authx"
	"github.com/chizw/ywty/server-go/internal/countries"
	"github.com/chizw/ywty/server-go/internal/db/types"
	"github.com/chizw/ywty/server-go/internal/model"
	"github.com/chizw/ywty/server-go/internal/setting"
	"github.com/chizw/ywty/server-go/internal/support/r"
	"github.com/chizw/ywty/server-go/internal/validate"
)

// ---------- GET /api/v2/configs ----------

func (d *deps) handleConfigs(w http.ResponseWriter, req *http.Request) {
	ctx := authx.From(req)

	var photoCount, photoSize int64
	d.gdb.Raw("SELECT count(*) FROM `photos` WHERE `deleted_at` IS NULL").Scan(&photoCount)
	d.gdb.Raw("SELECT COALESCE(sum(`size`), 0) FROM `photos` WHERE `deleted_at` IS NULL").Scan(&photoSize)

	s := func(group, name string) string {
		v, err := settingString(d.gdb, group, name)
		if err != nil {
			return ""
		}
		return v
	}
	b := func(group, name string) bool {
		v, _ := settingBoolP(d.gdb, group, name)
		return v
	}

	resp := map[string]any{
		"app": map[string]any{
			"name":                s(setting.GroupApp, "name"),
			"url":                 d.cfg.AppURL,
			"debug":               false,
			"icp_no":              s(setting.GroupApp, "icp_no"),
			"currency":            s(setting.GroupApp, "currency"),
			"enable_registration": b(setting.GroupApp, "enable_registration"),
			"guest_upload":        b(setting.GroupApp, "guest_upload"),
			"user_email_verify":   b(setting.GroupApp, "user_email_verify"),
			"user_phone_verify":   b(setting.GroupApp, "user_phone_verify"),
			"enable_site":         b(setting.GroupApp, "enable_site"),
			"enable_explore":      b(setting.GroupApp, "enable_explore"),
			"timestamp":           time.Now().Unix(),
			"is_logged_in":        ctx.User != nil,
			"photo_count":         photoCount,
			"photo_size":          photoSize,
			"countries":           countries.All(),
			"socialites":          d.socialiteDrivers(),
		},
		"site": map[string]any{
			"title":                          s(setting.GroupSite, "title"),
			"subtitle":                       s(setting.GroupSite, "subtitle"),
			"homepage_title":                 s(setting.GroupSite, "homepage_title"),
			"homepage_description":           s(setting.GroupSite, "homepage_description"),
			"notice":                         s(setting.GroupSite, "notice"),
			"custom_css":                     s(setting.GroupSite, "custom_css"),
			"custom_js":                      s(setting.GroupSite, "custom_js"),
			"homepage_background_image_url":  s(setting.GroupSite, "homepage_background_image_url"),
			"homepage_background_images":     d.storageURLs(s(setting.GroupSite, "homepage_background_images")),
			"auth_page_background_image_url": s(setting.GroupSite, "auth_page_background_image_url"),
			"auth_page_background_images":    d.storageURLs(s(setting.GroupSite, "auth_page_background_images")),
		},
	}
	r.Success(w, resp)
}

// socialiteDrivers 已配置的 OAuth 登录驱动（M5 实现具体流程）。
func (d *deps) socialiteDrivers() []map[string]any {
	type row struct {
		ID      int64
		Name    string
		Intro   string
		Options *string
	}
	var rows []row
	d.gdb.Raw(
		"SELECT `id`, `name`, `intro`, `options` FROM `drivers` WHERE `type` = 'socialite' AND `deleted_at` IS NULL",
	).Scan(&rows)
	out := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		provider := ""
		if row.Options != nil && *row.Options != "" {
			var opts map[string]any
			if json.Unmarshal([]byte(*row.Options), &opts) == nil {
				if p, ok := opts["provider"].(string); ok {
					provider = p
				}
			}
		}
		out = append(out, map[string]any{"id": row.ID, "name": row.Name, "intro": row.Intro, "provider": provider})
	}
	return out
}

// storageURLs 把站点设置里的背景图路径映射为可访问 URL（对齐 Storage::url 行为的近似实现）。
func (d *deps) storageURLs(raw string) []string {
	if raw == "" || raw == "null" {
		return []string{}
	}
	var list []string
	if json.Unmarshal([]byte(raw), &list) != nil {
		return []string{}
	}
	out := make([]string, 0, len(list))
	for _, v := range list {
		if v == "" {
			continue
		}
		if strings.HasPrefix(v, "http://") || strings.HasPrefix(v, "https://") {
			out = append(out, v)
			continue
		}
		out = append(out, d.cfg.AppURL+"/storage/"+strings.TrimPrefix(v, "/"))
	}
	return out
}

// ---------- GET /api/v2/group ----------

func (d *deps) handleGroup(w http.ResponseWriter, req *http.Request) {
	ctx := authx.From(req)
	if ctx.Group == nil {
		r.Error(w, "系统未初始化角色组")
		return
	}
	g := ctx.Group

	storages := []map[string]any{}
	var srows []struct {
		ID       int64
		Name     string
		Intro    string
		Provider string
	}
	d.gdb.Raw(
		"SELECT s.`id`, s.`name`, s.`intro`, s.`provider` FROM `storages` s "+
			"INNER JOIN `group_storage` gs ON gs.storage_id = s.id AND gs.group_id = ? "+
			"WHERE s.`deleted_at` IS NULL ORDER BY gs.`sort` ASC, s.`id` ASC", g.ID,
	).Scan(&srows)
	for _, s := range srows {
		storages = append(storages, map[string]any{"id": s.ID, "name": s.Name, "intro": s.Intro, "provider": s.Provider})
	}

	// 支付驱动（M4 完整实现 channels/methods）
	payments := []map[string]any{}
	var prows []struct {
		ID      int64
		Name    string
		Intro   string
		Options *string
	}
	d.gdb.Raw(
		"SELECT dr.`id`, dr.`name`, dr.`intro`, dr.`options` FROM `drivers` dr "+
			"INNER JOIN `group_driver` gd ON gd.driver_id = dr.id AND gd.type = 'payment' "+
			"WHERE gd.group_id = ? AND dr.`type` = 'payment' AND dr.`deleted_at` IS NULL "+
			"ORDER BY gd.`sort` ASC", g.ID,
	).Scan(&prows)
	for _, p := range prows {
		platform := ""
		channels := []any{}
		methods := []any{}
		if p.Options != nil && *p.Options != "" {
			var opts map[string]any
			if json.Unmarshal([]byte(*p.Options), &opts) == nil {
				if v, ok := opts["provider"].(string); ok {
					platform = v
				}
				if v, ok := opts["channels"].([]any); ok {
					channels = v
				}
			}
		}
		payments = append(payments, map[string]any{
			"id": p.ID, "name": p.Name, "intro": p.Intro,
			"platform": platform, "channels": channels, "methods": methods,
		})
	}

	r.Success(w, map[string]any{
		"group": map[string]any{
			"id": g.ID, "name": g.Name, "intro": g.Intro,
			"is_default": g.IsDefault, "is_guest": g.IsGuest, "options": jsonOrNull(g.Options),
		},
		"storages": storages,
		"payments": payments,
	})
}

func jsonOrNull(j types.JSON) any {
	if len(j) == 0 || string(j) == "null" {
		return nil
	}
	var v any
	if json.Unmarshal([]byte(j), &v) != nil {
		return nil
	}
	return v
}

// ---------- GET /api/v2/captcha ----------

func (d *deps) handleCaptcha(w http.ResponseWriter, req *http.Request) {
	d.captcha.Handler(w, req)
}

// ---------- POST /api/v2/mail/send、/api/v2/sms/send ----------

func (d *deps) handleMailCodeSend(w http.ResponseWriter, req *http.Request) {
	var body struct {
		Event   string `json:"event"`
		Email   string `json:"email"`
		Captcha string `json:"captcha"`
		Key     string `json:"captcha_key"`
	}
	if err := readBody(req, &body); err != nil {
		r.Error(w, "请求体解析失败")
		return
	}
	ctx := authx.From(req)

	// bind/verify 事件需要登录（对齐 EmailCodeSendRequest::authorize）
	if (body.Event == "bind" || body.Event == "verify") && ctx.User == nil {
		r.ErrorWithCode(w, http.StatusUnauthorized, "Unauthenticated.")
		return
	}

	v := validate.New()
	if !validate.In(body.Event, "register", "bind", "forget_password", "verify") {
		v.Add("event", "事件", "不存在。")
	}
	if !validate.Required(body.Email) || !validate.Email(body.Email) {
		v.Add("email", "邮箱", "必须是合法的邮箱。")
	}
	if !d.captcha.Verify(body.Key, body.Captcha) {
		v.Add("captcha", "图形验证码", "验证码错误。")
	}
	if v.Fail() {
		v.Respond(w)
		return
	}

	// 每 IP 120 秒内最多 3 次
	countKey := "mail_code_count:" + authx.ClientIP(req)
	if n, _ := d.cache.GetInt(countKey); n >= 3 {
		r.Error(w, "发送过于频繁，请稍后再试")
		return
	}
	d.cache.Increment(countKey, 120)

	siteName := d.cfg.AppName
	if t, err := settingString(d.gdb, setting.GroupSite, "title"); err == nil && t != "" {
		siteName = t
	}
	_ = d.queue.Dispatch("send_code_mail", map[string]any{
		"event": body.Event, "email": strings.ToLower(body.Email), "site_name": siteName,
	})
	r.Created(w, nil)
}

func (d *deps) handleSmsCodeSend(w http.ResponseWriter, req *http.Request) {
	var body struct {
		Event       string `json:"event"`
		Phone       string `json:"phone"`
		CountryCode string `json:"country_code"`
		Captcha     string `json:"captcha"`
		Key         string `json:"captcha_key"`
	}
	if err := readBody(req, &body); err != nil {
		r.Error(w, "请求体解析失败")
		return
	}
	ctx := authx.From(req)
	if (body.Event == "bind" || body.Event == "verify") && ctx.User == nil {
		r.ErrorWithCode(w, http.StatusUnauthorized, "Unauthenticated.")
		return
	}

	v := validate.New()
	if !validate.In(body.Event, "register", "bind", "forget_password", "verify") {
		v.Add("event", "事件", "不存在。")
	}
	if !validate.Required(body.Phone) {
		v.Add("phone", "手机号", "不能为空。")
	}
	if !d.captcha.Verify(body.Key, body.Captcha) {
		v.Add("captcha", "图形验证码", "验证码错误。")
	}
	if v.Fail() {
		v.Respond(w)
		return
	}

	countKey := "sms_code_count:" + authx.ClientIP(req)
	if n, _ := d.cache.GetInt(countKey); n >= 3 {
		r.Error(w, "发送过于频繁，请稍后再试")
		return
	}
	d.cache.Increment(countKey, 120)

	_ = d.queue.Dispatch("send_code_sms", map[string]any{
		"event": body.Event, "phone": body.Phone,
		"country_code": body.CountryCode,
	})
	r.Created(w, nil)
}

// ---------- GET /api/v2/token_permissions ----------

func (d *deps) handleTokenPermissions(w http.ResponseWriter, _ *http.Request) {
	type permItem struct {
		Value    string `json:"value"`
		Label    string `json:"label"`
		Detail   string `json:"detail"`
		Category string `json:"category"`
	}
	descriptions := map[authx.Permission]string{
		authx.PermUserProfileRead:   "读取用户资料",
		authx.PermUserProfileWrite:  "更新用户资料",
		authx.PermUserTokenRead:     "查看个人令牌",
		authx.PermUserTokenWrite:    "管理个人令牌",
		authx.PermUserGroupRead:     "查看用户角色组",
		authx.PermUserGroupWrite:    "管理用户角色组",
		authx.PermUserCapacityRead:  "查看用户容量",
		authx.PermUserCapacityWrite: "管理用户容量",
		authx.PermUserAlbumRead:     "查看相册",
		authx.PermUserAlbumWrite:    "管理相册",
		authx.PermUserPhotoRead:     "查看照片",
		authx.PermUserPhotoWrite:    "管理照片",
		authx.PermUserShareRead:     "查看分享",
		authx.PermUserShareWrite:    "管理分享",
		authx.PermUserTicketRead:    "查看工单",
		authx.PermUserTicketWrite:   "管理工单",
		authx.PermUserOrderRead:     "查看订单",
		authx.PermUserOrderWrite:    "管理订单",
		authx.PermOAuthRead:         "查看OAuth信息",
		authx.PermOAuthWrite:        "管理OAuth绑定",
		authx.PermExploreRead:       "浏览广场内容",
		authx.PermExploreWrite:      "管理广场互动",
		authx.PermUploadWrite:       "上传图片",
		authx.PermBasic:             "基础权限",
	}
	order := []authx.Permission{
		authx.PermUserProfileRead, authx.PermUserProfileWrite,
		authx.PermUserTokenRead, authx.PermUserTokenWrite,
		authx.PermUserGroupRead, authx.PermUserGroupWrite,
		authx.PermUserCapacityRead, authx.PermUserCapacityWrite,
		authx.PermUserAlbumRead, authx.PermUserAlbumWrite,
		authx.PermUserPhotoRead, authx.PermUserPhotoWrite,
		authx.PermUserShareRead, authx.PermUserShareWrite,
		authx.PermUserTicketRead, authx.PermUserTicketWrite,
		authx.PermUserOrderRead, authx.PermUserOrderWrite,
		authx.PermOAuthRead, authx.PermOAuthWrite,
		authx.PermExploreRead, authx.PermExploreWrite,
		authx.PermUploadWrite, authx.PermBasic,
	}
	perms := make([]permItem, 0, len(order))
	for _, p := range order {
		perms = append(perms, permItem{
			Value:    string(p),
			Label:    descriptions[p],
			Detail:   descriptions[p],
			Category: string(strings.SplitN(string(p), ":", 2)[0]),
		})
	}

	groups := map[string]map[string][]string{
		"user": {
			"用户资料": {string(authx.PermUserProfileRead), string(authx.PermUserProfileWrite)},
			"访问令牌": {string(authx.PermUserTokenRead), string(authx.PermUserTokenWrite)},
			"角色组":  {string(authx.PermUserGroupRead), string(authx.PermUserGroupWrite)},
			"存储容量": {string(authx.PermUserCapacityRead), string(authx.PermUserCapacityWrite)},
		},
		"content": {
			"相册管理": {string(authx.PermUserAlbumRead), string(authx.PermUserAlbumWrite)},
			"照片管理": {string(authx.PermUserPhotoRead), string(authx.PermUserPhotoWrite)},
			"内容分享": {string(authx.PermUserShareRead), string(authx.PermUserShareWrite)},
			"内容上传": {string(authx.PermUploadWrite)},
		},
		"service": {
			"工单服务": {string(authx.PermUserTicketRead), string(authx.PermUserTicketWrite)},
			"订单管理": {string(authx.PermUserOrderRead), string(authx.PermUserOrderWrite)},
		},
		"integration": {
			"OAuth集成": {string(authx.PermOAuthRead), string(authx.PermOAuthWrite)},
			"内容广场":    {string(authx.PermExploreRead), string(authx.PermExploreWrite)},
		},
		"basic": {
			"基础功能": {string(authx.PermBasic)},
		},
	}

	r.Success(w, map[string]any{"permissions": perms, "groups": groups})
}

// ---------- POST /api/v2/feedback ----------

func (d *deps) handleFeedback(w http.ResponseWriter, req *http.Request) {
	var body struct {
		Type    string `json:"type"`
		Title   string `json:"title"`
		Name    string `json:"name"`
		Email   string `json:"email"`
		Content string `json:"content"`
	}
	if err := readBody(req, &body); err != nil {
		r.Error(w, "请求体解析失败")
		return
	}
	v := validate.New()
	if !validate.In(body.Type, "general", "dmca") {
		v.Add("type", "类型", "不存在。")
	}
	if !validate.Required(body.Title) || len(body.Title) > 200 {
		v.Add("title", "标题", "不能为空。")
	}
	if !validate.Required(body.Name) || len(body.Name) > 80 {
		v.Add("name", "姓名", "不能为空。")
	}
	if !validate.Required(body.Email) || !validate.Email(body.Email) || len(body.Email) > 100 {
		v.Add("email", "邮箱", "必须是合法的邮箱。")
	}
	if !validate.Required(body.Content) || len(body.Content) > 2000 {
		v.Add("content", "反馈内容", "不能为空。")
	}
	if v.Fail() {
		v.Respond(w)
		return
	}
	ip := authx.ClientIP(req)
	if err := d.gdb.Create(&model.Feedback{
		Type: body.Type, Title: body.Title, Name: body.Name,
		Email: body.Email, Content: body.Content, IPAddress: &ip,
	}).Error; err != nil {
		r.Error(w, "提交失败，请稍后重试")
		return
	}
	r.Created(w, nil)
}
