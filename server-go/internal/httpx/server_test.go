package httpx_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

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
		AppKey:     "base64:MTIzNDU2Nzg5MDEyMzQ1Njc4OTAxMjM0NTY3ODkwMTI=",
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
