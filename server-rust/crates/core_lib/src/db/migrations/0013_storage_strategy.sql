-- 存储策略增强 · 对标 ywty
-- storages.access_url: 该策略绑定的访问域名（如 https://cdn.example.com）；
--                      空串 = 本地策略，跟随站点域名的 /uploads 路径
-- photos.storage_id:   图片归属的策略；NULL = 历史遗留（沿用旧 url 字段）
ALTER TABLE storages ADD COLUMN access_url VARCHAR(768) NOT NULL DEFAULT '';
ALTER TABLE photos ADD COLUMN storage_id INTEGER;

-- 种子：默认本地策略（幂等）
INSERT OR IGNORE INTO storages (id, name, intro, prefix, provider, access_url, options, created_at, updated_at)
VALUES (1, '本地存储', '跟随站点的本地磁盘', '', 'local', '', NULL, datetime('now'), datetime('now'));
