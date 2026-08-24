//! 容量服务

use crate::db::DbPool;

use crate::error::AppResult;

#[derive(Clone)]
pub struct CapacityService {
    pool: DbPool,
}

impl CapacityService {
    pub fn new(pool: DbPool) -> Self {
        Self { pool }
    }

    /// 获取用户容量信息
    pub async fn get_user_capacity(&self, user_id: i64) -> AppResult<serde_json::Value> {
        let user: Option<(i64, i64)> = sqlx::query_as(
            "SELECT capacity_used, capacity_max FROM users WHERE id = ? AND deleted_at IS NULL",
        )
        .bind(user_id)
        .fetch_optional(&self.pool)
        .await?;

        let (used, capacity) = user.unwrap_or_default();

        let max_image = 10 * 1024 * 1024; // 10MB 单图限制
        let unlimited = capacity <= 0;
        let remain = if unlimited {
            0
        } else {
            (capacity - used).max(0)
        };
        let used_percent = if capacity > 0 {
            ((used as f64 / capacity as f64) * 100.0) as i64
        } else {
            0
        };

        Ok(serde_json::json!({
            "used": used,
            "capacity": if unlimited { 0 } else { capacity },
            "max_image": max_image,
            "unlimited": unlimited,
            "used_percent": used_percent,
            "remain": remain,
        }))
    }

    /// 用户已用存储（字节），复用 users.capacity_used 统计
    pub async fn used_bytes(&self, user_id: i64) -> AppResult<i64> {
        let used: Option<i64> = sqlx::query_scalar(
            "SELECT capacity_used FROM users WHERE id = ? AND deleted_at IS NULL",
        )
        .bind(user_id)
        .fetch_optional(&self.pool)
        .await?;
        Ok(used.unwrap_or(0))
    }

    /// 生效配额（字节）：用户 override 优先，其次其角色组 max_storage，
    /// 再退回默认组；都无则 None = 不限
    pub async fn effective_limit_bytes(&self, user_id: i64) -> AppResult<Option<i64>> {
        // 1. 用户单独覆盖
        let row: Option<(Option<i64>, Option<String>)> = sqlx::query_as(
            "SELECT quota_override, role FROM users WHERE id = ? AND deleted_at IS NULL",
        )
        .bind(user_id)
        .fetch_optional(&self.pool)
        .await?;

        let (override_quota, role) = match row {
            Some((q, r)) => (q, r),
            None => return Ok(None),
        };
        if let Some(q) = override_quota {
            return Ok(Some(q));
        }

        // 2. 角色组：组名与用户 role 同名的组优先
        if let Some(role) = role {
            let limit: Option<Option<i64>> = sqlx::query_scalar(
                "SELECT max_storage FROM `groups` WHERE deleted_at IS NULL AND name = ? LIMIT 1",
            )
            .bind(&role)
            .fetch_optional(&self.pool)
            .await?;
            if let Some(limit) = limit {
                return Ok(limit);
            }
        }

        // 3. 默认组
        let limit: Option<Option<i64>> = sqlx::query_scalar(
            "SELECT max_storage FROM `groups` WHERE deleted_at IS NULL AND is_default = 1 ORDER BY id ASC LIMIT 1",
        )
        .fetch_optional(&self.pool)
        .await?;
        Ok(limit.flatten())
    }
}
