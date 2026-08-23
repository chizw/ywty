//! 订单服务

use sqlx::SqlitePool;

use crate::error::{AppError, AppResult};
use crate::models::order::{CreateOrderRequest, Order, ORDER_STATUS_PAID, ORDER_STATUS_UNPAID};
use crate::services::payment::{NotifyPayload, PaymentDriver, PaymentParams};

#[derive(Clone)]
pub struct OrderService {
    pool: SqlitePool,
}

impl OrderService {
    pub fn new(pool: SqlitePool) -> Self {
        Self { pool }
    }

    /// 生成订单号
    fn generate_trade_no() -> String {
        use base64::{engine::general_purpose::URL_SAFE_NO_PAD, Engine as _};
        let bytes: Vec<u8> = (0..12).map(|_| rand::random::<u8>()).collect();
        format!("T{}", URL_SAFE_NO_PAD.encode(&bytes))
    }

    /// 创建订单
    ///
    /// 金额取自套餐价格（plan_prices 中最低档），商品名取套餐名，不再写死 0。
    pub async fn create(&self, user_id: i64, req: &CreateOrderRequest) -> AppResult<Order> {
        let plan: Option<(String,)> =
            sqlx::query_as("SELECT name FROM plans WHERE id = ? AND deleted_at IS NULL")
                .bind(req.plan_id)
                .fetch_optional(&self.pool)
                .await?;
        let (plan_name,) =
            plan.ok_or_else(|| AppError::NotFound("套餐不存在或已下架".to_string()))?;

        let amount: i64 =
            sqlx::query_scalar("SELECT COALESCE(MIN(price), 0) FROM plan_prices WHERE plan_id = ?")
                .bind(req.plan_id)
                .fetch_one(&self.pool)
                .await?;

        let trade_no = Self::generate_trade_no();
        let out_trade_no = Self::generate_trade_no();
        let now = chrono::Utc::now().to_rfc3339();

        let result = sqlx::query(
            r#"
            INSERT INTO orders (plan_id, user_id, coupon_id, trade_no, out_trade_no, type, amount, deduct_amount, product, pay_method, status, created_at, updated_at)
            VALUES (?, ?, 0, ?, ?, 'plan', ?, 0, ?, ?, ?, ?, ?)
            "#,
        )
        .bind(req.plan_id)
        .bind(user_id)
        .bind(&trade_no)
        .bind(&out_trade_no)
        .bind(amount)
        .bind(&plan_name)
        .bind(req.pay_method.as_deref().unwrap_or(""))
        .bind(ORDER_STATUS_UNPAID)
        .bind(&now)
        .bind(&now)
        .execute(&self.pool)
        .await?;

        let id = result.last_insert_rowid();
        self.get_by_id(id).await
    }

    /// 获取订单详情
    pub async fn get_by_id(&self, id: i64) -> AppResult<Order> {
        let order: Option<Order> = sqlx::query_as("SELECT * FROM orders WHERE id = ?")
            .bind(id)
            .fetch_optional(&self.pool)
            .await?;

        order.ok_or_else(|| AppError::NotFound("订单不存在".to_string()))
    }

    /// 获取订单详情（仅所有者）
    pub async fn get(&self, user_id: i64, id: i64) -> AppResult<Order> {
        let order: Option<Order> =
            sqlx::query_as("SELECT * FROM orders WHERE id = ? AND user_id = ?")
                .bind(id)
                .bind(user_id)
                .fetch_optional(&self.pool)
                .await?;

        order.ok_or_else(|| AppError::NotFound("订单不存在".to_string()))
    }

    /// 列出用户订单
    pub async fn list(&self, user_id: i64, page: i32, per_page: i32) -> AppResult<Vec<Order>> {
        let page = if page < 1 { 1 } else { page };
        let per_page = if !(1..=100).contains(&per_page) {
            20
        } else {
            per_page
        };
        let offset = (page - 1) * per_page;

        let rows = sqlx::query_as(
            "SELECT * FROM orders WHERE user_id = ? ORDER BY id DESC LIMIT ? OFFSET ?",
        )
        .bind(user_id)
        .bind(per_page)
        .bind(offset)
        .fetch_all(&self.pool)
        .await?;

        Ok(rows)
    }

    /// 取消订单
    pub async fn cancel(&self, user_id: i64, id: i64) -> AppResult<()> {
        let now = chrono::Utc::now().to_rfc3339();
        let result = sqlx::query(
            "UPDATE orders SET status = 'canceled', canceled_at = ? WHERE id = ? AND user_id = ? AND status = 'unpaid'",
        )
        .bind(&now)
        .bind(id)
        .bind(user_id)
        .execute(&self.pool)
        .await?;

        if result.rows_affected() == 0 {
            return Err(AppError::NotFound("订单不存在或无法取消".to_string()));
        }
        Ok(())
    }

    /// 发起支付：校验订单归属与状态后，交由支付驱动生成支付参数
    ///
    /// 不直接改单，订单状态只由 [`Self::handle_notify`] 的回调链路推进。
    pub async fn pay_initiate(
        &self,
        user_id: i64,
        id: i64,
        driver: &dyn PaymentDriver,
    ) -> AppResult<(Order, PaymentParams)> {
        let order = self.get(user_id, id).await?;
        if order.status != ORDER_STATUS_UNPAID {
            return Err(AppError::Business(format!(
                "订单当前状态为 {}，无法发起支付",
                order.status
            )));
        }
        let params = driver.create_payment(&order).await?;
        Ok((order, params))
    }

    /// 处理验签通过的支付回调（状态机防重放）
    ///
    /// - 仅 `unpaid -> paid` 允许转换；
    /// - 已支付订单重复通知幂等返回成功；
    /// - 通知金额必须与订单金额一致。
    pub async fn handle_notify(
        &self,
        payload: &NotifyPayload,
        pay_method: &str,
    ) -> AppResult<PaidOutcome> {
        let order: Option<Order> = sqlx::query_as("SELECT * FROM orders WHERE trade_no = ?")
            .bind(&payload.trade_no)
            .fetch_optional(&self.pool)
            .await?;
        let order = order.ok_or_else(|| AppError::NotFound("订单不存在".to_string()))?;

        if payload.amount != order.amount {
            return Err(AppError::Business("通知金额与订单金额不一致".to_string()));
        }

        match order.status.as_str() {
            // 幂等：已支付订单重复通知直接返回成功
            ORDER_STATUS_PAID => Ok(PaidOutcome::AlreadyPaid(order)),
            ORDER_STATUS_UNPAID => {
                let now = chrono::Utc::now().to_rfc3339();
                let result = sqlx::query(
                    "UPDATE orders SET status = ?, pay_method = ?, paid_at = ?, updated_at = ? WHERE trade_no = ? AND status = ?",
                )
                .bind(ORDER_STATUS_PAID)
                .bind(pay_method)
                .bind(&now)
                .bind(&now)
                .bind(&payload.trade_no)
                .bind(ORDER_STATUS_UNPAID)
                .execute(&self.pool)
                .await?;

                if result.rows_affected() == 0 {
                    // 并发回调竞态：重查最新状态
                    let latest: Order = sqlx::query_as("SELECT * FROM orders WHERE trade_no = ?")
                        .bind(&payload.trade_no)
                        .fetch_one(&self.pool)
                        .await?;
                    if latest.status == ORDER_STATUS_PAID {
                        return Ok(PaidOutcome::AlreadyPaid(latest));
                    }
                    return Err(AppError::Business(format!(
                        "订单状态为 {}，无法完成支付",
                        latest.status
                    )));
                }

                let updated: Order = sqlx::query_as("SELECT * FROM orders WHERE trade_no = ?")
                    .bind(&payload.trade_no)
                    .fetch_one(&self.pool)
                    .await?;
                Ok(PaidOutcome::Paid(updated))
            }
            other => Err(AppError::Business(format!(
                "订单状态为 {other}，无法完成支付"
            ))),
        }
    }
}

/// 回调处理结果
#[derive(Debug)]
pub enum PaidOutcome {
    /// 本次通知完成支付
    Paid(Order),
    /// 订单此前已支付（幂等）
    AlreadyPaid(Order),
}
