-- 组配额 + 用户单独覆盖 · MariaDB/MySQL 版
ALTER TABLE `groups` ADD COLUMN max_storage BIGINT NULL;
ALTER TABLE users ADD COLUMN quota_override BIGINT NULL;
