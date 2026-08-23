//! 工单 / 回复模型

use serde::{Deserialize, Serialize};
use sqlx::FromRow;

/// 工单级别
pub const TICKET_LEVEL_LOW: &str = "low";
pub const TICKET_LEVEL_MEDIUM: &str = "medium";
pub const TICKET_LEVEL_HIGH: &str = "high";
pub const TICKET_LEVEL_URGENT: &str = "urgent";

/// 工单状态
pub const TICKET_STATUS_IN_PROGRESS: &str = "in_progress";
pub const TICKET_STATUS_RESOLVED: &str = "resolved";
pub const TICKET_STATUS_CLOSED: &str = "closed";

/// 工单类型
pub const TICKET_TYPE_BUG: &str = "bug";
pub const TICKET_TYPE_FEATURE: &str = "feature";
pub const TICKET_TYPE_COMPLAINT: &str = "complaint";
pub const TICKET_TYPE_OTHER: &str = "other";

/// 工单实体
#[derive(Debug, Clone, FromRow, Serialize, utoipa::ToSchema)]
pub struct Ticket {
    pub id: i64,
    pub user_id: i64,
    pub issue_no: String,
    pub title: String,
    #[serde(rename = "type")]
    pub ticket_type: String,
    pub level: String,
    pub status: String,
    pub created_at: String,
    pub updated_at: String,
    pub deleted_at: Option<String>,
}

/// 工单回复实体
#[derive(Debug, Clone, FromRow, Serialize, utoipa::ToSchema)]
pub struct TicketReply {
    pub id: i64,
    pub ticket_id: i64,
    pub user_id: i64,
    /// 是否为管理员回复
    pub is_admin: i64,
    pub content: String,
    pub is_notify: i64,
    pub read_at: Option<String>,
    pub created_at: String,
    pub updated_at: String,
}

/// 创建工单请求
#[derive(Debug, Clone, Deserialize, validator::Validate, utoipa::ToSchema)]
pub struct CreateTicketRequest {
    pub title: String,
    pub ticket_type: Option<String>,
    pub level: Option<String>,
    pub content: String,
}

/// 工单详情（含回复）
#[derive(Debug, Clone, Serialize)]
pub struct TicketDetail {
    pub ticket: Ticket,
    pub replies: Vec<TicketReply>,
}
