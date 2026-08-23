//! OAuth 处理器

use axum::extract::{Path, State};
use axum::Json;
use axum_extra::extract::CookieJar;

use crate::error::{AppError, AppResult};
use crate::handlers::{validate_req, CurrentUser};
use crate::services::oauth::OAuthUserInfo;
use crate::utils::response::ApiResponse;
use crate::AppState;

#[derive(Debug, serde::Deserialize, validator::Validate, utoipa::ToSchema)]
pub struct OAuthBindRequest {
    pub provider: String,
    pub provider_user_id: String,
}

#[derive(Debug, serde::Deserialize, utoipa::ToSchema)]
pub struct OAuthFindRequest {
    pub provider: String,
    pub openid: String,
}

/// 列出 OAuth 绑定
#[utoipa::path(
    get,
    path = "/api/v1/oauth",
    responses(
        (status = 200, description = "成功", body = serde_json::Value),
    ),
    tag = "OAuth"
)]
pub async fn list(
    State(state): State<AppState>,
    CurrentUser { user_id, .. }: CurrentUser,
) -> AppResult<Json<ApiResponse<serde_json::Value>>> {
    let accounts = state.oauth_svc.list(user_id).await?;
    Ok(Json(ApiResponse::success(
        serde_json::json!({ "data": accounts }),
    )))
}

/// 绑定 OAuth 账号
#[utoipa::path(
    post,
    path = "/api/v1/oauth",
    request_body = OAuthBindRequest,
    responses(
        (status = 200, description = "成功", body = serde_json::Value),
    ),
    tag = "OAuth"
)]
pub async fn bind(
    State(state): State<AppState>,
    CurrentUser { user_id, .. }: CurrentUser,
    Json(req): Json<OAuthBindRequest>,
) -> AppResult<Json<ApiResponse<serde_json::Value>>> {
    validate_req(&req)?;
    let account = state
        .oauth_svc
        .bind(user_id, &req.provider, &req.provider_user_id)
        .await?;
    Ok(Json(ApiResponse::success(
        serde_json::json!({ "data": account }),
    )))
}

/// 解绑 OAuth 账号
#[utoipa::path(
    delete,
    path = "/api/v1/oauth/:id",
    params(
        ("id" = i64, Path, description = "OAuth 绑定 ID"),
    ),
    responses(
        (status = 200, description = "成功", body = serde_json::Value),
    ),
    tag = "OAuth"
)]
pub async fn unbind(
    State(state): State<AppState>,
    CurrentUser { user_id, .. }: CurrentUser,
    Path(id): Path<i64>,
) -> AppResult<Json<ApiResponse<serde_json::Value>>> {
    state.oauth_svc.unbind(user_id, id).await?;
    Ok(Json(ApiResponse::success_with_message(
        serde_json::json!({}),
        "解绑成功",
    )))
}

/// 通过 OpenID 查找用户
#[utoipa::path(
    post,
    path = "/api/v1/oauth/find",
    request_body = OAuthFindRequest,
    responses(
        (status = 200, description = "成功", body = serde_json::Value),
    ),
    tag = "OAuth"
)]
pub async fn find_by_openid(
    State(state): State<AppState>,
    CurrentUser { .. }: CurrentUser,
    Json(req): Json<OAuthFindRequest>,
) -> AppResult<Json<ApiResponse<serde_json::Value>>> {
    let result = state
        .oauth_svc
        .find_by_open_id(&req.provider, &req.openid)
        .await?;
    Ok(Json(ApiResponse::success(
        serde_json::json!({ "data": result }),
    )))
}

/// 获取授权 URL（公开）
///
/// 返回 OAuth 提供商的授权 URL 和 CSRF state。
/// 如果 Redis 可用，state 会存入 Redis（600s TTL）；
/// 否则通过签名 Cookie 降级存储。
#[utoipa::path(
    get,
    path = "/api/v1/oauth/:provider/authorize",
    params(
        ("provider" = String, Path, description = "OAuth 提供商 (github/google)"),
    ),
    responses(
        (status = 200, description = "成功", body = serde_json::Value),
    ),
    tag = "OAuth"
)]
pub async fn authorize(
    State(state): State<AppState>,
    Path(provider): Path<String>,
    jar: CookieJar,
) -> AppResult<(CookieJar, Json<ApiResponse<serde_json::Value>>)> {
    let (url, oauth_state) = state.oauth_svc.authorize_url(&provider).await?;

    // 存储 state 用于回调验证
    let jar = if let Some(redis) = &state.redis {
        // Redis 模式：state → Redis，TTL 600s
        let mut redis = redis.clone();
        let key = format!("oauth:state:{}", oauth_state);
        match redis.set_ex(&key, "1", 600).await {
            Ok(_) => {
                tracing::debug!("OAuth state 已存入 Redis: {}", oauth_state);
                jar
            }
            Err(e) => {
                tracing::warn!("Redis 存储 OAuth state 失败，降级签名 Cookie: {}", e);
                set_state_cookie(jar, &oauth_state, &state)
            }
        }
    } else {
        // 无 Redis：签名 Cookie 降级
        set_state_cookie(jar, &oauth_state, &state)
    };

    Ok((
        jar,
        Json(ApiResponse::success(serde_json::json!({ "data": {
            "url": url,
            "state": oauth_state,
        } }))),
    ))
}

/// OAuth 回调（公开）
///
/// 验证 state，交换 code，获取用户信息，查找或创建用户，返回 JWT。
#[utoipa::path(
    get,
    path = "/api/v1/oauth/:provider/callback",
    params(
        ("provider" = String, Path, description = "OAuth 提供商 (github/google)"),
        ("code" = String, Query, description = "授权码"),
        ("state" = String, Query, description = "CSRF state"),
    ),
    responses(
        (status = 200, description = "成功", body = serde_json::Value),
    ),
    tag = "OAuth"
)]
pub async fn callback(
    State(state): State<AppState>,
    Path(provider): Path<String>,
    axum::extract::Query(params): axum::extract::Query<std::collections::HashMap<String, String>>,
    jar: CookieJar,
) -> AppResult<(CookieJar, Json<ApiResponse<serde_json::Value>>)> {
    let code = params.get("code").cloned().unwrap_or_default();
    let callback_state = params.get("state").cloned().unwrap_or_default();

    if code.is_empty() {
        return Err(AppError::Validation("缺少 code 参数".to_string()));
    }
    if callback_state.is_empty() {
        return Err(AppError::Validation("缺少 state 参数".to_string()));
    }

    // 验证 state（防 CSRF）
    let state_valid = if let Some(redis) = &state.redis {
        // Redis 模式：检查并删除 state
        let mut redis = redis.clone();
        let key = format!("oauth:state:{}", callback_state);
        match redis.del(&key).await {
            Ok(deleted) => deleted,
            Err(e) => {
                tracing::warn!("Redis 验证 OAuth state 失败: {}", e);
                // 降级到 Cookie 验证
                verify_state_cookie(&jar, &callback_state, &state)
            }
        }
    } else {
        // 无 Redis：签名 Cookie 验证
        verify_state_cookie(&jar, &callback_state, &state)
    };

    if !state_valid {
        return Err(AppError::Auth(
            "OAuth state 验证失败（可能已过期或被篡改）".to_string(),
        ));
    }

    // 交换 code 获取用户信息
    let user_info: OAuthUserInfo = state.oauth_svc.login_or_register(&provider, &code).await?;

    // 查找是否已绑定
    let (user_id, username, email) = match state
        .oauth_svc
        .find_by_open_id(&provider, &user_info.provider_user_id)
        .await?
    {
        Some((id, name, mail)) => {
            // 已绑定：更新头像（可能变了）
            (id, name, mail)
        }
        None => {
            // 未绑定：自动注册新用户
            state.oauth_svc.create_oauth_user(&user_info).await?
        }
    };

    // 查询用户角色
    let role: String = sqlx::query_scalar("SELECT role FROM users WHERE id = ?")
        .bind(user_id)
        .fetch_one(&state.db)
        .await
        .unwrap_or_else(|_| "user".to_string());

    // 生成 JWT
    let token_pair = state.jwt.generate_token_pair(user_id, &username, &role)?;

    // 清除 state cookie
    let jar = jar.remove(cookie::Cookie::new("oauth_state", ""));

    Ok((
        jar,
        Json(ApiResponse::success(serde_json::json!({ "data": {
            "user": {
                "id": user_id,
                "username": username,
                "email": email,
                "role": role,
                "avatar": user_info.avatar,
                "provider": provider,
            },
            "token": token_pair,
        } }))),
    ))
}

/// 设置签名 state Cookie（Redis 不可用时的降级方案）
fn set_state_cookie(jar: CookieJar, state: &str, app_state: &AppState) -> CookieJar {
    // 使用 JWT secret 签名 state
    let secret = &app_state.config.auth.jwt.secret;
    let signature = hmac_sha256(secret.as_bytes(), state.as_bytes());
    let cookie_value = format!("{}.{}", state, signature);

    let cookie = cookie::Cookie::build(("oauth_state", cookie_value))
        .path("/")
        .max_age(time::Duration::minutes(10))
        .http_only(true)
        .same_site(cookie::SameSite::Lax)
        .build();

    jar.add(cookie)
}

/// 验证签名 state Cookie
fn verify_state_cookie(jar: &CookieJar, expected_state: &str, app_state: &AppState) -> bool {
    let cookie_value = match jar.get("oauth_state").map(|c| c.value().to_string()) {
        Some(v) => v,
        None => return false,
    };

    // 分离 state 和签名
    let parts: Vec<&str> = cookie_value.rsplitn(2, '.').collect();
    if parts.len() != 2 {
        return false;
    }

    let state = parts[1];
    let signature = parts[0];

    // 验证 state 匹配
    if state != expected_state {
        return false;
    }

    // 验证签名
    let secret = &app_state.config.auth.jwt.secret;
    let expected_sig = hmac_sha256(secret.as_bytes(), state.as_bytes());
    signature == expected_sig
}

/// HMAC-SHA256 签名
fn hmac_sha256(key: &[u8], data: &[u8]) -> String {
    use hmac::{Hmac, Mac};
    use sha2::Sha256;

    type HmacSha256 = Hmac<Sha256>;

    let mut mac = HmacSha256::new_from_slice(key).expect("HMAC key length is valid");
    mac.update(data);
    let result = mac.finalize();
    hex::encode(result.into_bytes())
}
