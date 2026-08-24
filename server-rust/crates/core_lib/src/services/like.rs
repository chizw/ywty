//! 点赞/举报服务

use crate::db::DbPool;
use chrono::Utc;

use crate::error::{AppError, AppResult};
use crate::models::photo::Report;

#[derive(Clone)]
pub struct LikeService {
    pool: DbPool,
}

impl LikeService {
    pub fn new(pool: DbPool) -> Self {
        Self { pool }
    }

    /// 点赞（已点则取消）
    pub async fn toggle(&self, user_id: i64, target_type: &str, target_id: i64) -> AppResult<bool> {
        // 检查是否已点赞
        let existing: Option<(i64,)> = sqlx::query_as(
            "SELECT id FROM likes WHERE user_id = ? AND likeable_type = ? AND likeable_id = ?",
        )
        .bind(user_id)
        .bind(target_type)
        .bind(target_id)
        .fetch_optional(&self.pool)
        .await?;

        if let Some((id,)) = existing {
            // 已存在 -> 取消
            sqlx::query("DELETE FROM likes WHERE id = ?")
                .bind(id)
                .execute(&self.pool)
                .await?;
            Ok(false)
        } else {
            // 新建
            sqlx::query(
                "INSERT INTO likes (user_id, likeable_type, likeable_id, created_at) VALUES (?, ?, ?, ?)",
            )
            .bind(user_id)
            .bind(target_type)
            .bind(target_id)
            .bind(crate::db::now_str())
            .execute(&self.pool)
            .await?;
            Ok(true)
        }
    }

    /// 点赞数量
    pub async fn count(&self, target_type: &str, target_id: i64) -> i64 {
        let count: Option<(i64,)> = sqlx::query_as(
            "SELECT COUNT(*) FROM likes WHERE likeable_type = ? AND likeable_id = ?",
        )
        .bind(target_type)
        .bind(target_id)
        .fetch_optional(&self.pool)
        .await
        .unwrap_or(None);

        count.map(|c| c.0).unwrap_or(0)
    }

    /// 是否已点赞
    pub async fn liked(&self, user_id: i64, target_type: &str, target_id: i64) -> bool {
        let count: Option<(i64,)> = sqlx::query_as(
            "SELECT COUNT(*) FROM likes WHERE user_id = ? AND likeable_type = ? AND likeable_id = ?",
        )
        .bind(user_id)
        .bind(target_type)
        .bind(target_id)
        .fetch_optional(&self.pool)
        .await
        .unwrap_or(None);

        count.map(|c| c.0).unwrap_or(0) > 0
    }
}

#[derive(Clone)]
pub struct ReportService {
    pool: DbPool,
}

impl ReportService {
    pub fn new(pool: DbPool) -> Self {
        Self { pool }
    }

    /// 创建举报
    pub async fn create(
        &self,
        user_id: i64,
        _ip: &str,
        target_type: &str,
        target_id: i64,
        content: &str,
    ) -> AppResult<Report> {
        let now = crate::db::now_str();

        let result = sqlx::query(
            r#"
            INSERT INTO reports (user_id, reportable_type, reportable_id, reason, status, created_at, updated_at)
            VALUES (?, ?, ?, ?, 0, ?, ?)
            "#,
        )
        .bind(user_id)
        .bind(target_type)
        .bind(target_id)
        .bind(content)
        .bind(&now)
        .bind(&now)
        .execute(&self.pool)
        .await?;

        let id = crate::db::last_id(&result);

        Ok(Report {
            id,
            user_id,
            reportable_type: target_type.to_string(),
            reportable_id: target_id,
            reason: content.to_string(),
            status: 0,
            created_at: Utc::now(),
            updated_at: Utc::now(),
        })
    }

    /// 后台分页列表
    pub async fn admin_list(
        &self,
        page: i64,
        per_page: i64,
        status: Option<i32>,
    ) -> AppResult<(Vec<Report>, i64)> {
        // 构建查询
        let base = "FROM reports WHERE 1=1";
        let mut filter = String::new();
        if status.is_some() {
            filter.push_str(" AND status = ?");
        }

        // 总数
        let count_sql = format!("SELECT COUNT(*) {} {}", base, filter);
        let mut q = sqlx::query_as::<_, (i64,)>(&count_sql);
        if let Some(s) = status {
            q = q.bind(s);
        }
        let total: i64 = q.fetch_one(&self.pool).await?.0;

        // 分页
        let list_sql = format!(
            "SELECT * {} {} ORDER BY id DESC LIMIT ? OFFSET ?",
            base, filter
        );
        let offset = (page - 1) * per_page;
        let mut q = sqlx::query_as::<_, Report>(&list_sql);
        if let Some(s) = status {
            q = q.bind(s);
        }
        let rows = q.bind(per_page).bind(offset).fetch_all(&self.pool).await?;

        Ok((rows, total))
    }

    /// 更新状态
    pub async fn update_status(&self, id: i64, status: i64) -> AppResult<()> {
        let now = crate::db::now_str();
        let result = sqlx::query("UPDATE reports SET status = ?, updated_at = ? WHERE id = ?")
            .bind(status)
            .bind(&now)
            .bind(id)
            .execute(&self.pool)
            .await?;

        if result.rows_affected() == 0 {
            return Err(AppError::NotFound("举报不存在".to_string()));
        }

        Ok(())
    }
}
