//! 单页模型

use serde::{Deserialize, Serialize};
use sqlx::FromRow;

/// 页面类型
pub const PAGE_TYPE_INTERNAL: &str = "internal";
pub const PAGE_TYPE_EXTERNAL: &str = "external";

/// 单页实体
#[derive(Debug, Clone, FromRow, Serialize, utoipa::ToSchema)]
pub struct Page {
    pub id: i64,
    #[serde(rename = "type")]
    pub page_type: String,
    pub name: String,
    pub icon: String,
    pub title: String,
    pub content: Option<String>,
    pub keywords: Option<String>,
    pub description: Option<String>,
    pub slug: String,
    pub url: String,
    pub view_count: i64,
    pub sort: i64,
    pub is_show: i64,
    pub created_at: String,
    pub updated_at: String,
    pub deleted_at: Option<String>,
}

/// 创建页面请求
#[derive(Debug, Clone, Deserialize, validator::Validate, utoipa::ToSchema)]
pub struct CreatePageRequest {
    pub page_type: Option<String>,
    pub name: String,
    pub icon: Option<String>,
    pub title: Option<String>,
    pub content: Option<String>,
    pub keywords: Option<String>,
    pub description: Option<String>,
    pub slug: Option<String>,
    pub url: Option<String>,
    pub sort: Option<i64>,
    pub is_show: Option<i64>,
}

/// 更新页面请求
#[derive(Debug, Clone, Deserialize, validator::Validate, utoipa::ToSchema)]
pub struct UpdatePageRequest {
    pub page_type: Option<String>,
    pub name: Option<String>,
    pub icon: Option<String>,
    pub title: Option<String>,
    pub content: Option<String>,
    pub keywords: Option<String>,
    pub description: Option<String>,
    pub slug: Option<String>,
    pub url: Option<String>,
    pub sort: Option<i64>,
    pub is_show: Option<i64>,
}
