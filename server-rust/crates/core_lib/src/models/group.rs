//! 群组模型

use serde::{Deserialize, Serialize};
use sqlx::FromRow;

/// 群组实体
#[derive(Debug, Clone, FromRow, Serialize, utoipa::ToSchema)]
pub struct Group {
    pub id: i64,
    pub name: String,
    pub intro: String,
    pub options: Option<String>,
    pub is_default: i64,
    pub is_guest: i64,
    /// 存储配额（字节），None = 不限
    pub max_storage: Option<i64>,
    pub created_at: String,
    pub updated_at: String,
    pub deleted_at: Option<String>,
}

/// 创建群组请求
#[derive(Debug, Clone, Deserialize, validator::Validate, utoipa::ToSchema)]
pub struct CreateGroupRequest {
    pub name: String,
    pub intro: Option<String>,
    pub options: Option<serde_json::Value>,
    pub is_default: Option<i64>,
    pub is_guest: Option<i64>,
    /// 存储配额（字节），null = 不限
    pub max_storage: Option<i64>,
}

/// 更新群组请求
#[derive(Debug, Clone, Deserialize, validator::Validate, utoipa::ToSchema)]
pub struct UpdateGroupRequest {
    pub name: Option<String>,
    pub intro: Option<String>,
    pub options: Option<serde_json::Value>,
    pub is_default: Option<i64>,
    pub is_guest: Option<i64>,
    /// 存储配额（字节）：字段缺省 = 不修改；null = 清除配额（不限）
    #[schema(value_type = Option<i64>)]
    pub max_storage: Option<Option<i64>>,
}
