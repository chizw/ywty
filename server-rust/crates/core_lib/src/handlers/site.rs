//! 站点公开信息处理器（免登录）

use axum::extract::State;
use axum::Json;
use serde::Serialize;
use utoipa::ToSchema;

use crate::error::AppResult;
use crate::services::settings;
use crate::utils::response::ApiResponse;
use crate::AppState;

/// 站点公开信息
#[derive(Debug, Serialize, ToSchema)]
pub struct SiteInfo {
    pub name: String,
    pub description: String,
    pub keywords: String,
    pub footer: String,
    pub icp: String,
    pub allow_register: bool,
    pub require_email_verify: bool,
}

/// 公开：站点信息（供前端页头/页脚与 SEO 使用）
#[utoipa::path(
    get,
    path = "/api/v1/site/info",
    responses(
        (status = 200, description = "成功", body = SiteInfo),
    ),
    tag = "站点"
)]
pub async fn info(State(state): State<AppState>) -> AppResult<Json<ApiResponse<SiteInfo>>> {
    let db = &state.db;

    let name = settings::get(db, settings::keys::SITE_NAME)
        .await?
        .filter(|v| !v.trim().is_empty())
        .unwrap_or_else(|| "云雾图驿".to_string());
    let description = settings::get(db, settings::keys::SITE_DESCRIPTION)
        .await?
        .filter(|v| !v.trim().is_empty())
        .unwrap_or_else(|| "自托管图床 / 云相册".to_string());
    let keywords = settings::get(db, settings::keys::SITE_KEYWORDS)
        .await?
        .unwrap_or_default();
    let footer = settings::get(db, settings::keys::SITE_FOOTER)
        .await?
        .unwrap_or_default();
    let icp = settings::get(db, settings::keys::SITE_ICP)
        .await?
        .unwrap_or_default();
    let allow_register =
        settings::get_bool(db, settings::keys::SECURITY_ALLOW_REGISTER, true).await?;
    let require_email_verify =
        settings::get_bool(db, settings::keys::SECURITY_REQUIRE_EMAIL_VERIFY, true).await?;

    Ok(Json(ApiResponse::success(SiteInfo {
        name,
        description,
        keywords,
        footer,
        icp,
        allow_register,
        require_email_verify,
    })))
}
