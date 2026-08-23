-- ============================================================================
-- 工单回复：ticket_replies 表已在 0005_p2_tables.sql 中创建，
-- 此处补充 is_admin 标记用于区分管理员回复与用户回复。
-- ============================================================================

ALTER TABLE ticket_replies ADD COLUMN is_admin INTEGER NOT NULL DEFAULT 0;
