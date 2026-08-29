package db_test

import (
	"path/filepath"
	"testing"

	"github.com/chizw/ywty/server-go/internal/config"
	"github.com/chizw/ywty/server-go/internal/db"
	"github.com/chizw/ywty/server-go/internal/setting"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// newTestDB 打开一个临时 SQLite 库并完成迁移。
func newTestDB(t *testing.T) (*gorm.DB, *config.Config) {
	t.Helper()
	dir := testDir(t)
	cfg := &config.Config{
		DataDir:    dir,
		DBDriver:   "sqlite",
		DBPath:     filepath.Join(dir, "test.db"),
		UploadsDir: filepath.Join(dir, "uploads"),
	}
	gdb, err := gorm.Open(sqlite.Open(cfg.DBPath+"?_pragma=foreign_keys(1)"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() {
		if sqlDB, err := gdb.DB(); err == nil {
			_ = sqlDB.Close()
		}
	})
	if err := db.Migrate(gdb, "sqlite"); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return gdb, cfg
}

func TestMigrateCreatesAllTables(t *testing.T) {
	gdb, _ := newTestDB(t)

	wantTables := []string{
		"cache", "cache_locks", "jobs", "job_batches", "failed_jobs",
		"personal_access_tokens", "settings", "groups", "drivers", "storages",
		"group_storage", "users", "group_driver", "storage_driver",
		"password_reset_tokens", "sessions", "oauth", "albums", "photos",
		"album_photo", "tags", "taggables", "shares", "shareables",
		"violations", "notices", "pages", "feedbacks", "tickets",
		"ticket_replies", "reports", "likes", "plans", "plan_prices",
		"plan_groups", "plan_capacities", "coupons", "orders",
		"user_groups", "user_capacities", "migrations",
	}
	for _, table := range wantTables {
		var one int
		if err := gdb.Raw("SELECT 1 FROM `" + table + "` LIMIT 1").Scan(&one).Error; err != nil {
			t.Errorf("表 %s 不存在: %v", table, err)
		}
	}

	var batchCount int64
	gdb.Raw("SELECT count(*) FROM `migrations` WHERE `batch` = 1").Scan(&batchCount)
	if batchCount != 35 {
		t.Errorf("migrations 记录数 = %d, want 35", batchCount)
	}
}

func TestMigrateIdempotent(t *testing.T) {
	gdb, _ := newTestDB(t)
	if err := db.Migrate(gdb, "sqlite"); err != nil {
		t.Fatalf("二次 migrate 不应报错: %v", err)
	}
	var batchCount int64
	gdb.Raw("SELECT count(*) FROM `migrations` WHERE `batch` = 1").Scan(&batchCount)
	if batchCount != 35 {
		t.Errorf("重复迁移后 migrations 记录数 = %d, want 35", batchCount)
	}
}

func TestIsInstalled(t *testing.T) {
	gdb, cfg := newTestDB(t)
	if db.IsInstalled(gdb, cfg) {
		t.Fatal("空库（无 users 数据）不应视为已安装")
	}
	if err := db.MarkInstalled(cfg); err != nil {
		t.Fatal(err)
	}
	if !db.IsInstalled(gdb, cfg) {
		t.Fatal("写入 installed.lock 后应视为已安装")
	}
	if !db.AgreementAgreed(cfg) {
		t.Fatal("非空 installed.lock 应视为已同意协议")
	}
}

func TestSettingRoundtrip(t *testing.T) {
	gdb, _ := newTestDB(t)
	if _, err := setting.Seed(gdb, "测试图床", "https://img.example.com", "ENC"); err != nil {
		t.Fatalf("seed: %v", err)
	}

	title, err := setting.String(gdb, setting.GroupSite, "title")
	if err != nil || title != "测试图床" {
		t.Fatalf("site.title = %q, err = %v", title, err)
	}
	v, err := setting.Bool(gdb, setting.GroupApp, "enable_registration")
	if err != nil || !v {
		t.Fatalf("app.enable_registration = %v, err = %v", v, err)
	}
	n, err := setting.Int64(gdb, setting.GroupUser, "initial_capacity")
	if err != nil || n != 5120 {
		t.Fatalf("user.initial_capacity = %d, err = %v", n, err)
	}

	if err := setting.Set(gdb, setting.GroupApp, "enable_site", false); err != nil {
		t.Fatal(err)
	}
	v, _ = setting.Bool(gdb, setting.GroupApp, "enable_site")
	if v {
		t.Fatal("Set 之后应读到 false")
	}

	// 二次 Seed 不覆盖
	if _, err := setting.Seed(gdb, "另一个名字", "http://x", "E2"); err != nil {
		t.Fatal(err)
	}
	title, _ = setting.String(gdb, setting.GroupSite, "title")
	if title != "测试图床" {
		t.Fatalf("重复 Seed 不应覆盖已有值，got %q", title)
	}
}
