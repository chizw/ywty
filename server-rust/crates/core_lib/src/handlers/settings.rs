//! 全局设置管理处理器（管理员）

use axum::extract::State;
use axum::Json;
use std::collections::BTreeMap;

use crate::error::{AppError, AppResult};
use crate::handlers::AdminUser;
use crate::services::settings;
use crate::utils::response::ApiResponse;
use crate::AppState;

/// 获取全部设置（敏感项脱敏为"是否已设置"布尔值）
#[utoipa::path(
    get,
    path = "/api/v1/admin/settings",
    responses(
        (status = 200, description = "成功", body = serde_json::Value),
    ),
    tag = "管理后台"
)]
pub async fn get_settings(
    State(state): State<AppState>,
    _admin: AdminUser,
) -> AppResult<Json<ApiResponse<serde_json::Value>>> {
    let rows = settings::get_all(&state.db).await?;

    let mut map = serde_json::Map::new();
    for (key, value) in rows {
        if settings::keys::SENSITIVE_KEYS.contains(&key.as_str()) {
            // 敏感项不回传明文，仅返回是否已设置
            map.insert(key, serde_json::Value::Bool(!value.is_empty()));
        } else {
            map.insert(key, serde_json::Value::String(value));
        }
    }

    Ok(Json(ApiResponse::success(serde_json::Value::Object(map))))
}

/// 批量保存设置（body 为 `{key: value}`，仅接受白名单内的键）
#[utoipa::path(
    put,
    path = "/api/v1/admin/settings",
    request_body = serde_json::Value,
    responses(
        (status = 200, description = "成功", body = serde_json::Value),
    ),
    tag = "管理后台"
)]
pub async fn update_settings(
    State(state): State<AppState>,
    _admin: AdminUser,
    Json(body): Json<BTreeMap<String, serde_json::Value>>,
) -> AppResult<Json<ApiResponse<serde_json::Value>>> {
    let mut saved = 0usize;
    for (key, value) in body {
        // 仅允许写入白名单键
        if !settings::keys::ALLOWED_KEYS.contains(&key.as_str()) {
            return Err(AppError::Validation(format!("不支持的设置项: {}", key)));
        }
        let text = match value {
            serde_json::Value::String(s) => s,
            serde_json::Value::Bool(b) => b.to_string(),
            serde_json::Value::Number(n) => n.to_string(),
            serde_json::Value::Null => String::new(),
            other => {
                return Err(AppError::Validation(format!(
                    "设置项 {} 的值类型无效: {}",
                    key, other
                )))
            }
        };
        settings::set(&state.db, &key, &text).await?;
        saved += 1;
    }

    Ok(Json(ApiResponse::success(serde_json::json!({
        "saved": saved,
    }))))
}
