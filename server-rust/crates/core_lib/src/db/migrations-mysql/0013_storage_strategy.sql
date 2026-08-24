-- 存储策略增强 · MariaDB/MySQL 版（对标 ywty）
ALTER TABLE storages ADD COLUMN access_url VARCHAR(768) NOT NULL DEFAULT '';
ALTER TABLE photos ADD COLUMN storage_id BIGINT NULL;

-- 种子：默认本地策略（幂等）
INSERT IGNORE INTO storages (id, name, intro, prefix, provider, access_url, options, created_at, updated_at)
VALUES (1, '本地存储', '跟随站点的本地磁盘', '', 'local', '', NULL, NOW(), NOW());
