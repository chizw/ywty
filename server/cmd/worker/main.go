// Package main 队列 Worker 入口
// 启动 Asynq Server 消费缩略图、图片处理、自动删除等任务
package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/ywty/server/internal/config"
	"github.com/ywty/server/internal/database"
	"github.com/ywty/server/internal/jobs"
	"github.com/ywty/server/internal/logger"
	"github.com/ywty/server/internal/queue"
	"github.com/ywty/server/internal/storage"
)

func main() {
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

	logger.L.Info("worker starting",
		zString("concurrency", fmt.Sprintf("%d", cfg.Queue.Concurrency)),
		zString("queues", fmt.Sprintf("%v", cfg.Queue.Queues)),
	)

	db, err := database.New(cfg.Database)
	if err != nil {
		logger.L.Fatal("init database error", zapErr(err))
	}
	if err := database.ConfigurePool(db, cfg.Database); err != nil {
		logger.L.Fatal("configure pool error", zapErr(err))
	}
	if err := database.Ping(db); err != nil {
		logger.L.Fatal("ping database error", zapErr(err))
	}
	defer database.Close(db)
	logger.L.Info("database connected", zString("driver", cfg.Database.Driver))

	// 初始化存储驱动（Worker 也需要访问存储做缩略图/清理）
	driver, err := storage.Get(cfg.Storage.Driver.Provider, storage.Options{
		"root":       cfg.Storage.Driver.Root,
		"public_url": cfg.Storage.Driver.PublicURL,
		"visibility": cfg.Storage.Driver.Visibility,
	})
	if err != nil {
		logger.L.Fatal("init storage driver error", zapErr(err))
	}
	logger.L.Info("storage driver ready", zString("driver", driver.Name()))

	// 启动 asynq server
	srv := queue.NewServer(cfg.Redis, cfg.Queue.Concurrency, cfg.Queue.Queues)
	mux := queue.NewMux()
	(&jobs.Handlers{DB: db, Driver: driver}).Register(mux)

	// 优雅退出
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		logger.L.Info("worker shutting down...")
		srv.Shutdown()
	}()

	if err := srv.Run(mux); err != nil {
		logger.L.Fatal("worker run error", zapErr(err))
	}
	logger.L.Info("worker bye")
}
