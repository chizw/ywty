// Package setting 管理 settings 表（spatie/laravel-settings 兼容存储）。
// payload 列存 JSON 编码后的值本身（如 "xxx" / true / 5120 / []），
// 加密字段存 Laravel Crypt 格式（见 internal/support/laracrypt）。
package setting

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/chizw/ywty/server-go/internal/db/types"
	"gorm.io/gorm"
)

const (
	GroupApp   = "app"
	GroupSite  = "site"
	GroupAdmin = "admin"
	GroupUser  = "user"
)

// 默认设置项与 database/settings/*.php 的种子数据一致。
// license_key 在安装时以 Laravel Crypt 格式加密写入。
func defaultEntries(appName, appURL, licenseEncrypted string) []entry {
	return []entry{
		{GroupApp, "name", appName},
		{GroupApp, "url", appURL},
		{GroupApp, "license_key", licenseEncrypted},
		{GroupApp, "timezone", "Asia/Shanghai"},
		{GroupApp, "locale", "zh_CN"},
		{GroupApp, "currency", "CNY"},
		{GroupApp, "icp_no", ""},
		{GroupApp, "ip_gain_method", "auto"},
		{GroupApp, "enable_registration", true},
		{GroupApp, "guest_upload", true},
		{GroupApp, "user_email_verify", true},
		{GroupApp, "user_phone_verify", false},
		{GroupApp, "mail_from_address", "hello@example.com"},
		{GroupApp, "mail_from_name", appName},
		// Intervention\Image\Drivers\Imagick\Driver —— 保留原值仅为数据兼容，
		// Go 版图像管线见 internal/imageproc
		{GroupApp, "image_driver", "Intervention\\Image\\Drivers\\Imagick\\Driver"},
		{GroupApp, "enable_site", true},
		{GroupApp, "enable_stat_api", false},
		{GroupApp, "enable_stat_api_key", ""},
		{GroupApp, "enable_explore", false},

		{GroupSite, "theme", "default"},
		{GroupSite, "title", appName},
		{GroupSite, "subtitle", "您的云上相册。"},
		{GroupSite, "homepage_title", appName},
		{GroupSite, "homepage_description", "Your photo album on the cloud."},
		{GroupSite, "notice", ""},
		{GroupSite, "homepage_background_image_url", ""},
		{GroupSite, "homepage_background_images", []any{}},
		{GroupSite, "auth_page_background_image_url", ""},
		{GroupSite, "auth_page_background_images", []any{}},
		{GroupSite, "custom_css", ""},
		{GroupSite, "custom_js", ""},

		{GroupAdmin, "top_navigation", false},
		{GroupAdmin, "primary_color", "sky"},
		{GroupAdmin, "dark_mode", false},
		{GroupAdmin, "default_theme_mode", "system"},

		{GroupUser, "initial_capacity", 5120},
	}
}

type entry struct {
	Group   string
	Name    string
	Payload any
}

// Seed 幂等写入默认设置（已存在的不覆盖），返回是否新建了任何行。
// 未安装的库（settings 表为空）执行后即达到 PHP 版安装完成时的设置状态。
func Seed(gdb *gorm.DB, appName, appURL, licenseEncrypted string) (bool, error) {
	created := false
	for _, e := range defaultEntries(appName, appURL, licenseEncrypted) {
		payload, err := types.NewJSON(e.Payload)
		if err != nil {
			return created, err
		}
		var count int64
		if err := gdb.Table("settings").
			Where("`group` = ? AND name = ?", e.Group, e.Name).
			Count(&count).Error; err != nil {
			return created, err
		}
		if count > 0 {
			continue
		}
		if err := gdb.Exec(
			"INSERT INTO `settings` (`group`, `name`, `locked`, `payload`, `created_at`, `updated_at`) VALUES (?, ?, 0, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)",
			e.Group, e.Name, payload,
		).Error; err != nil {
			return created, err
		}
		created = true
	}
	return created, nil
}

// ---------- 读取 ----------

// raw 取原始 payload 文本。
func raw(gdb *gorm.DB, group, name string) (string, bool, error) {
	var payload *string
	err := gdb.Raw("SELECT `payload` FROM `settings` WHERE `group` = ? AND `name` = ?", group, name).
		Scan(&payload).Error
	if err != nil {
		return "", false, err
	}
	if payload == nil {
		return "", false, nil
	}
	return *payload, true, nil
}

func String(gdb *gorm.DB, group, name string) (string, error) {
	p, ok, err := raw(gdb, group, name)
	if err != nil || !ok {
		return "", err
	}
	var v string
	if err := json.Unmarshal([]byte(p), &v); err != nil {
		return "", fmt.Errorf("setting: %s.%s: %w", group, name, err)
	}
	return v, nil
}

func Bool(gdb *gorm.DB, group, name string) (bool, error) {
	p, ok, err := raw(gdb, group, name)
	if err != nil || !ok {
		return false, err
	}
	var v bool
	if err := json.Unmarshal([]byte(p), &v); err != nil {
		return false, fmt.Errorf("setting: %s.%s: %w", group, name, err)
	}
	return v, nil
}

func Int64(gdb *gorm.DB, group, name string) (int64, error) {
	p, ok, err := raw(gdb, group, name)
	if err != nil || !ok {
		return 0, err
	}
	var v int64
	if err := json.Unmarshal([]byte(p), &v); err != nil {
		return 0, fmt.Errorf("setting: %s.%s: %w", group, name, err)
	}
	return v, nil
}

// Set 写入（upsert）单条设置。
func Set(gdb *gorm.DB, group, name string, payload any) error {
	p, err := types.NewJSON(payload)
	if err != nil {
		return err
	}
	res := gdb.Exec(
		"UPDATE `settings` SET `payload` = ?, `updated_at` = CURRENT_TIMESTAMP WHERE `group` = ? AND `name` = ?",
		p, group, name,
	)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gdb.Exec(
			"INSERT INTO `settings` (`group`, `name`, `locked`, `payload`, `created_at`, `updated_at`) VALUES (?, ?, 0, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)",
			group, name, p,
		).Error
	}
	return nil
}

// ErrNotFound 表示设置项不存在。
var ErrNotFound = errors.New("setting: 未找到")
