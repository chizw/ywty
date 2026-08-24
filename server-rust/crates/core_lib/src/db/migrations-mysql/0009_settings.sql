-- 全局设置表（键值存储） · MariaDB/MySQL 版
CREATE TABLE IF NOT EXISTS settings (
    `key`      VARCHAR(190) PRIMARY KEY,
    value      TEXT NOT NULL,
    updated_at DATETIME NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
