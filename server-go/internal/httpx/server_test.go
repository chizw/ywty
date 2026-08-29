package httpx_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/chizw/ywty/server-go/internal/config"
	"github.com/chizw/ywty/server-go/internal/db"
	"github.com/chizw/ywty/server-go/internal/httpx"
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
	ts := httptest.NewServer(httpx.New(cfg, gdb))
	t.Cleanup(func() {
		ts.Close()
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
	e.ts = httptest.NewServer(httpx.New(e.cfg, e.gdb))
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
