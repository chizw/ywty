-- ============================================================================
-- ywty (ywty) 初始 schema · MySQL/MariaDB 版
-- 逐列对齐 database/migrations/*.php（Laravel 11）的最终状态，
-- 可直接替换 PHP 版并沿用现有数据库；新建库后由应用写入 Laravel 风格的
-- migrations 记录，保持两侧工具链互通。
-- 时间列：TIMESTAMP NULL，会话时区固定 +00:00（与 PHP config/database.php 一致）。
-- ============================================================================

CREATE TABLE IF NOT EXISTS `cache` (
    `key` VARCHAR(255) NOT NULL,
    `value` MEDIUMTEXT NOT NULL,
    `expiration` INT NOT NULL,
    PRIMARY KEY (`key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='缓存表';

CREATE TABLE IF NOT EXISTS `cache_locks` (
    `key` VARCHAR(255) NOT NULL,
    `owner` VARCHAR(255) NOT NULL,
    `expiration` INT NOT NULL,
    PRIMARY KEY (`key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='缓存锁表';

CREATE TABLE IF NOT EXISTS `jobs` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `queue` VARCHAR(255) NOT NULL,
    `payload` LONGTEXT NOT NULL,
    `attempts` TINYINT UNSIGNED NOT NULL,
    `reserved_at` INT UNSIGNED NULL,
    `available_at` INT UNSIGNED NOT NULL,
    `created_at` INT UNSIGNED NOT NULL,
    PRIMARY KEY (`id`),
    KEY `jobs_queue_index` (`queue`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `job_batches` (
    `id` VARCHAR(255) NOT NULL,
    `name` VARCHAR(255) NOT NULL,
    `total_jobs` INT NOT NULL,
    `pending_jobs` INT NOT NULL,
    `failed_jobs` INT NOT NULL,
    `failed_job_ids` LONGTEXT NOT NULL,
    `options` MEDIUMTEXT NULL,
    `cancelled_at` INT NULL,
    `created_at` INT NOT NULL,
    `finished_at` INT NULL,
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `failed_jobs` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `uuid` VARCHAR(255) NOT NULL,
    `connection` TEXT NOT NULL,
    `queue` TEXT NOT NULL,
    `payload` LONGTEXT NOT NULL,
    `exception` LONGTEXT NOT NULL,
    `failed_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `failed_jobs_uuid_unique` (`uuid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `personal_access_tokens` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `tokenable_type` VARCHAR(255) NOT NULL,
    `tokenable_id` BIGINT UNSIGNED NOT NULL,
    `name` VARCHAR(255) NOT NULL,
    `token` VARCHAR(64) NOT NULL,
    `abilities` TEXT NULL,
    `last_used_at` TIMESTAMP NULL DEFAULT NULL,
    `expires_at` TIMESTAMP NULL DEFAULT NULL,
    `created_at` TIMESTAMP NULL DEFAULT NULL,
    `updated_at` TIMESTAMP NULL DEFAULT NULL,
    PRIMARY KEY (`id`),
    UNIQUE KEY `personal_access_tokens_token_unique` (`token`),
    KEY `personal_access_tokens_tokenable_type_tokenable_id_index` (`tokenable_type`, `tokenable_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `settings` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `group` VARCHAR(255) NOT NULL,
    `name` VARCHAR(255) NOT NULL,
    `locked` TINYINT(1) NOT NULL DEFAULT 0,
    `payload` JSON NOT NULL,
    `created_at` TIMESTAMP NULL DEFAULT NULL,
    `updated_at` TIMESTAMP NULL DEFAULT NULL,
    PRIMARY KEY (`id`),
    UNIQUE KEY `settings_group_name_unique` (`group`, `name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `groups` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `name` VARCHAR(255) NOT NULL COMMENT '名称',
    `intro` VARCHAR(2000) NOT NULL DEFAULT '' COMMENT '描述',
    `options` JSON NULL COMMENT '配置',
    `is_default` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否为默认组',
    `is_guest` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否为游客组',
    `deleted_at` TIMESTAMP NULL DEFAULT NULL,
    `created_at` TIMESTAMP NULL DEFAULT NULL,
    `updated_at` TIMESTAMP NULL DEFAULT NULL,
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='角色组表';

CREATE TABLE IF NOT EXISTS `drivers` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `type` VARCHAR(255) NOT NULL COMMENT '驱动类型',
    `name` VARCHAR(255) NOT NULL COMMENT '名称',
    `intro` VARCHAR(2000) NOT NULL DEFAULT '' COMMENT '简介',
    `options` JSON NULL COMMENT '配置',
    `deleted_at` TIMESTAMP NULL DEFAULT NULL,
    `created_at` TIMESTAMP NULL DEFAULT NULL,
    `updated_at` TIMESTAMP NULL DEFAULT NULL,
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='驱动表';

CREATE TABLE IF NOT EXISTS `storages` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `name` VARCHAR(255) NOT NULL COMMENT '名称',
    `intro` VARCHAR(2000) NOT NULL DEFAULT '' COMMENT '描述',
    `prefix` VARCHAR(1000) NOT NULL DEFAULT '' COMMENT '储存前缀',
    `provider` VARCHAR(64) NOT NULL COMMENT '储存提供者',
    `options` JSON NULL COMMENT '储存配置',
    `deleted_at` TIMESTAMP NULL DEFAULT NULL,
    `created_at` TIMESTAMP NULL DEFAULT NULL,
    `updated_at` TIMESTAMP NULL DEFAULT NULL,
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='储存策略表';

CREATE TABLE IF NOT EXISTS `group_storage` (
    `group_id` BIGINT UNSIGNED NOT NULL COMMENT '角色组',
    `storage_id` BIGINT UNSIGNED NOT NULL COMMENT '储存',
    `sort` INT NOT NULL DEFAULT 0 COMMENT '排序值',
    PRIMARY KEY (`group_id`, `storage_id`),
    CONSTRAINT `group_storage_group_id_foreign` FOREIGN KEY (`group_id`) REFERENCES `groups` (`id`) ON DELETE CASCADE,
    CONSTRAINT `group_storage_storage_id_foreign` FOREIGN KEY (`storage_id`) REFERENCES `storages` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='组与储存中间表';

CREATE TABLE IF NOT EXISTS `users` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `avatar` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '头像',
    `name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '姓名',
    `username` VARCHAR(255) NOT NULL COMMENT '用户名',
    `phone` VARCHAR(64) NULL COMMENT '手机号',
    `email` VARCHAR(255) NULL COMMENT '邮箱',
    `password` VARCHAR(255) NOT NULL COMMENT '密码',
    `location` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '所在地',
    `url` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '个人网站',
    `company` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '所在公司',
    `company_title` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '工作职位',
    `tagline` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '个性签名',
    `bio` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '个人简介',
    `interests` JSON NULL COMMENT '兴趣标签',
    `socials` JSON NULL COMMENT '社交账号',
    `phone_verified_at` TIMESTAMP NULL DEFAULT NULL,
    `email_verified_at` TIMESTAMP NULL DEFAULT NULL,
    `remember_token` VARCHAR(100) NULL,
    `is_admin` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否为管理员',
    `options` JSON NULL COMMENT '配置',
    `login_ip` VARCHAR(45) NULL COMMENT '最后登录 IP',
    `register_ip` VARCHAR(45) NULL COMMENT '注册 IP',
    `country_code` VARCHAR(32) NULL COMMENT '国家',
    `status` VARCHAR(64) NOT NULL DEFAULT 'normal' COMMENT '状态',
    `deleted_at` TIMESTAMP NULL DEFAULT NULL,
    `created_at` TIMESTAMP NULL DEFAULT NULL,
    `updated_at` TIMESTAMP NULL DEFAULT NULL,
    PRIMARY KEY (`id`),
    UNIQUE KEY `users_phone_unique` (`phone`),
    UNIQUE KEY `users_email_unique` (`email`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';

CREATE TABLE IF NOT EXISTS `group_driver` (
    `type` VARCHAR(32) NOT NULL COMMENT '驱动类型',
    `group_id` BIGINT UNSIGNED NOT NULL COMMENT '角色组',
    `driver_id` BIGINT UNSIGNED NOT NULL COMMENT '驱动',
    `sort` INT NOT NULL DEFAULT 0 COMMENT '排序值',
    PRIMARY KEY (`group_id`, `driver_id`),
    CONSTRAINT `group_driver_group_id_foreign` FOREIGN KEY (`group_id`) REFERENCES `groups` (`id`) ON DELETE CASCADE,
    CONSTRAINT `group_driver_driver_id_foreign` FOREIGN KEY (`driver_id`) REFERENCES `drivers` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='组与驱动中间表';

CREATE TABLE IF NOT EXISTS `storage_driver` (
    `type` VARCHAR(32) NOT NULL COMMENT '驱动类型',
    `storage_id` BIGINT UNSIGNED NOT NULL COMMENT '储存',
    `driver_id` BIGINT UNSIGNED NOT NULL COMMENT '驱动',
    `sort` INT NOT NULL DEFAULT 0 COMMENT '排序值',
    PRIMARY KEY (`storage_id`, `driver_id`),
    CONSTRAINT `storage_driver_storage_id_foreign` FOREIGN KEY (`storage_id`) REFERENCES `storages` (`id`) ON DELETE CASCADE,
    CONSTRAINT `storage_driver_driver_id_foreign` FOREIGN KEY (`driver_id`) REFERENCES `drivers` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='储存与驱动中间表';

CREATE TABLE IF NOT EXISTS `password_reset_tokens` (
    `email` VARCHAR(255) NOT NULL,
    `token` VARCHAR(255) NOT NULL,
    `created_at` TIMESTAMP NULL DEFAULT NULL,
    PRIMARY KEY (`email`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='密码重置令牌表';

CREATE TABLE IF NOT EXISTS `sessions` (
    `id` VARCHAR(255) NOT NULL,
    `user_id` BIGINT UNSIGNED NULL,
    `ip_address` VARCHAR(45) NULL,
    `user_agent` TEXT NULL,
    `payload` LONGTEXT NOT NULL,
    `last_activity` INT NOT NULL,
    PRIMARY KEY (`id`),
    KEY `sessions_user_id_index` (`user_id`),
    KEY `sessions_last_activity_index` (`last_activity`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='会话表';

CREATE TABLE IF NOT EXISTS `oauth` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `driver_id` BIGINT UNSIGNED NOT NULL COMMENT '三方授权驱动ID',
    `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
    `openid` VARCHAR(255) NOT NULL COMMENT '三方授权ID',
    `avatar` VARCHAR(512) NOT NULL DEFAULT '' COMMENT '三方授权头像',
    `email` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '三方授权邮箱',
    `name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '三方授权名称',
    `nickname` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '三方授权昵称',
    `raw` JSON NULL COMMENT '三方授权原始信息',
    `created_at` TIMESTAMP NULL DEFAULT NULL,
    `updated_at` TIMESTAMP NULL DEFAULT NULL,
    PRIMARY KEY (`id`),
    CONSTRAINT `oauth_driver_id_foreign` FOREIGN KEY (`driver_id`) REFERENCES `drivers` (`id`) ON DELETE CASCADE,
    CONSTRAINT `oauth_user_id_foreign` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='三方授权表';

CREATE TABLE IF NOT EXISTS `albums` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `user_id` BIGINT UNSIGNED NULL COMMENT '用户',
    `name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '名称',
    `intro` VARCHAR(2000) NOT NULL DEFAULT '' COMMENT '介绍',
    `is_public` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否公开',
    `deleted_at` TIMESTAMP NULL DEFAULT NULL,
    `created_at` TIMESTAMP NULL DEFAULT NULL,
    `updated_at` TIMESTAMP NULL DEFAULT NULL,
    PRIMARY KEY (`id`),
    CONSTRAINT `albums_user_id_foreign` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='相册表';

CREATE TABLE IF NOT EXISTS `photos` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `user_id` BIGINT UNSIGNED NULL COMMENT '用户',
    `group_id` BIGINT UNSIGNED NULL COMMENT '角色组',
    `storage_id` BIGINT UNSIGNED NULL COMMENT '储存',
    `name` VARCHAR(255) NOT NULL COMMENT '文件别名',
    `intro` VARCHAR(2000) NOT NULL DEFAULT '' COMMENT '介绍',
    `filename` VARCHAR(255) NOT NULL COMMENT '文件原始名称',
    `pathname` VARCHAR(255) NOT NULL COMMENT '文件路径名称',
    `mimetype` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '媒体类型',
    `extension` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '文件后缀',
    `md5` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '文件MD5',
    `sha1` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '文件SHA1',
    `exif` JSON NULL COMMENT 'EXIF 信息',
    `size` DECIMAL(20,2) NOT NULL DEFAULT 0 COMMENT '大小(kb)',
    `width` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '宽度',
    `height` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '高度',
    `is_public` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否公开',
    `status` VARCHAR(64) NOT NULL DEFAULT 'normal' COMMENT '状态',
    `ip_address` VARCHAR(45) NULL COMMENT '上传IP',
    `expired_at` TIMESTAMP NULL DEFAULT NULL COMMENT '到期时间',
    `deleted_at` TIMESTAMP NULL DEFAULT NULL,
    `created_at` TIMESTAMP NULL DEFAULT NULL,
    `updated_at` TIMESTAMP NULL DEFAULT NULL,
    PRIMARY KEY (`id`),
    KEY `photos_user_id_index` (`user_id`),
    KEY `photos_user_id_created_at_index` (`user_id`, `created_at`),
    KEY `photos_ip_address_created_at_index` (`ip_address`, `created_at`),
    CONSTRAINT `photos_user_id_foreign` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE,
    CONSTRAINT `photos_group_id_foreign` FOREIGN KEY (`group_id`) REFERENCES `groups` (`id`) ON DELETE SET NULL,
    CONSTRAINT `photos_storage_id_foreign` FOREIGN KEY (`storage_id`) REFERENCES `storages` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='图片表';

CREATE TABLE IF NOT EXISTS `album_photo` (
    `album_id` BIGINT UNSIGNED NOT NULL COMMENT '相册',
    `photo_id` BIGINT UNSIGNED NOT NULL COMMENT '图片',
    `sort` INT NOT NULL DEFAULT 0 COMMENT '排序值',
    PRIMARY KEY (`album_id`, `photo_id`),
    CONSTRAINT `album_photo_album_id_foreign` FOREIGN KEY (`album_id`) REFERENCES `albums` (`id`) ON DELETE CASCADE,
    CONSTRAINT `album_photo_photo_id_foreign` FOREIGN KEY (`photo_id`) REFERENCES `photos` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='相册与图片中间表';

CREATE TABLE IF NOT EXISTS `tags` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `name` VARCHAR(255) NOT NULL COMMENT '名称',
    `created_at` TIMESTAMP NULL DEFAULT NULL,
    `updated_at` TIMESTAMP NULL DEFAULT NULL,
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='标签表';

CREATE TABLE IF NOT EXISTS `taggables` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `tag_id` BIGINT UNSIGNED NULL COMMENT '标签',
    `user_id` BIGINT UNSIGNED NULL COMMENT '用户',
    `taggable_type` VARCHAR(255) NOT NULL,
    `taggable_id` BIGINT UNSIGNED NOT NULL,
    PRIMARY KEY (`id`),
    KEY `taggables_taggable_type_taggable_id_index` (`taggable_type`, `taggable_id`),
    CONSTRAINT `taggables_tag_id_foreign` FOREIGN KEY (`tag_id`) REFERENCES `tags` (`id`) ON DELETE CASCADE,
    CONSTRAINT `taggables_user_id_foreign` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='标签表';

CREATE TABLE IF NOT EXISTS `shares` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户',
    `type` VARCHAR(32) NOT NULL DEFAULT 'album' COMMENT '分享类型',
    `slug` VARCHAR(255) NOT NULL COMMENT 'url slug',
    `content` TEXT NULL COMMENT '分享内容',
    `password` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '密码',
    `view_count` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '浏览量',
    `expired_at` TIMESTAMP NULL DEFAULT NULL COMMENT '到期时间',
    `created_at` TIMESTAMP NULL DEFAULT NULL,
    `updated_at` TIMESTAMP NULL DEFAULT NULL,
    PRIMARY KEY (`id`),
    UNIQUE KEY `shares_slug_unique` (`slug`),
    KEY `shares_user_id_index` (`user_id`),
    CONSTRAINT `shares_user_id_foreign` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='分享表';

CREATE TABLE IF NOT EXISTS `shareables` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `share_id` BIGINT UNSIGNED NOT NULL COMMENT '分享',
    `shareable_type` VARCHAR(255) NOT NULL,
    `shareable_id` BIGINT UNSIGNED NOT NULL,
    PRIMARY KEY (`id`),
    KEY `shareables_shareable_type_shareable_id_index` (`shareable_type`, `shareable_id`),
    CONSTRAINT `shareables_share_id_foreign` FOREIGN KEY (`share_id`) REFERENCES `shares` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='分享内容表';

CREATE TABLE IF NOT EXISTS `violations` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `user_id` BIGINT UNSIGNED NULL COMMENT '用户',
    `photo_id` BIGINT UNSIGNED NULL COMMENT '图片',
    `reason` VARCHAR(255) NOT NULL DEFAULT '违规原因',
    `status` VARCHAR(32) NOT NULL DEFAULT 'unhandled' COMMENT '状态',
    `handled_at` TIMESTAMP NULL DEFAULT NULL COMMENT '处理时间',
    `deleted_at` TIMESTAMP NULL DEFAULT NULL,
    `created_at` TIMESTAMP NULL DEFAULT NULL,
    `updated_at` TIMESTAMP NULL DEFAULT NULL,
    PRIMARY KEY (`id`),
    CONSTRAINT `violations_user_id_foreign` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE SET NULL,
    CONSTRAINT `violations_photo_id_foreign` FOREIGN KEY (`photo_id`) REFERENCES `photos` (`id`) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='图片违规记录表';

CREATE TABLE IF NOT EXISTS `notices` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `title` VARCHAR(255) NOT NULL COMMENT '标题',
    `content` LONGTEXT NULL COMMENT '内容',
    `view_count` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '阅读量',
    `sort` INT NOT NULL DEFAULT 0 COMMENT '排序值',
    `deleted_at` TIMESTAMP NULL DEFAULT NULL,
    `created_at` TIMESTAMP NULL DEFAULT NULL,
    `updated_at` TIMESTAMP NULL DEFAULT NULL,
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='系统公告表';

CREATE TABLE IF NOT EXISTS `pages` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `type` VARCHAR(32) NOT NULL DEFAULT 'internal' COMMENT '类型',
    `name` VARCHAR(255) NOT NULL COMMENT '名称',
    `icon` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '图标',
    `title` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '标题',
    `content` LONGTEXT NULL COMMENT '网页内容',
    `keywords` TEXT NULL COMMENT '网页关键字',
    `description` TEXT NULL COMMENT '网页描述',
    `slug` VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'url slug',
    `url` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '跳转url',
    `view_count` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '浏览量',
    `sort` INT NOT NULL DEFAULT 0 COMMENT '排序值',
    `is_show` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否显示',
    `created_at` TIMESTAMP NULL DEFAULT NULL,
    `updated_at` TIMESTAMP NULL DEFAULT NULL,
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='页面表';

CREATE TABLE IF NOT EXISTS `feedbacks` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `type` VARCHAR(32) NOT NULL DEFAULT 'general' COMMENT '类型',
    `title` VARCHAR(64) NOT NULL COMMENT '标题',
    `name` VARCHAR(64) NOT NULL COMMENT '姓名',
    `email` VARCHAR(128) NOT NULL COMMENT 'email',
    `content` LONGTEXT NOT NULL COMMENT '内容',
    `ip_address` VARCHAR(45) NULL COMMENT 'IP 地址',
    `created_at` TIMESTAMP NULL DEFAULT NULL,
    `updated_at` TIMESTAMP NULL DEFAULT NULL,
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='意见与反馈表';

CREATE TABLE IF NOT EXISTS `tickets` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户',
    `issue_no` VARCHAR(255) NOT NULL COMMENT '工单编号',
    `title` VARCHAR(255) NOT NULL COMMENT '标题',
    `level` VARCHAR(32) NOT NULL DEFAULT 'low' COMMENT '级别',
    `status` VARCHAR(32) NOT NULL DEFAULT 'in_progress' COMMENT '状态',
    `deleted_at` TIMESTAMP NULL DEFAULT NULL,
    `created_at` TIMESTAMP NULL DEFAULT NULL,
    `updated_at` TIMESTAMP NULL DEFAULT NULL,
    PRIMARY KEY (`id`),
    UNIQUE KEY `tickets_issue_no_unique` (`issue_no`),
    CONSTRAINT `tickets_user_id_foreign` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='工单表';

CREATE TABLE IF NOT EXISTS `ticket_replies` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `ticket_id` BIGINT UNSIGNED NOT NULL COMMENT '工单',
    `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户',
    `content` LONGTEXT NOT NULL COMMENT '内容',
    `is_notify` TINYINT(1) NOT NULL DEFAULT 1 COMMENT '是否需要接收通知',
    `read_at` TIMESTAMP NULL DEFAULT NULL COMMENT '已读时间',
    `deleted_at` TIMESTAMP NULL DEFAULT NULL,
    `created_at` TIMESTAMP NULL DEFAULT NULL,
    `updated_at` TIMESTAMP NULL DEFAULT NULL,
    PRIMARY KEY (`id`),
    CONSTRAINT `ticket_replies_ticket_id_foreign` FOREIGN KEY (`ticket_id`) REFERENCES `tickets` (`id`) ON DELETE CASCADE,
    CONSTRAINT `ticket_replies_user_id_foreign` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='工单回复记录表';

CREATE TABLE IF NOT EXISTS `reports` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `report_user_id` BIGINT UNSIGNED NULL COMMENT '被举报用户',
    `reportable_type` VARCHAR(32) NOT NULL,
    `reportable_id` BIGINT UNSIGNED NOT NULL,
    `content` VARCHAR(255) NULL COMMENT '原因',
    `status` VARCHAR(32) NOT NULL DEFAULT 'unhandled' COMMENT '状态',
    `handled_at` TIMESTAMP NULL DEFAULT NULL COMMENT '处理时间',
    `ip_address` VARCHAR(45) NULL COMMENT 'IP 地址',
    `deleted_at` TIMESTAMP NULL DEFAULT NULL,
    `created_at` TIMESTAMP NULL DEFAULT NULL,
    `updated_at` TIMESTAMP NULL DEFAULT NULL,
    PRIMARY KEY (`id`),
    CONSTRAINT `reports_report_user_id_foreign` FOREIGN KEY (`report_user_id`) REFERENCES `users` (`id`) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='举报记录表';

CREATE TABLE IF NOT EXISTS `likes` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户',
    `likeable_type` VARCHAR(32) NOT NULL,
    `likeable_id` BIGINT UNSIGNED NOT NULL,
    `created_at` TIMESTAMP NULL DEFAULT NULL,
    `updated_at` TIMESTAMP NULL DEFAULT NULL,
    PRIMARY KEY (`id`),
    CONSTRAINT `likes_user_id_foreign` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='点赞表';

CREATE TABLE IF NOT EXISTS `plans` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `type` VARCHAR(255) NOT NULL DEFAULT 'vip' COMMENT '类型',
    `name` VARCHAR(255) NOT NULL COMMENT '名称',
    `intro` TEXT NULL COMMENT '简介',
    `features` JSON NULL COMMENT '特点',
    `badge` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '徽章内容',
    `sort` INT NOT NULL DEFAULT 0 COMMENT '排序值',
    `is_up` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否上架',
    `deleted_at` TIMESTAMP NULL DEFAULT NULL,
    `created_at` TIMESTAMP NULL DEFAULT NULL,
    `updated_at` TIMESTAMP NULL DEFAULT NULL,
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='计划套餐表';

CREATE TABLE IF NOT EXISTS `plan_prices` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `plan_id` BIGINT UNSIGNED NOT NULL COMMENT '计划',
    `name` VARCHAR(255) NOT NULL COMMENT '名称',
    `duration` INT NOT NULL DEFAULT 0 COMMENT '时长(分钟)',
    `price` INT NOT NULL DEFAULT 0 COMMENT '价格(分)',
    `created_at` TIMESTAMP NULL DEFAULT NULL,
    `updated_at` TIMESTAMP NULL DEFAULT NULL,
    PRIMARY KEY (`id`),
    CONSTRAINT `plan_prices_plan_id_foreign` FOREIGN KEY (`plan_id`) REFERENCES `plans` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='计划套餐阶梯价格表';

CREATE TABLE IF NOT EXISTS `plan_groups` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `plan_id` BIGINT UNSIGNED NOT NULL COMMENT '计划',
    `group_id` BIGINT UNSIGNED NULL COMMENT '角色组',
    PRIMARY KEY (`id`),
    CONSTRAINT `plan_groups_plan_id_foreign` FOREIGN KEY (`plan_id`) REFERENCES `plans` (`id`) ON DELETE CASCADE,
    CONSTRAINT `plan_groups_group_id_foreign` FOREIGN KEY (`group_id`) REFERENCES `groups` (`id`) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='计划可用组表';

CREATE TABLE IF NOT EXISTS `plan_capacities` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `plan_id` BIGINT UNSIGNED NOT NULL COMMENT '计划',
    `capacity` DECIMAL(20,2) NULL DEFAULT 0 COMMENT '容量(kb)',
    PRIMARY KEY (`id`),
    CONSTRAINT `plan_capacities_plan_id_foreign` FOREIGN KEY (`plan_id`) REFERENCES `plans` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='计划可用容量表';

CREATE TABLE IF NOT EXISTS `coupons` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `type` VARCHAR(32) NOT NULL DEFAULT 'direct' COMMENT '折扣类型',
    `name` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '名称',
    `code` VARCHAR(255) NOT NULL COMMENT '券码',
    `value` DECIMAL(8,2) NOT NULL DEFAULT 0 COMMENT '金额或折扣率',
    `usage_limit` INT UNSIGNED NOT NULL DEFAULT 1 COMMENT '可使用次数',
    `expired_at` TIMESTAMP NULL DEFAULT NULL COMMENT '到期时间',
    `deleted_at` TIMESTAMP NULL DEFAULT NULL,
    `created_at` TIMESTAMP NULL DEFAULT NULL,
    `updated_at` TIMESTAMP NULL DEFAULT NULL,
    PRIMARY KEY (`id`),
    UNIQUE KEY `coupons_code_unique` (`code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='优惠券表';

CREATE TABLE IF NOT EXISTS `orders` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `plan_id` BIGINT UNSIGNED NULL COMMENT '计划',
    `user_id` BIGINT UNSIGNED NULL COMMENT '用户',
    `coupon_id` BIGINT UNSIGNED NULL COMMENT '优惠券',
    `trade_no` VARCHAR(255) NOT NULL COMMENT '系统订单号',
    `out_trade_no` VARCHAR(255) NOT NULL COMMENT '支付订单号',
    `type` VARCHAR(32) NOT NULL DEFAULT 'plan' COMMENT '类型',
    `amount` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '实际付款金额(分)',
    `deduct_amount` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '抵扣金额(分)',
    `snapshot` JSON NULL COMMENT '产品快照',
    `product` JSON NULL COMMENT '购买产品数据',
    `pay_method` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '支付方式',
    `status` VARCHAR(32) NOT NULL DEFAULT 'unpaid' COMMENT '状态',
    `paid_at` TIMESTAMP NULL DEFAULT NULL COMMENT '支付时间',
    `canceled_at` TIMESTAMP NULL DEFAULT NULL COMMENT '取消时间',
    `created_at` TIMESTAMP NULL DEFAULT NULL,
    `updated_at` TIMESTAMP NULL DEFAULT NULL,
    PRIMARY KEY (`id`),
    UNIQUE KEY `orders_trade_no_unique` (`trade_no`),
    UNIQUE KEY `orders_out_trade_no_unique` (`out_trade_no`),
    KEY `orders_user_id_index` (`user_id`),
    CONSTRAINT `orders_plan_id_foreign` FOREIGN KEY (`plan_id`) REFERENCES `plans` (`id`) ON DELETE SET NULL,
    CONSTRAINT `orders_user_id_foreign` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE SET NULL,
    CONSTRAINT `orders_coupon_id_foreign` FOREIGN KEY (`coupon_id`) REFERENCES `coupons` (`id`) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='订单表';

CREATE TABLE IF NOT EXISTS `user_groups` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户',
    `group_id` BIGINT UNSIGNED NOT NULL COMMENT '角色组',
    `order_id` BIGINT UNSIGNED NULL COMMENT '来源订单',
    `from` VARCHAR(32) NOT NULL DEFAULT 'system' COMMENT '来源',
    `expired_at` TIMESTAMP NULL DEFAULT NULL COMMENT '到期时间',
    `deleted_at` TIMESTAMP NULL DEFAULT NULL,
    `created_at` TIMESTAMP NULL DEFAULT NULL,
    `updated_at` TIMESTAMP NULL DEFAULT NULL,
    PRIMARY KEY (`id`),
    KEY `user_groups_expired_at_index` (`expired_at`),
    CONSTRAINT `user_groups_user_id_foreign` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE,
    CONSTRAINT `user_groups_group_id_foreign` FOREIGN KEY (`group_id`) REFERENCES `groups` (`id`) ON DELETE CASCADE,
    CONSTRAINT `user_groups_order_id_foreign` FOREIGN KEY (`order_id`) REFERENCES `orders` (`id`) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户角色组表';

CREATE TABLE IF NOT EXISTS `user_capacities` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户',
    `order_id` BIGINT UNSIGNED NULL COMMENT '来源订单',
    `capacity` DECIMAL(20,2) NULL DEFAULT 0 COMMENT '容量(kb)',
    `from` VARCHAR(32) NOT NULL DEFAULT 'system' COMMENT '来源',
    `expired_at` TIMESTAMP NULL DEFAULT NULL COMMENT '到期时间',
    `deleted_at` TIMESTAMP NULL DEFAULT NULL,
    `created_at` TIMESTAMP NULL DEFAULT NULL,
    `updated_at` TIMESTAMP NULL DEFAULT NULL,
    PRIMARY KEY (`id`),
    KEY `user_capacities_expired_at_index` (`expired_at`),
    KEY `user_capacities_capacity_index` (`capacity`),
    CONSTRAINT `user_capacities_user_id_foreign` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE,
    CONSTRAINT `user_capacities_order_id_foreign` FOREIGN KEY (`order_id`) REFERENCES `orders` (`id`) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户容量表';
