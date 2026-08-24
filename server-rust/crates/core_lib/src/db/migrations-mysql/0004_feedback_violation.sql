-- 意见反馈 + 违规记录表 · MariaDB/MySQL 版
CREATE TABLE IF NOT EXISTS feedbacks (
    id         BIGINT AUTO_INCREMENT PRIMARY KEY,
    type       VARCHAR(32) NOT NULL DEFAULT 'general',
    title      VARCHAR(255) NOT NULL,
    name       VARCHAR(190) NOT NULL,
    email      VARCHAR(190) NOT NULL,
    content    TEXT NOT NULL,
    ip_address VARCHAR(64) NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE INDEX idx_feedbacks_type ON feedbacks(type);
CREATE INDEX idx_feedbacks_created_at ON feedbacks(created_at);

CREATE TABLE IF NOT EXISTS violations (
    id         BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_id    BIGINT NOT NULL DEFAULT 0,
    photo_id   BIGINT NOT NULL DEFAULT 0,
    reason     TEXT NOT NULL,
    status     VARCHAR(16) NOT NULL DEFAULT 'unhandled',
    handled_at DATETIME NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE INDEX idx_violations_photo_id ON violations(photo_id);
CREATE INDEX idx_violations_status ON violations(status);
