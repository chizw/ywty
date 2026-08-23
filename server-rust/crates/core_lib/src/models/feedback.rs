//! 意见反馈 + 违规记录模型

use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};
use sqlx::FromRow;

/// 反馈类型
pub const FEEDBACK_TYPE_GENERAL: &str = "general";
pub const FEEDBACK_TYPE_BUG: &str = "bug";
pub const FEEDBACK_TYPE_SUGGEST: &str = "suggest";

/// 意见反馈
#[derive(Debug, Clone, FromRow, Serialize, utoipa::ToSchema)]
pub struct Feedback {
    pub id: i64,
    #[sqlx(rename = "type")]
    #[serde(rename = "type")]
    pub type_: String,
    pub title: String,
    pub name: String,
    pub email: String,
    pub content: String,
    pub ip_address: Option<String>,
    pub created_at: DateTime<Utc>,
}

/// 创建反馈请求
#[derive(Debug, Clone, Deserialize, validator::Validate, utoipa::ToSchema)]
pub struct CreateFeedbackRequest {
    #[serde(rename = "type")]
    #[validate(length(min = 1, max = 32, message = "类型无效"))]
    pub type_: String,
    #[validate(length(min = 1, max = 64, message = "标题不能为空"))]
    pub title: String,
    #[validate(length(min = 1, max = 64, message = "姓名不能为空"))]
    pub name: String,
    #[validate(email(message = "邮箱格式不正确"))]
    pub email: String,
    #[validate(length(min = 1, message = "内容不能为空"))]
    pub content: String,
}

/// 违规状态
pub const VIOLATION_STATUS_UNHANDLED: &str = "unhandled";
pub const VIOLATION_STATUS_HANDLED: &str = "handled";
pub const VIOLATION_STATUS_IGNORED: &str = "ignored";

/// 违规记录
#[derive(Debug, Clone, FromRow, Serialize, utoipa::ToSchema)]
pub struct Violation {
    pub id: i64,
    pub user_id: i64,
    pub photo_id: i64,
    pub reason: String,
    pub status: String,
    pub handled_at: Option<DateTime<Utc>>,
    pub created_at: DateTime<Utc>,
    pub updated_at: DateTime<Utc>,
}

/// 创建违规记录请求
#[derive(Debug, Clone, Deserialize, validator::Validate, utoipa::ToSchema)]
pub struct CreateViolationRequest {
    pub photo_id: i64,
    #[validate(length(min = 1, max = 255, message = "原因不能为空"))]
    pub reason: String,
}
