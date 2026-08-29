-- ============================================================================
-- ywty (ywty) 初始 schema · SQLite 版
-- 列名/约束与 database/migrations/*.php（Laravel 11）最终状态一致；
-- 类型按 Laravel sqlite 语法映射：string→TEXT, json→TEXT, boolean→TINYINT(1),
-- decimal→NUMERIC, timestamp→TIMESTAMP。
-- ============================================================================

CREATE TABLE IF NOT EXISTS "cache" (
    "key" VARCHAR(255) NOT NULL PRIMARY KEY,
    "value" MEDIUMTEXT NOT NULL,
    "expiration" INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS "cache_locks" (
    "key" VARCHAR(255) NOT NULL PRIMARY KEY,
    "owner" VARCHAR(255) NOT NULL,
    "expiration" INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS "jobs" (
    "id" INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
    "queue" VARCHAR(255) NOT NULL,
    "payload" LONGTEXT NOT NULL,
    "attempts" TINYINT UNSIGNED NOT NULL,
    "reserved_at" INTEGER UNSIGNED NULL,
    "available_at" INTEGER UNSIGNED NOT NULL,
    "created_at" INTEGER UNSIGNED NOT NULL
);
CREATE INDEX IF NOT EXISTS "jobs_queue_index" ON "jobs" ("queue");

CREATE TABLE IF NOT EXISTS "job_batches" (
    "id" VARCHAR(255) NOT NULL PRIMARY KEY,
    "name" VARCHAR(255) NOT NULL,
    "total_jobs" INTEGER NOT NULL,
    "pending_jobs" INTEGER NOT NULL,
    "failed_jobs" INTEGER NOT NULL,
    "failed_job_ids" LONGTEXT NOT NULL,
    "options" MEDIUMTEXT NULL,
    "cancelled_at" INTEGER NULL,
    "created_at" INTEGER NOT NULL,
    "finished_at" INTEGER NULL
);

CREATE TABLE IF NOT EXISTS "failed_jobs" (
    "id" INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
    "uuid" VARCHAR(255) NOT NULL UNIQUE,
    "connection" TEXT NOT NULL,
    "queue" TEXT NOT NULL,
    "payload" LONGTEXT NOT NULL,
    "exception" LONGTEXT NOT NULL,
    "failed_at" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS "personal_access_tokens" (
    "id" INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
    "tokenable_type" VARCHAR(255) NOT NULL,
    "tokenable_id" INTEGER UNSIGNED NOT NULL,
    "name" VARCHAR(255) NOT NULL,
    "token" VARCHAR(64) NOT NULL UNIQUE,
    "abilities" TEXT NULL,
    "last_used_at" TIMESTAMP NULL,
    "expires_at" TIMESTAMP NULL,
    "created_at" TIMESTAMP NULL,
    "updated_at" TIMESTAMP NULL
);
CREATE INDEX IF NOT EXISTS "personal_access_tokens_tokenable_type_tokenable_id_index"
    ON "personal_access_tokens" ("tokenable_type", "tokenable_id");

CREATE TABLE IF NOT EXISTS "settings" (
    "id" INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
    "group" VARCHAR(255) NOT NULL,
    "name" VARCHAR(255) NOT NULL,
    "locked" TINYINT(1) NOT NULL DEFAULT 0,
    "payload" JSON NOT NULL,
    "created_at" TIMESTAMP NULL,
    "updated_at" TIMESTAMP NULL,
    UNIQUE ("group", "name")
);

CREATE TABLE IF NOT EXISTS "groups" (
    "id" INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
    "name" VARCHAR(255) NOT NULL,
    "intro" VARCHAR(2000) NOT NULL DEFAULT '',
    "options" JSON NULL,
    "is_default" TINYINT(1) NOT NULL DEFAULT 0,
    "is_guest" TINYINT(1) NOT NULL DEFAULT 0,
    "deleted_at" TIMESTAMP NULL,
    "created_at" TIMESTAMP NULL,
    "updated_at" TIMESTAMP NULL
);

CREATE TABLE IF NOT EXISTS "drivers" (
    "id" INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
    "type" VARCHAR(255) NOT NULL,
    "name" VARCHAR(255) NOT NULL,
    "intro" VARCHAR(2000) NOT NULL DEFAULT '',
    "options" JSON NULL,
    "deleted_at" TIMESTAMP NULL,
    "created_at" TIMESTAMP NULL,
    "updated_at" TIMESTAMP NULL
);

CREATE TABLE IF NOT EXISTS "storages" (
    "id" INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
    "name" VARCHAR(255) NOT NULL,
    "intro" VARCHAR(2000) NOT NULL DEFAULT '',
    "prefix" VARCHAR(1000) NOT NULL DEFAULT '',
    "provider" VARCHAR(64) NOT NULL,
    "options" JSON NULL,
    "deleted_at" TIMESTAMP NULL,
    "created_at" TIMESTAMP NULL,
    "updated_at" TIMESTAMP NULL
);

CREATE TABLE IF NOT EXISTS "group_storage" (
    "group_id" INTEGER UNSIGNED NOT NULL,
    "storage_id" INTEGER UNSIGNED NOT NULL,
    "sort" INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY ("group_id", "storage_id"),
    FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON DELETE CASCADE,
    FOREIGN KEY ("storage_id") REFERENCES "storages" ("id") ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS "users" (
    "id" INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
    "avatar" VARCHAR(255) NOT NULL DEFAULT '',
    "name" VARCHAR(255) NOT NULL DEFAULT '',
    "username" VARCHAR(255) NOT NULL,
    "phone" VARCHAR(64) NULL UNIQUE,
    "email" VARCHAR(255) NULL UNIQUE,
    "password" VARCHAR(255) NOT NULL,
    "location" VARCHAR(64) NOT NULL DEFAULT '',
    "url" VARCHAR(255) NOT NULL DEFAULT '',
    "company" VARCHAR(128) NOT NULL DEFAULT '',
    "company_title" VARCHAR(128) NOT NULL DEFAULT '',
    "tagline" VARCHAR(255) NOT NULL DEFAULT '',
    "bio" VARCHAR(255) NOT NULL DEFAULT '',
    "interests" JSON NULL,
    "socials" JSON NULL,
    "phone_verified_at" TIMESTAMP NULL,
    "email_verified_at" TIMESTAMP NULL,
    "remember_token" VARCHAR(100) NULL,
    "is_admin" TINYINT(1) NOT NULL DEFAULT 0,
    "options" JSON NULL,
    "login_ip" VARCHAR(45) NULL,
    "register_ip" VARCHAR(45) NULL,
    "country_code" VARCHAR(32) NULL,
    "status" VARCHAR(64) NOT NULL DEFAULT 'normal',
    "deleted_at" TIMESTAMP NULL,
    "created_at" TIMESTAMP NULL,
    "updated_at" TIMESTAMP NULL
);

CREATE TABLE IF NOT EXISTS "group_driver" (
    "type" VARCHAR(32) NOT NULL,
    "group_id" INTEGER UNSIGNED NOT NULL,
    "driver_id" INTEGER UNSIGNED NOT NULL,
    "sort" INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY ("group_id", "driver_id"),
    FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON DELETE CASCADE,
    FOREIGN KEY ("driver_id") REFERENCES "drivers" ("id") ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS "storage_driver" (
    "type" VARCHAR(32) NOT NULL,
    "storage_id" INTEGER UNSIGNED NOT NULL,
    "driver_id" INTEGER UNSIGNED NOT NULL,
    "sort" INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY ("storage_id", "driver_id"),
    FOREIGN KEY ("storage_id") REFERENCES "storages" ("id") ON DELETE CASCADE,
    FOREIGN KEY ("driver_id") REFERENCES "drivers" ("id") ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS "password_reset_tokens" (
    "email" VARCHAR(255) NOT NULL PRIMARY KEY,
    "token" VARCHAR(255) NOT NULL,
    "created_at" TIMESTAMP NULL
);

CREATE TABLE IF NOT EXISTS "sessions" (
    "id" VARCHAR(255) NOT NULL PRIMARY KEY,
    "user_id" INTEGER UNSIGNED NULL,
    "ip_address" VARCHAR(45) NULL,
    "user_agent" TEXT NULL,
    "payload" LONGTEXT NOT NULL,
    "last_activity" INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS "sessions_user_id_index" ON "sessions" ("user_id");
CREATE INDEX IF NOT EXISTS "sessions_last_activity_index" ON "sessions" ("last_activity");

CREATE TABLE IF NOT EXISTS "oauth" (
    "id" INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
    "driver_id" INTEGER UNSIGNED NOT NULL,
    "user_id" INTEGER UNSIGNED NOT NULL,
    "openid" VARCHAR(255) NOT NULL,
    "avatar" VARCHAR(512) NOT NULL DEFAULT '',
    "email" VARCHAR(255) NOT NULL DEFAULT '',
    "name" VARCHAR(255) NOT NULL DEFAULT '',
    "nickname" VARCHAR(255) NOT NULL DEFAULT '',
    "raw" JSON NULL,
    "created_at" TIMESTAMP NULL,
    "updated_at" TIMESTAMP NULL,
    FOREIGN KEY ("driver_id") REFERENCES "drivers" ("id") ON DELETE CASCADE,
    FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS "albums" (
    "id" INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
    "user_id" INTEGER UNSIGNED NULL,
    "name" VARCHAR(255) NOT NULL DEFAULT '',
    "intro" VARCHAR(2000) NOT NULL DEFAULT '',
    "is_public" TINYINT(1) NOT NULL DEFAULT 0,
    "deleted_at" TIMESTAMP NULL,
    "created_at" TIMESTAMP NULL,
    "updated_at" TIMESTAMP NULL,
    FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS "photos" (
    "id" INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
    "user_id" INTEGER UNSIGNED NULL,
    "group_id" INTEGER UNSIGNED NULL,
    "storage_id" INTEGER UNSIGNED NULL,
    "name" VARCHAR(255) NOT NULL,
    "intro" VARCHAR(2000) NOT NULL DEFAULT '',
    "filename" VARCHAR(255) NOT NULL,
    "pathname" VARCHAR(255) NOT NULL,
    "mimetype" VARCHAR(64) NOT NULL DEFAULT '',
    "extension" VARCHAR(32) NOT NULL DEFAULT '',
    "md5" VARCHAR(32) NOT NULL DEFAULT '',
    "sha1" VARCHAR(255) NOT NULL DEFAULT '',
    "exif" JSON NULL,
    "size" NUMERIC NOT NULL DEFAULT 0,
    "width" INTEGER UNSIGNED NOT NULL DEFAULT 0,
    "height" INTEGER UNSIGNED NOT NULL DEFAULT 0,
    "is_public" TINYINT(1) NOT NULL DEFAULT 0,
    "status" VARCHAR(64) NOT NULL DEFAULT 'normal',
    "ip_address" VARCHAR(45) NULL,
    "expired_at" TIMESTAMP NULL,
    "deleted_at" TIMESTAMP NULL,
    "created_at" TIMESTAMP NULL,
    "updated_at" TIMESTAMP NULL
);
CREATE INDEX IF NOT EXISTS "photos_user_id_index" ON "photos" ("user_id");
CREATE INDEX IF NOT EXISTS "photos_user_id_created_at_index" ON "photos" ("user_id", "created_at");
CREATE INDEX IF NOT EXISTS "photos_ip_address_created_at_index" ON "photos" ("ip_address", "created_at");
CREATE INDEX IF NOT EXISTS "photos_group_id_index" ON "photos" ("group_id");
CREATE INDEX IF NOT EXISTS "photos_storage_id_index" ON "photos" ("storage_id");

CREATE TABLE IF NOT EXISTS "album_photo" (
    "album_id" INTEGER UNSIGNED NOT NULL,
    "photo_id" INTEGER UNSIGNED NOT NULL,
    "sort" INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY ("album_id", "photo_id"),
    FOREIGN KEY ("album_id") REFERENCES "albums" ("id") ON DELETE CASCADE,
    FOREIGN KEY ("photo_id") REFERENCES "photos" ("id") ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS "tags" (
    "id" INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
    "name" VARCHAR(255) NOT NULL,
    "created_at" TIMESTAMP NULL,
    "updated_at" TIMESTAMP NULL
);

CREATE TABLE IF NOT EXISTS "taggables" (
    "id" INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
    "tag_id" INTEGER UNSIGNED NULL,
    "user_id" INTEGER UNSIGNED NULL,
    "taggable_type" VARCHAR(255) NOT NULL,
    "taggable_id" INTEGER UNSIGNED NOT NULL,
    FOREIGN KEY ("tag_id") REFERENCES "tags" ("id") ON DELETE CASCADE,
    FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS "taggables_taggable_type_taggable_id_index"
    ON "taggables" ("taggable_type", "taggable_id");

CREATE TABLE IF NOT EXISTS "shares" (
    "id" INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
    "user_id" INTEGER UNSIGNED NOT NULL,
    "type" VARCHAR(32) NOT NULL DEFAULT 'album',
    "slug" VARCHAR(255) NOT NULL UNIQUE,
    "content" TEXT NULL,
    "password" VARCHAR(128) NOT NULL DEFAULT '',
    "view_count" INTEGER UNSIGNED NOT NULL DEFAULT 0,
    "expired_at" TIMESTAMP NULL,
    "created_at" TIMESTAMP NULL,
    "updated_at" TIMESTAMP NULL,
    FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS "shares_user_id_index" ON "shares" ("user_id");

CREATE TABLE IF NOT EXISTS "shareables" (
    "id" INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
    "share_id" INTEGER UNSIGNED NOT NULL,
    "shareable_type" VARCHAR(255) NOT NULL,
    "shareable_id" INTEGER UNSIGNED NOT NULL,
    FOREIGN KEY ("share_id") REFERENCES "shares" ("id") ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS "shareables_shareable_type_shareable_id_index"
    ON "shareables" ("shareable_type", "shareable_id");

CREATE TABLE IF NOT EXISTS "violations" (
    "id" INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
    "user_id" INTEGER UNSIGNED NULL,
    "photo_id" INTEGER UNSIGNED NULL,
    "reason" VARCHAR(255) NOT NULL DEFAULT '违规原因',
    "status" VARCHAR(32) NOT NULL DEFAULT 'unhandled',
    "handled_at" TIMESTAMP NULL,
    "deleted_at" TIMESTAMP NULL,
    "created_at" TIMESTAMP NULL,
    "updated_at" TIMESTAMP NULL,
    FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE SET NULL,
    FOREIGN KEY ("photo_id") REFERENCES "photos" ("id") ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS "notices" (
    "id" INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
    "title" VARCHAR(255) NOT NULL,
    "content" LONGTEXT NULL,
    "view_count" INTEGER UNSIGNED NOT NULL DEFAULT 0,
    "sort" INTEGER NOT NULL DEFAULT 0,
    "deleted_at" TIMESTAMP NULL,
    "created_at" TIMESTAMP NULL,
    "updated_at" TIMESTAMP NULL
);

CREATE TABLE IF NOT EXISTS "pages" (
    "id" INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
    "type" VARCHAR(32) NOT NULL DEFAULT 'internal',
    "name" VARCHAR(255) NOT NULL,
    "icon" VARCHAR(64) NOT NULL DEFAULT '',
    "title" VARCHAR(255) NOT NULL DEFAULT '',
    "content" LONGTEXT NULL,
    "keywords" TEXT NULL,
    "description" TEXT NULL,
    "slug" VARCHAR(255) NOT NULL DEFAULT '',
    "url" VARCHAR(255) NOT NULL DEFAULT '',
    "view_count" INTEGER UNSIGNED NOT NULL DEFAULT 0,
    "sort" INTEGER NOT NULL DEFAULT 0,
    "is_show" TINYINT(1) NOT NULL DEFAULT 0,
    "created_at" TIMESTAMP NULL,
    "updated_at" TIMESTAMP NULL
);

CREATE TABLE IF NOT EXISTS "feedbacks" (
    "id" INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
    "type" VARCHAR(32) NOT NULL DEFAULT 'general',
    "title" VARCHAR(64) NOT NULL,
    "name" VARCHAR(64) NOT NULL,
    "email" VARCHAR(128) NOT NULL,
    "content" LONGTEXT NOT NULL,
    "ip_address" VARCHAR(45) NULL,
    "created_at" TIMESTAMP NULL,
    "updated_at" TIMESTAMP NULL
);

CREATE TABLE IF NOT EXISTS "tickets" (
    "id" INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
    "user_id" INTEGER UNSIGNED NOT NULL,
    "issue_no" VARCHAR(255) NOT NULL UNIQUE,
    "title" VARCHAR(255) NOT NULL,
    "level" VARCHAR(32) NOT NULL DEFAULT 'low',
    "status" VARCHAR(32) NOT NULL DEFAULT 'in_progress',
    "deleted_at" TIMESTAMP NULL,
    "created_at" TIMESTAMP NULL,
    "updated_at" TIMESTAMP NULL,
    FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS "ticket_replies" (
    "id" INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
    "ticket_id" INTEGER UNSIGNED NOT NULL,
    "user_id" INTEGER UNSIGNED NOT NULL,
    "content" LONGTEXT NOT NULL,
    "is_notify" TINYINT(1) NOT NULL DEFAULT 1,
    "read_at" TIMESTAMP NULL,
    "deleted_at" TIMESTAMP NULL,
    "created_at" TIMESTAMP NULL,
    "updated_at" TIMESTAMP NULL,
    FOREIGN KEY ("ticket_id") REFERENCES "tickets" ("id") ON DELETE CASCADE,
    FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS "reports" (
    "id" INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
    "report_user_id" INTEGER UNSIGNED NULL,
    "reportable_type" VARCHAR(32) NOT NULL,
    "reportable_id" INTEGER UNSIGNED NOT NULL,
    "content" VARCHAR(255) NULL,
    "status" VARCHAR(32) NOT NULL DEFAULT 'unhandled',
    "handled_at" TIMESTAMP NULL,
    "ip_address" VARCHAR(45) NULL,
    "deleted_at" TIMESTAMP NULL,
    "created_at" TIMESTAMP NULL,
    "updated_at" TIMESTAMP NULL,
    FOREIGN KEY ("report_user_id") REFERENCES "users" ("id") ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS "likes" (
    "id" INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
    "user_id" INTEGER UNSIGNED NOT NULL,
    "likeable_type" VARCHAR(32) NOT NULL,
    "likeable_id" INTEGER UNSIGNED NOT NULL,
    "created_at" TIMESTAMP NULL,
    "updated_at" TIMESTAMP NULL,
    FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS "plans" (
    "id" INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
    "type" VARCHAR(255) NOT NULL DEFAULT 'vip',
    "name" VARCHAR(255) NOT NULL,
    "intro" TEXT NULL,
    "features" JSON NULL,
    "badge" VARCHAR(32) NOT NULL DEFAULT '',
    "sort" INTEGER NOT NULL DEFAULT 0,
    "is_up" TINYINT(1) NOT NULL DEFAULT 0,
    "deleted_at" TIMESTAMP NULL,
    "created_at" TIMESTAMP NULL,
    "updated_at" TIMESTAMP NULL
);

CREATE TABLE IF NOT EXISTS "plan_prices" (
    "id" INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
    "plan_id" INTEGER UNSIGNED NOT NULL,
    "name" VARCHAR(255) NOT NULL,
    "duration" INTEGER NOT NULL DEFAULT 0,
    "price" INTEGER NOT NULL DEFAULT 0,
    "created_at" TIMESTAMP NULL,
    "updated_at" TIMESTAMP NULL,
    FOREIGN KEY ("plan_id") REFERENCES "plans" ("id") ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS "plan_groups" (
    "id" INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
    "plan_id" INTEGER UNSIGNED NOT NULL,
    "group_id" INTEGER UNSIGNED NULL,
    FOREIGN KEY ("plan_id") REFERENCES "plans" ("id") ON DELETE CASCADE,
    FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS "plan_capacities" (
    "id" INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
    "plan_id" INTEGER UNSIGNED NOT NULL,
    "capacity" NUMERIC NULL DEFAULT 0,
    FOREIGN KEY ("plan_id") REFERENCES "plans" ("id") ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS "coupons" (
    "id" INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
    "type" VARCHAR(32) NOT NULL DEFAULT 'direct',
    "name" VARCHAR(32) NOT NULL DEFAULT '',
    "code" VARCHAR(255) NOT NULL UNIQUE,
    "value" NUMERIC NOT NULL DEFAULT 0,
    "usage_limit" INTEGER UNSIGNED NOT NULL DEFAULT 1,
    "expired_at" TIMESTAMP NULL,
    "deleted_at" TIMESTAMP NULL,
    "created_at" TIMESTAMP NULL,
    "updated_at" TIMESTAMP NULL
);

CREATE TABLE IF NOT EXISTS "orders" (
    "id" INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
    "plan_id" INTEGER UNSIGNED NULL,
    "user_id" INTEGER UNSIGNED NULL,
    "coupon_id" INTEGER UNSIGNED NULL,
    "trade_no" VARCHAR(255) NOT NULL UNIQUE,
    "out_trade_no" VARCHAR(255) NOT NULL UNIQUE,
    "type" VARCHAR(32) NOT NULL DEFAULT 'plan',
    "amount" INTEGER UNSIGNED NOT NULL DEFAULT 0,
    "deduct_amount" INTEGER UNSIGNED NOT NULL DEFAULT 0,
    "snapshot" JSON NULL,
    "product" JSON NULL,
    "pay_method" VARCHAR(255) NOT NULL DEFAULT '',
    "status" VARCHAR(32) NOT NULL DEFAULT 'unpaid',
    "paid_at" TIMESTAMP NULL,
    "canceled_at" TIMESTAMP NULL,
    "created_at" TIMESTAMP NULL,
    "updated_at" TIMESTAMP NULL,
    FOREIGN KEY ("plan_id") REFERENCES "plans" ("id") ON DELETE SET NULL,
    FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE SET NULL,
    FOREIGN KEY ("coupon_id") REFERENCES "coupons" ("id") ON DELETE SET NULL
);
CREATE INDEX IF NOT EXISTS "orders_user_id_index" ON "orders" ("user_id");

CREATE TABLE IF NOT EXISTS "user_groups" (
    "id" INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
    "user_id" INTEGER UNSIGNED NOT NULL,
    "group_id" INTEGER UNSIGNED NOT NULL,
    "order_id" INTEGER UNSIGNED NULL,
    "from" VARCHAR(32) NOT NULL DEFAULT 'system',
    "expired_at" TIMESTAMP NULL,
    "deleted_at" TIMESTAMP NULL,
    "created_at" TIMESTAMP NULL,
    "updated_at" TIMESTAMP NULL,
    FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE CASCADE,
    FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON DELETE CASCADE,
    FOREIGN KEY ("order_id") REFERENCES "orders" ("id") ON DELETE SET NULL
);
CREATE INDEX IF NOT EXISTS "user_groups_expired_at_index" ON "user_groups" ("expired_at");

CREATE TABLE IF NOT EXISTS "user_capacities" (
    "id" INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
    "user_id" INTEGER UNSIGNED NOT NULL,
    "order_id" INTEGER UNSIGNED NULL,
    "capacity" NUMERIC NULL DEFAULT 0,
    "from" VARCHAR(32) NOT NULL DEFAULT 'system',
    "expired_at" TIMESTAMP NULL,
    "deleted_at" TIMESTAMP NULL,
    "created_at" TIMESTAMP NULL,
    "updated_at" TIMESTAMP NULL,
    FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE CASCADE,
    FOREIGN KEY ("order_id") REFERENCES "orders" ("id") ON DELETE SET NULL
);
CREATE INDEX IF NOT EXISTS "user_capacities_expired_at_index" ON "user_capacities" ("expired_at");
CREATE INDEX IF NOT EXISTS "user_capacities_capacity_index" ON "user_capacities" ("capacity");
