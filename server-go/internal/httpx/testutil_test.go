package httpx_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
)

// installAndLogin 安装并以管理员登录，返回会话 cookie。
func (e *env) installAndLogin(t *testing.T, username, password string) (*http.Response, envelope, []*http.Cookie) {
	t.Helper()
	_, installEnv, _ := e.postJSONCookie(t, "/api/v2/install", map[string]any{
		"app_name": "测试站", "app_url": "https://img.example.com", "app_license_key": "K",
		"db_connection": "sqlite", "admin_username": username, "admin_email": "a@example.com",
		"admin_password": password,
	})
	if installEnv.Status != "success" {
		t.Fatalf("install: %s %s", installEnv.Status, installEnv.Message)
	}
	return e.postJSONCookie(t, "/api/v2/login", map[string]any{
		"username": username, "password": password,
	})
}

var _ = bytes.NewReader
var _ = json.Marshal
