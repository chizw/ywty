package db

import (
	_ "embed"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

//go:embed migrations/mysql/0001_init.sql
var mysqlInitSQL string

//go:embed migrations/sqlite/0001_init.sql
var sqliteInitSQL string

// laravelMigrations 是 database/migrations/ 的文件名清单（按执行顺序）。
// Go 版建库后会把这些名字写入 Laravel 风格的 migrations 表（batch 1），
// 保证 PHP 侧工具链（migrate:status 等）视角一致。
var laravelMigrations = []string{
	"0001_01_01_000001_create_cache_table.php",
	"0001_01_01_000002_create_jobs_table.php",
	"0001_01_01_000003_create_personal_access_tokens_table.php",
	"2022_12_14_083707_create_settings_table.php",
	"2024_04_24_161046_create_groups_table.php",
	"2024_04_24_161047_create_drivers_table.php",
	"2024_04_24_161050_create_storages_table.php",
	"2024_04_24_172030_create_group_storage_table.php",
	"2024_04_24_172040_create_users_table.php",
	"2024_04_24_172050_create_group_driver_table.php",
	"2024_04_24_172051_create_storage_driver_table.php",
	"2024_04_24_172200_create_albums_table.php",
	"2024_04_24_172219_create_photos_table.php",
	"2024_04_24_172220_create_album_photo_table.php",
	"2024_04_24_172221_create_tags_table.php",
	"2024_04_24_172222_create_taggables_table.php",
	"2024_04_24_172223_create_shares_table.php",
	"2024_04_24_172230_create_shareables_table.php",
	"2024_04_24_172337_create_violations_table.php",
	"2024_04_24_172555_create_notices_table.php",
	"2024_04_24_172714_create_pages_table.php",
	"2024_04_24_172745_create_feedbacks_table.php",
	"2024_04_24_172823_create_tickets_table.php",
	"2024_04_24_172919_create_ticket_replies_table.php",
	"2024_04_24_173007_create_reports_table.php",
	"2024_04_24_205207_create_likes_table.php",
	"2024_04_24_225049_create_plans_table.php",
	"2024_04_24_225050_create_plan_prices_table.php",
	"2024_04_24_225060_create_plan_groups_table.php",
	"2024_04_24_225300_create_plan_capacities_table.php",
	"2024_04_24_225418_create_coupons_table.php",
	"2024_04_24_225450_create_orders_table.php",
	"2024_04_24_225451_create_user_groups_table.php",
	"2024_04_24_225452_create_user_capacities_table.php",
	"2025_04_29_172035_modify_users_email_field.php",
}

// Migrate 初始化 schema：
//   - users 表已存在（PHP 版安装过 / Go 版装过）→ 跳过建表，仅做兼容性修补；
//   - 否则执行 0001_init.sql 全量建表，并写入 Laravel 风格 migrations 记录。
func Migrate(gdb *gorm.DB, driver string) error {
	if tableExists(gdb, "users") {
		if driver == "mysql" {
			// 兼容 2025_04_29 之前安装的 PHP 版：users.email 当时是 NOT NULL
			return ensureEmailNullable(gdb)
		}
		return nil
	}

	switch driver {
	case "mysql":
		if err := execStatements(gdb, mysqlInitSQL); err != nil {
			return fmt.Errorf("migrate: mysql 初始化失败: %w", err)
		}
	default:
		if err := execStatements(gdb, sqliteInitSQL); err != nil {
			return fmt.Errorf("migrate: sqlite 初始化失败: %w", err)
		}
	}
	return seedMigrationsTable(gdb, driver)
}

// ensureEmailNullable 对旧 PHP 库做 users.email 可空修补（对齐 2025_04_29 迁移后的状态）。
func ensureEmailNullable(gdb *gorm.DB) error {
	var nullable string
	err := gdb.Raw(
		`SELECT IS_NULLABLE FROM information_schema.COLUMNS
		 WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'users' AND COLUMN_NAME = 'email'`,
	).Scan(&nullable).Error
	if err != nil {
		return nil //nolint:nilerr // 查询失败不阻断启动
	}
	if strings.EqualFold(nullable, "NO") {
		if err := gdb.Exec("ALTER TABLE `users` MODIFY `email` VARCHAR(255) NULL COMMENT '邮箱'").Error; err != nil {
			return fmt.Errorf("migrate: users.email 修补失败: %w", err)
		}
	}
	return nil
}

func seedMigrationsTable(gdb *gorm.DB, driver string) error {
	if !tableExists(gdb, "migrations") {
		var stmt string
		if driver == "mysql" {
			stmt = "CREATE TABLE `migrations` (" +
				"`id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY, " +
				"`migration` VARCHAR(255) NOT NULL, `batch` INT NOT NULL) " +
				"ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci"
		} else {
			stmt = "CREATE TABLE `migrations` (" +
				"`id` INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL, " +
				"`migration` VARCHAR(255) NOT NULL, `batch` INTEGER NOT NULL)"
		}
		if err := gdb.Exec(stmt).Error; err != nil {
			return fmt.Errorf("migrate: 创建 migrations 表失败: %w", err)
		}
	}
	for _, name := range laravelMigrations {
		if err := gdb.Exec(
			"INSERT INTO `migrations` (`migration`, `batch`) VALUES (?, 1)", name,
		).Error; err != nil {
			return fmt.Errorf("migrate: 写入 migrations 记录失败: %w", err)
		}
	}
	return nil
}

// execStatements 把 SQL 文本按语句切分后逐条执行（database/sql 单次 Exec 不支持多语句）。
// 切分规则：单引号字符串外部的分号；本 schema 不含需要转义的引号内容。
func execStatements(gdb *gorm.DB, raw string) error {
	for _, stmt := range splitStatements(raw) {
		if err := gdb.Exec(stmt).Error; err != nil {
			return fmt.Errorf("执行语句失败 [%s...]: %w", truncate(stmt, 60), err)
		}
	}
	return nil
}

func splitStatements(raw string) []string {
	var (
		stmts   []string
		current strings.Builder
		inStr   bool
	)
	for i := 0; i < len(raw); i++ {
		c := raw[i]
		switch {
		case c == '\'':
			inStr = !inStr
			current.WriteByte(c)
		case c == ';' && !inStr:
			s := strings.TrimSpace(current.String())
			if s != "" {
				stmts = append(stmts, s)
			}
			current.Reset()
		default:
			current.WriteByte(c)
		}
	}
	if s := strings.TrimSpace(current.String()); s != "" {
		stmts = append(stmts, s)
	}
	return stmts
}

func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n])
}
