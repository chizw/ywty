//! 优惠券服务

use crate::db::DbPool;

use crate::error::{AppError, AppResult};
use crate::models::coupon::{
    AdminCouponRequest, Coupon, ValidateCouponRequest, ValidateCouponResult,
};

#[derive(Clone)]
pub struct CouponService {
    pool: DbPool,
}

impl CouponService {
    pub fn new(pool: DbPool) -> Self {
        Self { pool }
    }

    /// 列出优惠券
    pub async fn list(&self, page: i32, per_page: i32) -> AppResult<(Vec<Coupon>, i64)> {
        let page = if page < 1 { 1 } else { page };
        let per_page = if !(1..=100).contains(&per_page) {
            20
        } else {
            per_page
        };
        let offset = (page - 1) * per_page;

        let rows = sqlx::query_as(
            "SELECT * FROM coupons WHERE deleted_at IS NULL ORDER BY id DESC LIMIT ? OFFSET ?",
        )
        .bind(per_page)
        .bind(offset)
        .fetch_all(&self.pool)
        .await?;

        let total: i64 =
            sqlx::query_scalar("SELECT COUNT(*) FROM coupons WHERE deleted_at IS NULL")
                .fetch_one(&self.pool)
                .await?;

        Ok((rows, total))
    }

    /// 获取优惠券详情
    pub async fn get(&self, id: i64) -> AppResult<Coupon> {
        let coupon: Option<Coupon> =
            sqlx::query_as("SELECT * FROM coupons WHERE id = ? AND deleted_at IS NULL")
                .bind(id)
                .fetch_optional(&self.pool)
                .await?;

        coupon.ok_or_else(|| AppError::NotFound("优惠券不存在".to_string()))
    }

    /// 创建优惠券
    pub async fn create(&self, req: &AdminCouponRequest) -> AppResult<Coupon> {
        let now = crate::db::now_str();
        let coupon_type = req.coupon_type.as_deref().unwrap_or("direct");
        let name = req.name.as_deref().unwrap_or("");
        let code = req.code.as_deref().unwrap_or("");
        let value = req.value.unwrap_or(0.0);
        let usage_limit = req.usage_limit.unwrap_or(1);
        let expired_at = req.expired_at.clone().unwrap_or_default();

        let result = sqlx::query(
            r#"
            INSERT INTO coupons (type, name, code, value, usage_limit, used_count, expired_at, created_at, updated_at)
            VALUES (?, ?, ?, ?, ?, 0, ?, ?, ?)
            "#,
        )
        .bind(coupon_type)
        .bind(name)
        .bind(code)
        .bind(value)
        .bind(usage_limit)
        .bind(&expired_at)
        .bind(&now)
        .bind(&now)
        .execute(&self.pool)
        .await?;

        let id = crate::db::last_id(&result);
        self.get(id).await
    }

    /// 更新优惠券
    pub async fn update(&self, id: i64, req: &AdminCouponRequest) -> AppResult<Coupon> {
        let existing = self.get(id).await?;
        let now = crate::db::now_str();

        let coupon_type = req.coupon_type.clone().unwrap_or(existing.coupon_type);
        let name = req.name.clone().unwrap_or(existing.name);
        let code = req.code.clone().unwrap_or(existing.code);
        let value = req.value.unwrap_or(existing.value);
        let usage_limit = req.usage_limit.unwrap_or(existing.usage_limit);
        let expired_at = req.expired_at.clone().unwrap_or_default();

        sqlx::query(
            "UPDATE coupons SET type = ?, name = ?, code = ?, value = ?, usage_limit = ?, expired_at = ?, updated_at = ? WHERE id = ?",
        )
        .bind(&coupon_type)
        .bind(&name)
        .bind(&code)
        .bind(value)
        .bind(usage_limit)
        .bind(&expired_at)
        .bind(&now)
        .bind(id)
        .execute(&self.pool)
        .await?;

        self.get(id).await
    }

    /// 删除优惠券
    pub async fn delete(&self, id: i64) -> AppResult<()> {
        let now = crate::db::now_str();
        let result =
            sqlx::query("UPDATE coupons SET deleted_at = ? WHERE id = ? AND deleted_at IS NULL")
                .bind(&now)
                .bind(id)
                .execute(&self.pool)
                .await?;

        if result.rows_affected() == 0 {
            return Err(AppError::NotFound("优惠券不存在".to_string()));
        }
        Ok(())
    }

    /// 校验优惠券
    pub async fn validate(&self, req: &ValidateCouponRequest) -> AppResult<ValidateCouponResult> {
        let coupon: Option<Coupon> =
            sqlx::query_as("SELECT * FROM coupons WHERE code = ? AND deleted_at IS NULL")
                .bind(&req.code)
                .fetch_optional(&self.pool)
                .await?;

        let coupon = coupon.ok_or_else(|| AppError::NotFound("优惠券不存在".to_string()))?;

        // 检查过期
        if let Some(ref exp) = coupon.expired_at {
            if let Ok(exp_time) = chrono::DateTime::parse_from_rfc3339(exp) {
                if exp_time < chrono::Utc::now() {
                    return Err(AppError::Business("优惠券已过期".to_string()));
                }
            }
        }

        // 检查使用次数
        if coupon.used_count >= coupon.usage_limit {
            return Err(AppError::Business("优惠券已达使用上限".to_string()));
        }

        // 计算折扣（简化：direct 类型直接减）
        let discount = coupon.value as i64;

        Ok(ValidateCouponResult { coupon, discount })
    }
}
