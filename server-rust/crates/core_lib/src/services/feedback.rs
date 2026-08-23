//! 意见反馈 + 违规记录服务

use chrono::Utc;
use sqlx::SqlitePool;

use crate::error::{AppError, AppResult};
use crate::models::feedback::{CreateFeedbackRequest, CreateViolationRequest, Feedback, Violation};

#[derive(Clone)]
pub struct FeedbackService {
    pool: SqlitePool,
}

impl FeedbackService {
    pub fn new(pool: SqlitePool) -> Self {
        Self { pool }
    }

    /// 创建反馈（公开）
    pub async fn create(&self, ip: &str, req: &CreateFeedbackRequest) -> AppResult<Feedback> {
        let now = Utc::now().to_rfc3339();

        let result = sqlx::query(
            r#"
            INSERT INTO feedbacks (type, title, name, email, content, ip_address, created_at)
            VALUES (?, ?, ?, ?, ?, ?, ?)
            "#,
        )
        .bind(&req.type_)
        .bind(&req.title)
        .bind(&req.name)
        .bind(&req.email)
        .bind(&req.content)
        .bind(ip)
        .bind(&now)
        .execute(&self.pool)
        .await?;

        let id = result.last_insert_rowid();

        Ok(Feedback {
            id,
            type_: req.type_.clone(),
            title: req.title.clone(),
            name: req.name.clone(),
            email: req.email.clone(),
            content: req.content.clone(),
            ip_address: Some(ip.to_string()),
            created_at: Utc::now(),
        })
    }

    /// 后台分页列表
    pub async fn list(&self, page: i64, per_page: i64) -> AppResult<(Vec<Feedback>, i64)> {
        let base = "FROM feedbacks WHERE 1=1";

        // 总数
        let total: i64 = sqlx::query_as::<_, (i64,)>(&format!("SELECT COUNT(*) {}", base))
            .fetch_one(&self.pool)
            .await?
            .0;

        // 分页
        let offset = (page - 1) * per_page;
        let rows = sqlx::query_as::<_, Feedback>(&format!(
            "SELECT * {} ORDER BY id DESC LIMIT ? OFFSET ?",
            base
        ))
        .bind(per_page)
        .bind(offset)
        .fetch_all(&self.pool)
        .await?;

        Ok((rows, total))
    }

    /// 删除反馈
    pub async fn delete(&self, id: i64) -> AppResult<()> {
        let result = sqlx::query("DELETE FROM feedbacks WHERE id = ?")
            .bind(id)
            .execute(&self.pool)
            .await?;

        if result.rows_affected() == 0 {
            return Err(AppError::NotFound("反馈不存在".to_string()));
        }

        Ok(())
    }
}

#[derive(Clone)]
pub struct ViolationService {
    pool: SqlitePool,
}

impl ViolationService {
    pub fn new(pool: SqlitePool) -> Self {
        Self { pool }
    }

    /// 创建违规记录
    pub async fn create(&self, user_id: i64, req: &CreateViolationRequest) -> AppResult<Violation> {
        let now = Utc::now().to_rfc3339();

        let result = sqlx::query(
            r#"
            INSERT INTO violations (user_id, photo_id, reason, status, created_at, updated_at)
            VALUES (?, ?, ?, 'unhandled', ?, ?)
            "#,
        )
        .bind(user_id)
        .bind(req.photo_id)
        .bind(&req.reason)
        .bind(&now)
        .bind(&now)
        .execute(&self.pool)
        .await?;

        let id = result.last_insert_rowid();

        Ok(Violation {
            id,
            user_id,
            photo_id: req.photo_id,
            reason: req.reason.clone(),
            status: "unhandled".to_string(),
            handled_at: None,
            created_at: Utc::now(),
            updated_at: Utc::now(),
        })
    }

    /// 后台分页列表
    pub async fn list(
        &self,
        page: i64,
        per_page: i64,
        status: Option<&str>,
    ) -> AppResult<(Vec<Violation>, i64)> {
        let base = "FROM violations WHERE 1=1";
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
        let mut q = sqlx::query_as::<_, Violation>(&list_sql);
        if let Some(s) = status {
            q = q.bind(s);
        }
        let rows = q.bind(per_page).bind(offset).fetch_all(&self.pool).await?;

        Ok((rows, total))
    }

    /// 更新状态
    pub async fn update_status(&self, id: i64, status: &str) -> AppResult<()> {
        // 验证状态值
        match status {
            "unhandled" | "handled" | "ignored" => {}
            _ => return Err(AppError::Validation("无效的状态值".to_string())),
        }

        let now = Utc::now().to_rfc3339();
        let result = sqlx::query(
            "UPDATE violations SET status = ?, handled_at = ?, updated_at = ? WHERE id = ?",
        )
        .bind(status)
        .bind(&now)
        .bind(&now)
        .bind(id)
        .execute(&self.pool)
        .await?;

        if result.rows_affected() == 0 {
            return Err(AppError::NotFound("违规记录不存在".to_string()));
        }

        Ok(())
    }
}
