//! 公告模型

use serde::{Deserialize, Serialize};
use sqlx::FromRow;

/// 公告实体
#[derive(Debug, Clone, FromRow, Serialize, utoipa::ToSchema)]
pub struct Notice {
    pub id: i64,
    pub title: String,
    pub content: Option<String>,
    pub view_count: i64,
    pub sort: i64,
    pub created_at: String,
    pub updated_at: String,
    pub deleted_at: Option<String>,
}

/// 创建公告请求
#[derive(Debug, Clone, Deserialize, validator::Validate, utoipa::ToSchema)]
pub struct CreateNoticeRequest {
    pub title: String,
    pub content: Option<String>,
    pub sort: Option<i64>,
}

/// 更新公告请求
#[derive(Debug, Clone, Deserialize, validator::Validate, utoipa::ToSchema)]
pub struct UpdateNoticeRequest {
    pub title: Option<String>,
    pub content: Option<String>,
    pub sort: Option<i64>,
}
