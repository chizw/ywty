use serde::Deserialize;
use std::path::Path;

/// 应用配置
///
/// 手动实现 `Debug`：密钥类字段（JWT / 数据库 / Redis / 存储 / OAuth / SMTP）一律脱敏，
/// 防止调试日志或 panic 信息倾泻敏感配置。
#[derive(Clone, Deserialize)]
pub struct AppConfig {
    pub app: AppConfigApp,
    pub database: DatabaseConfig,
    pub redis: Option<RedisConfig>,
    pub auth: AuthConfig,
    pub storage: StorageConfig,
    pub queue: Option<QueueConfig>,
    pub ratelimit: Option<RateLimitConfig>,
    pub notify: Option<NotifyConfig>,
    pub oauth: Option<OAuthConfig>,
    pub watermark: Option<WatermarkConfig>,
}

impl std::fmt::Debug for AppConfig {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        let mail = self
            .notify
            .as_ref()
            .and_then(|n| n.mail.as_ref())
            .map(|m| &m.host);
        f.debug_struct("AppConfig")
            .field("app", &self.app)
            .field("database", &"<redacted>")
            .field("redis", &self.redis.as_ref().map(|r| &r.addr))
            .field("auth.jwt.secret", &"<redacted>")
            .field("storage.driver", &self.storage.driver)
            .field("storage.root", &self.storage.root)
            .field("oauth", &self.oauth.as_ref().map(|_| "<configured>"))
            .field("notify.mail.host", &mail)
            .field(
                "watermark.enabled",
                &self.watermark.as_ref().map(|w| w.enabled),
            )
            .finish()
    }
}

#[derive(Debug, Clone, Deserialize)]
pub struct AppConfigApp {
    pub name: String,
    pub env: String,
    pub host: String,
    pub port: u16,
    pub timezone: String,
    pub base_url: Option<String>,
}

#[derive(Debug, Clone, Deserialize)]
pub struct DatabaseConfig {
    pub driver: DatabaseDriver,
    pub host: Option<String>,
    pub port: Option<u16>,
    pub username: Option<String>,
    pub password: Option<String>,
    pub database: Option<String>,
    pub path: Option<String>,
    pub max_open_conns: Option<u32>,
    pub max_idle_conns: Option<u32>,
}

#[derive(Debug, Clone, Deserialize)]
#[serde(rename_all = "lowercase")]
pub enum DatabaseDriver {
    Sqlite,
    Mysql,
}

#[derive(Debug, Clone, Deserialize)]
pub struct RedisConfig {
    pub addr: String,
    pub password: Option<String>,
    pub db: Option<i64>,
}

#[derive(Debug, Clone, Deserialize)]
pub struct AuthConfig {
    pub jwt: JwtConfig,
}

#[derive(Debug, Clone, Deserialize)]
pub struct JwtConfig {
    pub secret: String,
    pub access_expire: i64,
    pub refresh_expire: i64,
}

#[derive(Debug, Clone, Deserialize)]
pub struct StorageConfig {
    pub driver: StorageDriver,
    pub root: Option<String>,
    pub url: Option<String>,
    pub s3: Option<S3Config>,
    pub oss: Option<OssConfig>,
    pub cos: Option<CosConfig>,
    pub qiniu: Option<QiniuConfig>,
    pub upyun: Option<UpyunConfig>,
}

#[derive(Debug, Clone, Deserialize)]
#[serde(rename_all = "lowercase")]
pub enum StorageDriver {
    Local,
    S3,
    Oss,
    Cos,
    Qiniu,
}

#[derive(Debug, Clone, Deserialize)]
pub struct S3Config {
    pub region: String,
    pub bucket: String,
    pub access_key: String,
    pub secret_key: String,
    pub endpoint: Option<String>,
}

#[derive(Debug, Clone, Deserialize)]
pub struct OssConfig {
    pub endpoint: String,
    pub bucket: String,
    pub access_key: String,
    pub secret_key: String,
}

#[derive(Debug, Clone, Deserialize)]
pub struct CosConfig {
    pub region: String,
    pub bucket: String,
    pub secret_id: String,
    pub secret_key: String,
}

#[derive(Debug, Clone, Deserialize)]
pub struct QiniuConfig {
    pub zone: String,
    pub bucket: String,
    pub access_key: String,
    pub secret_key: String,
    pub domain: String, // 七牛绑定的自定义域名（用于生成公开 URL）
}

#[derive(Debug, Clone, Deserialize)]
pub struct UpyunConfig {
    pub bucket: String,
    pub operator: String,
    pub password: String,
}

#[derive(Debug, Clone, Deserialize)]
pub struct QueueConfig {
    pub concurrency: Option<usize>,
    pub queues: Option<Vec<String>>,
}

#[derive(Debug, Clone, Deserialize)]
pub struct RateLimitConfig {
    pub enable: Option<bool>,
    pub upload_per_minute: Option<u32>,
    pub api_per_minute: Option<u32>,
}

#[derive(Debug, Clone, Deserialize)]
pub struct NotifyConfig {
    pub mail: Option<MailConfig>,
}

#[derive(Debug, Clone, Deserialize)]
pub struct MailConfig {
    pub driver: String,
    pub host: String,
    pub port: u16,
    pub username: String,
    pub password: String,
    pub from: String,
}

/// OAuth 全局配置
#[derive(Debug, Clone, Deserialize)]
pub struct OAuthConfig {
    pub github: Option<OAuthProviderConfig>,
    pub google: Option<OAuthProviderConfig>,
}

/// 单个 OAuth 提供商配置
#[derive(Debug, Clone, Deserialize)]
pub struct OAuthProviderConfig {
    pub client_id: String,
    pub client_secret: String,
    pub redirect_uri: String,
}

/// 水印配置
#[derive(Debug, Clone, Deserialize)]
pub struct WatermarkConfig {
    pub enabled: bool,
    pub text: String,
    pub position: String,
    pub opacity: u8,
    pub font_size: f32,
    pub font_path: Option<String>,
    pub mode: String,
}

impl Default for WatermarkConfig {
    fn default() -> Self {
        Self {
            enabled: false,
            text: "{site_name}".to_string(),
            position: "bottom-right".to_string(),
            opacity: 128,
            font_size: 24.0,
            font_path: None,
            mode: "site_name".to_string(),
        }
    }
}

impl AppConfig {
    /// 从文件加载配置（环境变量覆盖见 [`apply_env_overrides_with`]）
    pub fn from_file<P: AsRef<Path>>(path: P) -> crate::AppResult<Self> {
        let builder = config::Config::builder().add_source(config::File::from(path.as_ref()));

        builder
            .build()
            .map_err(|e| crate::AppError::Config(e.to_string()))?
            .try_deserialize()
            .map_err(|e| crate::AppError::Config(e.to_string()))
            .map(|mut cfg: AppConfig| {
                apply_env_overrides(&mut cfg);
                cfg
            })
    }

    /// 获取数据库连接字符串
    pub fn database_url(&self) -> String {
        match self.database.driver {
            DatabaseDriver::Sqlite => self
                .database
                .path
                .clone()
                .unwrap_or_else(|| "ywty.db".to_string()),
            DatabaseDriver::Mysql => {
                let host = self.database.host.clone().unwrap_or_default();
                let port = self.database.port.unwrap_or(3306);
                let username = self.database.username.clone().unwrap_or_default();
                let password = self.database.password.clone().unwrap_or_default();
                let database = self.database.database.clone().unwrap_or_default();
                format!("mysql://{username}:{password}@{host}:{port}/{database}")
            }
        }
    }

    /// 获取服务器监听地址
    pub fn listen_addr(&self) -> String {
        format!("{}:{}", self.app.host, self.app.port)
    }

    /// 存储公开访问基址（全站唯一推导点）
    ///
    /// 优先级：`storage.url` 显式配置 > `{app.base_url}/uploads` > `http://localhost:{port}/uploads`。
    /// 本地驱动产物存放在 uploads 目录并经该前缀对外提供。
    pub fn storage_public_url(&self) -> String {
        if let Some(url) = self
            .storage
            .url
            .as_deref()
            .map(str::trim)
            .filter(|s| !s.is_empty())
        {
            return url.trim_end_matches('/').to_string();
        }
        match self
            .app
            .base_url
            .as_deref()
            .map(str::trim)
            .filter(|s| !s.is_empty())
        {
            Some(base) => format!("{}/uploads", base.trim_end_matches('/')),
            None => format!("http://localhost:{}/uploads", self.app.port),
        }
    }
}

/// 短环境变量覆盖层（在 YAML 配置之后应用，便于 Docker 零配置部署）
///
/// 规则：
/// - 仅当环境变量存在且取值非空白时才覆盖对应字段；
/// - 数值 / 枚举解析失败时忽略该变量，保留原值。
///
/// | 环境变量 | 覆盖字段 |
/// |---|---|
/// | `PORT` / `HOST` / `APP_ENV` / `APP_NAME` / `APP_URL` | app.port / app.host / app.env / app.name / app.base_url |
/// | `DB_DRIVER`(`sqlite`\|`mysql`) / `DB_PATH` / `DB_HOST` / `DB_PORT` / `DB_USER` / `DB_PASSWORD` / `DB_NAME` | database.driver / path / host / port / username / password / database |
/// | `REDIS_ADDR` / `REDIS_PASSWORD` / `REDIS_DB` | redis.addr / password / db |
/// | `JWT_SECRET` | auth.jwt.secret |
/// | `STORAGE_DRIVER`(`local`\|`s3`\|`oss`\|`cos`\|`qiniu`) / `STORAGE_ROOT` / `STORAGE_URL` | storage.driver / root / url |
/// | `RATELIMIT_ENABLE`(true\|false) | ratelimit.enable |
fn apply_env_overrides_with<F>(cfg: &mut AppConfig, getenv: F)
where
    F: Fn(&str) -> Option<String>,
{
    // 统一取值：trim 后为空视为未设置（Docker compose 常产生空值变量）
    let env = |key: &str| -> Option<String> {
        getenv(key)
            .map(|v| v.trim().to_string())
            .filter(|v| !v.is_empty())
    };

    if let Some(v) = env("PORT") {
        if let Ok(port) = v.parse::<u16>() {
            cfg.app.port = port;
        }
    }
    if let Some(v) = env("HOST") {
        cfg.app.host = v;
    }
    if let Some(v) = env("APP_ENV") {
        cfg.app.env = v;
    }
    if let Some(v) = env("APP_NAME") {
        cfg.app.name = v;
    }
    if let Some(v) = env("APP_URL") {
        cfg.app.base_url = Some(v);
    }

    if let Some(v) = env("DB_DRIVER") {
        match v.to_ascii_lowercase().as_str() {
            "sqlite" => cfg.database.driver = DatabaseDriver::Sqlite,
            "mysql" => cfg.database.driver = DatabaseDriver::Mysql,
            _ => {}
        }
    }
    if let Some(v) = env("DB_PATH") {
        cfg.database.path = Some(v);
    }
    if let Some(v) = env("DB_HOST") {
        cfg.database.host = Some(v);
    }
    if let Some(v) = env("DB_PORT") {
        if let Ok(port) = v.parse::<u16>() {
            cfg.database.port = Some(port);
        }
    }
    if let Some(v) = env("DB_USER") {
        cfg.database.username = Some(v);
    }
    if let Some(v) = env("DB_PASSWORD") {
        cfg.database.password = Some(v);
    }
    if let Some(v) = env("DB_NAME") {
        cfg.database.database = Some(v);
    }

    if let Some(addr) = env("REDIS_ADDR") {
        let redis = cfg.redis.get_or_insert(RedisConfig {
            addr: String::new(),
            password: None,
            db: None,
        });
        redis.addr = addr;
    }
    if let Some(redis) = cfg.redis.as_mut() {
        if let Some(v) = env("REDIS_PASSWORD") {
            redis.password = Some(v);
        }
        if let Some(v) = env("REDIS_DB") {
            if let Ok(db) = v.parse::<i64>() {
                redis.db = Some(db);
            }
        }
    }

    if let Some(v) = env("JWT_SECRET") {
        cfg.auth.jwt.secret = v;
    }

    if let Some(v) = env("STORAGE_DRIVER") {
        match v.to_ascii_lowercase().as_str() {
            "local" => cfg.storage.driver = StorageDriver::Local,
            "s3" => cfg.storage.driver = StorageDriver::S3,
            "oss" => cfg.storage.driver = StorageDriver::Oss,
            "cos" => cfg.storage.driver = StorageDriver::Cos,
            "qiniu" => cfg.storage.driver = StorageDriver::Qiniu,
            _ => {}
        }
    }
    if let Some(v) = env("STORAGE_ROOT") {
        cfg.storage.root = Some(v);
    }
    if let Some(v) = env("STORAGE_URL") {
        cfg.storage.url = Some(v);
    }
    if let Some(v) = env("RATELIMIT_ENABLE") {
        let enabled = matches!(v.to_ascii_lowercase().as_str(), "1" | "true" | "on" | "yes");
        match cfg.ratelimit.as_mut() {
            Some(rl) => rl.enable = Some(enabled),
            None => {
                cfg.ratelimit = Some(RateLimitConfig {
                    enable: Some(enabled),
                    upload_per_minute: None,
                    api_per_minute: None,
                })
            }
        }
    }
}

/// 生产入口：从真实进程环境读取短变量并覆盖配置
pub fn apply_env_overrides(cfg: &mut AppConfig) {
    apply_env_overrides_with(cfg, |key| std::env::var(key).ok());
}

impl Default for AppConfig {
    fn default() -> Self {
        Self {
            app: AppConfigApp {
                name: "云雾图驿".to_string(),
                env: "dev".to_string(),
                host: "0.0.0.0".to_string(),
                port: 3000,
                timezone: "Asia/Shanghai".to_string(),
                base_url: None,
            },
            database: DatabaseConfig {
                driver: DatabaseDriver::Sqlite,
                host: None,
                port: None,
                username: None,
                password: None,
                database: None,
                path: Some("ywty.db".to_string()),
                max_open_conns: Some(100),
                max_idle_conns: Some(10),
            },
            redis: None,
            auth: AuthConfig {
                jwt: JwtConfig {
                    secret: "please-change-me-in-production".to_string(),
                    access_expire: 7200,
                    refresh_expire: 2592000,
                },
            },
            storage: StorageConfig {
                driver: StorageDriver::Local,
                root: Some("./uploads".to_string()),
                url: None,
                s3: None,
                oss: None,
                cos: None,
                qiniu: None,
                upyun: None,
            },
            queue: None,
            ratelimit: None,
            notify: None,
            oauth: None,
            watermark: Some(WatermarkConfig::default()),
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::collections::HashMap;

    /// 构造 mock 取值闭包（避免 std::env::set_var 的 unsafe 与并行测试竞争）
    fn env_from<'a>(pairs: &'a [(&'a str, &'a str)]) -> impl Fn(&str) -> Option<String> + 'a {
        let map: HashMap<String, String> = pairs
            .iter()
            .map(|(k, v)| (k.to_string(), v.to_string()))
            .collect();
        move |key: &str| map.get(key).cloned()
    }

    #[test]
    fn short_env_overrides_apply() {
        let mut cfg = AppConfig::default();
        apply_env_overrides_with(
            &mut cfg,
            env_from(&[
                ("PORT", "8123"),
                ("JWT_SECRET", "unit-test-secret"),
                ("APP_URL", "https://img.example.com"),
            ]),
        );

        assert_eq!(cfg.app.port, 8123);
        assert_eq!(cfg.auth.jwt.secret, "unit-test-secret");
        assert_eq!(cfg.app.base_url.as_deref(), Some("https://img.example.com"));
    }

    #[test]
    fn invalid_or_blank_values_are_ignored() {
        let mut cfg = AppConfig::default();
        apply_env_overrides_with(
            &mut cfg,
            env_from(&[("PORT", "not-a-number"), ("JWT_SECRET", "   ")]),
        );

        assert_eq!(cfg.app.port, 3000);
        assert_eq!(cfg.auth.jwt.secret, "please-change-me-in-production");
    }

    #[test]
    fn storage_public_url_prefers_explicit_then_base_url() {
        let mut cfg = AppConfig::default();
        assert_eq!(
            cfg.storage_public_url(),
            format!("http://localhost:{}/uploads", cfg.app.port)
        );

        cfg.app.base_url = Some("http://example.com/".to_string());
        assert_eq!(cfg.storage_public_url(), "http://example.com/uploads");

        cfg.storage.url = Some("https://cdn.example.com".to_string());
        assert_eq!(cfg.storage_public_url(), "https://cdn.example.com");
    }
}
