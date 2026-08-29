// ywty 单二进制：HTTP 服务 + 数据库队列 worker + 调度器（worker 见后续里程碑）。
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/chizw/ywty/server-go/internal/config"
	"github.com/chizw/ywty/server-go/internal/db"
	"github.com/chizw/ywty/server-go/internal/httpx"
	"github.com/chizw/ywty/server-go/internal/install"
)

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, nil)))

	cfg, err := config.Load()
	if err != nil {
		slog.Error("配置加载失败", "err", err)
		os.Exit(1)
	}

	gdb, err := db.Open(cfg)
	if err != nil {
		slog.Error("数据库连接失败", "err", err)
		os.Exit(1)
	}

	if err := db.Migrate(gdb, cfg.DBDriver); err != nil {
		slog.Error("数据库迁移失败", "err", err)
		os.Exit(1)
	}

	if installed, err := install.AutoInstallFromEnv(gdb, cfg); err != nil {
		slog.Warn("自动安装未执行", "err", err)
	} else if installed {
		slog.Info("程序安装完成")
	}

	srv := &http.Server{
		Addr:              cfg.Addr(),
		Handler:           httpx.New(cfg, gdb),
		ReadHeaderTimeout: 30 * time.Second,
	}

	go func() {
		slog.Info("服务启动", "addr", cfg.Addr(), "db", cfg.DBDriver, "app_url", cfg.AppURL)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("HTTP 服务异常退出", "err", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	slog.Info("收到退出信号，优雅关闭中...")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("优雅关闭失败", "err", err)
	}
}
