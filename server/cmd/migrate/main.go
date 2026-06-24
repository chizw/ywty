// Package main 数据迁移 CLI（子命令模式）
// 用法：
//
//	go run ./cmd/migrate up       # 自动迁移 + 初始化种子（默认）
//	go run ./cmd/migrate down     # 删除所有表（危险！需确认）
//	go run ./cmd/migrate status   # 列出所有已创建的表
//	go run ./cmd/migrate seed     # 仅执行种子数据
//	go run ./cmd/migrate          # 无子命令等价于 up（向后兼容）
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/ywty/server/internal/config"
	"github.com/ywty/server/internal/database"
	"github.com/ywty/server/internal/logger"
	"github.com/ywty/server/internal/migrate"
	dbseed "github.com/ywty/server/internal/seed"
	"gorm.io/gorm"
)

func main() {
	// 解析子命令（向后兼容：无子命令时默认 up）
	sub := "up"
	if len(os.Args) >= 2 {
		switch os.Args[1] {
		case "up", "down", "status", "seed":
			sub = os.Args[1]
			// 移除子命令参数，让 flag.Parse 正常工作
			os.Args = append(os.Args[:1], os.Args[2:]...)
		}
	}
	flag.Parse()

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

	logger.L.Info("migrate start",
		zString("driver", cfg.Database.Driver),
		zString("command", sub),
	)

	db, err := database.New(cfg.Database)
	if err != nil {
		logger.L.Fatal("init database error", zapErr(err))
	}
	defer database.Close(db)
	if err := database.ConfigurePool(db, cfg.Database); err != nil {
		logger.L.Fatal("configure pool error", zapErr(err))
	}

	switch sub {
	case "up":
		runUp(db)
	case "down":
		runDown(db)
	case "status":
		runStatus(db)
	case "seed":
		runSeed(db)
	}
}

// runUp 执行自动迁移 + 种子（默认行为）
func runUp(db *gorm.DB) {
	if err := migrate.AutoMigrate(db); err != nil {
		logger.L.Fatal("auto migrate error", zapErr(err))
	}
	logger.L.Info("auto migrate success")
	if err := dbseed.Run(db); err != nil {
		logger.L.Fatal("seed error", zapErr(err))
	}
	logger.L.Info("seed success")
}

// runDown 删除所有表（需确认）
func runDown(db *gorm.DB) {
	if !confirm("即将删除所有表，此操作不可恢复！输入 yes 继续：") {
		logger.L.Warn("down cancelled")
		return
	}
	if err := dropAll(db); err != nil {
		logger.L.Fatal("drop tables error", zapErr(err))
	}
	logger.L.Warn("all tables dropped")
}

// runStatus 列出所有已创建的表
func runStatus(db *gorm.DB) {
	tables, err := db.Migrator().GetTables()
	if err != nil {
		logger.L.Fatal("get tables error", zapErr(err))
	}
	logger.L.Info("tables status", zInt("count", len(tables)))
	for _, t := range tables {
		fmt.Println("  -", t)
	}
}

// runSeed 仅执行种子数据
func runSeed(db *gorm.DB) {
	if err := dbseed.Run(db); err != nil {
		logger.L.Fatal("seed error", zapErr(err))
	}
	logger.L.Info("seed success")
}

// confirm 从 stdin 读取确认输入
func confirm(prompt string) bool {
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		return false
	}
	return strings.TrimSpace(strings.ToLower(line)) == "yes"
}

// dropAll 反向删除所有模型对应的表（按依赖倒序）
func dropAll(db *gorm.DB) error {
	models := migrate.AllModels()
	for i := len(models) - 1; i >= 0; i-- {
		if err := db.Migrator().DropTable(models[i]); err != nil {
			return err
		}
	}
	return nil
}
