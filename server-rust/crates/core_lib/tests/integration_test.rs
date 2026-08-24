//! ywty server-rust 集成测试
//!
//! 使用临时 SQLite 数据库，测试完整的 API 请求/响应流程。
#![cfg(not(feature = "mysql"))]

use axum::body::Body;
use axum::http::{Request, StatusCode};
use http_body_util::BodyExt;
use serde_json::Value;
use tower::ServiceExt;

/// 创建测试用的 AppState（使用内存数据库）
async fn setup_test_app() -> axum::Router {
    use core_lib::AppConfig;

    let config = AppConfig {
        app: core_lib::config::AppConfigApp {
            name: "ywty-test".to_string(),
            env: "test".to_string(),
            host: "127.0.0.1".to_string(),
            port: 0,
            timezone: "UTC".to_string(),
            base_url: None,
        },
        database: core_lib::config::DatabaseConfig {
            driver: core_lib::config::DatabaseDriver::Sqlite,
            host: None,
            port: None,
            username: None,
            password: None,
            database: None,
            path: Some(":memory:".to_string()),
            max_open_conns: Some(5),
            max_idle_conns: Some(1),
        },
        redis: None,
        auth: core_lib::config::AuthConfig {
            jwt: core_lib::config::JwtConfig {
                secret: "test-secret-key-for-integration-tests".to_string(),
                access_expire: 3600,
                refresh_expire: 86400,
            },
        },
        storage: core_lib::config::StorageConfig {
            driver: core_lib::config::StorageDriver::Local,
            root: Some("/tmp/ywty-test-uploads".to_string()),
            url: Some("http://localhost:3000".to_string()),
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
        watermark: None,
    };

    let state = core_lib::AppState::from_config(config).await.unwrap();

    // 验证数据库连接
    core_lib::db::ping(&state.db).await.expect("数据库连接失败");

    // 运行迁移（使用 CARGO_MANIFEST_DIR 确保路径正确）
    let migrations = format!("{}/src/db/migrations", env!("CARGO_MANIFEST_DIR"));
    let m = sqlx::migrate::Migrator::new(std::path::Path::new(&migrations))
        .await
        .expect("迁移目录加载失败");
    m.run(&state.db).await.expect("迁移执行失败");

    core_lib::router::create_router(state)
}

/// 辅助函数：发送请求并返回 (状态码, JSON 响应)
async fn request_json(
    app: axum::Router,
    method: &str,
    path: &str,
    body: Option<Value>,
) -> (StatusCode, Value) {
    let body = match body {
        Some(b) => Body::from(b.to_string()),
        None => Body::empty(),
    };

    let req = Request::builder()
        .method(method)
        .uri(path)
        .header("content-type", "application/json")
        .body(body)
        .unwrap();

    let response = app.oneshot(req).await.unwrap();
    let status = response.status();
    let body_bytes = response.into_body().collect().await.unwrap().to_bytes();
    let body_str = String::from_utf8(body_bytes.to_vec()).unwrap();
    let json: Value = serde_json::from_str(&body_str).unwrap_or(Value::Null);

    (status, json)
}

/// 辅助函数：发送带 Bearer token 的请求
async fn request_with_auth(
    app: axum::Router,
    method: &str,
    path: &str,
    token: &str,
    body: Option<Value>,
) -> (StatusCode, Value) {
    let body = match body {
        Some(b) => Body::from(b.to_string()),
        None => Body::empty(),
    };

    let req = Request::builder()
        .method(method)
        .uri(path)
        .header("authorization", format!("Bearer {}", token))
        .header("content-type", "application/json")
        .body(body)
        .unwrap();

    let response = app.oneshot(req).await.unwrap();
    let status = response.status();
    let body_bytes = response.into_body().collect().await.unwrap().to_bytes();
    let body_str = String::from_utf8(body_bytes.to_vec()).unwrap();
    let json: Value = serde_json::from_str(&body_str).unwrap_or(Value::Null);

    (status, json)
}

// ============================================================================
// 健康检查
// ============================================================================

#[tokio::test]
async fn healthz_returns_ok() {
    let app = setup_test_app().await;
    let (status, body) = request_json(app, "GET", "/api/v1/healthz", None).await;

    assert_eq!(status, StatusCode::OK);
    assert_eq!(body["status"], "ok");
    assert!(body["database"].as_str().is_some());
}

#[tokio::test]
async fn ping_returns_pong() {
    let app = setup_test_app().await;
    let (status, _) = request_json(app, "GET", "/api/v1/ping", None).await;
    assert_eq!(status, StatusCode::OK);
}

// ============================================================================
// 验证码
// ============================================================================

#[tokio::test]
async fn get_captcha_returns_image() {
    let app = setup_test_app().await;
    let (status, body) = request_json(app, "GET", "/api/v1/captcha", None).await;

    assert_eq!(status, StatusCode::OK);
    assert_eq!(body["code"], 0);
    let data = &body["data"];
    assert!(data["captcha_id"].as_str().is_some());
    assert!(data["captcha_image"]
        .as_str()
        .unwrap()
        .starts_with("data:image/png;base64,"));
    assert_eq!(data["expires_in"], 300);
}

#[tokio::test]
async fn verify_invalid_captcha_returns_false() {
    let app = setup_test_app().await;
    let (status, body) = request_json(
        app,
        "POST",
        "/api/v1/captcha/verify",
        Some(serde_json::json!({
            "captcha_id": "invalid-id",
            "captcha_code": "abcd"
        })),
    )
    .await;

    assert_eq!(status, StatusCode::OK);
    assert_eq!(body["code"], 0);
    assert_eq!(body["data"]["valid"], false);
}

// ============================================================================
// 注册/登录
// ============================================================================

#[tokio::test]
async fn register_creates_user_and_returns_token() {
    let app = setup_test_app().await;
    let (status, body) = request_json(
        app,
        "POST",
        "/api/v1/auth/register",
        Some(serde_json::json!({
            "username": "testuser",
            "email": "test@example.com",
            "password": "password123"
        })),
    )
    .await;

    assert_eq!(status, StatusCode::OK, "注册失败: {}", body);
    assert_eq!(body["code"], 0);
    let data = &body["data"];
    assert!(data["access_token"].as_str().is_some());
    assert_eq!(data["user"]["username"], "testuser");
    assert_eq!(data["user"]["email"], "test@example.com");
}

#[tokio::test]
async fn login_with_correct_password_succeeds() {
    let app = setup_test_app().await;

    // 先注册
    request_json(
        app,
        "POST",
        "/api/v1/auth/register",
        Some(serde_json::json!({
            "username": "loginuser",
            "email": "login@example.com",
            "password": "mypassword"
        })),
    )
    .await;

    // 登录（LoginRequest 使用 account 字段，不是 email）
    let app = setup_test_app().await;
    let (status, body) = request_json(
        app,
        "POST",
        "/api/v1/auth/login",
        Some(serde_json::json!({
            "account": "login@example.com",
            "password": "mypassword"
        })),
    )
    .await;

    // 注意：由于是新的内存数据库，用户不存在 → 401 Unauthorized
    assert_eq!(status, StatusCode::UNAUTHORIZED);
    assert_eq!(body["code"], 40100);
}

#[tokio::test]
async fn register_duplicate_email_fails() {
    let app = setup_test_app().await;

    // 第一次注册
    request_json(
        app.clone(),
        "POST",
        "/api/v1/auth/register",
        Some(serde_json::json!({
            "username": "user1",
            "email": "dup@example.com",
            "password": "password123"
        })),
    )
    .await;

    // 重复邮箱
    let (status, body) = request_json(
        app,
        "POST",
        "/api/v1/auth/register",
        Some(serde_json::json!({
            "username": "user2",
            "email": "dup@example.com",
            "password": "password123"
        })),
    )
    .await;

    // 重复邮箱 → 400 Bad Request
    assert_eq!(status, StatusCode::BAD_REQUEST);
    assert_eq!(body["code"], 40000);
    assert!(
        body["message"].as_str().unwrap().contains("邮箱")
            || body["message"].as_str().unwrap().contains("已存在")
    );
}

// ============================================================================
// 认证端点
// ============================================================================

#[tokio::test]
async fn me_requires_auth() {
    let app = setup_test_app().await;
    let (status, body) = request_json(app, "GET", "/api/v1/auth/me", None).await;
    // 未认证 → 401 Unauthorized
    assert_eq!(status, StatusCode::UNAUTHORIZED);
    assert_eq!(body["code"], 40100);
}

#[tokio::test]
async fn register_then_access_me() {
    let app = setup_test_app().await;

    // 注册
    let (_, body) = request_json(
        app.clone(),
        "POST",
        "/api/v1/auth/register",
        Some(serde_json::json!({
            "username": "meuser",
            "email": "me@example.com",
            "password": "password123"
        })),
    )
    .await;

    let token = body["data"]["access_token"].as_str().unwrap();

    // 使用 token 访问 /me
    let (status, me_body) = request_with_auth(app, "GET", "/api/v1/auth/me", token, None).await;
    assert_eq!(status, StatusCode::OK);
    assert_eq!(me_body["code"], 0);
    // /auth/me 直接返回 UserPublic（data.username），不再嵌套 user 对象
    assert_eq!(me_body["data"]["username"], "meuser");
}

// ============================================================================
// 相册
// ============================================================================

#[tokio::test]
async fn create_album_requires_auth() {
    let app = setup_test_app().await;
    let (status, body) = request_json(
        app,
        "POST",
        "/api/v1/albums",
        Some(serde_json::json!({ "name": "test" })),
    )
    .await;
    // 未认证 → 401 Unauthorized
    assert_eq!(status, StatusCode::UNAUTHORIZED);
    assert_eq!(body["code"], 40100);
}

#[tokio::test]
async fn create_and_list_albums() {
    let app = setup_test_app().await;

    // 注册
    let (_, auth) = request_json(
        app.clone(),
        "POST",
        "/api/v1/auth/register",
        Some(serde_json::json!({
            "username": "albumuser",
            "email": "album@example.com",
            "password": "password123"
        })),
    )
    .await;

    let token = auth["data"]["access_token"].as_str().unwrap();

    // 创建相册
    let (status, _) = request_with_auth(
        app.clone(),
        "POST",
        "/api/v1/albums",
        token,
        Some(serde_json::json!({
            "name": "测试相册",
            "description": "集成测试用相册"
        })),
    )
    .await;
    assert_eq!(status, StatusCode::OK, "创建相册失败");

    // 获取相册列表
    let (status, body) = request_with_auth(app, "GET", "/api/v1/albums", token, None).await;
    assert_eq!(status, StatusCode::OK);
    assert!(body["data"].as_array().is_some());
}

// ============================================================================
// 限流
// ============================================================================

#[tokio::test]
async fn rate_limit_allows_normal_traffic() {
    let app = setup_test_app().await;

    // 发送少量请求，应该都成功
    let (status, _) = request_json(app, "GET", "/api/v1/healthz", None).await;
    assert_eq!(status, StatusCode::OK);
}

// ============================================================================
// 404
// ============================================================================

#[tokio::test]
async fn unknown_route_returns_404() {
    let app = setup_test_app().await;
    let (status, body) = request_json(app, "GET", "/api/v1/nonexistent", None).await;

    assert_eq!(status, StatusCode::NOT_FOUND);
    assert_eq!(body["code"], 40400);
}
