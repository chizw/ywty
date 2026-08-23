//! 驱动管理处理器

use axum::extract::State;
use axum::Json;

use crate::error::AppResult;
use crate::utils::response::ApiResponse;
use crate::AppState;

/// 列出所有可用驱动
#[utoipa::path(
    get,
    path = "/api/v1/admin/drivers",
    responses(
        (status = 200, description = "成功", body = serde_json::Value),
    ),
    tag = "存储"
)]
pub async fn list(
    State(_state): State<AppState>,
) -> AppResult<Json<ApiResponse<serde_json::Value>>> {
    // 返回静态驱动列表（实际应从各模块收集）
    Ok(Json(ApiResponse::success(serde_json::json!({
        "data": {
            "storage": ["local"],
            "mail": ["smtp"],
            "social": [],
            // 图片处理为占位：仅本地缩略图/水印，custom_http 未实现
            "process": [],
        }
    }))))
}
