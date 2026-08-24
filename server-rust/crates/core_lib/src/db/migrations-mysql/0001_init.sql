-- ============================================================================
-- ywty 初始迁移 (0001_init) · MariaDB/MySQL 版
-- 覆盖 5 个核心域：用户/相册/图片/认证/存储
-- 时间列统一 DATETIME（UTC，会话时区由应用层固定为 +00:00）
-- ============================================================================

-- ----------------------------------------------------------------------------
-- 用户表
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS users (
    id               BIGINT AUTO_INCREMENT PRIMARY KEY,
    uuid             VARCHAR(64)  NOT NULL UNIQUE,
    username         VARCHAR(190) NOT NULL,
    email            VARCHAR(190) NOT NULL UNIQUE,
    password         TEXT         NOT NULL,
    avatar           TEXT NULL,
    bio              TEXT         NOT NULL DEFAULT '',
    role             VARCHAR(32)  NOT NULL DEFAULT 'user',
    status           INT          NOT NULL DEFAULT 1,
    capacity_used    BIGINT       NOT NULL DEFAULT 0,
    capacity_max     BIGINT       NOT NULL DEFAULT 104857600,
    plan_id          BIGINT NULL,
    plan_expires_at  DATETIME NULL,
    email_verified_at DATETIME NULL,
    last_login_at    DATETIME NULL,
    last_login_ip    VARCHAR(64) NULL,
    created_at       DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at       DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at       DATETIME NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_role ON users(role);
CREATE INDEX idx_users_deleted_at ON users(deleted_at);

-- ----------------------------------------------------------------------------
-- 验证码表
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS verify_codes (
    id         BIGINT AUTO_INCREMENT PRIMARY KEY,
    channel    VARCHAR(16) NOT NULL,
    account    VARCHAR(190) NOT NULL,
    event      VARCHAR(32) NOT NULL,
    code       VARCHAR(16) NOT NULL,
    ip_address VARCHAR(64) NULL,
    used_at    BIGINT NULL,
    expired_at BIGINT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE INDEX idx_verify_codes_channel ON verify_codes(channel);
CREATE INDEX idx_verify_codes_account ON verify_codes(account);
CREATE INDEX idx_verify_codes_event ON verify_codes(event);
CREATE INDEX idx_verify_codes_expired_at ON verify_codes(expired_at);

-- ----------------------------------------------------------------------------
-- 相册表
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS albums (
    id              BIGINT AUTO_INCREMENT PRIMARY KEY,
    uuid            VARCHAR(64) NOT NULL UNIQUE,
    user_id         BIGINT NOT NULL,
    name            VARCHAR(190) NOT NULL,
    description     TEXT NULL,
    cover_photo_id  BIGINT NULL,
    is_public       TINYINT(1) NOT NULL DEFAULT 0,
    photo_count     INT NOT NULL DEFAULT 0,
    views           BIGINT NOT NULL DEFAULT 0,
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at      DATETIME NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE INDEX idx_albums_user_id ON albums(user_id);
CREATE INDEX idx_albums_deleted_at ON albums(deleted_at);

-- ----------------------------------------------------------------------------
-- 相册-图片关联表
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS album_photos (
    album_id    BIGINT NOT NULL,
    photo_id    BIGINT NOT NULL,
    sort_order  INT NOT NULL DEFAULT 0,
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (album_id, photo_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE INDEX idx_album_photos_photo_id ON album_photos(photo_id);

-- ----------------------------------------------------------------------------
-- 图片表
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS photos (
    id             BIGINT AUTO_INCREMENT PRIMARY KEY,
    uuid           VARCHAR(64) NOT NULL UNIQUE,
    user_id        BIGINT NOT NULL,
    album_id       BIGINT NULL,
    filename       VARCHAR(255) NOT NULL,
    original_name  VARCHAR(255) NOT NULL,
    path           VARCHAR(512) NOT NULL,
    url            VARCHAR(768) NOT NULL,
    thumbnail_url  VARCHAR(768) NULL,
    size           BIGINT NOT NULL,
    width          INT NULL,
    height         INT NULL,
    mime_type      VARCHAR(128) NOT NULL,
    md5            VARCHAR(64) NULL,
    sha1           VARCHAR(64) NULL,
    exif           TEXT NULL,
    is_public      TINYINT(1) NOT NULL DEFAULT 0,
    views          BIGINT NOT NULL DEFAULT 0,
    likes          BIGINT NOT NULL DEFAULT 0,
    status         INT NOT NULL DEFAULT 1,
    expired_at     DATETIME NULL,
    created_at     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at     DATETIME NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE INDEX idx_photos_user_id ON photos(user_id);
CREATE INDEX idx_photos_album_id ON photos(album_id);
CREATE INDEX idx_photos_is_public ON photos(is_public);
CREATE INDEX idx_photos_status ON photos(status);
CREATE INDEX idx_photos_deleted_at ON photos(deleted_at);

-- ----------------------------------------------------------------------------
-- 标签表
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS tags (
    id           BIGINT AUTO_INCREMENT PRIMARY KEY,
    name         VARCHAR(190) NOT NULL UNIQUE,
    slug         VARCHAR(190) NOT NULL UNIQUE,
    photo_count  INT NOT NULL DEFAULT 0,
    created_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ----------------------------------------------------------------------------
-- 图片-标签关联表
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS photo_tags (
    photo_id   BIGINT NOT NULL,
    tag_id     BIGINT NOT NULL,
    PRIMARY KEY (photo_id, tag_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE INDEX idx_photo_tags_tag_id ON photo_tags(tag_id);

-- ----------------------------------------------------------------------------
-- 分享表
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS shares (
    id              BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_id         BIGINT NOT NULL,
    shareable_type  VARCHAR(16) NOT NULL,
    shareable_id    BIGINT NOT NULL,
    slug            VARCHAR(64) NOT NULL UNIQUE,
    password        TEXT NULL,
    views           BIGINT NOT NULL DEFAULT 0,
    expires_at      DATETIME NULL,
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at      DATETIME NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE INDEX idx_shares_user_id ON shares(user_id);
CREATE INDEX idx_shares_slug ON shares(slug);

-- ----------------------------------------------------------------------------
-- 点赞表（多态）
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS likes (
    id            BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_id       BIGINT NOT NULL,
    likeable_type VARCHAR(16) NOT NULL,
    likeable_id   BIGINT NOT NULL,
    created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE UNIQUE INDEX idx_likes_unique ON likes(user_id, likeable_type, likeable_id);
CREATE INDEX idx_likes_target ON likes(likeable_type, likeable_id);

-- ----------------------------------------------------------------------------
-- 举报表
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS reports (
    id              BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_id         BIGINT NOT NULL,
    reportable_type VARCHAR(16) NOT NULL,
    reportable_id   BIGINT NOT NULL,
    reason          TEXT NOT NULL,
    status          INT NOT NULL DEFAULT 0,
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE INDEX idx_reports_target ON reports(reportable_type, reportable_id);

-- ----------------------------------------------------------------------------
-- OAuth 账号绑定表
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS oauth_accounts (
    id                BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_id           BIGINT NOT NULL,
    provider          VARCHAR(32) NOT NULL,
    provider_user_id  VARCHAR(190) NOT NULL,
    access_token      TEXT NULL,
    refresh_token     TEXT NULL,
    expires_at        DATETIME NULL,
    created_at        DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at        DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE UNIQUE INDEX idx_oauth_unique ON oauth_accounts(provider, provider_user_id);
CREATE INDEX idx_oauth_user_id ON oauth_accounts(user_id);

-- ----------------------------------------------------------------------------
-- 个人访问令牌表
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS personal_access_tokens (
    id            BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_id       BIGINT NOT NULL,
    name          VARCHAR(190) NOT NULL,
    token         VARCHAR(128) NOT NULL UNIQUE,
    scopes        TEXT NULL,
    last_used_at  DATETIME NULL,
    expires_at    DATETIME NULL,
    created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at    DATETIME NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE INDEX idx_pat_user_id ON personal_access_tokens(user_id);
CREATE INDEX idx_pat_expires_at ON personal_access_tokens(expires_at);

-- ----------------------------------------------------------------------------
-- 存储驱动配置表
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS storages (
    id         BIGINT AUTO_INCREMENT PRIMARY KEY,
    name       VARCHAR(190) NOT NULL,
    intro      TEXT NOT NULL,
    prefix     VARCHAR(255) NOT NULL DEFAULT '',
    provider   VARCHAR(64) NOT NULL,
    options    TEXT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE INDEX idx_storages_provider ON storages(provider);
