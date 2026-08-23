//! 单页服务

use sqlx::SqlitePool;

use crate::error::{AppError, AppResult};
use crate::models::page::{CreatePageRequest, Page, UpdatePageRequest};

#[derive(Clone)]
pub struct PageService {
    pool: SqlitePool,
}

impl PageService {
    pub fn new(pool: SqlitePool) -> Self {
        Self { pool }
    }

    /// 列出公开页面
    pub async fn list_public(&self) -> AppResult<Vec<Page>> {
        let rows = sqlx::query_as(
            "SELECT * FROM pages WHERE deleted_at IS NULL AND is_show = 1 ORDER BY sort ASC",
        )
        .fetch_all(&self.pool)
        .await?;

        Ok(rows)
    }

    /// 通过 slug 获取页面
    pub async fn get_by_slug(&self, slug: &str) -> AppResult<Page> {
        let page: Option<Page> = sqlx::query_as(
            "SELECT * FROM pages WHERE slug = ? AND deleted_at IS NULL AND is_show = 1",
        )
        .bind(slug)
        .fetch_optional(&self.pool)
        .await?;

        let page = page.ok_or_else(|| AppError::NotFound("页面不存在".to_string()))?;

        // 增加浏览计数
        sqlx::query("UPDATE pages SET view_count = view_count + 1 WHERE id = ?")
            .bind(page.id)
            .execute(&self.pool)
            .await?;

        Ok(page)
    }

    /// 管理端列表
    pub async fn admin_list(&self, page: i32, per_page: i32) -> AppResult<(Vec<Page>, i64)> {
        let page = if page < 1 { 1 } else { page };
        let per_page = if !(1..=100).contains(&per_page) {
            20
        } else {
            per_page
        };
        let offset = (page - 1) * per_page;

        let rows = sqlx::query_as(
            "SELECT * FROM pages WHERE deleted_at IS NULL ORDER BY sort ASC LIMIT ? OFFSET ?",
        )
        .bind(per_page)
        .bind(offset)
        .fetch_all(&self.pool)
        .await?;

        let total: i64 = sqlx::query_scalar("SELECT COUNT(*) FROM pages WHERE deleted_at IS NULL")
            .fetch_one(&self.pool)
            .await?;

        Ok((rows, total))
    }

    /// 管理端获取详情
    pub async fn admin_get(&self, id: i64) -> AppResult<Page> {
        let page: Option<Page> =
            sqlx::query_as("SELECT * FROM pages WHERE id = ? AND deleted_at IS NULL")
                .bind(id)
                .fetch_optional(&self.pool)
                .await?;

        page.ok_or_else(|| AppError::NotFound("页面不存在".to_string()))
    }

    /// 创建页面
    pub async fn admin_create(&self, req: &CreatePageRequest) -> AppResult<Page> {
        let now = chrono::Utc::now().to_rfc3339();
        let page_type = req.page_type.as_deref().unwrap_or("internal");
        let name = &req.name;
        let icon = req.icon.as_deref().unwrap_or("");
        let title = req.title.as_deref().unwrap_or("");
        let content = req.content.clone().unwrap_or_default();
        let keywords = req.keywords.clone().unwrap_or_default();
        let description = req.description.clone().unwrap_or_default();
        let slug = req.slug.as_deref().unwrap_or("");
        let url = req.url.as_deref().unwrap_or("");
        let sort = req.sort.unwrap_or(0);
        let is_show = req.is_show.unwrap_or(0);

        let result = sqlx::query(
            r#"
            INSERT INTO pages (type, name, icon, title, content, keywords, description, slug, url, view_count, sort, is_show, created_at, updated_at)
            VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 0, ?, ?, ?, ?)
            "#,
        )
        .bind(page_type)
        .bind(name)
        .bind(icon)
        .bind(title)
        .bind(&content)
        .bind(&keywords)
        .bind(&description)
        .bind(slug)
        .bind(url)
        .bind(sort)
        .bind(is_show)
        .bind(&now)
        .bind(&now)
        .execute(&self.pool)
        .await?;

        let id = result.last_insert_rowid();
        self.admin_get(id).await
    }

    /// 更新页面
    pub async fn admin_update(&self, id: i64, req: &UpdatePageRequest) -> AppResult<Page> {
        let existing = self.admin_get(id).await?;
        let now = chrono::Utc::now().to_rfc3339();

        let page_type = req.page_type.clone().unwrap_or(existing.page_type);
        let name = req.name.clone().unwrap_or(existing.name);
        let icon = req.icon.clone().unwrap_or(existing.icon);
        let title = req.title.clone().unwrap_or(existing.title);
        let content = req.content.clone().unwrap_or_default();
        let keywords = req.keywords.clone().unwrap_or_default();
        let description = req.description.clone().unwrap_or_default();
        let slug = req.slug.clone().unwrap_or(existing.slug);
        let url = req.url.clone().unwrap_or(existing.url);
        let sort = req.sort.unwrap_or(existing.sort);
        let is_show = req.is_show.unwrap_or(existing.is_show);

        sqlx::query(
            "UPDATE pages SET type = ?, name = ?, icon = ?, title = ?, content = ?, keywords = ?, description = ?, slug = ?, url = ?, sort = ?, is_show = ?, updated_at = ? WHERE id = ?",
        )
        .bind(&page_type)
        .bind(&name)
        .bind(&icon)
        .bind(&title)
        .bind(&content)
        .bind(&keywords)
        .bind(&description)
        .bind(&slug)
        .bind(&url)
        .bind(sort)
        .bind(is_show)
        .bind(&now)
        .bind(id)
        .execute(&self.pool)
        .await?;

        self.admin_get(id).await
    }

    /// 删除页面
    pub async fn admin_delete(&self, id: i64) -> AppResult<()> {
        let now = chrono::Utc::now().to_rfc3339();
        let result =
            sqlx::query("UPDATE pages SET deleted_at = ? WHERE id = ? AND deleted_at IS NULL")
                .bind(&now)
                .bind(id)
                .execute(&self.pool)
                .await?;

        if result.rows_affected() == 0 {
            return Err(AppError::NotFound("页面不存在".to_string()));
        }
        Ok(())
    }
}
