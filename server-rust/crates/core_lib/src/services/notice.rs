//! 公告服务

use sqlx::SqlitePool;

use crate::error::{AppError, AppResult};
use crate::models::notice::{CreateNoticeRequest, Notice, UpdateNoticeRequest};

#[derive(Clone)]
pub struct NoticeService {
    pool: SqlitePool,
}

impl NoticeService {
    pub fn new(pool: SqlitePool) -> Self {
        Self { pool }
    }

    /// 列出公开公告
    pub async fn list_public(&self, page: i32, per_page: i32) -> AppResult<(Vec<Notice>, i64)> {
        let page = if page < 1 { 1 } else { page };
        let per_page = if !(1..=100).contains(&per_page) {
            20
        } else {
            per_page
        };
        let offset = (page - 1) * per_page;

        let rows = sqlx::query_as(
            "SELECT * FROM notices WHERE deleted_at IS NULL ORDER BY sort DESC, id DESC LIMIT ? OFFSET ?",
        )
        .bind(per_page)
        .bind(offset)
        .fetch_all(&self.pool)
        .await?;

        let total: i64 =
            sqlx::query_scalar("SELECT COUNT(*) FROM notices WHERE deleted_at IS NULL")
                .fetch_one(&self.pool)
                .await?;

        Ok((rows, total))
    }

    /// 获取公开公告详情
    pub async fn get_public(&self, id: i64) -> AppResult<Notice> {
        let notice: Option<Notice> =
            sqlx::query_as("SELECT * FROM notices WHERE id = ? AND deleted_at IS NULL")
                .bind(id)
                .fetch_optional(&self.pool)
                .await?;

        let notice = notice.ok_or_else(|| AppError::NotFound("公告不存在".to_string()))?;

        // 增加浏览计数
        sqlx::query("UPDATE notices SET view_count = view_count + 1 WHERE id = ?")
            .bind(id)
            .execute(&self.pool)
            .await?;

        Ok(notice)
    }

    /// 管理端列表
    pub async fn admin_list(&self, page: i32, per_page: i32) -> AppResult<(Vec<Notice>, i64)> {
        self.list_public(page, per_page).await
    }

    /// 创建公告
    pub async fn admin_create(&self, req: &CreateNoticeRequest) -> AppResult<Notice> {
        let now = chrono::Utc::now().to_rfc3339();
        let sort = req.sort.unwrap_or(0);

        let result = sqlx::query(
            r#"
            INSERT INTO notices (title, content, view_count, sort, created_at, updated_at)
            VALUES (?, ?, 0, ?, ?, ?)
            "#,
        )
        .bind(&req.title)
        .bind(&req.content)
        .bind(sort)
        .bind(&now)
        .bind(&now)
        .execute(&self.pool)
        .await?;

        let id = result.last_insert_rowid();
        self.admin_get(id).await
    }

    /// 获取公告（管理端）
    pub async fn admin_get(&self, id: i64) -> AppResult<Notice> {
        let notice: Option<Notice> =
            sqlx::query_as("SELECT * FROM notices WHERE id = ? AND deleted_at IS NULL")
                .bind(id)
                .fetch_optional(&self.pool)
                .await?;

        notice.ok_or_else(|| AppError::NotFound("公告不存在".to_string()))
    }

    /// 更新公告
    pub async fn admin_update(&self, id: i64, req: &UpdateNoticeRequest) -> AppResult<Notice> {
        let existing = self.admin_get(id).await?;
        let now = chrono::Utc::now().to_rfc3339();

        let title = req.title.clone().unwrap_or(existing.title);
        let content = req.content.clone().unwrap_or_default();
        let sort = req.sort.unwrap_or(existing.sort);

        sqlx::query(
            "UPDATE notices SET title = ?, content = ?, sort = ?, updated_at = ? WHERE id = ?",
        )
        .bind(&title)
        .bind(&content)
        .bind(sort)
        .bind(&now)
        .bind(id)
        .execute(&self.pool)
        .await?;

        self.admin_get(id).await
    }

    /// 删除公告
    pub async fn admin_delete(&self, id: i64) -> AppResult<()> {
        let now = chrono::Utc::now().to_rfc3339();
        let result =
            sqlx::query("UPDATE notices SET deleted_at = ? WHERE id = ? AND deleted_at IS NULL")
                .bind(&now)
                .bind(id)
                .execute(&self.pool)
                .await?;

        if result.rows_affected() == 0 {
            return Err(AppError::NotFound("公告不存在".to_string()));
        }
        Ok(())
    }
}
