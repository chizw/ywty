// Package install 程序安装流程：系统设置种子、默认角色组与储存策略、超级管理员。
package install

import (
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/chizw/ywty/server-go/internal/config"
	"github.com/chizw/ywty/server-go/internal/db"
	"github.com/chizw/ywty/server-go/internal/db/types"
	"github.com/chizw/ywty/server-go/internal/model"
	"github.com/chizw/ywty/server-go/internal/setting"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Options 安装参数（与原版安装接口字段一致）。
type Options struct {
	AppName       string
	AppURL        string
	LicenseKey    string
	AdminUsername string
	AdminEmail    string
	AdminPassword string
}

var usernamePattern = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
var emailPattern = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)

// Validate 等价 InstallRequest::rules()（db_* 字段仅做兼容校验，Go 版数据库连接由环境变量决定）。
func (o *Options) Validate() map[string][]string {
	errs := map[string][]string{}
	req := func(ok bool, field, attr string) {
		if !ok {
			errs[field] = append(errs[field], attr+" 不能为空。")
		}
	}
	req(o.AppName != "", "app_name", "应用名称")
	req(o.AppURL != "", "app_url", "应用 URL")
	req(o.LicenseKey != "", "app_license_key", "授权密钥")
	req(o.AdminUsername != "", "admin_username", "管理员用户名")
	req(o.AdminEmail != "", "admin_email", "管理员邮箱")
	req(o.AdminPassword != "", "admin_password", "管理员密码")
	if o.AdminUsername != "" && !usernamePattern.MatchString(o.AdminUsername) {
		errs["admin_username"] = append(errs["admin_username"], "管理员用户名 格式不正确。")
	}
	if o.AdminEmail != "" && !emailPattern.MatchString(o.AdminEmail) {
		errs["admin_email"] = append(errs["admin_email"], "管理员邮箱 必须是合法的邮箱。")
	}
	return errs
}

// verifyLicense 授权校验（当前版本恒通过）。
func verifyLicense(string, string) bool { return true }

// Run 执行安装。前提：schema 已就绪（db.Migrate 已执行）、未安装。
func Run(gdb *gorm.DB, cfg *config.Config, o Options) error {
	if db.IsInstalled(gdb, cfg) {
		return errors.New("程序已安装，如需重新安装请删除 installed.lock 后重试")
	}
	if !verifyLicense(o.LicenseKey, o.AppURL) {
		return errors.New("许可证验证失败，请检查应用域名、许可证密钥是否正确！")
	}

	appURL := o.AppURL
	if appURL == "" {
		appURL = cfg.AppURL
	}

	now := time.Now().UTC()

	// 1. 系统设置（license_key 明文存储）
	if _, err := setting.Seed(gdb, o.AppName, appURL, o.LicenseKey); err != nil {
		return fmt.Errorf("写入系统设置失败: %w", err)
	}

	// 2. 默认角色组
	allowTypes := []string{
		"jpg", "jpeg", "webp", "avif", "bmp", "gif", "png", "tif", "tiff",
		"jp2", "j2k", "jp2k", "jpf", "jpm", "jpg2", "j2c", "jpc", "jpx", "heic", "heif",
	}
	groupOptions, err := types.NewJSON(map[string]any{
		"max_upload_size":         5120,
		"file_expire_seconds":     0,
		"limit_concurrent_upload": 4,
		"limit_per_minute":        20,
		"limit_per_hour":          100,
		"limit_per_day":           300,
		"limit_per_week":          600,
		"limit_per_month":         1000,
		"allow_file_types":        allowTypes,
	})
	if err != nil {
		return err
	}
	group := model.Group{
		Name:      "系统默认组",
		Intro:     "这是系统默认角色组",
		Options:   groupOptions,
		IsDefault: true,
		IsGuest:   true,
	}
	if err := gdb.Create(&group).Error; err != nil {
		return fmt.Errorf("创建默认角色组失败: %w", err)
	}

	// 3. 本地储存策略
	storageOptions, err := types.NewJSON(map[string]any{
		"public_url":         appURL,
		"naming_rule":        "{Ymd}/{md5}",
		"generate_thumbnail": true,
		"thumbnail_max_size": 800,
		"thumbnail_quality":  90,
		"root":               cfg.UploadsDir,
	})
	if err != nil {
		return err
	}
	storage := model.Storage{
		Name:     "本地储存",
		Intro:    "这是本地储存驱动",
		Prefix:   "uploads",
		Provider: "local",
		Options:  storageOptions,
	}
	if err := gdb.Create(&storage).Error; err != nil {
		return fmt.Errorf("创建本地储存失败: %w", err)
	}
	if err := gdb.Create(&model.GroupStorage{GroupID: group.ID, StorageID: storage.ID}).Error; err != nil {
		return fmt.Errorf("关联角色组与储存失败: %w", err)
	}

	// 4. 默认页面
	pageContent := "关于我们"
	if err := gdb.Create(&model.Page{
		Type:    "internal",
		Name:    "关于我们",
		Icon:    "fa-users",
		Title:   "关于我们",
		Content: &pageContent,
		Slug:    "about",
		URL:     "",
		IsShow:  true,
	}).Error; err != nil {
		return fmt.Errorf("创建默认页面失败: %w", err)
	}

	// 5. 超级管理员 + 默认组 + 默认容量
	hash, err := bcrypt.GenerateFromPassword([]byte(o.AdminPassword), 12)
	if err != nil {
		return fmt.Errorf("加密密码失败: %w", err)
	}
	user := model.User{
		Name:            o.AdminUsername,
		Username:        o.AdminUsername,
		Email:           &o.AdminEmail,
		Password:        string(hash),
		IsAdmin:         true,
		Status:          "normal",
		EmailVerifiedAt: &now,
		Options: types.MustJSON(map[string]any{
			"language":                 "zh-CN",
			"show_original_photos":     false,
			"encode_copied_url":        true,
			"auto_upload_after_select": false,
		}),
	}
	if err := gdb.Create(&user).Error; err != nil {
		return fmt.Errorf("创建超级管理员失败: %w", err)
	}
	if err := gdb.Create(&model.UserGroup{
		UserID:  user.ID,
		GroupID: group.ID,
		From:    "system",
	}).Error; err != nil {
		return fmt.Errorf("分配用户角色组失败: %w", err)
	}
	if err := gdb.Create(&model.UserCapacity{
		UserID:   user.ID,
		Capacity: 1073741824, // 1T，单位 KB
		From:     "system",
	}).Error; err != nil {
		return fmt.Errorf("分配用户容量失败: %w", err)
	}

	return db.MarkInstalled(cfg)
}

// AutoInstallFromEnv 对齐原版容器启动安装：提供 APP_URL + APP_LICENSE_KEY +
// ADMIN_* 环境变量时自动安装；否则跳过并提示走 /api/v2/install。
func AutoInstallFromEnv(gdb *gorm.DB, cfg *config.Config) (bool, error) {
	if db.IsInstalled(gdb, cfg) {
		return false, nil
	}
	if cfg.LicenseKey == "" || cfg.AdminUsername == "" || cfg.AdminEmail == "" || cfg.AdminPassword == "" {
		return false, errors.New("程序未安装且缺少安装环境变量（APP_LICENSE_KEY、ADMIN_USERNAME、ADMIN_EMAIL、ADMIN_PASSWORD），可调用 POST /api/v2/install 完成安装")
	}
	o := Options{
		AppName:       cfg.AppName,
		AppURL:        cfg.AppURL,
		LicenseKey:    cfg.LicenseKey,
		AdminUsername: cfg.AdminUsername,
		AdminEmail:    cfg.AdminEmail,
		AdminPassword: cfg.AdminPassword,
	}
	if errs := o.Validate(); len(errs) > 0 {
		return false, fmt.Errorf("安装环境变量校验失败: %v", errs)
	}
	return true, Run(gdb, cfg, o)
}
