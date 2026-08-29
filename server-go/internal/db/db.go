// Package db 负责 GORM 连接建立、schema 迁移与安装状态判断。
package db

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/chizw/ywty/server-go/internal/config"
	"github.com/glebarez/sqlite"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Open 按配置建立数据库连接。
// MySQL 会话时区固定 +00:00（与原版数据库会话一致）；
// SQLite 开启 WAL、外键与 busy_timeout，保证并发写可靠。
func Open(cfg *config.Config) (*gorm.DB, error) {
	var dial gorm.Dialector

	switch cfg.DBDriver {
	case "mysql":
		dsn := fmt.Sprintf(
			"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=UTC&time_zone=%s&interpolateParams=true",
			url.QueryEscape(cfg.DBUser), url.QueryEscape(cfg.DBPassword),
			cfg.DBHost, cfg.DBPort, url.QueryEscape(cfg.DBName),
			url.QueryEscape("'+00:00'"),
		)
		dial = mysql.Open(dsn)
	default: // sqlite
		if dir := filepath.Dir(cfg.DBPath); dir != "" && dir != "." {
			if err := os.MkdirAll(dir, 0o755); err != nil {
				return nil, fmt.Errorf("db: 无法创建数据目录: %w", err)
			}
		}
		dsn := cfg.DBPath +
			"?_pragma=foreign_keys(1)" +
			"&_pragma=journal_mode(WAL)" +
			"&_pragma=busy_timeout(10000)" +
			"&_pragma=synchronous(NORMAL)"
		dial = sqlite.Open(dsn)
	}

	gdb, err := gorm.Open(dial, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})
	if err != nil {
		return nil, fmt.Errorf("db: 连接失败: %w", err)
	}

	sqlDB, err := gdb.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(50)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(time.Hour)

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("db: ping 失败: %w", err)
	}
	return gdb, nil
}

// IsInstalled 判断程序是否已安装：installed.lock 存在（Go 版写在 DATA_DIR，
// 兼容原版放在运行目录），或 users 表已有数据（原版安装过的库必含管理员行；
// 仅建表不算安装，Go 版全新安装流程先建表后种子数据）。
func IsInstalled(gdb *gorm.DB, cfg *config.Config) bool {
	for _, p := range []string{
		filepath.Join(cfg.DataDir, "installed.lock"),
		"installed.lock",
	} {
		if _, err := os.Stat(p); err == nil {
			return true
		}
	}
	if !tableExists(gdb, "users") {
		return false
	}
	var count int64
	gdb.Raw("SELECT count(*) FROM `users`").Scan(&count)
	return count > 0
}

// MarkInstalled 写入安装完成标记。
func MarkInstalled(cfg *config.Config) error {
	if err := os.MkdirAll(cfg.DataDir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(cfg.DataDir, "installed.lock"), []byte("agreed"), 0o644)
}

// AgreementAgreed 判断协议同意状态：installed.lock 非空即已同意。
func AgreementAgreed(cfg *config.Config) bool {
	for _, p := range []string{
		filepath.Join(cfg.DataDir, "installed.lock"),
		"installed.lock",
	} {
		if b, err := os.ReadFile(p); err == nil && len(b) > 0 {
			return true
		}
	}
	return false
}

func tableExists(gdb *gorm.DB, table string) bool {
	var count int64
	switch gdb.Dialector.Name() {
	case "sqlite":
		gdb.Raw(`SELECT count(*) FROM sqlite_master WHERE type = 'table' AND name = ?`, table).Scan(&count)
	default:
		gdb.Raw(`SELECT count(*) FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = ?`).
			Scan(&count)
	}
	return count > 0
}
