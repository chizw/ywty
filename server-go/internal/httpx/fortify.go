package httpx

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/chizw/ywty/server-go/internal/authx"
	"github.com/chizw/ywty/server-go/internal/cache"
	"github.com/chizw/ywty/server-go/internal/captchax"
	"github.com/chizw/ywty/server-go/internal/config"
	"github.com/chizw/ywty/server-go/internal/countries"
	"github.com/chizw/ywty/server-go/internal/db/types"
	"github.com/chizw/ywty/server-go/internal/mailx"
	"github.com/chizw/ywty/server-go/internal/model"
	"github.com/chizw/ywty/server-go/internal/queue"
	"github.com/chizw/ywty/server-go/internal/setting"
	"github.com/chizw/ywty/server-go/internal/support/r"
	"github.com/chizw/ywty/server-go/internal/validate"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// deps 处理器共享依赖。
type deps struct {
	cfg     *config.Config
	gdb     *gorm.DB
	cache   *cache.Cache
	queue   *queue.Queue
	captcha *captchax.Service
}

// ---------- POST /api/v2/login（Fortify 等价） ----------

func (d *deps) handleLogin(w http.ResponseWriter, req *http.Request) {
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := readBody(req, &body); err != nil || strings.TrimSpace(body.Username) == "" {
		r.ErrorData(w, http.StatusUnprocessableEntity, "The given data was invalid.",
			map[string]any{"errors": map[string][]string{"username": {"用户名 不能为空。"}}})
		return
	}
	body.Username = strings.ToLower(strings.TrimSpace(body.Username))

	// 登录限流：username+ip 每分钟 5 次（对齐 fortify 'login' limiter）
	ip := authx.ClientIP(req)
	limiterKey := "login:" + body.Username + "|" + ip
	if n, _ := d.cache.GetInt(limiterKey); n >= 5 {
		r.ErrorWithCode(w, http.StatusTooManyRequests, "Too Many Attempts.")
		return
	}
	d.cache.Increment(limiterKey, 60)

	var user model.User
	err := d.gdb.Where("username = ?", body.Username).First(&user).Error
	if err != nil || user.ID == 0 ||
		bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(body.Password)) != nil {
		r.ErrorData(w, http.StatusUnprocessableEntity, "The given data was invalid.",
			map[string]any{"errors": map[string][]string{
				"username": {"These credentials do not match our records."},
			}})
		return
	}

	sid, err := authx.CreateSession(d.gdb, user.ID, ip, req.UserAgent())
	if err != nil {
		r.Error(w, "登录会话创建失败")
		return
	}
	d.cache.Forget(limiterKey)

	http.SetCookie(w, &http.Cookie{
		Name:     authx.SessionCookieName(d.cfg),
		Value:    sid,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   120 * 60,
	})
	// Fortify LoginResponse: response()->json(['two_factor' => false])
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"two_factor": false})
}

// ---------- POST /api/v2/logout ----------

func (d *deps) handleLogout(w http.ResponseWriter, req *http.Request) {
	if c, err := req.Cookie(authx.SessionCookieName(d.cfg)); err == nil {
		authx.DestroySession(d.gdb, c.Value)
	}
	http.SetCookie(w, &http.Cookie{
		Name: authx.SessionCookieName(d.cfg), Value: "", Path: "/", MaxAge: -1,
	})
	w.WriteHeader(http.StatusNoContent)
}

// ---------- POST /api/v2/register（CreateNewUser 等价） ----------

func (d *deps) handleRegister(w http.ResponseWriter, req *http.Request) {
	var body map[string]any
	if err := readBody(req, &body); err != nil {
		r.Error(w, "请求体解析失败")
		return
	}
	str := func(k string) string {
		if v, ok := body[k].(string); ok {
			return strings.TrimSpace(v)
		}
		return ""
	}

	enableReg, _ := setting.Bool(d.gdb, setting.GroupApp, "enable_registration")
	if !enableReg {
		r.Error(w, "系统已关闭注册功能")
		return
	}
	emailVerify, _ := setting.Bool(d.gdb, setting.GroupApp, "user_email_verify")
	phoneVerify, _ := setting.Bool(d.gdb, setting.GroupApp, "user_phone_verify")

	username := strings.ToLower(str("username"))
	name := str("name")
	email := strings.ToLower(str("email"))
	phone := str("phone")
	countryCode := str("country_code")
	password := str("password")
	code := str("code")

	v := validate.New()
	if !validate.Required(username) || !validate.AlphaDash(username) || !validate.MaxLen(username, 200) {
		v.Add("username", "用户名", "格式不正确。")
	} else if unique := d.checkUnique("users", "username", username, 0); !unique {
		v.Add("username", "用户名", "已被占用。")
	}
	if !validate.Required(name) || !validate.MaxLen(name, 255) {
		v.Add("name", "昵称", "不能为空。")
	}
	if email == "" && phone == "" {
		v.Add("email", "邮箱", "不能为空。")
	}
	if email != "" {
		if !validate.Email(email) || !validate.MaxLen(email, 255) {
			v.Add("email", "邮箱", "必须是合法的邮箱。")
		} else if unique := d.checkUnique("users", "email", email, 0); !unique {
			v.Add("email", "邮箱", "已被占用。")
		}
	}
	if phone != "" {
		if !phoneVerify {
			v.Add("phone", "手机号", "系统暂不支持手机号注册")
		} else {
			if countryCode == "" {
				v.Add("country_code", "国家代码", "不能为空。")
			} else if !countries.IsValidCountryCode(countryCode) {
				v.Add("country_code", "国家代码", "不存在。")
			}
			if unique := d.checkUnique("users", "phone", phone, 0); !unique {
				v.Add("phone", "手机号", "已被占用。")
			}
		}
	}
	if !validate.Required(password) || len(password) < 8 {
		v.Add("password", "密码", "至少需要 8 个字符。")
	}
	if pc, ok := body["password_confirmation"].(string); ok && pc != password {
		v.Add("password", "密码", "两次输入的密码不一致。")
	}
	if email != "" && emailVerify {
		if !mailx.VerifyCode(d.cache, mailx.CodeKey("register", email), code) {
			v.Add("code", "验证码", "Invalid verification code.")
		}
	}
	if phone != "" && phoneVerify {
		if !smsVerify(d.cache, "register", phone, code) {
			v.Add("code", "验证码", "Invalid verification code.")
		}
	}
	if v.Fail() {
		v.Respond(w)
		return
	}

	// 创建用户（UserService::store）
	now := time.Now().UTC()
	ip := authx.ClientIP(req)
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), 12)
	user := model.User{
		Username: username,
		Name:     name,
		Status:   "normal",
		Password: string(hash),
		Options: types.MustJSON(map[string]any{
			"language":                 "zh-CN",
			"show_original_photos":     false,
			"encode_copied_url":        true,
			"auto_upload_after_select": false,
		}),
		RegisterIP: &ip,
	}
	if email != "" {
		user.Email = &email
		user.EmailVerifiedAt = &now
	}
	if phone != "" {
		user.Phone = &phone
		user.PhoneVerifiedAt = &now
		user.CountryCode = &countryCode
	}
	if err := d.gdb.Create(&user).Error; err != nil {
		r.Error(w, "注册失败，请稍后重试")
		return
	}

	// 默认角色组 + 初始容量
	var defaultGroupID int64
	d.gdb.Raw("SELECT `id` FROM `groups` WHERE `is_default` = 1 AND `deleted_at` IS NULL ORDER BY `id` LIMIT 1").Scan(&defaultGroupID)
	if defaultGroupID > 0 {
		_ = d.gdb.Create(&model.UserGroup{UserID: user.ID, GroupID: defaultGroupID, From: "system"}).Error
	}
	if initialCapacity, err := setting.Int64(d.gdb, setting.GroupUser, "initial_capacity"); err == nil {
		_ = d.gdb.Create(&model.UserCapacity{UserID: user.ID, Capacity: float64(initialCapacity), From: "system"}).Error
	}

	// 自动登录（Fortify guard->login）
	if sid, err := authx.CreateSession(d.gdb, user.ID, ip, req.UserAgent()); err == nil {
		http.SetCookie(w, &http.Cookie{
			Name: authx.SessionCookieName(d.cfg), Value: sid, Path: "/",
			HttpOnly: true, SameSite: http.SameSiteLaxMode, MaxAge: 120 * 60,
		})
	}
	// Fortify RegisterResponse: new JsonResponse('', 201)
	w.WriteHeader(http.StatusCreated)
}

// ---------- POST /api/v2/mail/reset_password、/api/v2/sms/reset_password ----------

func (d *deps) handleMailResetPassword(w http.ResponseWriter, req *http.Request) {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Code     string `json:"code"`
	}
	if err := readBody(req, &body); err != nil {
		r.Error(w, "请求体解析失败")
		return
	}
	body.Email = strings.ToLower(strings.TrimSpace(body.Email))

	v := validate.New()
	var user model.User
	if !validate.Required(body.Email) || !validate.Email(body.Email) {
		v.Add("email", "邮箱", "不能为空。")
	} else if err := d.gdb.Where("email = ?", body.Email).First(&user).Error; err != nil {
		v.Add("email", "邮箱", "不存在。")
	}
	if !validate.Required(body.Password) || len(body.Password) < 8 {
		v.Add("password", "密码", "至少需要 8 个字符。")
	}
	if !mailx.VerifyCode(d.cache, mailx.CodeKey("forget_password", body.Email), body.Code) {
		v.Add("code", "验证码", "Invalid verification code.")
	}
	if v.Fail() {
		v.Respond(w)
		return
	}
	hash, _ := bcrypt.GenerateFromPassword([]byte(body.Password), 12)
	if err := d.gdb.Model(&model.User{}).Where("id = ?", user.ID).
		Updates(map[string]any{"password": string(hash), "updated_at": time.Now().UTC()}).Error; err != nil {
		r.Error(w, "重置失败，请稍后重试")
		return
	}
	r.Created(w, nil)
}

func (d *deps) handleSmsResetPassword(w http.ResponseWriter, req *http.Request) {
	var body struct {
		Phone       string `json:"phone"`
		CountryCode string `json:"country_code"`
		Password    string `json:"password"`
		Code        string `json:"code"`
	}
	if err := readBody(req, &body); err != nil {
		r.Error(w, "请求体解析失败")
		return
	}
	v := validate.New()
	var user model.User
	if !validate.Required(body.Phone) {
		v.Add("phone", "手机号", "不能为空。")
	} else if err := d.gdb.Where("phone = ?", body.Phone).First(&user).Error; err != nil {
		v.Add("phone", "手机号", "不存在。")
	}
	if !validate.Required(body.Password) || len(body.Password) < 8 {
		v.Add("password", "密码", "至少需要 8 个字符。")
	}
	if !smsVerify(d.cache, "forget_password", body.Phone, body.Code) {
		v.Add("code", "验证码", "Invalid verification code.")
	}
	if v.Fail() {
		v.Respond(w)
		return
	}
	hash, _ := bcrypt.GenerateFromPassword([]byte(body.Password), 12)
	if err := d.gdb.Model(&model.User{}).Where("id = ?", user.ID).
		Updates(map[string]any{"password": string(hash), "updated_at": time.Now().UTC()}).Error; err != nil {
		r.Error(w, "重置失败，请稍后重试")
		return
	}
	r.Created(w, nil)
}

// ---------- 小工具 ----------

func readBody(req *http.Request, dst any) error {
	defer func() { _ = req.Body.Close() }()
	return json.NewDecoder(req.Body).Decode(dst)
}

func (d *deps) checkUnique(table, column, value string, exceptID int64) bool {
	var count int64
	q := "SELECT count(*) FROM `" + table + "` WHERE `" + column + "` = ? AND deleted_at IS NULL"
	args := []any{value}
	if exceptID > 0 {
		q += " AND `id` <> ?"
		args = append(args, exceptID)
	}
	d.gdb.Raw(q, args...).Scan(&count)
	return count == 0
}

// smsVerify 短信验证码（M5 接入真实短信后生效，缓存键与 PHP SmsService 一致）。
func smsVerify(c *cache.Cache, event, phone, code string) bool {
	stored, ok := c.Get("sms_code:" + event + ":" + phone)
	return ok && stored == code
}
