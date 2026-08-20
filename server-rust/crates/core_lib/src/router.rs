//! 路由注册（对应 Go 的 `internal/router/router.go`）

use axum::{
    middleware,
    response::IntoResponse,
    routing::{delete, get, patch, post},
    Router,
};
use tower::ServiceBuilder;
use tower_http::compression::CompressionLayer;

use crate::middleware::{auth_middleware, cors_layer, rate_limit_middleware, request_id_middleware};
use crate::handlers::{album, auth, photo, storage, user};
use crate::AppState;

/// 构建应用路由
pub fn create_router(state: AppState) -> Router {
    // 创建限流状态
    let rate_limit_state = crate::middleware::create_rate_limit_state();
    crate::middleware::start_rate_limit_cleanup(rate_limit_state.clone());
    // 公开路由
    let public_routes = Router::new()
        .route("/healthz", get(health_check))
        .route("/ping", get(ping))
        // 认证公开端点
        .route("/auth/register", post(auth::register))
        .route("/auth/login", post(auth::login))
        .route("/auth/refresh", post(auth::refresh))
        .route("/auth/reset-password", post(auth::reset_password))
        .route("/captcha", get(auth::get_captcha))
        .route("/captcha/verify", post(auth::verify_captcha))
        .route("/verify-codes", post(auth::send_verify_code))
        // 公开探索
        .route("/public/photos", get(photo::list_public))
        // 公开分享页
        .route("/s/:slug", get(share_view))
        // 公开图片重定向
        .route("/i/:id", get(photo_redirect));

    // 需要认证的路由
    let auth_routes = Router::new()
        // 认证域
        .route("/auth/logout", post(auth::logout))
        .route("/auth/me", get(auth::me))
        // 用户域
        .route("/user/profile", get(user::get_profile).patch(user::update_profile))
        .route("/user/change-password", post(user::change_password))
        .route("/user/change-email", post(user::change_email))
        // 图片域
        .route("/photos", get(photo::list).post(photo::upload))
        .route(
            "/photos/:id",
            get(photo::get).patch(photo::update).delete(photo::delete),
        )
        .route("/photos/batch-delete", post(photo::batch_delete))
        .route("/photos/batch-update", patch(photo::batch_update))
        .route("/photos/:id/move-to-album", post(photo::move_to_album))
        .route("/photos/:id/copy", post(photo::copy))
        // 相册域
        .route("/albums", get(album::list).post(album::create))
        .route(
            "/albums/:id",
            get(album::get).patch(album::update).delete(album::delete),
        )
        .route(
            "/albums/:id/photos",
            get(album::list_photos).post(album::add_photos),
        )
        .route(
            "/albums/:id/photos/:photo_id",
            delete(album::remove_photo),
        )
        // 存储域
        .route("/storage/sign", get(storage::sign))
        // 认证中间件
        .layer(middleware::from_fn_with_state(
            state.jwt.as_ref().clone(),
            auth_middleware,
        ));

    // 合并路由
    Router::new()
        .nest("/api/v1", public_routes.merge(auth_routes))
        .layer(
            ServiceBuilder::new()
                .layer(cors_layer())
                .layer(middleware::from_fn(request_id_middleware))
                .layer(CompressionLayer::new())
                .layer(middleware::from_fn_with_state(
                    rate_limit_state.clone(),
                    rate_limit_middleware,
                )),
        )
        .fallback(
            |axum::extract::State(_state): axum::extract::State<AppState>,
             req: axum::extract::Request| async move {
                tracing::warn!("FALLBACK hit: {} {}", req.method(), req.uri());
                (
                    axum::http::StatusCode::NOT_FOUND,
                    axum::Json(serde_json::json!({
                        "code": "NOT_FOUND",
                        "message": "路由不存在"
                    })),
                )
                    .into_response()
            },
        )
        .with_state(state)
}

/// 健康检查
async fn health_check(
    axum::extract::State(state): axum::extract::State<AppState>,
) -> axum::response::Result<axum::Json<serde_json::Value>> {
    use crate::db::ping;

    let db_status = match ping(&state.db).await {
        Ok(_) => "ok",
        Err(_) => "error",
    };

    Ok(axum::Json(serde_json::json!({
        "status": "ok",
        "app": state.config.app.name,
        "env": state.config.app.env,
        "database": db_status,
        "time": chrono::Utc::now().to_rfc3339(),
    })))
}

/// Ping
async fn ping() -> &'static str {
    "pong"
}

/// 分享页 - 通过 slug 查看分享内容
async fn share_view(
    axum::extract::State(state): axum::extract::State<AppState>,
    axum::extract::Path(slug): axum::extract::Path<String>,
) -> (axum::http::StatusCode, axum::Json<serde_json::Value>) {
    // 查询分享记录
    let share: Option<(i64, String, i64, Option<String>, Option<i64>)> = sqlx::query_as(
        "SELECT id, shareable_type, shareable_id, password, expires_at FROM shares WHERE slug = ? AND deleted_at IS NULL"
    )
    .bind(&slug)
    .fetch_optional(&state.db)
    .await
    .unwrap_or(None);

    let (share_id, shareable_type, shareable_id, password, expires_at) = match share {
        Some(s) => s,
        None => {
            return (axum::http::StatusCode::NOT_FOUND, axum::Json(serde_json::json!({
                "code": "NOT_FOUND",
                "message": "分享不存在或已删除"
            })));
        }
    };

    // 检查是否过期
    if let Some(exp) = expires_at {
        let now = chrono::Utc::now().timestamp();
        if exp < now {
            return (axum::http::StatusCode::GONE, axum::Json(serde_json::json!({
                "code": "EXPIRED",
                "message": "分享已过期"
            })));
        }
    }

    // 增加浏览计数
    let _ = sqlx::query("UPDATE shares SET views = views + 1 WHERE id = ?")
        .bind(share_id)
        .execute(&state.db)
        .await;

    // 根据类型返回内容
    match shareable_type.as_str() {
        "photo" => {
            let photo: Option<(String, String, Option<String>, i64, Option<i32>, Option<i32>)> = sqlx::query_as(
                "SELECT uuid, url, thumbnail_url, size, width, height FROM photos WHERE id = ? AND deleted_at IS NULL"
            )
            .bind(shareable_id)
            .fetch_optional(&state.db)
            .await
            .unwrap_or(None);

            match photo {
                Some((uuid, url, thumbnail_url, size, width, height)) => {
                    let requires_password = password.is_some();
                    (axum::http::StatusCode::OK, axum::Json(serde_json::json!({
                        "type": "photo",
                        "share_id": share_id,
                        "slug": slug,
                        "uuid": uuid,
                        "url": url,
                        "thumbnail_url": thumbnail_url,
                        "size": size,
                        "width": width,
                        "height": height,
                        "requires_password": requires_password,
                    })))
                }
                None => (axum::http::StatusCode::NOT_FOUND, axum::Json(serde_json::json!({
                    "code": "NOT_FOUND",
                    "message": "分享的图片不存在"
                }))),
            }
        }
        "album" => {
            let album: Option<(String, String, Option<String>, i64)> = sqlx::query_as(
                "SELECT uuid, name, description, photo_count FROM albums WHERE id = ? AND deleted_at IS NULL"
            )
            .bind(shareable_id)
            .fetch_optional(&state.db)
            .await
            .unwrap_or(None);

            match album {
                Some((uuid, name, description, photo_count)) => {
                    let requires_password = password.is_some();
                    (axum::http::StatusCode::OK, axum::Json(serde_json::json!({
                        "type": "album",
                        "share_id": share_id,
                        "slug": slug,
                        "uuid": uuid,
                        "name": name,
                        "description": description,
                        "photo_count": photo_count,
                        "requires_password": requires_password,
                    })))
                }
                None => (axum::http::StatusCode::NOT_FOUND, axum::Json(serde_json::json!({
                    "code": "NOT_FOUND",
                    "message": "分享的相册不存在"
                }))),
            }
        }
        _ => (axum::http::StatusCode::BAD_REQUEST, axum::Json(serde_json::json!({
            "code": "UNSUPPORTED_TYPE",
            "message": format!("不支持的分享类型: {}", shareable_type)
        }))),
    }
}

/// 图片直链重定向 - 返回图片的实际 URL（可用于 <img> src）
async fn photo_redirect(
    axum::extract::State(state): axum::extract::State<AppState>,
    axum::extract::Path(id): axum::extract::Path<i64>,
) -> axum::response::Response {
    use axum::response::IntoResponse;

    // 查询图片
    let photo: Option<(String, i32)> = sqlx::query_as(
        "SELECT url, status FROM photos WHERE id = ? AND deleted_at IS NULL"
    )
    .bind(id)
    .fetch_optional(&state.db)
    .await
    .unwrap_or(None);

    let (url, status) = match photo {
        Some(p) => p,
        None => {
            return (axum::http::StatusCode::NOT_FOUND, axum::Json(serde_json::json!({
                "code": "NOT_FOUND",
                "message": "图片不存在"
            }))).into_response();
        }
    };

    // 检查状态
    if status != 1 {
        return (axum::http::StatusCode::FORBIDDEN, axum::Json(serde_json::json!({
            "code": "UNAVAILABLE",
            "message": "图片暂不可用"
        }))).into_response();
    }

    // 增加浏览计数（异步，不阻塞重定向）
    let state_clone = state.db.clone();
    let id_clone = id;
    tokio::spawn(async move {
        let _ = sqlx::query("UPDATE photos SET views = views + 1 WHERE id = ?")
            .bind(id_clone)
            .execute(&state_clone)
            .await;
    });

    // 302 重定向到实际 URL
    axum::response::Redirect::temporary(&url).into_response()
}

#[cfg(test)]
mod tests {
    use super::*;
    use axum::extract::Path;
    use axum::routing::get;
    use tower::ServiceExt;

    async fn echo_id(Path(id): Path<String>) -> String {
        format!("PARAM-{}", id)
    }

    async fn echo_slug(Path(slug): Path<String>) -> String {
        format!("SLUG-{}", slug)
    }

    /// axum 使用 matchit 原生的 `:id` 语法（绕过 axum 的 `{id}` 解析器）
    #[tokio::test]
    async fn axum_with_native_colon_syntax() {
        let app = Router::new()
            .route("/lit", get(|| async { "LITERAL-OK" }))
            .route("/users/:id", get(echo_id));

        let resp = app
            .clone()
            .oneshot(
                axum::http::Request::builder()
                    .uri("/lit")
                    .body(axum::body::Body::empty())
                    .unwrap(),
            )
            .await
            .unwrap();
        assert_eq!(resp.status(), 200);

        let resp = app
            .oneshot(
                axum::http::Request::builder()
                    .uri("/users/42")
                    .body(axum::body::Body::empty())
                    .unwrap(),
            )
            .await
            .unwrap();
        assert_eq!(
            resp.status(),
            200,
            "axum with native :id syntax should match /users/42"
        );
        let body = axum::body::to_bytes(resp.into_body(), usize::MAX).await.unwrap();
        assert_eq!(&body[..], b"PARAM-42");
    }

    /// 直接测试 matchit 参数路由匹配（无网络、无服务器）
    #[tokio::test]
    async fn param_route_matches_at_top_level() {
        let app = Router::new()
            .route("/lit", get(|| async { "LITERAL-OK" }))
            .route("/users/:id", get(echo_id));

        // literal
        let resp = app
            .clone()
            .oneshot(
                axum::http::Request::builder()
                    .uri("/lit")
                    .body(axum::body::Body::empty())
                    .unwrap(),
            )
            .await
            .unwrap();
        assert_eq!(resp.status(), 200);

        // param
        let resp = app
            .oneshot(
                axum::http::Request::builder()
                    .uri("/users/42")
                    .body(axum::body::Body::empty())
                    .unwrap(),
            )
            .await
            .unwrap();
        assert_eq!(resp.status(), 200, "param route /users/42 should match");
        let body = axum::body::to_bytes(resp.into_body(), usize::MAX).await.unwrap();
        assert_eq!(&body[..], b"PARAM-42");
    }

    /// 测试 nest 内的参数路由
    #[tokio::test]
    async fn param_route_matches_under_nest() {
        let app = Router::new().nest(
            "/api/v1",
            Router::new()
                .route("/healthz", get(|| async { "HEALTHZ-OK" }))
                .route("/s/:slug", get(echo_slug)),
        );

        // literal nested
        let resp = app
            .clone()
            .oneshot(
                axum::http::Request::builder()
                    .uri("/api/v1/healthz")
                    .body(axum::body::Body::empty())
                    .unwrap(),
            )
            .await
            .unwrap();
        assert_eq!(resp.status(), 200);

        // param nested
        let resp = app
            .oneshot(
                axum::http::Request::builder()
                    .uri("/api/v1/s/my-slug")
                    .body(axum::body::Body::empty())
                    .unwrap(),
            )
            .await
            .unwrap();
        assert_eq!(
            resp.status(),
            200,
            "nested param route /api/v1/s/my-slug should match"
        );
        let body = axum::body::to_bytes(resp.into_body(), usize::MAX).await.unwrap();
        assert_eq!(&body[..], b"SLUG-my-slug");
    }

    /// 直接驱动 matchit 原生 API（使用 matchit 原生的 `:id` 语法）
    #[test]
    fn matchit_raw_param_match() {
        let mut router = matchit::Router::new();
        router.insert("/lit", "literal").unwrap();
        router.insert("/users/:id", "param").unwrap();

        // literal
        let m = router.at("/lit").unwrap();
        assert_eq!(*m.value, "literal");

        // param — 使用 matchit 原生 `:id` 语法
        let m = router.at("/users/42").unwrap();
        assert_eq!(*m.value, "param");
        assert_eq!(m.params.get("id"), Some("42"));
    }
}
