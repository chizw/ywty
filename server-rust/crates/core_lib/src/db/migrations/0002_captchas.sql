-- ============================================================================
-- 图片验证码表
-- ============================================================================
CREATE TABLE IF NOT EXISTS captchas (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    captcha_id TEXT NOT NULL UNIQUE,
    code       TEXT NOT NULL,
    ip_address TEXT,
    used_at    INTEGER,
    expired_at INTEGER NOT NULL,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_captchas_captcha_id ON captchas(captcha_id);
CREATE INDEX IF NOT EXISTS idx_captchas_expired_at ON captchas(expired_at);
