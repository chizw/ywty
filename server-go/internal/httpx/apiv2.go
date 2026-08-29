package httpx

import (
	"encoding/json"
	"net/http"

	"github.com/chizw/ywty/server-go/internal/appfiles"
	"github.com/chizw/ywty/server-go/internal/config"
	"github.com/chizw/ywty/server-go/internal/db"
	"github.com/chizw/ywty/server-go/internal/install"
	"github.com/chizw/ywty/server-go/internal/support/r"
	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

// apiRouter 挂载 /api 下所有版本路由。
//
// 注意：未匹配路径（含 /api/v2/__nope__）返回 HTTP 200 + {"status":"error"} envelope，
// 这是 CI 冒烟断言的约定（PHP 版此处为 404，除此外行为一致）。
func apiRouter(cfg *config.Config, gdb *gorm.DB) http.Handler {
	r := chi.NewRouter()
	notFound := apiFallback(cfg, gdb)
	r.NotFound(notFound)
	r.MethodNotAllowed(notFound)

	r.Route("/v2", func(v2 chi.Router) {
		v2.NotFound(notFound)
		v2.MethodNotAllowed(notFound)

		// 安装前置接口（未安装时也可访问）
		v2.Get("/check-installation", handleCheckInstallation(cfg, gdb))
		v2.Get("/license", handleLicense)
		v2.Get("/changelog", handleChangelog)
		v2.Post("/install", handleInstall(cfg, gdb))

		// 已安装才可访问的业务接口（后续里程碑逐个挂载）
		v2.Group(func(g chi.Router) {
			g.Use(requireInstalled(cfg, gdb))
			_ = g // M1+：configs/group/captcha/upload/user/* 等业务路由
		})
	})

	// /api/v1 legacy（PicGo 等旧客户端）——M2 实现
	r.Route("/v1", func(v1 chi.Router) {
		v1.NotFound(notFound)
		v1.MethodNotAllowed(notFound)
	})

	return r
}

// apiFallback 未匹配路径：未安装时与 PHP boot 行为一致返回 "Application is not installed."，
// 已安装返回 Not Found（HTTP 200 envelope，CI 冒烟约定）。
func apiFallback(cfg *config.Config, gdb *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		if !db.IsInstalled(gdb, cfg) {
			r.Error(w, "Application is not installed.")
			return
		}
		r.Error(w, "Not Found")
	}
}

// requireInstalled 未安装时返回与 PHP 一致的错误 envelope（HTTP 200）。
func requireInstalled(cfg *config.Config, gdb *gorm.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			if !db.IsInstalled(gdb, cfg) {
				r.Error(w, "Application is not installed.")
				return
			}
			next.ServeHTTP(w, req)
		})
	}
}

// ---------- HomeController (M0) ----------

func handleCheckInstallation(cfg *config.Config, gdb *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		r.Success(w, map[string]any{"installed": db.IsInstalled(gdb, cfg)})
	}
}

func handleLicense(w http.ResponseWriter, _ *http.Request) {
	r.Success(w, map[string]any{"content": appfiles.ReadFile("LICENSE.md")})
}

func handleChangelog(w http.ResponseWriter, _ *http.Request) {
	r.Success(w, map[string]any{"content": appfiles.ReadFile("CHANGELOG.md")})
}

// handleInstall 等价 HomeController::install：校验参数后执行安装，
// 返回 201 + 安装日志文本（data.content）。
func handleInstall(cfg *config.Config, gdb *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		var body map[string]any
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			r.ErrorData(w, http.StatusUnprocessableEntity, "The given data was invalid.",
				map[string]any{"errors": map[string][]string{"app_name": {"应用名称 不能为空。"}}})
			return
		}
		str := func(k string) string {
			if v, ok := body[k].(string); ok {
				return v
			}
			return ""
		}
		opts := install.Options{
			AppName:       str("app_name"),
			AppURL:        str("app_url"),
			LicenseKey:    str("app_license_key"),
			AdminUsername: str("admin_username"),
			AdminEmail:    str("admin_email"),
			AdminPassword: str("admin_password"),
			// db_* 字段按 openapi 兼容接收；Go 版数据库连接由环境变量决定
		}
		if _, ok := body["admin_password_confirmation"]; ok &&
			body["admin_password_confirmation"] != body["admin_password"] {
			r.ErrorData(w, http.StatusUnprocessableEntity, "The given data was invalid.",
				map[string]any{"errors": map[string][]string{
					"admin_password": {"管理员密码 两次输入的密码不一致。"},
				}})
			return
		}
		if errs := opts.Validate(); len(errs) > 0 {
			r.ErrorData(w, http.StatusUnprocessableEntity, "The given data was invalid.",
				map[string]any{"errors": errs})
			return
		}
		if err := install.Run(gdb, cfg, opts); err != nil {
			r.Error(w, err.Error())
			return
		}
		r.Created(w, map[string]any{
			"content": "安装完成，请前往 " + opts.AppURL + "/admin 进入后台。",
		})
	}
}
