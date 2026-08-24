-- 用户手机号支持 · MariaDB/MySQL 版
ALTER TABLE users ADD COLUMN phone VARCHAR(32) NULL;
ALTER TABLE users ADD COLUMN phone_verified_at DATETIME NULL;

CREATE INDEX idx_users_phone ON users(phone);
