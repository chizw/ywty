-- 保证系统始终存在超级管理员
--
-- 迁移 0006 只对当时已存在的用户生效；此后在空表上由种子创建的默认 admin
-- （以及任何历史遗留库）可能没有任何 is_super_admin = 1 的账号，
-- 导致「设为管理员」等仅超管可用的能力对所有人不可用。
-- 此迁移在"系统中不存在任何超管"时，提升最小 id 的活跃管理员为超管。

UPDATE users
SET is_super_admin = 1, role = 'admin'
WHERE id = (
    SELECT MIN(id) FROM users
    WHERE deleted_at IS NULL AND role IN ('admin', 'super_admin')
)
AND NOT EXISTS (
    SELECT 1 FROM users WHERE is_super_admin = 1 AND deleted_at IS NULL
);
