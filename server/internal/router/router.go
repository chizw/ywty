// Package router 统一路由注册
package router

import (
	"net/http"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/ywty/server/internal/auth"
	"github.com/ywty/server/internal/config"
	"github.com/ywty/server/internal/handler"
	"github.com/ywty/server/internal/logger"
	"github.com/ywty/server/internal/middleware"
	"github.com/ywty/server/internal/notify"
	"github.com/ywty/server/internal/queue"
	"github.com/ywty/server/internal/rbac"
	"github.com/ywty/server/internal/response"
	"github.com/ywty/server/internal/service"
	"github.com/ywty/server/internal/storage"
)

// Options 路由选项
type Options struct {
	Cfg         *config.Config
	DB          *gorm.DB
	StoreDriver storage.Driver
	Queue       *queue.Client
}

// New 构建根引擎
func New(opt *Options) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	r.Use(gin.Recovery())
	r.Use(middleware.RequestID())
	r.Use(middleware.CORS(
		opt.Cfg.CORS.AllowOrigins,
		opt.Cfg.CORS.AllowMethods,
		opt.Cfg.CORS.AllowHeaders,
		opt.Cfg.CORS.ExposeHeaders,
		opt.Cfg.CORS.AllowCredentials,
		opt.Cfg.CORS.MaxAge,
	))

	// 初始化通知驱动
	initNotify(opt.Cfg)

	// 健康检查
	r.GET("/healthz", func(c *gin.Context) {
		if err := dbPing(opt.DB); err != nil {
			response.Fail(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"app":     opt.Cfg.App.Name,
			"env":     opt.Cfg.App.Env,
			"version": opt.Cfg.App.Version,
			"time":    time.Now().Format(time.RFC3339),
		})
	})

	// 本地存储对外访问（仅当 driver 为 local 时挂载）
	mountLocalStatic(r, opt.Cfg.Storage.Driver.Root)

	// 公开图片重定向：GET /i/:id
	photoSvc := service.NewPhotoService(opt.DB, opt.StoreDriver, opt.Cfg.App.BaseURL)
	photoSvc.SetQueue(opt.Queue)
	r.GET("/i/:id", handler.PhotoRedirect(photoSvc))
	// 公开分享页：GET /s/:slug
	r.GET("/s/:slug", handler.ShareView(service.NewShareService(opt.DB)))

	// 限流
	apiRL := middleware.NewRateLimiter(opt.Cfg.RateLimit.APIPerMinute)
	uploadRL := middleware.NewRateLimiter(opt.Cfg.RateLimit.UploadPerMinute)

	// JWT 颁发器
	issuer := auth.NewIssuer(opt.Cfg.Auth.JWT)

	// Casbin RBAC 鉴权
	enforcer, err := rbac.Init(opt.DB)
	if err != nil {
		logger.L.Error("casbin init failed", zap.Error(err))
	}
	if enforcer != nil {
		if err := rbac.AutoSeed(enforcer); err != nil {
			logger.L.Error("casbin seed failed", zap.Error(err))
		}
	}
	casbinMW := rbac.Middleware(enforcer)

	// 业务服务
	authSvc := service.NewAuthService(opt.DB, issuer)
	userSvc := service.NewUserService(opt.DB)
	photoSvcFull := photoSvc
	albumSvc := service.NewAlbumService(opt.DB)
	tagSvc := service.NewTagService(opt.DB)
	shareSvc := service.NewShareService(opt.DB)
	likeSvc := service.NewLikeService(opt.DB)
	reportSvc := service.NewReportService(opt.DB)
	violationSvc := service.NewViolationService(opt.DB)
	feedbackSvc := service.NewFeedbackService(opt.DB)
	tokenSvc := service.NewTokenService(opt.DB)
	verifySvc := service.NewVerifyCodeService(opt.DB)
	oauthSvc := service.NewOAuthService(opt.DB, issuer)
	captchaSvc := service.NewCaptchaService()
	orderSvc := service.NewOrderService(opt.DB, opt.Cfg.App.BaseURL)
	planSvc := service.NewPlanService(opt.DB)
	couponSvc := service.NewCouponService(opt.DB)
	ticketSvc := service.NewTicketService(opt.DB)
	noticeSvc := service.NewNoticeService(opt.DB)
	pageSvc := service.NewPageService(opt.DB)
	groupSvc := service.NewGroupService(opt.DB)
	licenseSvc := service.NewLicenseService(opt.DB)

	authH := handler.NewAuthHandler(authSvc, handler.NewCaptchaHandler(captchaSvc))
	userH := handler.NewUserHandler(userSvc)
	photoH := handler.NewPhotoHandler(photoSvcFull)
	albumH := handler.NewAlbumHandler(albumSvc)
	tagH := handler.NewTagHandler(tagSvc)
	shareH := handler.NewShareHandler(shareSvc)
	likeH := handler.NewLikeHandler(likeSvc)
	reportH := handler.NewReportHandler(reportSvc)
	violationH := handler.NewViolationHandler(violationSvc)
	feedbackH := handler.NewFeedbackHandler(feedbackSvc)
	tokenH := handler.NewTokenHandler(tokenSvc)
	verifyH := handler.NewVerifyCodeHandler(verifySvc)
	oauthH := handler.NewOAuthHandler(oauthSvc)
	adminH := handler.NewAdminHandler(opt.DB, userSvc, authSvc, enforcer)
	capH := handler.NewCapacityHandler(service.NewCapacityService(opt.DB))
	storeH := handler.NewStorageHandler(opt.DB)
	driversH := handler.NewDriversHandler()
	planH := handler.NewPlanHandler(planSvc)
	couponH := handler.NewCouponHandler(couponSvc)
	orderH := handler.NewOrderHandler(orderSvc)
	ticketH := handler.NewTicketHandler(ticketSvc)
	noticeH := handler.NewNoticeHandler(noticeSvc)
	pageH := handler.NewPageHandler(pageSvc)
	_ = handler.NewGroupHandler(groupSvc) // 群组接口在 adminH 内部已挂载
	licenseH := handler.NewLicenseHandler(licenseSvc)

	// API 路由组
	v1 := r.Group("/api/v1")
	v1.Use(apiRL.Middleware())
	{
		v1.GET("/ping", func(c *gin.Context) {
			response.Success(c, gin.H{"message": "pong", "version": "v1"})
		})

		// 公开
		verifyH.Mount(v1)
		authH.Mount(v1, nil) // auth.Mount 内部处理
		// 公开图片列表（无需鉴权）
		v1.GET("/public/photos", photoH.ListPublic)

		// 鉴权
		authMW := auth.Middleware(opt.DB, issuer)
		authCasbinMW := func(c *gin.Context) {
			authMW(c)
			if c.IsAborted() {
				return
			}
			casbinMW(c)
		}
		photoAuthMW := func(c *gin.Context) {
			authMW(c)
			if c.IsAborted() {
				return
			}
			casbinMW(c)
			if c.IsAborted() {
				return
			}
			uploadRL.Middleware()(c)
		}
		userH.Mount(v1, authCasbinMW)
		photoH.Mount(v1, photoAuthMW) // 上传链路鉴权 + 额外限流
		albumH.Mount(v1, authCasbinMW)
		tagH.Mount(v1, authCasbinMW)
		shareH.Mount(v1, authCasbinMW)
		likeH.Mount(v1, authCasbinMW)
		reportH.Mount(v1, authCasbinMW)
		violationH.Mount(v1, authCasbinMW)
		feedbackH.Mount(v1, authCasbinMW)
		tokenH.Mount(v1, authCasbinMW)
		oauthH.Mount(v1, authCasbinMW)
		adminH.Mount(v1, authCasbinMW) // 内部再叠加 AdminOnly
		capH.Mount(v1, authCasbinMW)
		storeH.Mount(v1, authCasbinMW)
		driversH.Mount(v1, authCasbinMW)
		planH.Mount(v1, authCasbinMW)   // 公开接口不挂 mw，管理接口内部叠加 AdminOnly
		couponH.Mount(v1, authCasbinMW) // 用户验证接口不挂 mw，管理接口内部叠加 AdminOnly
		orderH.Mount(v1, authCasbinMW)  // 支付回调公开，其余需鉴权
		noticeH.Mount(v1, authCasbinMW) // 公开接口不挂 mw，管理接口内部叠加 AdminOnly
		pageH.Mount(v1, authCasbinMW)   // 公开接口不挂 mw，管理接口内部叠加 AdminOnly
		ticketH.Mount(v1, authCasbinMW) // 用户侧挂 mw，管理接口内部叠加 AdminOnly
		licenseH.Mount(v1, authCasbinMW)
	}

	v2 := r.Group("/api/v2")
	v2.Use(apiRL.Middleware())
	{
		v2.GET("/ping", func(c *gin.Context) {
			response.Success(c, gin.H{"message": "pong", "version": "v2"})
		})
		// v2 与 v1 业务接口对齐
		verifyH.Mount(v2)
		authH.Mount(v2, nil)
		authMW := auth.Middleware(opt.DB, issuer)
		authCasbinMW := func(c *gin.Context) {
			authMW(c)
			if c.IsAborted() {
				return
			}
			casbinMW(c)
		}
		photoAuthMW := func(c *gin.Context) {
			authMW(c)
			if c.IsAborted() {
				return
			}
			casbinMW(c)
			if c.IsAborted() {
				return
			}
			uploadRL.Middleware()(c)
		}
		userH.Mount(v2, authCasbinMW)
		photoH.Mount(v2, photoAuthMW)
		albumH.Mount(v2, authCasbinMW)
		tagH.Mount(v2, authCasbinMW)
		shareH.Mount(v2, authCasbinMW)
		likeH.Mount(v2, authCasbinMW)
		reportH.Mount(v2, authCasbinMW)
		violationH.Mount(v2, authCasbinMW)
		feedbackH.Mount(v2, authCasbinMW)
		tokenH.Mount(v2, authCasbinMW)
		oauthH.Mount(v2, authCasbinMW)
		adminH.Mount(v2, authCasbinMW)
		capH.Mount(v2, authCasbinMW)
		storeH.Mount(v2, authCasbinMW)
		driversH.Mount(v2, authCasbinMW)
		planH.Mount(v2, authCasbinMW)   // 公开接口不挂 mw，管理接口内部叠加 AdminOnly
		couponH.Mount(v2, authCasbinMW) // 用户验证接口不挂 mw，管理接口内部叠加 AdminOnly
		orderH.Mount(v2, authCasbinMW)
		noticeH.Mount(v2, authCasbinMW) // 公开接口不挂 mw，管理接口内部叠加 AdminOnly
		pageH.Mount(v2, authCasbinMW)   // 公开接口不挂 mw，管理接口内部叠加 AdminOnly
		ticketH.Mount(v2, authCasbinMW) // 用户侧挂 mw，管理接口内部叠加 AdminOnly
	}

	return r
}

func dbPing(db *gorm.DB) error {
	if db == nil {
		return nil
	}
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}

// initNotify 初始化邮件/短信驱动
func initNotify(cfg *config.Config) {
	// Mail
	switch cfg.Mail.Provider {
	case "smtp":
		notify.SetMailer(notify.NewSMTPMailer(
			cfg.Mail.Host, cfg.Mail.Port, cfg.Mail.Username, cfg.Mail.Password, cfg.Mail.From, cfg.Mail.FromName,
		))
	case "aliyun_directmail":
		notify.SetMailer(notify.NewAliyunDirectMailer(
			cfg.Mail.Username, cfg.Mail.Password, cfg.Mail.Host, cfg.Mail.From, cfg.Mail.FromName,
		))
	default:
		notify.SetMailer(notify.NewLogMailer())
	}
	// SMS
	switch cfg.SMS.Provider {
	case "aliyun":
		notify.SetSMSer(notify.NewAliyunSMSer(
			cfg.SMS.AccessKey, cfg.SMS.SecretKey, cfg.SMS.Sign, cfg.SMS.Region,
		))
	case "tencent":
		notify.SetSMSer(notify.NewTencentSMSer(
			cfg.SMS.AccessKey, cfg.SMS.SecretKey, cfg.SMS.Sign, "", cfg.SMS.Region,
		))
	case "twilio":
		// Twilio 需要 SID/Token/From 三项 - 这里简化：使用通用字段
		notify.SetSMSer(notify.NewTwilioSMSer(
			cfg.SMS.AccessKey, cfg.SMS.SecretKey, cfg.SMS.Sign,
		))
	default:
		notify.SetSMSer(notify.NewLogSMSer())
	}
}

// mountLocalStatic 将本地存储物理目录挂到 /uploads/*
func mountLocalStatic(r *gin.Engine, root string) {
	if root == "" {
		return
	}
	abs, err := filepath.Abs(root)
	if err != nil {
		abs = root
	}
	r.Static("/uploads", abs)
	logger.L.Info("local static mounted", zap.String("url", "/uploads"), zap.String("root", abs))
}
