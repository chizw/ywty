-- ============================================================================
-- P2 域：订单 / 套餐 / 优惠券 / 公告 / 页面 / 工单 / 群组 / 许可证
-- ============================================================================

-- ----------------------------------------------------------------------------
-- 订单表
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS orders (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    plan_id       INTEGER NOT NULL DEFAULT 0,
    user_id       INTEGER NOT NULL,
    coupon_id     INTEGER NOT NULL DEFAULT 0,
    trade_no      TEXT NOT NULL UNIQUE,
    out_trade_no  TEXT NOT NULL UNIQUE,
    type          TEXT NOT NULL DEFAULT 'plan',
    amount        INTEGER NOT NULL DEFAULT 0,
    deduct_amount INTEGER NOT NULL DEFAULT 0,
    snapshot      TEXT,
    product       TEXT,
    pay_method    TEXT NOT NULL DEFAULT '',
    status        TEXT NOT NULL DEFAULT 'unpaid',
    paid_at       TEXT,
    canceled_at   TEXT,
    created_at    TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at    TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_orders_user_id ON orders(user_id);
CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(status);
CREATE INDEX IF NOT EXISTS idx_orders_trade_no ON orders(trade_no);

-- ----------------------------------------------------------------------------
-- 套餐表
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS plans (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    type       TEXT NOT NULL DEFAULT 'vip',
    name       TEXT NOT NULL,
    intro      TEXT,
    features   TEXT,
    badge      TEXT NOT NULL DEFAULT '',
    sort       INTEGER NOT NULL DEFAULT 0,
    is_up      INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now')),
    deleted_at TEXT
);

CREATE TABLE IF NOT EXISTS plan_prices (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    plan_id    INTEGER NOT NULL,
    name       TEXT NOT NULL,
    duration   INTEGER NOT NULL DEFAULT 0,
    price      INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_plan_prices_plan_id ON plan_prices(plan_id);

-- ----------------------------------------------------------------------------
-- 优惠券表
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS coupons (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    type        TEXT NOT NULL DEFAULT 'direct',
    name        TEXT NOT NULL DEFAULT '',
    code        TEXT NOT NULL UNIQUE,
    value       REAL NOT NULL DEFAULT 0,
    usage_limit INTEGER NOT NULL DEFAULT 1,
    used_count  INTEGER NOT NULL DEFAULT 0,
    expired_at  TEXT,
    created_at  TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at  TEXT NOT NULL DEFAULT (datetime('now')),
    deleted_at  TEXT
);

CREATE INDEX IF NOT EXISTS idx_coupons_code ON coupons(code);
CREATE INDEX IF NOT EXISTS idx_coupons_expired_at ON coupons(expired_at);

-- ----------------------------------------------------------------------------
-- 公告表
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS notices (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    title      TEXT NOT NULL,
    content    TEXT,
    view_count INTEGER NOT NULL DEFAULT 0,
    sort       INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now')),
    deleted_at TEXT
);

-- ----------------------------------------------------------------------------
-- 页面表
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS pages (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    type        TEXT NOT NULL DEFAULT 'internal',
    name        TEXT NOT NULL,
    icon        TEXT NOT NULL DEFAULT '',
    title       TEXT NOT NULL DEFAULT '',
    content     TEXT,
    keywords    TEXT,
    description TEXT,
    slug        TEXT NOT NULL DEFAULT '',
    url         TEXT NOT NULL DEFAULT '',
    view_count  INTEGER NOT NULL DEFAULT 0,
    sort        INTEGER NOT NULL DEFAULT 0,
    is_show     INTEGER NOT NULL DEFAULT 0,
    created_at  TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at  TEXT NOT NULL DEFAULT (datetime('now')),
    deleted_at  TEXT
);

CREATE INDEX IF NOT EXISTS idx_pages_slug ON pages(slug);

-- ----------------------------------------------------------------------------
-- 工单表
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS tickets (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id    INTEGER NOT NULL,
    issue_no   TEXT NOT NULL UNIQUE,
    title      TEXT NOT NULL,
    type       TEXT NOT NULL DEFAULT 'other',
    level      TEXT NOT NULL DEFAULT 'low',
    status     TEXT NOT NULL DEFAULT 'in_progress',
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now')),
    deleted_at TEXT
);

CREATE INDEX IF NOT EXISTS idx_tickets_user_id ON tickets(user_id);
CREATE INDEX IF NOT EXISTS idx_tickets_status ON tickets(status);

CREATE TABLE IF NOT EXISTS ticket_replies (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    ticket_id  INTEGER NOT NULL,
    user_id    INTEGER NOT NULL,
    content    TEXT NOT NULL,
    is_notify  INTEGER NOT NULL DEFAULT 1,
    read_at    TEXT,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_ticket_replies_ticket_id ON ticket_replies(ticket_id);

-- ----------------------------------------------------------------------------
-- 群组表
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS groups (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    name       TEXT NOT NULL,
    intro      TEXT NOT NULL DEFAULT '',
    options    TEXT,
    is_default INTEGER NOT NULL DEFAULT 0,
    is_guest   INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now')),
    deleted_at TEXT
);

-- ----------------------------------------------------------------------------
-- 许可证表
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS licenses (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    key        TEXT NOT NULL UNIQUE,
    status     TEXT NOT NULL DEFAULT 'inactive',
    activated_at TEXT,
    expires_at TEXT,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);
