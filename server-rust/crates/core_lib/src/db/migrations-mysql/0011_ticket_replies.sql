-- 工单回复 is_admin 标记 · MariaDB/MySQL 版
ALTER TABLE ticket_replies ADD COLUMN is_admin TINYINT(1) NOT NULL DEFAULT 0;
