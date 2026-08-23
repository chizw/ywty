//! 套餐 / 价格模型

use serde::{Deserialize, Serialize};
use sqlx::FromRow;

/// 套餐类型
pub const PLAN_TYPE_VIP: &str = "vip";
pub const PLAN_TYPE_DEFAULT: &str = "default";

/// 套餐实体
#[derive(Debug, Clone, FromRow, Serialize, utoipa::ToSchema)]
pub struct Plan {
    pub id: i64,
    #[serde(rename = "type")]
    pub plan_type: String,
    pub name: String,
    pub intro: Option<String>,
    pub features: Option<String>,
    pub badge: String,
    pub sort: i64,
    pub is_up: i64,
    pub created_at: String,
    pub updated_at: String,
    pub deleted_at: Option<String>,
}

/// 套餐阶梯价格
#[derive(Debug, Clone, FromRow, Serialize, utoipa::ToSchema)]
pub struct PlanPrice {
    pub id: i64,
    pub plan_id: i64,
    pub name: String,
    pub duration: i64,
    pub price: i64,
    pub created_at: String,
    pub updated_at: String,
}

/// 套餐详情（含价格列表）
#[derive(Debug, Clone, Serialize, utoipa::ToSchema)]
pub struct PlanDetail {
    pub plan: Plan,
    pub prices: Vec<PlanPrice>,
}

/// 管理端套餐请求
#[derive(Debug, Clone, Deserialize, validator::Validate, utoipa::ToSchema)]
pub struct AdminPlanRequest {
    pub plan_type: Option<String>,
    pub name: Option<String>,
    pub intro: Option<String>,
    pub features: Option<String>,
    pub badge: Option<String>,
    pub sort: Option<i64>,
    pub is_up: Option<i64>,
    pub prices: Option<Vec<PlanPriceRequest>>,
}

/// 价格请求
#[derive(Debug, Clone, Deserialize, Serialize, utoipa::ToSchema)]
pub struct PlanPriceRequest {
    pub name: String,
    pub duration: i64,
    pub price: i64,
}
