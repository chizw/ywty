package httpx_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"image"
	"image/color"
	"image/png"
	"mime/multipart"
	"time"

	"github.com/chizw/ywty/server-go/internal/cache"
	"github.com/chizw/ywty/server-go/internal/captchax"
	"github.com/chizw/ywty/server-go/internal/config"
	"github.com/chizw/ywty/server-go/internal/db"
	"github.com/chizw/ywty/server-go/internal/httpx"
	"github.com/chizw/ywty/server-go/internal/jobs"
	"github.com/chizw/ywty/server-go/internal/queue"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

type env struct {
	cfg *config.Config
	gdb *gorm.DB
	ts  *httptest.Server
}

func newEnv(t *testing.T) *env {
	t.Helper()
	dir := testDir(t)
	cfg := &config.Config{
		DataDir:    dir,
		DBDriver:   "sqlite",
		DBPath:     filepath.Join(dir, "test.db"),
		UploadsDir: filepath.Join(dir, "uploads"),
		StaticDir:  filepath.Join(dir, "public"),
		AppURL:     "http://localhost:3000",
	}
	// 伪造前端主题产物，验证静态托管与 SPA fallback
	if err := os.MkdirAll(cfg.StaticDir, 0o755); err != nil {
		t.Fatal(err)
	}
	index := "<!doctype html><html><head><title>raw</title></head><body><div id=\"app\"></div></body></html>"
	if err := os.WriteFile(filepath.Join(cfg.StaticDir, "index.html"), []byte(index), 0o644); err != nil {
		t.Fatal(err)
	}

	gdb, err := gorm.Open(sqlite.Open(cfg.DBPath+"?_pragma=foreign_keys(1)"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Migrate(gdb, "sqlite"); err != nil {
		t.Fatal(err)
	}
	c := cache.New(gdb)
	q := queue.New(gdb)
	jobs.Register(q, gdb, c, cfg)
	q.Start(200 * time.Millisecond)
	cp := captchax.New(c)
	ts := httptest.NewServer(httpx.NewServices(cfg, gdb, c, q, cp))
	t.Cleanup(func() {
		ts.Close()
		q.Stop()
		if sqlDB, err := gdb.DB(); err == nil {
			_ = sqlDB.Close()
		}
	})
	return &env{cfg: cfg, gdb: gdb, ts: ts}
}

type envelope struct {
	Status  string          `json:"status"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
	Time    int64           `json:"time"`
}

func (e *env) get(t *testing.T, path string) (*http.Response, envelope) {
	t.Helper()
	return e.do(t, http.MethodGet, path, nil)
}

func (e *env) postJSON(t *testing.T, path string, body any) (*http.Response, envelope) {
	t.Helper()
	b, _ := json.Marshal(body)
	return e.do(t, http.MethodPost, path, b)
}

func (e *env) do(t *testing.T, method, path string, body []byte) (*http.Response, envelope) {
	t.Helper()
	var req *http.Request
	if body != nil {
		req, _ = http.NewRequest(method, e.ts.URL+path, bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req, _ = http.NewRequest(method, e.ts.URL+path, nil)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("%s %s: %v", method, path, err)
	}
	t.Cleanup(func() { _ = resp.Body.Close() })
	var env envelope
	_ = json.NewDecoder(resp.Body).Decode(&env)
	return resp, env
}

func TestHealthz(t *testing.T) {
	e := newEnv(t)
	resp, _ := e.get(t, "/healthz")
	if resp.StatusCode != 200 {
		t.Fatalf("healthz status = %d", resp.StatusCode)
	}
}

func TestCheckInstallationFlow(t *testing.T) {
	e := newEnv(t)

	resp, env := e.get(t, "/api/v2/check-installation")
	if resp.StatusCode != 200 || env.Status != "success" {
		t.Fatalf("check-installation: %d %s", resp.StatusCode, env.Status)
	}
	var data struct {
		Installed bool `json:"installed"`
	}
	_ = json.Unmarshal(env.Data, &data)
	if data.Installed {
		t.Fatal("初始应为未安装")
	}

	// 未安装时业务接口返回与 PHP 一致的错误 envelope
	resp, env = e.get(t, "/api/v2/configs")
	if resp.StatusCode != 200 || env.Status != "error" || env.Message != "Application is not installed." {
		t.Fatalf("not-installed guard: %d %s %q", resp.StatusCode, env.Status, env.Message)
	}

	// 参数缺失 → 422 + errors
	resp, env = e.postJSON(t, "/api/v2/install", map[string]any{})
	if resp.StatusCode != 422 || env.Status != "error" {
		t.Fatalf("install validation: %d %s", resp.StatusCode, env.Status)
	}

	// 正常安装
	resp, env = e.postJSON(t, "/api/v2/install", map[string]any{
		"app_name":        "测试站",
		"app_url":         "https://img.example.com",
		"app_license_key": "KEY-001",
		"db_connection":   "sqlite",
		"admin_username":  "admin",
		"admin_email":     "admin@example.com",
		"admin_password":  "password123",
	})
	if resp.StatusCode != 201 || env.Status != "success" {
		t.Fatalf("install: %d %s %s", resp.StatusCode, env.Status, env.Message)
	}

	resp, env = e.get(t, "/api/v2/check-installation")
	_ = json.Unmarshal(env.Data, &data)
	if !data.Installed {
		t.Fatal("安装后 check-installation 应为 true")
	}
}

func TestAPIFallbackEnvelope(t *testing.T) {
	e := newEnv(t)

	// 未安装：与 PHP boot 行为一致
	resp, env := e.get(t, "/api/v2/__nope__")
	if resp.StatusCode != 200 || env.Status != "error" || env.Message != "Application is not installed." {
		t.Fatalf("api fallback(未安装): %d %s %q", resp.StatusCode, env.Status, env.Message)
	}

	// 安装后：Not Found envelope（HTTP 200，CI 冒烟约定）
	e.postJSON(t, "/api/v2/install", map[string]any{
		"app_name":        "站",
		"app_url":         "https://img.example.com",
		"app_license_key": "K",
		"db_connection":   "sqlite",
		"admin_username":  "admin3",
		"admin_email":     "a3@example.com",
		"admin_password":  "password123",
	})
	resp, env = e.get(t, "/api/v2/__nope__")
	if resp.StatusCode != 200 || env.Status != "error" || env.Message != "Not Found" {
		t.Fatalf("api fallback(已安装): %d %s %q", resp.StatusCode, env.Status, env.Message)
	}
}

func TestStaticAndSPA(t *testing.T) {
	e := newEnv(t)

	// 静态文件
	_ = os.WriteFile(filepath.Join(e.cfg.StaticDir, "favicon.ico"), []byte("ico"), 0o644)
	resp, err := http.Get(e.ts.URL + "/favicon.ico")
	if err != nil || resp.StatusCode != 200 {
		t.Fatalf("静态文件应可访问: %d %v", resp.StatusCode, err)
	}
	_ = resp.Body.Close()

	// SPA fallback：未知路径返回 index.html（直接读原始 body，不走 envelope 解码）
	resp, err = http.Get(e.ts.URL + "/some/spa/route")
	if err != nil || resp.StatusCode != 200 {
		t.Fatalf("SPA fallback status = %d %v", resp.StatusCode, err)
	}
	body := new(bytes.Buffer)
	_, _ = body.ReadFrom(resp.Body)
	_ = resp.Body.Close()
	if !bytes.Contains(body.Bytes(), []byte(`<div id="app">`)) {
		t.Fatalf("SPA fallback 应返回 index.html 内容: %q", body.String())
	}

	// 安装后标题注入
	e.postJSON(t, "/api/v2/install", map[string]any{
		"app_name":        "标题测试站",
		"app_url":         "https://img.example.com",
		"app_license_key": "K",
		"db_connection":   "sqlite",
		"admin_username":  "admin2",
		"admin_email":     "a2@example.com",
		"admin_password":  "password123",
	})
	// 清除标题缓存：新建 server 实例
	e.ts.Close()
	c := cache.New(e.gdb)
	q := queue.New(e.gdb)
	jobs.Register(q, e.gdb, c, e.cfg)
	q.Start(200 * time.Millisecond)
	e.ts = httptest.NewServer(httpx.NewServices(e.cfg, e.gdb, c, q, captchax.New(c)))
	t.Cleanup(e.ts.Close)
	resp, err = http.Get(e.ts.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	body.Reset()
	_, _ = body.ReadFrom(resp.Body)
	_ = resp.Body.Close()
	if !bytes.Contains(body.Bytes(), []byte("<title>标题测试站 - 您的云上相册。</title>")) {
		t.Fatalf("标题未注入: %s", body.String())
	}
}

// ---------- M1 集成链路 ----------

func (e *env) postJSONCookie(t *testing.T, path string, body any) (*http.Response, envelope, []*http.Cookie) {
	t.Helper()
	b, _ := json.Marshal(body)
	req, _ := http.NewRequest(http.MethodPost, e.ts.URL+path, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("%s: %v", path, err)
	}
	var env envelope
	_ = json.NewDecoder(resp.Body).Decode(&env)
	_ = resp.Body.Close()
	return resp, env, resp.Cookies()
}

func TestAuthFlow(t *testing.T) {
	e := newEnv(t)

	// 安装
	_, installEnv, _ := e.postJSONCookie(t, "/api/v2/install", map[string]any{
		"app_name": "测试站", "app_url": "https://img.example.com", "app_license_key": "K",
		"db_connection": "sqlite", "admin_username": "admin", "admin_email": "a@example.com",
		"admin_password": "password123",
	})
	if installEnv.Status != "success" {
		t.Fatalf("install: %s", installEnv.Message)
	}

	// configs：is_logged_in=false、注册开启
	resp, cfgEnv := e.get(t, "/api/v2/configs")
	if resp.StatusCode != 200 || cfgEnv.Status != "success" {
		t.Fatalf("configs: %d %s", resp.StatusCode, cfgEnv.Status)
	}
	var cfgData struct {
		App struct {
			IsLoggedIn         bool  `json:"is_logged_in"`
			EnableRegistration bool  `json:"enable_registration"`
			PhotoCount         int   `json:"photo_count"`
			Countries          []any `json:"countries"`
		} `json:"app"`
		Site struct {
			Title string `json:"title"`
		} `json:"site"`
	}
	_ = json.Unmarshal(cfgEnv.Data, &cfgData)
	if cfgData.App.IsLoggedIn || !cfgData.App.EnableRegistration || len(cfgData.App.Countries) == 0 {
		t.Fatalf("configs 数据异常: %+v", cfgData)
	}

	// 直接注入注册验证码（队列任务需 SMTP，测试环境绕过）
	_ = putCode(e.gdb, "mail_code:register:newuser@example.com", "123456")

	// 注册
	resp, regEnv, cookies := e.postJSONCookie(t, "/api/v2/register", map[string]any{
		"username": "NewUser", "name": "新用户", "email": "newuser@example.com",
		"password": "password123", "password_confirmation": "password123", "code": "123456",
	})
	if resp.StatusCode != 201 {
		t.Fatalf("register: %d %s %s", resp.StatusCode, regEnv.Status, regEnv.Message)
	}
	if len(cookies) == 0 {
		t.Fatal("注册后应下发会话 cookie")
	}

	// 会话访问 profile
	req, _ := http.NewRequest(http.MethodGet, e.ts.URL+"/api/v2/user/profile", nil)
	req.AddCookie(cookies[0])
	resp2, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	var profEnv envelope
	_ = json.NewDecoder(resp2.Body).Decode(&profEnv)
	_ = resp2.Body.Close()
	if resp2.StatusCode != 200 || profEnv.Status != "success" {
		t.Fatalf("profile via session: %d %s", resp2.StatusCode, profEnv.Message)
	}
	var prof struct {
		Username  string         `json:"username"`
		Email     string         `json:"email"`
		IsAdmin   bool           `json:"is_admin"`
		UserGroup map[string]any `json:"group"`
	}
	_ = json.Unmarshal(profEnv.Data, &prof)
	if prof.Username != "newuser" || prof.IsAdmin { // 注册用户名会被小写化
		t.Fatalf("profile 内容异常: %+v", prof)
	}

	// 创建令牌（会话认证有全部权限）
	tokBody, _ := json.Marshal(map[string]any{"name": "picgo"})
	req3, _ := http.NewRequest(http.MethodPost, e.ts.URL+"/api/v2/user/tokens", bytes.NewReader(tokBody))
	req3.Header.Set("Content-Type", "application/json")
	req3.AddCookie(cookies[0])
	resp3, err := http.DefaultClient.Do(req3)
	if err != nil {
		t.Fatal(err)
	}
	var tokEnv envelope
	_ = json.NewDecoder(resp3.Body).Decode(&tokEnv)
	_ = resp3.Body.Close()
	if resp3.StatusCode != 200 {
		t.Fatalf("token create: %d %s", resp3.StatusCode, tokEnv.Message)
	}
	var tok struct {
		Token     string   `json:"token"`
		Abilities []string `json:"abilities"`
	}
	_ = json.Unmarshal(tokEnv.Data, &tok)
	if tok.Token == "" {
		t.Fatal("应返回明文 token")
	}

	// 用令牌访问 profile
	req4, _ := http.NewRequest(http.MethodGet, e.ts.URL+"/api/v2/user/profile", nil)
	req4.Header.Set("Authorization", "Bearer "+tok.Token)
	resp4, err := http.DefaultClient.Do(req4)
	if err != nil {
		t.Fatal(err)
	}
	var profEnv2 envelope
	_ = json.NewDecoder(resp4.Body).Decode(&profEnv2)
	_ = resp4.Body.Close()
	if resp4.StatusCode != 200 || profEnv2.Status != "success" {
		t.Fatalf("profile via token: %d %s", resp4.StatusCode, profEnv2.Message)
	}

	// 登录 + 错误密码
	resp5, _, loginCookies := e.postJSONCookie(t, "/api/v2/login", map[string]any{
		"username": "newuser", "password": "password123",
	})
	if resp5.StatusCode != 200 {
		t.Fatalf("login: %d", resp5.StatusCode)
	}
	if len(loginCookies) == 0 {
		t.Fatal("登录应下发会话 cookie")
	}
	resp6, badEnv, _ := e.postJSONCookie(t, "/api/v2/login", map[string]any{
		"username": "newuser", "password": "wrongpass",
	})
	if resp6.StatusCode != 422 || badEnv.Status != "error" {
		t.Fatalf("bad login: %d %s", resp6.StatusCode, badEnv.Message)
	}

	// 无令牌访问受保护接口
	resp7, _ := e.get(t, "/api/v2/user/profile")
	if resp7.StatusCode != 401 {
		t.Fatalf("unauthenticated profile: %d", resp7.StatusCode)
	}
}

// ---------- M2 上传链路 ----------

func makePNG(t *testing.T, w, h int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x * 7 % 255), G: uint8(y * 5 % 255), B: 128, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func multipartBody(t *testing.T, fieldName, filename string, data []byte, fields map[string]string) (*bytes.Buffer, string) {
	t.Helper()
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	fw, err := mw.CreateFormFile(fieldName, filename)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := fw.Write(data); err != nil {
		t.Fatal(err)
	}
	for k, v := range fields {
		_ = mw.WriteField(k, v)
	}
	_ = mw.Close()
	return &buf, mw.FormDataContentType()
}

func TestUploadFlow(t *testing.T) {
	e := newEnv(t)

	// 安装 + 管理员登录
	_, _, cookies := e.installAndLogin(t, "admin", "password123")

	// 上传 PNG
	pngData := makePNG(t, 64, 48)
	body, ctype := multipartBody(t, "file", "测试图片.png", pngData, map[string]string{
		"storage_id": "1", "is_public": "1",
	})
	req, _ := http.NewRequest(http.MethodPost, e.ts.URL+"/api/v2/upload", body)
	req.Header.Set("Content-Type", ctype)
	req.AddCookie(cookies[0])
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	var upEnv envelope
	_ = json.NewDecoder(resp.Body).Decode(&upEnv)
	_ = resp.Body.Close()
	if resp.StatusCode != 200 || upEnv.Status != "success" {
		t.Fatalf("upload: %d %s %s", resp.StatusCode, upEnv.Status, upEnv.Message)
	}
	var up struct {
		ID        int64  `json:"id"`
		Pathname  string `json:"pathname"`
		Extension string `json:"extension"`
		PublicURL string `json:"public_url"`
		Width     int    `json:"width"`
		Height    int    `json:"height"`
		MD5       string `json:"md5"`
	}
	_ = json.Unmarshal(upEnv.Data, &up)
	if up.ID == 0 || up.Pathname == "" || up.Width != 64 || up.Height != 48 {
		t.Fatalf("upload 数据异常: %+v", up)
	}

	// 物理文件存在（默认组本地储存 root = {dir}/uploads）
	physical := filepath.Join(e.cfg.UploadsDir, filepath.FromSlash(up.Pathname))
	if _, err := os.Stat(physical); err != nil {
		t.Fatalf("物理文件不存在: %s", physical)
	}

	// 图片直出路由（prefix=uploads）
	imgResp, err := http.Get(e.ts.URL + "/uploads/" + up.Pathname)
	if err != nil {
		t.Fatal(err)
	}
	imgBody := new(bytes.Buffer)
	_, _ = imgBody.ReadFrom(imgResp.Body)
	_ = imgResp.Body.Close()
	if imgResp.StatusCode != 200 || !bytes.Equal(imgBody.Bytes(), pngData) {
		t.Fatalf("图片直出异常: %d len=%d", imgResp.StatusCode, imgBody.Len())
	}

	// 缩略图任务（队列 200ms 轮询）
	deadline := time.Now().Add(3 * time.Second)
	thumbOK := false
	for time.Now().Before(deadline) {
		tr, err := http.Get(e.ts.URL + "/storage/thumbnails/" + up.Pathname)
		if err == nil {
			if tr.StatusCode == 200 {
				thumbOK = true
			}
			_ = tr.Body.Close()
		}
		if thumbOK {
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
	if !thumbOK {
		t.Fatal("缩略图未生成")
	}

	// 列表
	req2, _ := http.NewRequest(http.MethodGet, e.ts.URL+"/api/v2/user/photos", nil)
	req2.AddCookie(cookies[0])
	resp2, _ := http.DefaultClient.Do(req2)
	var listEnv envelope
	_ = json.NewDecoder(resp2.Body).Decode(&listEnv)
	_ = resp2.Body.Close()
	var list struct {
		Data []map[string]any `json:"data"`
		Meta struct {
			Total int64 `json:"total"`
		} `json:"meta"`
	}
	_ = json.Unmarshal(listEnv.Data, &list)
	if list.Meta.Total != 1 || len(list.Data) != 1 {
		t.Fatalf("photos 列表异常: %+v", list)
	}

	// 更新 + 删除
	ub, _ := json.Marshal(map[string]any{"ids": []int64{up.ID}, "name": "新名字"})
	req3, _ := http.NewRequest(http.MethodPut, e.ts.URL+"/api/v2/photos/update", bytes.NewReader(ub))
	req3.Header.Set("Content-Type", "application/json")
	req3.AddCookie(cookies[0])
	resp3, _ := http.DefaultClient.Do(req3)
	_ = resp3.Body.Close()

	db2, _ := json.Marshal(map[string]any{"ids": []int64{up.ID}})
	req4, _ := http.NewRequest(http.MethodDelete, e.ts.URL+"/api/v2/photos", bytes.NewReader(db2))
	req4.Header.Set("Content-Type", "application/json")
	req4.AddCookie(cookies[0])
	resp4, _ := http.DefaultClient.Do(req4)
	var delEnv envelope
	_ = json.NewDecoder(resp4.Body).Decode(&delEnv)
	_ = resp4.Body.Close()
	if resp4.StatusCode != 200 {
		t.Fatalf("删除失败: %d %s", resp4.StatusCode, delEnv.Message)
	}
	if _, err := os.Stat(physical); !os.IsNotExist(err) {
		t.Fatal("删除后物理文件应被清理")
	}

	// legacy v1：strategies + profile + 上传
	req5, _ := http.NewRequest(http.MethodGet, e.ts.URL+"/api/v1/strategies", nil)
	req5.AddCookie(cookies[0])
	resp5, _ := http.DefaultClient.Do(req5)
	var sEnv struct {
		Status bool `json:"status"`
		Data   struct {
			Strategies []map[string]any `json:"strategies"`
		} `json:"data"`
	}
	_ = json.NewDecoder(resp5.Body).Decode(&sEnv)
	_ = resp5.Body.Close()
	if !sEnv.Status || len(sEnv.Data.Strategies) == 0 {
		t.Fatalf("v1 strategies 异常: %+v", sEnv)
	}

	body6, ctype6 := multipartBody(t, "file", "legacy.jpg", makePNG(t, 32, 32), nil)
	req6, _ := http.NewRequest(http.MethodPost, e.ts.URL+"/api/v1/upload", body6)
	req6.Header.Set("Content-Type", ctype6)
	req6.AddCookie(cookies[0])
	resp6, _ := http.DefaultClient.Do(req6)
	var legacy struct {
		Status bool           `json:"status"`
		Data   map[string]any `json:"data"`
	}
	_ = json.NewDecoder(resp6.Body).Decode(&legacy)
	_ = resp6.Body.Close()
	if !legacy.Status || legacy.Data["key"] == nil {
		t.Fatalf("v1 upload 异常: %+v", legacy)
	}
}

func TestShareFlow(t *testing.T) {
	e := newEnv(t)
	_, _, cookies := e.installAndLogin(t, "admin", "password123")

	// 上传公开图片
	body, ctype := multipartBody(t, "file", "share.png", makePNG(t, 40, 40), map[string]string{
		"storage_id": "1", "is_public": "1",
	})
	req, _ := http.NewRequest(http.MethodPost, e.ts.URL+"/api/v2/upload", body)
	req.Header.Set("Content-Type", ctype)
	req.AddCookie(cookies[0])
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	var upEnv envelope
	_ = json.NewDecoder(resp.Body).Decode(&upEnv)
	_ = resp.Body.Close()
	var up struct {
		ID int64 `json:"id"`
	}
	_ = json.Unmarshal(upEnv.Data, &up)

	// 创建分享
	sb, _ := json.Marshal(map[string]any{"type": "photo", "ids": []int64{up.ID}, "content": "看看这张"})
	req2, _ := http.NewRequest(http.MethodPost, e.ts.URL+"/api/v2/user/shares", bytes.NewReader(sb))
	req2.Header.Set("Content-Type", "application/json")
	req2.AddCookie(cookies[0])
	resp2, _ := http.DefaultClient.Do(req2)
	var shEnv envelope
	_ = json.NewDecoder(resp2.Body).Decode(&shEnv)
	_ = resp2.Body.Close()
	if resp2.StatusCode != 201 || shEnv.Status != "success" {
		t.Fatalf("创建分享失败: %d %s", resp2.StatusCode, shEnv.Message)
	}
	var sh struct {
		Slug string `json:"slug"`
	}
	_ = json.Unmarshal(shEnv.Data, &sh)
	if sh.Slug == "" {
		t.Fatal("应返回 slug")
	}

	// 公开访问分享
	resp3, showEnv := e.get(t, "/api/v2/shares/"+sh.Slug)
	if resp3.StatusCode != 200 || showEnv.Status != "success" {
		t.Fatalf("访问分享失败: %d %s", resp3.StatusCode, showEnv.Message)
	}
	var show struct {
		IsValid bool   `json:"is_valid"`
		Content string `json:"content"`
	}
	_ = json.Unmarshal(showEnv.Data, &show)
	if !show.IsValid || show.Content != "看看这张" {
		t.Fatalf("分享内容异常: %+v", show)
	}

	// 分享图片列表
	_, phEnv := e.get(t, "/api/v2/shares/"+sh.Slug+"/photos")
	var ph struct {
		IsValid bool             `json:"is_valid"`
		Data    []map[string]any `json:"data"`
		Meta    map[string]any   `json:"meta"`
	}
	_ = json.Unmarshal(phEnv.Data, &ph)
	if !ph.IsValid || len(ph.Data) != 1 {
		t.Fatalf("分享图片列表异常: %+v raw=%s", ph, string(phEnv.Data))
	}

	// 广场默认关闭
	resp5, expEnv := e.get(t, "/api/v2/explore/photos")
	if resp5.StatusCode != 404 || expEnv.Message != "Gallery feature is disabled" {
		t.Fatalf("explore 开关异常: %d %s", resp5.StatusCode, expEnv.Message)
	}
}

func TestOrderFlow(t *testing.T) {
	e := newEnv(t)
	_, _, cookies := e.installAndLogin(t, "admin", "password123")

	// 造套餐数据：vip 套餐 → 默认组，价格 0（自动完成）与 1000 分两档
	var groupID int64
	e.gdb.Raw("SELECT `id` FROM `groups` WHERE `is_default` = 1 LIMIT 1").Scan(&groupID)
	now := time.Now().UTC()
	e.gdb.Exec("INSERT INTO `plans` (`type`, `name`, `intro`, `features`, `badge`, `sort`, `is_up`, `created_at`, `updated_at`) "+
		"VALUES ('vip', 'VIP套餐', '测试', '[\"功能A\"]', '热', 0, 1, ?, ?)", now, now)
	var planID int64
	e.gdb.Raw("SELECT LAST_INSERT_ROWID()").Scan(&planID)
	e.gdb.Exec("INSERT INTO `plan_groups` (`plan_id`, `group_id`) VALUES (?, ?)", planID, groupID)
	e.gdb.Exec("INSERT INTO `plan_prices` (`plan_id`, `name`, `duration`, `price`, `created_at`, `updated_at`) "+
		"VALUES (?, '月付', 43200, 0, ?, ?)", planID, now, now)
	var priceID int64
	e.gdb.Raw("SELECT LAST_INSERT_ROWID()").Scan(&priceID)

	// plans 公共列表
	resp, plansEnv := e.get(t, "/api/v2/plans")
	if resp.StatusCode != 200 || plansEnv.Status != "success" {
		t.Fatalf("plans: %d %s", resp.StatusCode, plansEnv.Message)
	}

	// 0 元订单自动完成
	ob, _ := json.Marshal(map[string]any{"price_id": priceID})
	req, _ := http.NewRequest(http.MethodPost, e.ts.URL+"/api/v2/user/orders", bytes.NewReader(ob))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(cookies[0])
	resp2, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	var ordEnv envelope
	_ = json.NewDecoder(resp2.Body).Decode(&ordEnv)
	_ = resp2.Body.Close()
	if resp2.StatusCode != 201 || ordEnv.Status != "success" {
		t.Fatalf("创建订单失败: %d %s", resp2.StatusCode, ordEnv.Message)
	}
	var ord struct {
		TradeNo string `json:"trade_no"`
		IsPaid  bool   `json:"is_paid"`
	}
	_ = json.Unmarshal(ordEnv.Data, &ord)
	if !ord.IsPaid {
		t.Fatalf("0 元订单应自动支付: %+v", ord)
	}

	// 用户组发放
	req3, _ := http.NewRequest(http.MethodGet, e.ts.URL+"/api/v2/user/groups", nil)
	req3.AddCookie(cookies[0])
	resp3, _ := http.DefaultClient.Do(req3)
	var groupsEnv envelope
	_ = json.NewDecoder(resp3.Body).Decode(&groupsEnv)
	_ = resp3.Body.Close()
	var groups struct {
		Data []map[string]any `json:"data"`
	}
	_ = json.Unmarshal(groupsEnv.Data, &groups)
	if len(groups.Data) == 0 {
		t.Fatalf("0 元订单后应有角色组: %s", string(groupsEnv.Data))
	}
	g := groups.Data[0]
	if g["from"] != "subscribe" {
		t.Fatalf("来源应为 subscribe: %+v", g)
	}

	// 付费订单：preview + store + pay（未配置驱动应报错）
	e.gdb.Exec("INSERT INTO `plan_prices` (`plan_id`, `name`, `duration`, `price`, `created_at`, `updated_at`) "+
		"VALUES (?, '年付', 525600, 1000, ?, ?)", planID, now, now)
	var paidPriceID int64
	e.gdb.Raw("SELECT LAST_INSERT_ROWID()").Scan(&paidPriceID)

	pb, _ := json.Marshal(map[string]any{"price_id": paidPriceID})
	req4, _ := http.NewRequest(http.MethodPost, e.ts.URL+"/api/v2/orders/preview", bytes.NewReader(pb))
	req4.Header.Set("Content-Type", "application/json")
	req4.AddCookie(cookies[0])
	resp4, _ := http.DefaultClient.Do(req4)
	var pvEnv envelope
	_ = json.NewDecoder(resp4.Body).Decode(&pvEnv)
	_ = resp4.Body.Close()
	var pv struct {
		Amount int64 `json:"amount"`
	}
	_ = json.Unmarshal(pvEnv.Data, &pv)
	if pv.Amount != 1000 {
		t.Fatalf("preview 金额异常: %+v", pv)
	}

	req5, _ := http.NewRequest(http.MethodPost, e.ts.URL+"/api/v2/user/orders", bytes.NewReader(pb))
	req5.Header.Set("Content-Type", "application/json")
	req5.AddCookie(cookies[0])
	resp5, _ := http.DefaultClient.Do(req5)
	var ord2Env envelope
	_ = json.NewDecoder(resp5.Body).Decode(&ord2Env)
	_ = resp5.Body.Close()
	var ord2 struct {
		TradeNo string `json:"trade_no"`
		IsPaid  bool   `json:"is_paid"`
	}
	_ = json.Unmarshal(ord2Env.Data, &ord2)
	if ord2.IsPaid {
		t.Fatal("付费订单不应自动支付")
	}

	// 支付：组未配置支付驱动 → 业务错误
	payb, _ := json.Marshal(map[string]any{"platform": "epay", "channel": "alipay", "method": "web"})
	req6, _ := http.NewRequest(http.MethodPost, e.ts.URL+"/api/v2/orders/"+ord2.TradeNo+"/pay", bytes.NewReader(payb))
	req6.Header.Set("Content-Type", "application/json")
	req6.AddCookie(cookies[0])
	resp6, _ := http.DefaultClient.Do(req6)
	var payEnv envelope
	_ = json.NewDecoder(resp6.Body).Decode(&payEnv)
	_ = resp6.Body.Close()
	if payEnv.Status != "error" || payEnv.Message != "未配置支付驱动，请联系管理员" {
		t.Fatalf("pay 应报未配置驱动: %s", payEnv.Message)
	}

	// 取消订单
	req7, _ := http.NewRequest(http.MethodPut, e.ts.URL+"/api/v2/orders/"+ord2.TradeNo+"/cancel", nil)
	req7.AddCookie(cookies[0])
	resp7, _ := http.DefaultClient.Do(req7)
	_ = resp7.Body.Close()
	if resp7.StatusCode != 204 {
		t.Fatalf("取消订单失败: %d", resp7.StatusCode)
	}
}
