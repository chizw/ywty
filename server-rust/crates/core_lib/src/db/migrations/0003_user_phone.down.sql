-- 回滚 0003_user_phone 迁移
-- SQLite 不支持 DROP COLUMN（旧版本），此处仅标记
-- 完整回滚需要重建表
DROP INDEX IF EXISTS idx_users_phone;
