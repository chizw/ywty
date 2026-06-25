// Package main API 服务入口
package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ywty/server/internal/config"
	"github.com/ywty/server/internal/database"
	"github.com/ywty/server/internal/logger"
	"github.com/ywty/server/internal/queue"
	"github.com/ywty/server/internal/router"
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

	logger.L.Info("starting 云雾图驿 api server",
		zString("app", cfg.App.Name),
		zString("env", cfg.App.Env),
		zInt("port", cfg.App.Port),
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
	logger.L.Info("database connected", zString("driver", cfg.Database.Driver))

	// 初始化默认存储驱动
	driver, err := storage.Get(cfg.Storage.Driver.Provider, storage.Options{
		"root":       cfg.Storage.Driver.Root,
		"public_url": cfg.Storage.Driver.PublicURL,
		"visibility": cfg.Storage.Driver.Visibility,
	})
	if err != nil {
		logger.L.Fatal("init storage driver error", zapErr(err))
	}
	logger.L.Info("storage driver ready", zString("driver", driver.Name()))

	// 初始化队列客户端（异步任务：缩略图/处理/自动删除）
	// Redis 为空时跳过队列（纯 SQLite 模式下队列功能不可用）
	var qClient *queue.Client
	if cfg.Redis.Addr != "" {
		qClient = queue.NewClient(cfg.Redis)
		defer qClient.Close()
		logger.L.Info("queue client ready", zString("addr", cfg.Redis.Addr))
	} else {
		logger.L.Info("redis not configured, queue disabled")
	}

	r := router.New(&router.Options{
		Cfg:         cfg,
		DB:          db,
		StoreDriver: driver,
		Queue:       qClient,
	})

	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.App.Host, cfg.App.Port),
		Handler:      r,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		logger.L.Info("listening", zString("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.L.Fatal("server error", zapErr(err))
		}
	}()

	// 优雅退出
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.L.Info("shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.L.Fatal("shutdown error", zapErr(err))
	}
	_ = database.Close(db)
	logger.L.Info("bye")
}
