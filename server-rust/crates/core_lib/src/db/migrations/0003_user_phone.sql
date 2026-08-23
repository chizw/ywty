-- ============================================================================
-- 用户手机号支持
-- 支持：reset-password/phone、change-phone
-- ============================================================================

ALTER TABLE users ADD COLUMN phone TEXT;
ALTER TABLE users ADD COLUMN phone_verified_at TEXT;

CREATE INDEX IF NOT EXISTS idx_users_phone ON users(phone);
