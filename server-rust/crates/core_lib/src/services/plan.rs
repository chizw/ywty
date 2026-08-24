//! 套餐服务

use crate::db::DbPool;

use crate::error::{AppError, AppResult};
use crate::models::plan::{AdminPlanRequest, Plan, PlanDetail, PlanPrice};

#[derive(Clone)]
pub struct PlanService {
    pool: DbPool,
}

impl PlanService {
    pub fn new(pool: DbPool) -> Self {
        Self { pool }
    }

    /// 列出套餐（公开：仅上架）
    pub async fn list_public(&self) -> AppResult<Vec<Plan>> {
        let rows = sqlx::query_as(
            "SELECT * FROM plans WHERE deleted_at IS NULL AND is_up = 1 ORDER BY sort ASC",
        )
        .fetch_all(&self.pool)
        .await?;

        Ok(rows)
    }

    /// 列出上架套餐（含价格档，单次查询避免 N+1）
    pub async fn list_public_with_prices(&self) -> AppResult<Vec<PlanDetail>> {
        let plans = self.list_public().await?;
        if plans.is_empty() {
            return Ok(Vec::new());
        }

        let ids: Vec<i64> = plans.iter().map(|p| p.id).collect();
        let placeholders = vec!["?"; ids.len()].join(", ");
        let sql = format!(
            "SELECT * FROM plan_prices WHERE plan_id IN ({placeholders}) ORDER BY price ASC"
        );
        let mut query = sqlx::query_as::<_, PlanPrice>(&sql);
        for id in &ids {
            query = query.bind(id);
        }
        let all_prices = query.fetch_all(&self.pool).await?;

        Ok(plans
            .into_iter()
            .map(|plan| {
                let prices = all_prices
                    .iter()
                    .filter(|p| p.plan_id == plan.id)
                    .cloned()
                    .collect();
                PlanDetail { plan, prices }
            })
            .collect())
    }

    /// 列出所有套餐（管理端）
    pub async fn list_all(&self) -> AppResult<Vec<Plan>> {
        let rows = sqlx::query_as("SELECT * FROM plans WHERE deleted_at IS NULL ORDER BY sort ASC")
            .fetch_all(&self.pool)
            .await?;

        Ok(rows)
    }

    /// 获取套餐详情
    pub async fn get(&self, id: i64) -> AppResult<PlanDetail> {
        let plan: Option<Plan> =
            sqlx::query_as("SELECT * FROM plans WHERE id = ? AND deleted_at IS NULL")
                .bind(id)
                .fetch_optional(&self.pool)
                .await?;

        let plan = plan.ok_or_else(|| AppError::NotFound("套餐不存在".to_string()))?;

        let prices =
            sqlx::query_as("SELECT * FROM plan_prices WHERE plan_id = ? ORDER BY price ASC")
                .bind(id)
                .fetch_all(&self.pool)
                .await?;

        Ok(PlanDetail { plan, prices })
    }

    /// 创建套餐
    pub async fn create(&self, req: &AdminPlanRequest) -> AppResult<Plan> {
        let now = crate::db::now_str();
        let name = req.name.as_deref().unwrap_or("");
        let plan_type = req.plan_type.as_deref().unwrap_or("vip");
        let intro = req.intro.as_deref().unwrap_or("");
        let features = req.features.as_deref().unwrap_or("");
        let badge = req.badge.as_deref().unwrap_or("");
        let sort = req.sort.unwrap_or(0);
        let is_up = req.is_up.unwrap_or(0);

        let result = sqlx::query(
            r#"
            INSERT INTO plans (type, name, intro, features, badge, sort, is_up, created_at, updated_at)
            VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
            "#,
        )
        .bind(plan_type)
        .bind(name)
        .bind(intro)
        .bind(features)
        .bind(badge)
        .bind(sort)
        .bind(is_up)
        .bind(&now)
        .bind(&now)
        .execute(&self.pool)
        .await?;

        let id = crate::db::last_id(&result);

        // 插入价格
        if let Some(prices) = &req.prices {
            for price in prices {
                sqlx::query(
                    "INSERT INTO plan_prices (plan_id, name, duration, price, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)",
                )
                .bind(id)
                .bind(&price.name)
                .bind(price.duration)
                .bind(price.price)
                .bind(&now)
                .bind(&now)
                .execute(&self.pool)
                .await?;
            }
        }

        self.get(id).await.map(|d| d.plan)
    }

    /// 更新套餐
    pub async fn update(&self, id: i64, req: &AdminPlanRequest) -> AppResult<Plan> {
        // 检查存在
        let existing: Option<Plan> =
            sqlx::query_as("SELECT * FROM plans WHERE id = ? AND deleted_at IS NULL")
                .bind(id)
                .fetch_optional(&self.pool)
                .await?;

        if existing.is_none() {
            return Err(AppError::NotFound("套餐不存在".to_string()));
        }

        let now = crate::db::now_str();

        // 动态更新字段
        let mut updates = Vec::new();
        if let Some(ref t) = req.plan_type {
            updates.push(format!("type = '{}'", t));
        }
        if let Some(ref name) = req.name {
            updates.push(format!("name = '{}'", name));
        }
        if let Some(ref intro) = req.intro {
            updates.push(format!("intro = '{}'", intro));
        }
        if let Some(ref features) = req.features {
            updates.push(format!("features = '{}'", features));
        }
        if let Some(ref badge) = req.badge {
            updates.push(format!("badge = '{}'", badge));
        }
        if let Some(sort) = req.sort {
            updates.push(format!("sort = {}", sort));
        }
        if let Some(is_up) = req.is_up {
            updates.push(format!("is_up = {}", is_up));
        }

        if !updates.is_empty() {
            let sql = format!(
                "UPDATE plans SET {}, updated_at = '{}' WHERE id = {}",
                updates.join(", "),
                now,
                id
            );
            sqlx::query(&sql).execute(&self.pool).await?;
        }

        // 更新价格（如果有）
        if let Some(prices) = &req.prices {
            // 删除旧价格
            sqlx::query("DELETE FROM plan_prices WHERE plan_id = ?")
                .bind(id)
                .execute(&self.pool)
                .await?;
            // 插入新价格
            for price in prices {
                sqlx::query(
                    "INSERT INTO plan_prices (plan_id, name, duration, price, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)",
                )
                .bind(id)
                .bind(&price.name)
                .bind(price.duration)
                .bind(price.price)
                .bind(&now)
                .bind(&now)
                .execute(&self.pool)
                .await?;
            }
        }

        self.get(id).await.map(|d| d.plan)
    }

    /// 删除套餐（软删除）
    pub async fn delete(&self, id: i64) -> AppResult<()> {
        let now = crate::db::now_str();
        let result =
            sqlx::query("UPDATE plans SET deleted_at = ? WHERE id = ? AND deleted_at IS NULL")
                .bind(&now)
                .bind(id)
                .execute(&self.pool)
                .await?;

        if result.rows_affected() == 0 {
            return Err(AppError::NotFound("套餐不存在".to_string()));
        }
        Ok(())
    }

    /// 切换上架状态
    pub async fn toggle_up(&self, id: i64) -> AppResult<Plan> {
        let now = crate::db::now_str();
        sqlx::query(
            "UPDATE plans SET is_up = CASE WHEN is_up = 1 THEN 0 ELSE 1 END, updated_at = ? WHERE id = ? AND deleted_at IS NULL",
        )
        .bind(&now)
        .bind(id)
        .execute(&self.pool)
        .await?;

        let plan: Plan = sqlx::query_as("SELECT * FROM plans WHERE id = ?")
            .bind(id)
            .fetch_one(&self.pool)
            .await?;

        Ok(plan)
    }
}
