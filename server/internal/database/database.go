// Package database 多驱动 DB 工厂（MySQL / SQLite）
package database

import (
	"fmt"
	"path/filepath"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	// 纯 Go SQLite 驱动（无需 CGO）
	"github.com/glebarez/sqlite"

	"github.com/ywty/server/internal/config"
)

// New 根据配置创建 GORM DB 实例
func New(cfg config.Database) (*gorm.DB, error) {
	gormCfg := &gorm.Config{
		Logger:                                   logger.Default.LogMode(logger.Warn),
		PrepareStmt:                              true,
		DisableForeignKeyConstraintWhenMigrating: true,
		NowFunc: func() time.Time {
			return time.Now().Local()
		},
	}

	switch cfg.Driver {
	case "mysql":
		dsn := cfg.DSN()
		db, err := gorm.Open(mysql.Open(dsn), gormCfg)
		if err != nil {
			return nil, fmt.Errorf("open mysql: %w", err)
		}
		return db, nil

	case "sqlite":
		path := cfg.Path
		if path == "" {
			path = "storage/dev.db"
		}
		// 确保父目录存在
		_ = mkdirAll(filepath.Dir(path))
		db, err := gorm.Open(sqlite.Open(path+"?_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)"), gormCfg)
		if err != nil {
			return nil, fmt.Errorf("open sqlite: %w", err)
		}
		return db, nil

	default:
		return nil, fmt.Errorf("unsupported database driver: %s", cfg.Driver)
	}
}

// ConfigurePool 应用连接池配置
func ConfigurePool(db *gorm.DB, cfg config.Database) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	if cfg.MaxOpenConns > 0 {
		sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	}
	if cfg.MaxIdleConns > 0 {
		sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	}
	if cfg.ConnMaxLifetime > 0 {
		sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetimeDuration())
	}
	if cfg.ConnMaxIdleTime > 0 {
		sqlDB.SetConnMaxIdleTime(cfg.ConnMaxIdleTimeDuration())
	}
	return nil
}

// Ping 健康检查
func Ping(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}

// Close 关闭连接
func Close(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
