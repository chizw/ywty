-- 图片验证码表 · MariaDB/MySQL 版
CREATE TABLE IF NOT EXISTS captchas (
    id         BIGINT AUTO_INCREMENT PRIMARY KEY,
    captcha_id VARCHAR(64) NOT NULL UNIQUE,
    code       VARCHAR(16) NOT NULL,
    ip_address VARCHAR(64) NULL,
    used_at    BIGINT NULL,
    expired_at BIGINT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE INDEX idx_captchas_captcha_id ON captchas(captcha_id);
CREATE INDEX idx_captchas_expired_at ON captchas(expired_at);
