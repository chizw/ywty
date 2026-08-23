//! 优惠券模型

use serde::{Deserialize, Serialize};
use sqlx::FromRow;

/// 优惠券类型
pub const COUPON_TYPE_DIRECT: &str = "direct";
pub const COUPON_TYPE_PERCENT: &str = "percent";

/// 优惠券实体
#[derive(Debug, Clone, FromRow, Serialize, utoipa::ToSchema)]
pub struct Coupon {
    pub id: i64,
    #[serde(rename = "type")]
    pub coupon_type: String,
    pub name: String,
    pub code: String,
    pub value: f64,
    pub usage_limit: i64,
    pub used_count: i64,
    pub expired_at: Option<String>,
    pub created_at: String,
    pub updated_at: String,
    pub deleted_at: Option<String>,
}

/// 管理端优惠券请求
#[derive(Debug, Clone, Deserialize, validator::Validate, utoipa::ToSchema)]
pub struct AdminCouponRequest {
    pub coupon_type: Option<String>,
    pub name: Option<String>,
    pub code: Option<String>,
    pub value: Option<f64>,
    pub usage_limit: Option<i64>,
    pub expired_at: Option<String>,
}

/// 校验优惠券请求
#[derive(Debug, Clone, Deserialize, validator::Validate, utoipa::ToSchema)]
pub struct ValidateCouponRequest {
    pub code: String,
    pub amount: Option<i64>,
}

/// 校验结果
#[derive(Debug, Clone, Serialize)]
pub struct ValidateCouponResult {
    pub coupon: Coupon,
    pub discount: i64,
}
