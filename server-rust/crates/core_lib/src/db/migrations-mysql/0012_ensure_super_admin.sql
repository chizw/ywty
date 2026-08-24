-- 保证系统始终存在超级管理员 · MariaDB/MySQL 版
--
-- MariaDB 不允许 UPDATE 的子查询直接引用目标表，
-- 将条件全部放入派生表后 JOIN 完成更新。
UPDATE users u
JOIN (
    SELECT MIN(id) AS mid FROM users
    WHERE deleted_at IS NULL
      AND role IN ('admin', 'super_admin')
      AND NOT EXISTS (
          SELECT 1 FROM users s
          WHERE s.is_super_admin = 1 AND s.deleted_at IS NULL
      )
) t ON u.id = t.mid
SET u.is_super_admin = 1, u.role = 'admin';
