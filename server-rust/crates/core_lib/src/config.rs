use serde::Deserialize;
use std::path::Path;

/// 应用配置
#[derive(Debug, Clone, Deserialize)]
pub struct AppConfig {
    pub app: AppConfigApp,
    pub database: DatabaseConfig,
    pub redis: Option<RedisConfig>,
    pub auth: AuthConfig,
    pub storage: StorageConfig,
    pub queue: Option<QueueConfig>,
    pub ratelimit: Option<RateLimitConfig>,
    pub notify: Option<NotifyConfig>,
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
    Upyun,
    Ftp,
    Sftp,
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
    pub sms: Option<SmsConfig>,
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

#[derive(Debug, Clone, Deserialize)]
pub struct SmsConfig {
    pub driver: String,
    pub access_key: Option<String>,
    pub secret_key: Option<String>,
}

impl AppConfig {
    /// 从文件加载配置
    pub fn from_file<P: AsRef<Path>>(path: P) -> crate::AppResult<Self> {
        let builder = config::Config::builder()
            .add_source(config::File::from(path.as_ref()))
            .add_source(
                config::Environment::with_prefix("YWTY")
                    .separator("__")
                    .try_parsing(true),
            );

        builder
            .build()
            .map_err(|e| crate::AppError::Config(e.to_string()))?
            .try_deserialize()
            .map_err(|e| crate::AppError::Config(e.to_string()))
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
        }
    }
}
