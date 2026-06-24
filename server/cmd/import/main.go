// Package main 老数据导入工具
// 从 Laravel/MySQL 老库导入数据到新库
//
// 用法：
//
//	# 用 DSN 指定老库（需包含 parseTime=true 以正确解析时间字段）
//	go run ./cmd/import -source-dsn "user:pass@tcp(127.0.0.1:3306)/old_ywty?charset=utf8mb4&parseTime=true&loc=Local"
//
//	# 用配置文件指定老库
//	go run ./cmd/import -source-config ./configs/old.yaml
//
//	# 同时指定老库和新库 DSN
//	go run ./cmd/import -source-dsn "..." -target-dsn "user:pass@tcp(127.0.0.1:3306)/ywty?charset=utf8mb4&parseTime=true&loc=Local"
//
// 注意：
//   - 导入前请确保目标库已执行 `migrate up` 创建表结构
//   - 已存在的记录（按主键 ID 判断）会被跳过
//   - 老库的 created_at/updated_at 为 datetime 字符串，通过 DSN 的 parseTime=true 自动解析为 time.Time
//   - PhoneVerifiedAt/EmailVerifiedAt/ExpiredAt 等字段在新模型中为 *int64（unix 秒），老库需同为整型
package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/ywty/server/internal/config"
	"github.com/ywty/server/internal/database"
	"github.com/ywty/server/internal/logger"
	"github.com/ywty/server/internal/model"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Setting 设置表（老库通用 key-value 结构；若 schema 不同请调整）
type Setting struct {
	ID        uint64 `gorm:"primaryKey;autoIncrement"`
	Key       string `gorm:"column:key;size:255;not null"`
	Value     string `gorm:"type:text"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

// TableName 指定表名
func (Setting) TableName() string { return "settings" }

// stats 导入统计
type stats struct {
	name     string
	total    int
	imported int
	skipped  int
	err      error
}

func main() {
	sourceDSN := flag.String("source-dsn", "", "老库 MySQL DSN（需包含 parseTime=true）")
	sourceConfig := flag.String("source-config", "", "老库配置文件路径（YAML，使用其 database 段）")
	targetDSN := flag.String("target-dsn", "", "新库 DSN（默认用配置文件；指定时按 MySQL 处理）")
	batchSize := flag.Int("batch-size", 500, "批量读取/写入大小")
	flag.Parse()

	if *sourceDSN == "" && *sourceConfig == "" {
		fmt.Println("必须指定 -source-dsn 或 -source-config")
		flag.Usage()
		os.Exit(1)
	}

	// 加载新库配置（用于 target 默认连接 + 日志）
	cfg, err := config.Load()
	if err != nil {
		fmt.Println("load config error:", err)
		os.Exit(1)
	}
	if err := logger.Init(cfg.Log); err != nil {
		fmt.Println("init logger error:", err)
		os.Exit(1)
	}
	defer logger.Sync()

	logger.L.Info("import start",
		zString("source-dsn", maskDSN(*sourceDSN)),
		zString("source-config", *sourceConfig),
	)

	// 连接源库（老 Laravel MySQL）
	source, err := openSource(*sourceDSN, *sourceConfig)
	if err != nil {
		logger.L.Fatal("open source db error", zapErr(err))
	}
	defer database.Close(source)
	logger.L.Info("source db connected")

	// 连接目标库（新库）
	target, err := openTarget(*targetDSN, cfg)
	if err != nil {
		logger.L.Fatal("open target db error", zapErr(err))
	}
	defer database.Close(target)
	if err := database.ConfigurePool(target, cfg.Database); err != nil {
		logger.L.Fatal("configure target pool error", zapErr(err))
	}
	logger.L.Info("target db connected")

	// 按依赖顺序导入核心表
	allStats := []stats{
		importTable[model.User](source, target, "users", *batchSize),
		importTable[model.Group](source, target, "groups", *batchSize),
		importTable[model.Storage](source, target, "storages", *batchSize),
		importTable[model.Album](source, target, "albums", *batchSize),
		importTable[model.Photo](source, target, "photos", *batchSize),
		importTable[model.Tag](source, target, "tags", *batchSize),
		importTable[model.Share](source, target, "shares", *batchSize),
		importTable[Setting](source, target, "settings", *batchSize),
	}

	// 打印导入统计
	printStats(allStats)
}

// importTable 从源库读取数据导入目标库，跳过已存在的 ID
// 使用 Unscoped 读取包含软删除的全部记录，OnConflict DoNothing 跳过已存在主键
func importTable[T any](source, target *gorm.DB, name string, batchSize int) stats {
	s := stats{name: name}

	// 检查源表是否存在，不存在则跳过
	if !source.Migrator().HasTable(name) {
		logger.L.Warn("source table not found, skipping", zString("table", name))
		return s
	}

	var items []T
	result := source.Unscoped().FindInBatches(&items, batchSize, func(tx *gorm.DB, batch int) error {
		s.total += len(items)
		// 批量插入，冲突时跳过（基于主键 ID）
		cr := target.Clauses(clause.OnConflict{DoNothing: true}).Create(&items)
		if cr.Error != nil {
			return cr.Error
		}
		s.imported += int(cr.RowsAffected)
		s.skipped += len(items) - int(cr.RowsAffected)
		return nil
	})

	if result.Error != nil {
		s.err = result.Error
	}
	logger.L.Info("import table done",
		zString("table", name),
		zInt("total", s.total),
		zInt("imported", s.imported),
		zInt("skipped", s.skipped),
	)
	return s
}

// openSource 打开老库（优先用 DSN，其次用配置文件）
func openSource(dsn, configPath string) (*gorm.DB, error) {
	if dsn != "" {
		return gorm.Open(mysql.Open(dsn), &gorm.Config{
			DisableForeignKeyConstraintWhenMigrating: true,
		})
	}
	oldCfg, err := config.Load(configPath)
	if err != nil {
		return nil, fmt.Errorf("load source config: %w", err)
	}
	return database.New(oldCfg.Database)
}

// openTarget 打开新库（优先用 DSN，其次用配置文件）
func openTarget(dsn string, cfg *config.Config) (*gorm.DB, error) {
	if dsn != "" {
		return gorm.Open(mysql.Open(dsn), &gorm.Config{
			DisableForeignKeyConstraintWhenMigrating: true,
		})
	}
	return database.New(cfg.Database)
}

// printStats 打印导入统计表格
func printStats(all []stats) {
	fmt.Println()
	fmt.Println("导入统计:")
	fmt.Println("+------------+--------+----------+---------+")
	fmt.Println("| 表名       | 总数   | 已导入   | 已跳过  |")
	fmt.Println("+------------+--------+----------+---------+")
	for _, s := range all {
		fmt.Printf("| %-10s | %6d | %8d | %7d |\n", s.name, s.total, s.imported, s.skipped)
		if s.err != nil {
			logger.L.Error("import table error", zString("table", s.name), zapErr(s.err))
		}
	}
	fmt.Println("+------------+--------+----------+---------+")
}

// maskDSN 隐藏 DSN 中的密码部分用于日志输出
func maskDSN(dsn string) string {
	if dsn == "" {
		return ""
	}
	// 简单截断，避免日志泄露密码
	if len(dsn) > 40 {
		return dsn[:15] + "..." + dsn[len(dsn)-15:]
	}
	return dsn
}
