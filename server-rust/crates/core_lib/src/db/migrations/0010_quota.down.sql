-- 回滚 0010_quota 迁移
-- SQLite 3.35+ 支持 DROP COLUMN
ALTER TABLE users DROP COLUMN quota_override;
ALTER TABLE groups DROP COLUMN max_storage;
