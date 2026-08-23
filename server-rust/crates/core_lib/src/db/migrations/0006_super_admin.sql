-- 超级管理员：新增标记列
ALTER TABLE users ADD COLUMN is_super_admin INTEGER NOT NULL DEFAULT 0;

-- 首个注册用户（系统默认）自动成为超级管理员
UPDATE users
SET is_super_admin = 1, role = 'admin'
WHERE id = (SELECT MIN(id) FROM users WHERE deleted_at IS NULL);
