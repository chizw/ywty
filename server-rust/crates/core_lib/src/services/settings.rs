//! 全局设置服务（settings 表键值存储）
//!
//! 提供对 `settings` 表的读写：以 `&DbPool` 为参数的自由函数，
//! 以及常用类型的辅助读取（bool / i64）。键名统一在 [`keys`] 模块中维护。

use crate::db::DbPool;

use crate::error::AppResult;

/// 设置键名常量
pub mod keys {
    // ---- 邮件服务器（SMTP）----
    pub const MAIL_SMTP_HOST: &str = "mail.smtp.host";
    pub const MAIL_SMTP_PORT: &str = "mail.smtp.port";
    pub const MAIL_SMTP_USERNAME: &str = "mail.smtp.username";
    pub const MAIL_SMTP_PASSWORD: &str = "mail.smtp.password";
    pub const MAIL_SMTP_FROM: &str = "mail.smtp.from";
    pub const MAIL_SMTP_SSL: &str = "mail.smtp.ssl";

    // ---- 安全开关 ----
    /// 注册是否需要邮箱验证码（默认 true）
    pub const SECURITY_REQUIRE_EMAIL_VERIFY: &str = "security.require_email_verify";
    /// 是否开放找回密码（默认 true）
    pub const SECURITY_ALLOW_PASSWORD_RESET: &str = "security.allow_password_reset";
    /// 是否允许注册（默认 true）
    pub const SECURITY_ALLOW_REGISTER: &str = "security.allow_register";
    /// 是否强制图形验证码（注册/找回密码，默认 false）
    pub const SECURITY_REQUIRE_CAPTCHA: &str = "security.require_captcha";

    // ---- 站点信息 ----
    /// 站点名称（默认"云雾图驿"）
    pub const SITE_NAME: &str = "site.name";
    /// 站点描述（默认"自托管图床 / 云相册"）
    pub const SITE_DESCRIPTION: &str = "site.description";
    /// SEO 关键词（可空）
    pub const SITE_KEYWORDS: &str = "site.keywords";
    /// 页脚文本（可空）
    pub const SITE_FOOTER: &str = "site.footer";
    /// ICP 备案号（可空）
    pub const SITE_ICP: &str = "site.icp";

    // ---- OAuth 第三方登录 ----
    pub const OAUTH_GITHUB_CLIENT_ID: &str = "oauth.github.client_id";
    pub const OAUTH_GITHUB_CLIENT_SECRET: &str = "oauth.github.client_secret";
    pub const OAUTH_GITHUB_REDIRECT_URI: &str = "oauth.github.redirect_uri";
    pub const OAUTH_GOOGLE_CLIENT_ID: &str = "oauth.google.client_id";
    pub const OAUTH_GOOGLE_CLIENT_SECRET: &str = "oauth.google.client_secret";
    pub const OAUTH_GOOGLE_REDIRECT_URI: &str = "oauth.google.redirect_uri";

    /// 敏感键清单：管理接口返回时脱敏为"是否已设置"布尔值
    pub const SENSITIVE_KEYS: &[&str] = &[
        MAIL_SMTP_PASSWORD,
        OAUTH_GITHUB_CLIENT_SECRET,
        OAUTH_GOOGLE_CLIENT_SECRET,
    ];

    /// 允许通过管理接口写入的键白名单
    pub const ALLOWED_KEYS: &[&str] = &[
        MAIL_SMTP_HOST,
        MAIL_SMTP_PORT,
        MAIL_SMTP_USERNAME,
        MAIL_SMTP_PASSWORD,
        MAIL_SMTP_FROM,
        MAIL_SMTP_SSL,
        SECURITY_REQUIRE_EMAIL_VERIFY,
        SECURITY_ALLOW_PASSWORD_RESET,
        SECURITY_ALLOW_REGISTER,
        SECURITY_REQUIRE_CAPTCHA,
        SITE_NAME,
        SITE_DESCRIPTION,
        SITE_KEYWORDS,
        SITE_FOOTER,
        SITE_ICP,
        OAUTH_GITHUB_CLIENT_ID,
        OAUTH_GITHUB_CLIENT_SECRET,
        OAUTH_GITHUB_REDIRECT_URI,
        OAUTH_GOOGLE_CLIENT_ID,
        OAUTH_GOOGLE_CLIENT_SECRET,
        OAUTH_GOOGLE_REDIRECT_URI,
    ];
}

/// 读取单个设置项，不存在时返回 None
pub async fn get(pool: &DbPool, key: &str) -> AppResult<Option<String>> {
    let row: Option<(String,)> = sqlx::query_as("SELECT value FROM settings WHERE `key` = ?")
        .bind(key)
        .fetch_optional(pool)
        .await?;
    Ok(row.map(|r| r.0))
}

/// 写入（upsert）单个设置项
pub async fn set(pool: &DbPool, key: &str, value: &str) -> AppResult<()> {
    let now = crate::db::now_str();

    // 冲突处理语法两方言不同：SQLite 用 ON CONFLICT，MySQL/MariaDB 用 ON DUPLICATE KEY
    #[cfg(feature = "mysql")]
    let sql = r#"
        INSERT INTO settings (`key`, value, updated_at)
        VALUES (?, ?, ?)
        ON DUPLICATE KEY UPDATE value = VALUES(value), updated_at = VALUES(updated_at)
        "#;
    #[cfg(not(feature = "mysql"))]
    let sql = r#"
        INSERT INTO settings (`key`, value, updated_at)
        VALUES (?, ?, ?)
        ON CONFLICT(`key`) DO UPDATE SET value = excluded.value, updated_at = excluded.updated_at
        "#;

    sqlx::query(sql)
        .bind(key)
        .bind(value)
        .bind(&now)
        .execute(pool)
        .await?;
    Ok(())
}

/// 读取全部设置项（键升序）
pub async fn get_all(pool: &DbPool) -> AppResult<Vec<(String, String)>> {
    let rows: Vec<(String, String)> =
        sqlx::query_as("SELECT `key`, value FROM settings ORDER BY `key`")
            .fetch_all(pool)
            .await?;
    Ok(rows)
}

/// 读取布尔设置：值为 "1"/"true"/"on"/"yes" 视为 true；不存在时返回 default
pub async fn get_bool(pool: &DbPool, key: &str, default: bool) -> AppResult<bool> {
    match get(pool, key).await? {
        Some(v) => {
            let v = v.trim().to_ascii_lowercase();
            Ok(matches!(v.as_str(), "1" | "true" | "on" | "yes"))
        }
        None => Ok(default),
    }
}

/// 读取整数设置：解析失败或不存在时返回 default
pub async fn get_i64(pool: &DbPool, key: &str, default: i64) -> AppResult<i64> {
    match get(pool, key).await? {
        Some(v) => Ok(v.trim().parse().unwrap_or(default)),
        None => Ok(default),
    }
}
