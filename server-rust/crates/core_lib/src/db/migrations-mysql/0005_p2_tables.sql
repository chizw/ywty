-- ============================================================================
-- P2 域：订单 / 套餐 / 优惠券 / 公告 / 页面 / 工单 / 群组 / 许可证 · MariaDB 版
-- 注意：groups 为 MariaDB 保留字，全程使用反引号
-- ============================================================================

-- ----------------------------------------------------------------------------
-- 订单表
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS orders (
    id            BIGINT AUTO_INCREMENT PRIMARY KEY,
    plan_id       BIGINT NOT NULL DEFAULT 0,
    user_id       BIGINT NOT NULL,
    coupon_id     BIGINT NOT NULL DEFAULT 0,
    trade_no      VARCHAR(64) NOT NULL UNIQUE,
    out_trade_no  VARCHAR(64) NOT NULL UNIQUE,
    type          VARCHAR(32) NOT NULL DEFAULT 'plan',
    amount        BIGINT NOT NULL DEFAULT 0,
    deduct_amount BIGINT NOT NULL DEFAULT 0,
    snapshot      TEXT NULL,
    product       VARCHAR(255) NULL,
    pay_method    VARCHAR(32) NOT NULL DEFAULT '',
    status        VARCHAR(16) NOT NULL DEFAULT 'unpaid',
    paid_at       DATETIME NULL,
    canceled_at   DATETIME NULL,
    created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE INDEX idx_orders_user_id ON orders(user_id);
CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_orders_trade_no ON orders(trade_no);

-- ----------------------------------------------------------------------------
-- 套餐表
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS plans (
    id         BIGINT AUTO_INCREMENT PRIMARY KEY,
    type       VARCHAR(32) NOT NULL DEFAULT 'vip',
    name       VARCHAR(190) NOT NULL,
    intro      TEXT NULL,
    features   TEXT NULL,
    badge      VARCHAR(64) NOT NULL DEFAULT '',
    sort       INT NOT NULL DEFAULT 0,
    is_up      TINYINT(1) NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS plan_prices (
    id         BIGINT AUTO_INCREMENT PRIMARY KEY,
    plan_id    BIGINT NOT NULL,
    name       VARCHAR(190) NOT NULL,
    duration   INT NOT NULL DEFAULT 0,
    price      BIGINT NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE INDEX idx_plan_prices_plan_id ON plan_prices(plan_id);

-- ----------------------------------------------------------------------------
-- 优惠券表
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS coupons (
    id          BIGINT AUTO_INCREMENT PRIMARY KEY,
    type        VARCHAR(32) NOT NULL DEFAULT 'direct',
    name        VARCHAR(190) NOT NULL DEFAULT '',
    code        VARCHAR(64) NOT NULL UNIQUE,
    value       DOUBLE NOT NULL DEFAULT 0,
    usage_limit INT NOT NULL DEFAULT 1,
    used_count  INT NOT NULL DEFAULT 0,
    expired_at  DATETIME NULL,
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at  DATETIME NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE INDEX idx_coupons_code ON coupons(code);
CREATE INDEX idx_coupons_expired_at ON coupons(expired_at);

-- ----------------------------------------------------------------------------
-- 公告表
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS notices (
    id         BIGINT AUTO_INCREMENT PRIMARY KEY,
    title      VARCHAR(255) NOT NULL,
    content    TEXT NULL,
    view_count BIGINT NOT NULL DEFAULT 0,
    sort       INT NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ----------------------------------------------------------------------------
-- 页面表
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS pages (
    id          BIGINT AUTO_INCREMENT PRIMARY KEY,
    type        VARCHAR(32) NOT NULL DEFAULT 'internal',
    name        VARCHAR(190) NOT NULL,
    icon        VARCHAR(190) NOT NULL DEFAULT '',
    title       VARCHAR(255) NOT NULL DEFAULT '',
    content     LONGTEXT NULL,
    keywords    VARCHAR(255) NULL,
    description TEXT NULL,
    slug        VARCHAR(190) NOT NULL DEFAULT '',
    url         VARCHAR(768) NOT NULL DEFAULT '',
    view_count  BIGINT NOT NULL DEFAULT 0,
    sort        INT NOT NULL DEFAULT 0,
    is_show     TINYINT(1) NOT NULL DEFAULT 0,
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at  DATETIME NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE INDEX idx_pages_slug ON pages(slug);

-- ----------------------------------------------------------------------------
-- 工单表
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS tickets (
    id         BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_id    BIGINT NOT NULL,
    issue_no   VARCHAR(64) NOT NULL UNIQUE,
    title      VARCHAR(255) NOT NULL,
    type       VARCHAR(32) NOT NULL DEFAULT 'other',
    level      VARCHAR(16) NOT NULL DEFAULT 'low',
    status     VARCHAR(16) NOT NULL DEFAULT 'in_progress',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE INDEX idx_tickets_user_id ON tickets(user_id);
CREATE INDEX idx_tickets_status ON tickets(status);

CREATE TABLE IF NOT EXISTS ticket_replies (
    id         BIGINT AUTO_INCREMENT PRIMARY KEY,
    ticket_id  BIGINT NOT NULL,
    user_id    BIGINT NOT NULL,
    content    TEXT NOT NULL,
    is_notify  TINYINT(1) NOT NULL DEFAULT 1,
    read_at    DATETIME NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE INDEX idx_ticket_replies_ticket_id ON ticket_replies(ticket_id);

-- ----------------------------------------------------------------------------
-- 群组表（groups 为 MariaDB 保留字）
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `groups` (
    id         BIGINT AUTO_INCREMENT PRIMARY KEY,
    name       VARCHAR(190) NOT NULL,
    intro      TEXT NOT NULL,
    options    TEXT NULL,
    is_default TINYINT(1) NOT NULL DEFAULT 0,
    is_guest   TINYINT(1) NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ----------------------------------------------------------------------------
-- 许可证表
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS licenses (
    id           BIGINT AUTO_INCREMENT PRIMARY KEY,
    `key`        VARCHAR(190) NOT NULL UNIQUE,
    status       VARCHAR(16) NOT NULL DEFAULT 'inactive',
    activated_at DATETIME NULL,
    expires_at   DATETIME NULL,
    created_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
