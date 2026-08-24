//! 路由注册

use axum::{
    middleware,
    response::IntoResponse,
    routing::{delete, get, patch, post},
    Router,
};
use tower::ServiceBuilder;
use tower_http::compression::CompressionLayer;

use crate::handlers::{
    admin, album, auth, capacity, coupon, drivers, feedback, group, like, notice, oauth, order,
    page, photo, plan, settings, share, site, storage, storage_admin, tag, ticket, token, user,
};
use crate::middleware::{
    admin_guard, auth_middleware, cors_layer, rate_limit_middleware, request_id_middleware,
};
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
        // 站点公开信息
        .route("/site/info", get(site::info))
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
        .route("/i/:id", get(photo_redirect))
        // 公开反馈提交
        .route("/feedback", post(feedback::create_feedback))
        // 公开套餐
        .route("/plans", get(plan::list_public))
        .route("/plans/:id", get(plan::get_public))
        // 公开公告
        .route("/notices", get(notice::list_public))
        .route("/notices/:id", get(notice::get_public))
        // 公开页面
        .route("/pages", get(page::list_public))
        .route("/pages/:slug", get(page::get_public))
        // 公开优惠券校验
        .route("/coupons/validate", post(coupon::validate))
        // OAuth 公开端点
        .route("/oauth/providers", get(oauth::providers))
        .route("/oauth/:provider/authorize", get(oauth::authorize))
        .route("/oauth/:provider/callback", get(oauth::callback))
        // 支付回调（公开）
        .route("/orders/notify", post(order::notify));

    // 需要认证的路由
    let auth_routes = Router::new()
        // 认证域
        .route("/auth/logout", post(auth::logout))
        .route("/auth/me", get(auth::me))
        // 用户域
        .route(
            "/user/profile",
            get(user::get_profile).patch(user::update_profile),
        )
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
        .route("/albums/:id/photos/:photo_id", delete(album::remove_photo))
        // 存储域
        .route("/storage/sign", get(storage::sign))
        // 分享域
        .route("/shares", get(share::list).post(share::create))
        .route("/shares/:id", patch(share::update).delete(share::delete))
        // 标签域
        .route("/tags", get(tag::list).post(tag::create))
        .route("/tags/:id", delete(tag::delete))
        .route("/tags/attach", post(tag::attach))
        .route("/tags/detach", post(tag::detach))
        // 点赞域
        .route("/likes", post(like::toggle).get(like::status))
        // 举报域
        .route("/reports", post(like::create_report))
        .route("/admin/reports", get(like::admin_list_reports))
        .route(
            "/admin/reports/:id",
            patch(like::admin_update_report_status),
        )
        // 违规记录域
        .route("/violations", post(feedback::create_violation))
        .route("/admin/violations", get(feedback::admin_list_violations))
        .route(
            "/admin/violations/:id",
            patch(feedback::admin_update_violation_status),
        )
        // 反馈管理域
        .route("/admin/feedbacks", get(feedback::admin_list_feedbacks))
        .route(
            "/admin/feedbacks/:id",
            delete(feedback::admin_delete_feedback),
        )
        // 订单域
        .route("/orders", get(order::list).post(order::create))
        .route("/orders/:id", get(order::get))
        .route("/orders/:id/pay", post(order::pay))
        .route("/orders/:id/cancel", post(order::cancel))
        // 工单域
        .route("/tickets", get(ticket::list).post(ticket::create))
        .route("/tickets/:id", get(ticket::get))
        .route(
            "/tickets/:id/replies",
            post(ticket::reply).get(ticket::list_replies),
        )
        .route("/tickets/:id/close", post(ticket::close))
        // OAuth 需鉴权
        .route("/oauth", get(oauth::list).post(oauth::bind))
        .route("/oauth/:id", delete(oauth::unbind))
        .route("/oauth/find", post(oauth::find_by_openid))
        // API Token
        .route("/tokens", get(token::list).post(token::create))
        .route("/tokens/:id", delete(token::revoke))
        // 容量
        .route("/capacity", get(capacity::get))
        // 管理后台
        .route("/admin/stats", get(admin::stats))
        .route(
            "/admin/settings",
            get(settings::get_settings).put(settings::update_settings),
        )
        .route("/admin/users", get(admin::list_users))
        .route(
            "/admin/users/:id",
            patch(admin::update_user).delete(admin::delete_user),
        )
        .route("/admin/photos", get(admin::list_all_photos))
        .route("/admin/photos/:id", delete(admin::delete_photo))
        .route("/admin/groups", get(group::list).post(group::create))
        .route(
            "/admin/groups/:id",
            get(group::get).put(group::update).delete(group::delete),
        )
        // 管理端：套餐
        .route(
            "/admin/plans",
            get(plan::admin_list).post(plan::admin_create),
        )
        .route(
            "/admin/plans/:id",
            get(plan::admin_get)
                .patch(plan::admin_update)
                .delete(plan::admin_delete),
        )
        .route("/admin/plans/:id/toggle", post(plan::admin_toggle_up))
        // 管理端：优惠券
        .route(
            "/admin/coupons",
            get(coupon::admin_list).post(coupon::admin_create),
        )
        .route(
            "/admin/coupons/:id",
            get(coupon::admin_get)
                .patch(coupon::admin_update)
                .delete(coupon::admin_delete),
        )
        // 管理端：公告
        .route(
            "/admin/notices",
            get(notice::admin_list).post(notice::admin_create),
        )
        .route(
            "/admin/notices/:id",
            patch(notice::admin_update).delete(notice::admin_delete),
        )
        // 管理端：页面
        .route(
            "/admin/pages",
            get(page::admin_list).post(page::admin_create),
        )
        .route(
            "/admin/pages/:id",
            get(page::admin_get)
                .patch(page::admin_update)
                .delete(page::admin_delete),
        )
        // 管理端：工单
        .route("/admin/tickets", get(ticket::admin_list))
        .route("/admin/tickets/:id", get(ticket::admin_get))
        .route(
            "/admin/tickets/:id/replies",
            post(ticket::admin_reply).get(ticket::admin_list_replies),
        )
        .route(
            "/admin/tickets/:id/status",
            patch(ticket::admin_update_status),
        )
        // 管理端：驱动
        .route("/admin/drivers", get(drivers::list))
        // 管理端：存储
        .route("/admin/storage/drivers", get(storage_admin::list_drivers))
        .route("/admin/storages", get(storage_admin::list_storages))
        .route(
            "/admin/storages/create",
            post(storage_admin::create_storage),
        )
        .route(
            "/admin/storages/update/:id",
            patch(storage_admin::update_storage),
        )
        .route(
            "/admin/storages/delete/:id",
            delete(storage_admin::delete_storage),
        )
        .route("/admin/storage/copy", post(storage_admin::copy))
        // RBAC 管理（占位）
        .route(
            "/admin/rbac/policies",
            get(admin::list_rbac_policies).post(admin::add_rbac_policy),
        )
        .route(
            "/admin/rbac/policies/delete",
            post(admin::delete_rbac_policy),
        )
        .route(
            "/admin/rbac/roles",
            get(admin::list_rbac_roles).post(admin::assign_rbac_role),
        )
        // 中间件顺序（外层先执行）：auth_middleware 先注入 Claims，admin_guard 再校验角色
        .layer(middleware::from_fn_with_state(state.clone(), admin_guard))
        .layer(middleware::from_fn_with_state(
            state.clone(),
            auth_middleware,
        ));

    // Swagger UI 和 OpenAPI JSON
    let swagger = crate::openapi::swagger_routes();

    // 合并路由
    Router::new()
        .nest("/api/v1", public_routes.merge(auth_routes))
        .merge(swagger)
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
                        "code": 40400,
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
) -> axum::Json<serde_json::Value> {
    use crate::db::ping;

    let db_status = match ping(&state.db).await {
        Ok(_) => "ok",
        Err(_) => "error",
    };

    axum::Json(serde_json::json!({
        "status": "ok",
        "app": state.config.app.name,
        "env": state.config.app.env,
        "database": db_status,
        "time": crate::db::now_str(),
    }))
}

/// Ping
async fn ping() -> &'static str {
    "pong"
}

/// 分享页 - 通过 slug 查看分享内容（密码保护在服务端强制校验）
async fn share_view(
    axum::extract::State(state): axum::extract::State<AppState>,
    axum::extract::Path(slug): axum::extract::Path<String>,
    axum::extract::Query(params): axum::extract::Query<std::collections::HashMap<String, String>>,
) -> (axum::http::StatusCode, axum::Json<serde_json::Value>) {
    // 查询分享记录
    type ShareRow = (i64, String, i64, Option<String>, Option<i64>);
    let share: Option<ShareRow> = sqlx::query_as(
        "SELECT id, shareable_type, shareable_id, password, expires_at FROM shares WHERE slug = ? AND deleted_at IS NULL"
    )
    .bind(&slug)
    .fetch_optional(&state.db)
    .await
    .unwrap_or(None);

    let (share_id, shareable_type, shareable_id, password, expires_at) = match share {
        Some(s) => s,
        None => {
            return (
                axum::http::StatusCode::NOT_FOUND,
                axum::Json(serde_json::json!({
                    "code": 40400,
                    "message": "分享不存在或已删除"
                })),
            );
        }
    };

    // 检查是否过期
    if let Some(exp) = expires_at {
        let now = chrono::Utc::now().timestamp();
        if exp < now {
            return (
                axum::http::StatusCode::GONE,
                axum::Json(serde_json::json!({
                    "code": 40400,
                    "message": "分享已过期"
                })),
            );
        }
    }

    // 校验访问密码（服务端强制，密码错误不返回任何内容）
    if let Some(hash) = &password {
        let provided = params.get("password").cloned().unwrap_or_default();
        let ok = crate::auth::password::verify_password(&provided, hash).unwrap_or(false);
        if !ok {
            return (
                axum::http::StatusCode::FORBIDDEN,
                axum::Json(serde_json::json!({
                    "code": 40300,
                    "message": "访问密码错误"
                })),
            );
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
            type PhotoRow = (
                String,
                String,
                Option<String>,
                i64,
                Option<i32>,
                Option<i32>,
            );
            let photo: Option<PhotoRow> = sqlx::query_as(
                "SELECT uuid, url, thumbnail_url, size, width, height FROM photos WHERE id = ? AND deleted_at IS NULL"
            )
            .bind(shareable_id)
            .fetch_optional(&state.db)
            .await
            .unwrap_or(None);

            match photo {
                Some((uuid, url, thumbnail_url, size, width, height)) => {
                    let requires_password = password.is_some();
                    (
                        axum::http::StatusCode::OK,
                        axum::Json(serde_json::json!({
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
                        })),
                    )
                }
                None => (
                    axum::http::StatusCode::NOT_FOUND,
                    axum::Json(serde_json::json!({
                        "code": 40400,
                        "message": "分享的图片不存在"
                    })),
                ),
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
                    (
                        axum::http::StatusCode::OK,
                        axum::Json(serde_json::json!({
                            "type": "album",
                            "share_id": share_id,
                            "slug": slug,
                            "uuid": uuid,
                            "name": name,
                            "description": description,
                            "photo_count": photo_count,
                            "requires_password": requires_password,
                        })),
                    )
                }
                None => (
                    axum::http::StatusCode::NOT_FOUND,
                    axum::Json(serde_json::json!({
                        "code": 40400,
                        "message": "分享的相册不存在"
                    })),
                ),
            }
        }
        _ => (
            axum::http::StatusCode::BAD_REQUEST,
            axum::Json(serde_json::json!({
                "code": 40000,
                "message": format!("不支持的分享类型: {}", shareable_type)
            })),
        ),
    }
}

/// 图片直链重定向 - 返回图片的实际 URL（可用于 <img> src）
async fn photo_redirect(
    axum::extract::State(state): axum::extract::State<AppState>,
    axum::extract::Path(id): axum::extract::Path<i64>,
) -> axum::response::Response {
    use axum::response::IntoResponse;

    // 仅公开图片可直链访问（is_public = 1）
    let photo: Option<(String, i32)> = sqlx::query_as(
        "SELECT url, status FROM photos WHERE id = ? AND is_public = 1 AND deleted_at IS NULL",
    )
    .bind(id)
    .fetch_optional(&state.db)
    .await
    .unwrap_or(None);

    let (url, status) = match photo {
        Some(p) => p,
        None => {
            return (
                axum::http::StatusCode::NOT_FOUND,
                axum::Json(serde_json::json!({
                    "code": 40400,
                    "message": "图片不存在"
                })),
            )
                .into_response();
        }
    };

    // 检查状态
    if status != 1 {
        return (
            axum::http::StatusCode::FORBIDDEN,
            axum::Json(serde_json::json!({
                "code": 40300,
                "message": "图片暂不可用"
            })),
        )
            .into_response();
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
        let body = axum::body::to_bytes(resp.into_body(), usize::MAX)
            .await
            .unwrap();
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
        let body = axum::body::to_bytes(resp.into_body(), usize::MAX)
            .await
            .unwrap();
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
        let body = axum::body::to_bytes(resp.into_body(), usize::MAX)
            .await
            .unwrap();
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
