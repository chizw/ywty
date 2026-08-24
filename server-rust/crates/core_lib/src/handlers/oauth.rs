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
///
/// `mode=bind`（需携带登录 Cookie）：回调后绑定到当前用户而非登录/注册。
#[utoipa::path(
    get,
    path = "/api/v1/oauth/:provider/authorize",
    params(
        ("provider" = String, Path, description = "OAuth 提供商 (github/google)"),
        ("mode" = String, Query, description = "login(默认) 或 bind"),
    ),
    responses(
        (status = 200, description = "成功", body = serde_json::Value),
    ),
    tag = "OAuth"
)]
pub async fn authorize(
    State(state): State<AppState>,
    Path(provider): Path<String>,
    axum::extract::Query(params): axum::extract::Query<std::collections::HashMap<String, String>>,
    headers: axum::http::HeaderMap,
    jar: CookieJar,
) -> AppResult<(CookieJar, Json<ApiResponse<serde_json::Value>>)> {
    // bind 模式必须已登录（前端 fetch 会携带 Bearer token）
    let mode = match params.get("mode").map(|s| s.as_str()) {
        Some("bind") => {
            let token = headers
                .get(axum::http::header::AUTHORIZATION)
                .and_then(|v| v.to_str().ok())
                .and_then(|v| v.strip_prefix("Bearer "))
                .ok_or_else(|| AppError::Auth("绑定第三方账号前请先登录".to_string()))?;
            let claims = state.jwt.verify_token(token)?;
            format!("bind:{}", claims.sub)
        }
        _ => "login".to_string(),
    };

    let (url, oauth_state) = state.oauth_svc.authorize_url(&provider).await?;

    // 存储 state 用于回调验证
    let jar = if let Some(redis) = &state.redis {
        // Redis 模式：state → mode，TTL 600s
        let mut redis = redis.clone();
        let key = format!("oauth:state:{}", oauth_state);
        match redis.set_ex(&key, &mode, 600).await {
            Ok(_) => {
                tracing::debug!("OAuth state 已存入 Redis: {}", oauth_state);
                jar
            }
            Err(e) => {
                tracing::warn!("Redis 存储 OAuth state 失败，降级签名 Cookie: {}", e);
                set_state_cookie(jar, &oauth_state, &mode, &state)
            }
        }
    } else {
        // 无 Redis：签名 Cookie 降级
        set_state_cookie(jar, &oauth_state, &mode, &state)
    };

    Ok((
        jar,
        Json(ApiResponse::success(serde_json::json!({ "data": {
            "url": url,
            "state": oauth_state,
        } }))),
    ))
}

/// 列出已配置的 OAuth 提供商（公开，供登录页/绑定页渲染入口）
#[utoipa::path(
    get,
    path = "/api/v1/oauth/providers",
    responses(
        (status = 200, description = "成功", body = serde_json::Value),
    ),
    tag = "OAuth"
)]
pub async fn providers(
    State(state): State<AppState>,
) -> AppResult<Json<ApiResponse<serde_json::Value>>> {
    let items = state
        .oauth_svc
        .configured_providers()
        .await
        .into_iter()
        .map(|(provider, name)| serde_json::json!({ "provider": provider, "name": name }))
        .collect::<Vec<_>>();
    Ok(Json(ApiResponse::success(
        serde_json::json!({ "providers": items }),
    )))
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
    headers: axum::http::HeaderMap,
    jar: CookieJar,
) -> AppResult<axum::response::Response> {
    let code = params.get("code").cloned().unwrap_or_default();
    let callback_state = params.get("state").cloned().unwrap_or_default();

    if code.is_empty() {
        return Err(AppError::Validation("缺少 code 参数".to_string()));
    }
    if callback_state.is_empty() {
        return Err(AppError::Validation("缺少 state 参数".to_string()));
    }

    // 验证 state（防 CSRF），同时取出发起时的模式（login / bind:<uid>）
    let stored_mode = if let Some(redis) = &state.redis {
        let mut redis = redis.clone();
        let key = format!("oauth:state:{}", callback_state);
        match redis.get(&key).await {
            Ok(v) => {
                // 一次性消费
                let _ = redis.del(&key).await;
                v
            }
            Err(e) => {
                tracing::warn!("Redis 验证 OAuth state 失败: {}", e);
                read_state_cookie(&jar, &callback_state, &state)
            }
        }
    } else {
        read_state_cookie(&jar, &callback_state, &state)
    };

    let Some(mode) = stored_mode else {
        return Err(AppError::Auth(
            "OAuth state 验证失败（可能已过期或被篡改）".to_string(),
        ));
    };

    // 交换 code 获取用户信息
    let user_info: OAuthUserInfo = state.oauth_svc.login_or_register(&provider, &code).await?;

    // 清除 state cookie
    let jar = jar.remove(cookie::Cookie::new("oauth_state", ""));

    let accept_html = headers
        .get(axum::http::header::ACCEPT)
        .and_then(|v| v.to_str().ok())
        .map(|a| a.contains("text/html"))
        .unwrap_or(false);

    // ---------- 绑定模式：挂到已登录用户 ----------
    if let Some(uid) = mode.strip_prefix("bind:") {
        let user_id: i64 = uid
            .parse()
            .map_err(|_| AppError::Auth("绑定状态无效".to_string()))?;
        state
            .oauth_svc
            .bind(user_id, &provider, &user_info.provider_user_id)
            .await?;
        return oauth_finish(
            &state,
            jar,
            serde_json::json!({ "message": "绑定成功", "provider": provider }),
            Some(format!("/dashboard/oauth?bound={}", provider)),
            accept_html,
        );
    }

    // ---------- 登录模式：查找或创建用户 ----------
    let (user_id, username, email) = match state
        .oauth_svc
        .find_by_open_id(&provider, &user_info.provider_user_id)
        .await?
    {
        Some((id, name, mail)) => (id, name, mail),
        None => state.oauth_svc.create_oauth_user(&user_info).await?,
    };

    let role: String = sqlx::query_scalar("SELECT role FROM users WHERE id = ?")
        .bind(user_id)
        .fetch_one(&state.db)
        .await
        .unwrap_or_else(|_| "user".to_string());

    let token_pair = state.jwt.generate_token_pair(user_id, &username, &role)?;

    let payload = serde_json::json!({
        "user": {
            "id": user_id,
            "username": username,
            "email": email,
            "role": role,
            "avatar": user_info.avatar,
            "provider": provider,
        },
        "token": token_pair,
    });

    oauth_finish(&state, jar, payload, None, accept_html)
}

/// 统一收尾：浏览器 HTML 导航 302 到前端回调页（hash 携带数据）；程序化请求返回 JSON。
fn oauth_finish(
    state: &AppState,
    jar: CookieJar,
    payload: serde_json::Value,
    redirect_path: Option<String>,
    accept_html: bool,
) -> AppResult<axum::response::Response> {
    use axum::response::IntoResponse;

    if accept_html {
        let base = state
            .config
            .app
            .base_url
            .clone()
            .unwrap_or_else(|| format!("http://localhost:{}", state.config.app.port))
            .trim_end_matches('/')
            .to_string();
        let target = match redirect_path {
            Some(p) => format!("{}{}", base, p),
            None => {
                let token = payload
                    .get("token")
                    .cloned()
                    .unwrap_or(serde_json::json!({}));
                let access = token
                    .get("access_token")
                    .and_then(|v| v.as_str())
                    .unwrap_or("");
                let refresh = token
                    .get("refresh_token")
                    .and_then(|v| v.as_str())
                    .unwrap_or("");
                let user = payload
                    .get("user")
                    .cloned()
                    .unwrap_or(serde_json::json!({}));
                format!(
                    "{}/auth/callback#access_token={}&refresh_token={}&user={}",
                    base,
                    access,
                    refresh,
                    urlencoding::encode(&user.to_string())
                )
            }
        };
        let redirect = axum::response::Redirect::to(&target);
        return Ok((jar, redirect).into_response());
    }

    Ok((jar, Json(ApiResponse::success(payload))).into_response())
}

/// 设置签名 state Cookie（Redis 不可用时的降级方案）
///
/// cookie 值格式：`{signature}.{state}.{mode}`
fn set_state_cookie(jar: CookieJar, state: &str, mode: &str, app_state: &AppState) -> CookieJar {
    // 使用 JWT secret 签名 state
    let secret = &app_state.config.auth.jwt.secret;
    let payload = format!("{}.{}", state, mode);
    let signature = hmac_sha256(secret.as_bytes(), payload.as_bytes());
    let cookie_value = format!("{}.{}", signature, payload);

    let cookie = cookie::Cookie::build(("oauth_state", cookie_value))
        .path("/")
        .max_age(time::Duration::minutes(10))
        .http_only(true)
        .same_site(cookie::SameSite::Lax)
        .build();

    jar.add(cookie)
}

/// 验证签名 state Cookie，返回其中的模式（login / bind:<uid>）
fn read_state_cookie(
    jar: &CookieJar,
    expected_state: &str,
    app_state: &AppState,
) -> Option<String> {
    let cookie_value = jar.get("oauth_state").map(|c| c.value().to_string())?;

    // 分离签名与载荷
    let parts: Vec<&str> = cookie_value.rsplitn(2, '.').collect();
    if parts.len() != 2 {
        return None;
    }
    let (signature, payload) = (parts[0], parts[1]);

    // 验证签名
    let secret = &app_state.config.auth.jwt.secret;
    let expected_sig = hmac_sha256(secret.as_bytes(), payload.as_bytes());
    if signature != expected_sig {
        return None;
    }

    // 载荷 = state.mode（state 本身不含点）
    let mut iter = payload.splitn(2, '.');
    let state_part = iter.next()?;
    let mode_part = iter.next().unwrap_or("login");
    if state_part != expected_state {
        return None;
    }
    Some(mode_part.to_string())
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
