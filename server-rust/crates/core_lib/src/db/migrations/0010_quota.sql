-- ============================================================================
-- 0010_quota: 组配额 + 用户单独覆盖
-- groups.max_storage: 角色组存储配额（字节），NULL = 不限
-- users.quota_override: 用户单独配额覆盖（字节），NULL = 跟随角色组
-- ============================================================================

ALTER TABLE groups ADD COLUMN max_storage INTEGER;
ALTER TABLE users ADD COLUMN quota_override INTEGER;
