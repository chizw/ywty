//! 订单模型

use serde::{Deserialize, Serialize};
use sqlx::FromRow;

/// 订单类型
pub const ORDER_TYPE_PLAN: &str = "plan";
pub const ORDER_TYPE_CHARGE: &str = "charge";

/// 订单状态
pub const ORDER_STATUS_UNPAID: &str = "unpaid";
pub const ORDER_STATUS_PAID: &str = "paid";
pub const ORDER_STATUS_CANCELED: &str = "canceled";
pub const ORDER_STATUS_REFUNDED: &str = "refunded";

/// 订单实体
#[derive(Debug, Clone, FromRow, Serialize, utoipa::ToSchema)]
pub struct Order {
    pub id: i64,
    pub plan_id: i64,
    pub user_id: i64,
    pub coupon_id: i64,
    pub trade_no: String,
    pub out_trade_no: String,
    #[serde(rename = "type")]
    pub order_type: String,
    pub amount: i64,
    pub deduct_amount: i64,
    pub snapshot: Option<String>,
    pub product: Option<String>,
    pub pay_method: String,
    pub status: String,
    pub paid_at: Option<String>,
    pub canceled_at: Option<String>,
    pub created_at: String,
    pub updated_at: String,
}

/// 创建订单请求
#[derive(Debug, Clone, Deserialize, validator::Validate, utoipa::ToSchema)]
pub struct CreateOrderRequest {
    pub plan_id: i64,
    pub coupon_code: Option<String>,
    pub pay_method: Option<String>,
}
