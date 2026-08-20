//! 应用入口（配置加载、状态构建、迁移、路由、服务启动）
//!
//! 对应 Go 后端的 `cmd/server/main.go`。二进制 crate `api` 仅保留一个
//! 调用 [`run`] 的 `thin main`，使最终二进制 crate 自身的代码量保持极小，
//! 以规避 Rust 1.97/LLVM 22 在大体积二进制 crate 上的代码生成崩溃问题。

use std::sync::Arc;

use anyhow::Result;
use tokio::net::TcpListener;
use tokio::signal;
use tracing_subscriber::{layer::SubscriberExt, util::SubscriberInitExt};

use crate::auth::JwtAuth;
use crate::config::AppConfig;
use crate::db;
use crate::router;

/// 应用共享状态（所有服务与连接池的聚合）
#[derive(Clone)]
pub struct AppState {
    pub config: Arc<AppConfig>,
    pub db: sqlx::SqlitePool,
    pub jwt: Arc<JwtAuth>,
    // 服务
    pub auth_svc: crate::services::auth::AuthService,
    pub user_svc: crate::services::user::UserService,
    pub photo_svc: crate::services::photo::PhotoService,
    pub album_svc: crate::services::album::AlbumService,
    pub storage_svc: crate::services::storage::StorageService,
}

impl AppState {
    /// 从配置构建应用状态
    pub async fn from_config(config: AppConfig) -> crate::error::AppResult<Self> {
        let db_pool = db::create_pool(&config.database).await?;
        let jwt = Arc::new(JwtAuth::new(&config.auth.jwt));

        let public_url = config
            .storage
            .url
            .clone()
            .unwrap_or_else(|| "http://localhost:3000".to_string());
        let upload_root = config
            .storage
            .root
            .clone()
            .unwrap_or_else(|| "./uploads".to_string());

        // 构建邮件服务
        let mail_svc = config
            .notify
            .as_ref()
            .and_then(|n| n.mail.as_ref())
            .map(|mail_cfg| crate::services::mail::MailService::new(mail_cfg))
            .unwrap_or_else(crate::services::mail::MailService::disabled);

        Ok(Self {
            config: Arc::new(config),
            db: db_pool.clone(),
            jwt: jwt.clone(),
            auth_svc: crate::services::auth::AuthService::new(
                db_pool.clone(),
                jwt.as_ref().clone(),
                mail_svc,
            ),
            user_svc: crate::services::user::UserService::new(db_pool.clone()),
            photo_svc: crate::services::photo::PhotoService::new(
                db_pool.clone(),
                public_url.clone(),
                upload_root.clone(),
            ),
            album_svc: crate::services::album::AlbumService::new(db_pool.clone()),
            storage_svc: crate::services::storage::StorageService::new(
                db_pool,
                public_url,
                upload_root,
            ),
        })
    }
}

/// 运行应用（加载配置、建状态、迁移、监听、服务）
pub async fn run() -> Result<()> {
    // 加载环境变量
    dotenvy::dotenv().ok();

    // 加载配置
    let config = if std::path::Path::new("config.yaml").exists() {
        AppConfig::from_file("config.yaml")?
    } else {
        AppConfig::default()
    };

    // 初始化日志
    tracing_subscriber::registry()
        .with(
            tracing_subscriber::EnvFilter::try_from_default_env().unwrap_or_else(|_| {
                format!("{}=debug,tower_http=debug", env!("CARGO_CRATE_NAME")).into()
            }),
        )
        .with(tracing_subscriber::fmt::layer())
        .init();

    tracing::info!("🚀 {} 服务启动中...", config.app.name);
    tracing::info!("📡 监听地址: {}", config.listen_addr());

    // 构建应用状态
    let state = AppState::from_config(config).await?;

    // 运行数据库迁移
    tracing::info!("🔄 运行数据库迁移...");
    if let Err(e) = sqlx::migrate!("./src/db/migrations")
        .run(&state.db)
        .await
    {
        tracing::warn!("数据库迁移警告（可忽略）: {}", e);
    }

    // 构建路由（create_router 取得 state 的所有权，故监听地址需提前取出）
    let listen_addr = state.config.listen_addr();
    let app = router::create_router(state);

    // 启动服务
    let listener = TcpListener::bind(&listen_addr).await?;
    tracing::info!("✅ 服务已启动: http://{}", listen_addr);

    axum::serve(listener, app)
        .with_graceful_shutdown(shutdown_signal())
        .await?;

    Ok(())
}

/// 优雅关闭信号处理
async fn shutdown_signal() {
    let ctrl_c = async {
        signal::ctrl_c()
            .await
            .expect("Failed to install Ctrl+C handler");
    };

    #[cfg(unix)]
    let terminate = async {
        signal::unix::signal(signal::unix::SignalKind::terminate())
            .expect("Failed to install signal handler")
            .recv()
            .await;
    };

    #[cfg(not(unix))]
    let terminate = std::future::pending::<()>();

    tokio::select! {
        _ = ctrl_c => tracing::info!("收到 Ctrl+C 信号，正在关闭..."),
        _ = terminate => tracing::info!("收到 SIGTERM 信号，正在关闭..."),
    }
}
