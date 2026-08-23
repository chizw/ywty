-- ============================================================================
-- ywty 初始迁移 (0001_init)
-- 覆盖 5 个核心域：用户/相册/图片/认证/存储
-- ============================================================================

-- ----------------------------------------------------------------------------
-- 用户表
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS users (
    id               INTEGER PRIMARY KEY AUTOINCREMENT,
    uuid             TEXT NOT NULL UNIQUE,
    username         TEXT NOT NULL,
    email            TEXT NOT NULL UNIQUE,
    password         TEXT NOT NULL,
    avatar           TEXT,
    bio              TEXT NOT NULL DEFAULT '',
    role             TEXT NOT NULL DEFAULT 'user',
    status           INTEGER NOT NULL DEFAULT 1,
    capacity_used    INTEGER NOT NULL DEFAULT 0,
    capacity_max     INTEGER NOT NULL DEFAULT 104857600,
    plan_id          INTEGER,
    plan_expires_at  TEXT,
    email_verified_at TEXT,
    last_login_at    TEXT,
    last_login_ip    TEXT,
    created_at       TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at       TEXT NOT NULL DEFAULT (datetime('now')),
    deleted_at       TEXT
);

CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);
CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users(deleted_at);

-- ----------------------------------------------------------------------------
-- 验证码表
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS verify_codes (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    channel    TEXT NOT NULL,
    account    TEXT NOT NULL,
    event      TEXT NOT NULL,
    code       TEXT NOT NULL,
    ip_address TEXT,
    used_at    INTEGER,
    expired_at INTEGER NOT NULL,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now')),
    deleted_at TEXT
);

CREATE INDEX IF NOT EXISTS idx_verify_codes_channel ON verify_codes(channel);
CREATE INDEX IF NOT EXISTS idx_verify_codes_account ON verify_codes(account);
CREATE INDEX IF NOT EXISTS idx_verify_codes_event ON verify_codes(event);
CREATE INDEX IF NOT EXISTS idx_verify_codes_expired_at ON verify_codes(expired_at);

-- ----------------------------------------------------------------------------
-- 相册表
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS albums (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    uuid            TEXT NOT NULL UNIQUE,
    user_id         INTEGER NOT NULL,
    name            TEXT NOT NULL,
    description     TEXT,
    cover_photo_id  INTEGER,
    is_public       INTEGER NOT NULL DEFAULT 0,
    photo_count     INTEGER NOT NULL DEFAULT 0,
    views           INTEGER NOT NULL DEFAULT 0,
    created_at      TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at      TEXT NOT NULL DEFAULT (datetime('now')),
    deleted_at      TEXT
);

CREATE INDEX IF NOT EXISTS idx_albums_user_id ON albums(user_id);
CREATE INDEX IF NOT EXISTS idx_albums_deleted_at ON albums(deleted_at);

-- ----------------------------------------------------------------------------
-- 相册-图片关联表
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS album_photos (
    album_id    INTEGER NOT NULL,
    photo_id    INTEGER NOT NULL,
    sort_order  INTEGER NOT NULL DEFAULT 0,
    created_at  TEXT NOT NULL DEFAULT (datetime('now')),
    PRIMARY KEY (album_id, photo_id)
);

CREATE INDEX IF NOT EXISTS idx_album_photos_photo_id ON album_photos(photo_id);

-- ----------------------------------------------------------------------------
-- 图片表
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS photos (
    id             INTEGER PRIMARY KEY AUTOINCREMENT,
    uuid           TEXT NOT NULL UNIQUE,
    user_id        INTEGER NOT NULL,
    album_id       INTEGER,
    filename       TEXT NOT NULL,
    original_name  TEXT NOT NULL,
    path           TEXT NOT NULL,
    url            TEXT NOT NULL,
    thumbnail_url  TEXT,
    size           INTEGER NOT NULL,
    width          INTEGER,
    height         INTEGER,
    mime_type      TEXT NOT NULL,
    md5            TEXT,
    sha1           TEXT,
    exif           TEXT,
    is_public      INTEGER NOT NULL DEFAULT 0,
    views          INTEGER NOT NULL DEFAULT 0,
    likes          INTEGER NOT NULL DEFAULT 0,
    status         INTEGER NOT NULL DEFAULT 1,
    expired_at     TEXT,
    created_at     TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at     TEXT NOT NULL DEFAULT (datetime('now')),
    deleted_at     TEXT
);

CREATE INDEX IF NOT EXISTS idx_photos_user_id ON photos(user_id);
CREATE INDEX IF NOT EXISTS idx_photos_album_id ON photos(album_id);
CREATE INDEX IF NOT EXISTS idx_photos_is_public ON photos(is_public);
CREATE INDEX IF NOT EXISTS idx_photos_status ON photos(status);
CREATE INDEX IF NOT EXISTS idx_photos_deleted_at ON photos(deleted_at);

-- ----------------------------------------------------------------------------
-- 标签表
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS tags (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    name         TEXT NOT NULL UNIQUE,
    slug         TEXT NOT NULL UNIQUE,
    photo_count  INTEGER NOT NULL DEFAULT 0,
    created_at   TEXT NOT NULL DEFAULT (datetime('now'))
);

-- ----------------------------------------------------------------------------
-- 图片-标签关联表
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS photo_tags (
    photo_id   INTEGER NOT NULL,
    tag_id     INTEGER NOT NULL,
    PRIMARY KEY (photo_id, tag_id)
);

CREATE INDEX IF NOT EXISTS idx_photo_tags_tag_id ON photo_tags(tag_id);

-- ----------------------------------------------------------------------------
-- 分享表
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS shares (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id         INTEGER NOT NULL,
    shareable_type  TEXT NOT NULL,
    shareable_id    INTEGER NOT NULL,
    slug            TEXT NOT NULL UNIQUE,
    password        TEXT,
    views           INTEGER NOT NULL DEFAULT 0,
    expires_at      TEXT,
    created_at      TEXT NOT NULL DEFAULT (datetime('now')),
    deleted_at      TEXT
);

CREATE INDEX IF NOT EXISTS idx_shares_user_id ON shares(user_id);
CREATE INDEX IF NOT EXISTS idx_shares_slug ON shares(slug);

-- ----------------------------------------------------------------------------
-- 点赞表 (多态)
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS likes (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id       INTEGER NOT NULL,
    likeable_type TEXT NOT NULL,
    likeable_id   INTEGER NOT NULL,
    created_at    TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_likes_unique ON likes(user_id, likeable_type, likeable_id);
CREATE INDEX IF NOT EXISTS idx_likes_target ON likes(likeable_type, likeable_id);

-- ----------------------------------------------------------------------------
-- 举报表
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS reports (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id         INTEGER NOT NULL,
    reportable_type TEXT NOT NULL,
    reportable_id   INTEGER NOT NULL,
    reason          TEXT NOT NULL,
    status          INTEGER NOT NULL DEFAULT 0,
    created_at      TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at      TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_reports_target ON reports(reportable_type, reportable_id);

-- ----------------------------------------------------------------------------
-- OAuth 账号绑定表
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS oauth_accounts (
    id                INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id           INTEGER NOT NULL,
    provider          TEXT NOT NULL,
    provider_user_id  TEXT NOT NULL,
    access_token      TEXT,
    refresh_token     TEXT,
    expires_at        TEXT,
    created_at        TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at        TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_oauth_unique ON oauth_accounts(provider, provider_user_id);
CREATE INDEX IF NOT EXISTS idx_oauth_user_id ON oauth_accounts(user_id);

-- ----------------------------------------------------------------------------
-- 个人访问令牌表
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS personal_access_tokens (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id       INTEGER NOT NULL,
    name          TEXT NOT NULL,
    token         TEXT NOT NULL UNIQUE,
    scopes        TEXT,
    last_used_at  TEXT,
    expires_at    TEXT,
    created_at    TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at    TEXT NOT NULL DEFAULT (datetime('now')),
    deleted_at    TEXT
);

CREATE INDEX IF NOT EXISTS idx_pat_user_id ON personal_access_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_pat_expires_at ON personal_access_tokens(expires_at);

-- ----------------------------------------------------------------------------
-- 存储驱动配置表
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS storages (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    name       TEXT NOT NULL,
    intro      TEXT NOT NULL DEFAULT '',
    prefix     TEXT NOT NULL DEFAULT '',
    provider   TEXT NOT NULL,
    options    TEXT,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now')),
    deleted_at TEXT
);

CREATE INDEX IF NOT EXISTS idx_storages_provider ON storages(provider);
