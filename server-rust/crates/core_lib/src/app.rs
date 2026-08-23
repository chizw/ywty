//! 应用入口（配置加载、状态构建、迁移、路由、服务启动）
//!
//! 二进制 crate `api` 仅保留一个调用 [`run`] 的 `thin main`，
//! 使最终二进制 crate 自身的代码量保持极小，
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
    pub share_svc: crate::services::share::ShareService,
    pub tag_svc: crate::services::tag::TagService,
    pub like_svc: crate::services::like::LikeService,
    pub report_svc: crate::services::like::ReportService,
    pub feedback_svc: crate::services::feedback::FeedbackService,
    pub violation_svc: crate::services::feedback::ViolationService,
    // P2 服务
    pub admin_svc: crate::services::admin::AdminService,
    pub order_svc: crate::services::order::OrderService,
    pub plan_svc: crate::services::plan::PlanService,
    pub coupon_svc: crate::services::coupon::CouponService,
    pub notice_svc: crate::services::notice::NoticeService,
    pub page_svc: crate::services::page::PageService,
    pub ticket_svc: crate::services::ticket::TicketService,
    pub group_svc: crate::services::group::GroupService,
    pub token_svc: crate::services::token::TokenService,
    pub capacity_svc: crate::services::capacity::CapacityService,
    pub oauth_svc: crate::services::oauth::OAuthService,
    pub storage_admin_svc: crate::services::storage_admin::StorageAdminService,
    // Redis 缓存（可选，未配置时降级为 None）
    pub redis: Option<crate::services::redis_cache::RedisCache>,
    // 图片处理队列
    pub queue: Option<crate::services::queue::ImageQueue>,
    // 多存储驱动
    pub storage_driver: Arc<dyn crate::services::storage_driver::StorageDriver>,
}

impl AppState {
    /// 从配置构建应用状态
    pub async fn from_config(config: AppConfig) -> crate::error::AppResult<Self> {
        let db_pool = db::create_pool(&config.database).await?;
        let jwt = Arc::new(JwtAuth::new(&config.auth.jwt));

        // 存储公开基址：统一由 AppConfig::storage_public_url 推导（唯一推导点）
        let public_url = config.storage_public_url();

        // 构建邮件服务
        let mail_svc = config
            .notify
            .as_ref()
            .and_then(|n| n.mail.as_ref())
            .map(crate::services::mail::MailService::new)
            .unwrap_or_else(crate::services::mail::MailService::disabled);

        // 提取配置（在 config 被 move 之前）
        let redis_cfg = config.redis.clone();
        let oauth_cfg = config.oauth.clone();
        let storage_driver = crate::services::storage_driver::create_driver(&config.storage)?;
        let redis = crate::services::redis_cache::RedisCache::new(
            redis_cfg.as_ref().map(|r| r.addr.as_str()).unwrap_or(""),
            redis_cfg.as_ref().and_then(|r| r.password.as_deref()),
            redis_cfg.as_ref().and_then(|r| r.db),
        )
        .await
        .unwrap_or_else(|e| {
            tracing::warn!("⚠️ Redis 初始化失败，降级为无缓存模式: {}", e);
            None
        });

        // 队列 worker 数量：从配置读取，默认 4
        let worker_count = config
            .queue
            .as_ref()
            .and_then(|q| q.concurrency)
            .unwrap_or(4);

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
            ),
            album_svc: crate::services::album::AlbumService::new(db_pool.clone()),
            storage_svc: crate::services::storage::StorageService::new(public_url),
            share_svc: crate::services::share::ShareService::new(db_pool.clone()),
            tag_svc: crate::services::tag::TagService::new(db_pool.clone()),
            like_svc: crate::services::like::LikeService::new(db_pool.clone()),
            report_svc: crate::services::like::ReportService::new(db_pool.clone()),
            feedback_svc: crate::services::feedback::FeedbackService::new(db_pool.clone()),
            violation_svc: crate::services::feedback::ViolationService::new(db_pool.clone()),
            // P2 服务
            admin_svc: crate::services::admin::AdminService::new(db_pool.clone()),
            order_svc: crate::services::order::OrderService::new(db_pool.clone()),
            plan_svc: crate::services::plan::PlanService::new(db_pool.clone()),
            coupon_svc: crate::services::coupon::CouponService::new(db_pool.clone()),
            notice_svc: crate::services::notice::NoticeService::new(db_pool.clone()),
            page_svc: crate::services::page::PageService::new(db_pool.clone()),
            ticket_svc: crate::services::ticket::TicketService::new(db_pool.clone()),
            group_svc: crate::services::group::GroupService::new(db_pool.clone()),
            token_svc: crate::services::token::TokenService::new(db_pool.clone()),
            capacity_svc: crate::services::capacity::CapacityService::new(db_pool.clone()),
            oauth_svc: crate::services::oauth::OAuthService::new(
                db_pool.clone(),
                oauth_cfg.as_ref().and_then(|o| o.github.clone()),
                oauth_cfg.as_ref().and_then(|o| o.google.clone()),
            ),
            storage_admin_svc: crate::services::storage_admin::StorageAdminService::new(
                db_pool.clone(),
            ),
            redis,
            queue: Some(crate::services::queue::ImageQueue::new(
                db_pool,
                worker_count,
            )),
            storage_driver,
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
    if config.auth.jwt.secret == "please-change-me-in-production" {
        tracing::warn!(
            "⚠️ 正在使用默认 JWT 密钥，令牌可被伪造；请通过 JWT_SECRET 环境变量或 config.yaml 设置强随机密钥"
        );
    }

    // 构建应用状态
    let state = AppState::from_config(config).await?;

    // 运行数据库迁移
    tracing::info!("🔄 运行数据库迁移...");
    if let Err(e) = sqlx::migrate!("./src/db/migrations").run(&state.db).await {
        tracing::warn!("数据库迁移警告（可忽略）: {}", e);
    }

    // 种子默认管理员（仅当用户表为空时执行）
    seed_admin_user(&state.db).await;

    // 构建路由（create_router 取得 state 的所有权，故监听地址需提前取出）
    let listen_addr = state.config.listen_addr();
    let app = router::create_router(state);

    // 启动服务
    let listener = TcpListener::bind(&listen_addr).await?;
    tracing::info!("✅ 服务已启动: http://{}", listen_addr);

    axum::serve(
        listener,
        app.into_make_service_with_connect_info::<std::net::SocketAddr>(),
    )
    .with_graceful_shutdown(shutdown_signal())
    .await?;

    Ok(())
}

/// 种子默认管理员
///
/// 当用户表为空时，自动创建默认 admin 账号。
/// 默认凭据：admin / admin123456（生产环境请尽快修改）
async fn seed_admin_user(db: &sqlx::SqlitePool) {
    let count: i64 = match sqlx::query_scalar("SELECT COUNT(*) FROM users")
        .fetch_one(db)
        .await
    {
        Ok(n) => n,
        Err(_) => return,
    };

    if count > 0 {
        return;
    }

    tracing::info!("🌱 用户表为空，创建默认管理员...");

    let uuid = uuid::Uuid::new_v4().to_string();
    let password_hash = match crate::auth::password::hash_password("admin123456") {
        Ok(h) => h,
        Err(e) => {
            tracing::warn!("⚠️ 密码哈希失败，跳过种子: {}", e);
            return;
        }
    };

    let now = chrono::Utc::now().to_rfc3339();
    let _ = sqlx::query(
        r#"
        INSERT INTO users (uuid, username, email, password, role, is_super_admin, status, created_at, updated_at)
        VALUES (?, 'admin', 'admin@ywty.local', ?, 'admin', 1, 1, ?, ?)
        "#,
    )
    .bind(&uuid)
    .bind(&password_hash)
    .bind(&now)
    .bind(&now)
    .execute(db)
    .await;

    tracing::info!("✅ 默认管理员已创建: admin（默认密码，请尽快修改）");
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
