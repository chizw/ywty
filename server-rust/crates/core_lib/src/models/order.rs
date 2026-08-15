use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};
use sqlx::FromRow;
use uuid::Uuid;

/// 套餐实体
#[derive(Debug, Clone, FromRow, Serialize)]
pub struct Plan {
    pub id: i64,
    pub name: String,
    pub description: Option<String>,
    pub price: i64,
    pub original_price: Option<i64>,
    pub duration_days: i32,
    pub capacity: i64,
    pub features: Option<String>,
    pub is_active: bool,
    pub sort_order: i32,
    pub created_at: DateTime<Utc>,
    pub updated_at: DateTime<Utc>,
}

/// 订单实体
#[derive(Debug, Clone, FromRow, Serialize)]
pub struct Order {
    pub id: i64,
    pub order_no: String,
    pub user_id: i64,
    pub plan_id: i64,
    pub amount: i64,
    pub discount_amount: i64,
    pub paid_amount: i64,
    pub status: String,
    pub payment_method: Option<String>,
    pub payment_no: Option<String>,
    pub paid_at: Option<DateTime<Utc>>,
    pub expired_at: Option<DateTime<Utc>>,
    pub created_at: DateTime<Utc>,
    pub updated_at: DateTime<Utc>,
}

/// 创建订单请求
#[derive(Debug, Clone, Deserialize)]
pub struct CreateOrderRequest {
    pub plan_id: i64,
    pub coupon_code: Option<String>,
    pub payment_method: String,
}

/// 订单响应
#[derive(Debug, Clone, Serialize)]
pub struct OrderResponse {
    pub order_no: String,
    pub amount: i64,
    pub payment_method: String,
    pub payment_url: Option<String>,
    pub qr_code: Option<String>,
    pub expired_at: DateTime<Utc>,
}

/// 优惠券实体
#[derive(Debug, Clone, FromRow, Serialize)]
pub struct Coupon {
    pub id: i64,
    pub code: String,
    pub name: String,
    pub discount_type: String,
    pub discount_value: i64,
    pub min_amount: i64,
    pub max_uses: i32,
    pub used_count: i32,
    pub starts_at: DateTime<Utc>,
    pub expires_at: DateTime<Utc>,
    pub is_active: bool,
    pub created_at: DateTime<Utc>,
}

impl Order {
    pub fn new(user_id: i64, plan_id: i64, amount: i64) -> Self {
        Self {
            id: 0,
            order_no: format!("ORD{}", Uuid::new_v4().to_string().replace("-", "").to_uppercase()[..20].to_string()),
            user_id,
            plan_id,
            amount,
            discount_amount: 0,
            paid_amount: amount,
            status: "pending".to_string(),
            payment_method: None,
            payment_no: None,
            paid_at: None,
            expired_at: Some(Utc::now() + chrono::Duration::minutes(30)),
            created_at: Utc::now(),
            updated_at: Utc::now(),
        }
    }
}
