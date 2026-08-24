-- 超级管理员标记 · MariaDB/MySQL 版
-- 注意：MariaDB 不允许 UPDATE 的子查询直接引用目标表，改用派生表 JOIN
ALTER TABLE users ADD COLUMN is_super_admin TINYINT(1) NOT NULL DEFAULT 0;

-- 首个注册用户（系统默认）自动成为超级管理员
UPDATE users u
JOIN (
    SELECT MIN(id) AS mid FROM users WHERE deleted_at IS NULL
) t ON u.id = t.mid
SET u.is_super_admin = 1, u.role = 'admin';
