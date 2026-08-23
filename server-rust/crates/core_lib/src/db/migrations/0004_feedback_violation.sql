-- ============================================================================
-- 意见反馈 + 违规记录表
-- ============================================================================

CREATE TABLE IF NOT EXISTS feedbacks (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    type       TEXT NOT NULL DEFAULT 'general',
    title      TEXT NOT NULL,
    name       TEXT NOT NULL,
    email      TEXT NOT NULL,
    content    TEXT NOT NULL,
    ip_address TEXT,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_feedbacks_type ON feedbacks(type);
CREATE INDEX IF NOT EXISTS idx_feedbacks_created_at ON feedbacks(created_at);

CREATE TABLE IF NOT EXISTS violations (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id    INTEGER NOT NULL DEFAULT 0,
    photo_id   INTEGER NOT NULL DEFAULT 0,
    reason     TEXT NOT NULL,
    status     TEXT NOT NULL DEFAULT 'unhandled',
    handled_at TEXT,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_violations_photo_id ON violations(photo_id);
CREATE INDEX IF NOT EXISTS idx_violations_status ON violations(status);
