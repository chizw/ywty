package httpx

import (
	"encoding/json"
	"net/http"

	"github.com/chizw/ywty/server-go/internal/appfiles"
	"github.com/chizw/ywty/server-go/internal/authx"
	"github.com/chizw/ywty/server-go/internal/cache"
	"github.com/chizw/ywty/server-go/internal/captchax"
	"github.com/chizw/ywty/server-go/internal/config"
	"github.com/chizw/ywty/server-go/internal/db"
	"github.com/chizw/ywty/server-go/internal/install"
	"github.com/chizw/ywty/server-go/internal/queue"
	"github.com/chizw/ywty/server-go/internal/setting"
	"github.com/chizw/ywty/server-go/internal/support/r"
	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

// NewServices 组装完整应用（main 调用）。
func NewServices(cfg *config.Config, gdb *gorm.DB, c *cache.Cache, q *queue.Queue, cp *captchax.Service) http.Handler {
	return New(cfg, gdb, &deps{cfg: cfg, gdb: gdb, cache: c, queue: q, captcha: cp})
}

// apiRouter 挂载 /api 下所有版本路由。
//
// 注意：未匹配路径（含 /api/v2/__nope__）返回 HTTP 200 + {"status":"error"} envelope，
// 这是 CI 冒烟断言的约定（PHP 版此处为 404，除此外行为一致）。
func apiRouter(cfg *config.Config, gdb *gorm.DB, d *deps) http.Handler {
	notFound := func(w http.ResponseWriter, _ *http.Request) {
		if !db.IsInstalled(gdb, cfg) {
			r.Error(w, "Application is not installed.")
			return
		}
		r.Error(w, "Not Found")
	}

	r := chi.NewRouter()
	r.NotFound(notFound)
	r.MethodNotAllowed(notFound)

	// ---------- /api/v2 ----------
	v2 := chi.NewRouter()
	v2.NotFound(notFound)
	v2.MethodNotAllowed(notFound)

	// 安装前置接口（未安装时也可访问）
	v2.Get("/check-installation", handleCheckInstallation(cfg, gdb))
	v2.Get("/license", handleLicense)
	v2.Get("/changelog", handleChangelog)
	v2.Post("/install", handleInstall(cfg, gdb))

	// 已安装才可访问的业务接口
	v2.Group(func(g chi.Router) {
		g.Use(requireInstalled(cfg, gdb))

		// Fortify 等价
		g.Post("/login", d.handleLogin)
		g.Post("/logout", d.handleLogout)
		g.Post("/register", d.handleRegister)

		// 公共
		g.Group(func(pub chi.Router) {
			pub.Use(authx.OptionalAuth(gdb, cfg), authx.Initialize(gdb, cfg))
			pub.Get("/configs", d.handleConfigs)
			pub.Get("/group", d.handleGroup)
			pub.Get("/captcha", d.handleCaptcha)
			pub.Get("/token_permissions", d.handleTokenPermissions)
			pub.Post("/feedback", d.handleFeedback)

			// 验证码发送（限流 5 次/分钟）
			pub.Group(func(th chi.Router) {
				th.Use(rateLimit(d.cache, 5, 60))
				th.Post("/sms/send", d.handleSmsCodeSend)
				th.Post("/mail/send", d.handleMailCodeSend)
			})

			// 找回密码
			pub.Post("/sms/reset_password", d.handleSmsResetPassword)
			pub.Post("/mail/reset_password", d.handleMailResetPassword)

			// 分享/广场等公共路由在 M3 挂载；M2 挂载 upload
		})

		// 登录用户（auth + 令牌权限检查）
		g.Group(func(u chi.Router) {
			u.Use(authx.Auth(gdb, cfg), authx.Initialize(gdb, cfg), authx.CheckTokenPermission)

			u.Get("/user/profile", d.handleProfile)
			u.Post("/user/profile", d.handleUpdateProfile)
			u.Post("/user/setting", d.handleUpdateSetting)
			u.Post("/user/bind_phone", d.handleBindPhone)
			u.Post("/user/bind_email", d.handleBindEmail)

			u.Get("/user/groups", d.handleUserGroups)
			u.Delete("/user/groups/{id}", d.handleUserGroupDestroy)
			u.Get("/user/capacities", d.handleUserCapacities)
			u.Delete("/user/capacities/{id}", d.handleUserCapacityDestroy)

			u.Get("/user/tokens", d.handleTokensIndex)
			u.Post("/user/tokens", d.handleTokensStore)
			u.Delete("/user/tokens/{id}", d.handleTokensDestroy)
			u.Get("/user/tokens/permissions", d.handleTokensPermissions)

			// 相册/照片/分享/工单/订单路由在 M2-M4 挂载
		})
	})
	r.Mount("/v2", v2)

	// ---------- /api/v1 legacy（PicGo 等旧客户端，M2 实现） ----------
	v1 := chi.NewRouter()
	v1.NotFound(notFound)
	v1.MethodNotAllowed(notFound)
	v1.Group(func(g chi.Router) {
		g.Use(requireInstalled(cfg, gdb), authx.Initialize(gdb, cfg))
	})
	r.Mount("/v1", v1)

	return r
}

// rateLimit 极简固定窗口限流（throttle:N, ttl 秒）。
func rateLimit(c *cache.Cache, max int, ttl int) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			key := "throttle:" + req.URL.Path + ":" + authx.ClientIP(req)
			if n, _ := c.GetInt(key); n >= max {
				r.ErrorWithCode(w, http.StatusTooManyRequests, "Too Many Attempts.")
				return
			}
			c.Increment(key, ttl)
			next.ServeHTTP(w, req)
		})
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

// ---------- 设置读取小工具（httpx 内共享） ----------

func settingString(gdb *gorm.DB, group, name string) (string, error) {
	return setting.String(gdb, group, name)
}

func settingBoolP(gdb *gorm.DB, group, name string) (bool, error) {
	return setting.Bool(gdb, group, name)
}
