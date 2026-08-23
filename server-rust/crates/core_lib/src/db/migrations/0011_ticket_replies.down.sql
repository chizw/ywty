-- 回滚：移除工单回复的管理员标记
ALTER TABLE ticket_replies DROP COLUMN is_admin;
