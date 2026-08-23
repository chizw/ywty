//! 标签服务

use chrono::Utc;
use sqlx::SqlitePool;

use crate::error::{AppError, AppResult};
use crate::models::photo::Tag;

#[derive(Clone)]
pub struct TagService {
    pool: SqlitePool,
}

impl TagService {
    pub fn new(pool: SqlitePool) -> Self {
        Self { pool }
    }

    /// 生成 slug（简单 URL-safe 名称）
    fn make_slug(name: &str) -> String {
        name.trim()
            .to_lowercase()
            .replace(' ', "-")
            .replace(|c: char| !c.is_alphanumeric() && c != '-', "")
    }

    /// 创建标签（同名复用）
    pub async fn create(&self, name: &str) -> AppResult<Tag> {
        let name = name.trim();
        if name.is_empty() {
            return Err(AppError::Validation("标签名不能为空".to_string()));
        }

        // 查找是否已存在
        let existing: Option<Tag> = sqlx::query_as("SELECT * FROM tags WHERE name = ?")
            .bind(name)
            .fetch_optional(&self.pool)
            .await?;

        if let Some(tag) = existing {
            return Ok(tag);
        }

        let slug = Self::make_slug(name);
        let now = Utc::now().to_rfc3339();

        let result = sqlx::query("INSERT INTO tags (name, slug, created_at) VALUES (?, ?, ?)")
            .bind(name)
            .bind(&slug)
            .bind(&now)
            .execute(&self.pool)
            .await?;

        let id = result.last_insert_rowid();

        Ok(Tag {
            id,
            name: name.to_string(),
            slug,
            photo_count: 0,
            created_at: Utc::now(),
        })
    }

    /// 列出标签
    pub async fn list(&self) -> AppResult<Vec<Tag>> {
        let rows = sqlx::query_as::<_, Tag>("SELECT * FROM tags ORDER BY id DESC LIMIT 200")
            .fetch_all(&self.pool)
            .await?;

        Ok(rows)
    }

    /// 删除标签
    pub async fn delete(&self, id: i64) -> AppResult<()> {
        let result = sqlx::query("DELETE FROM tags WHERE id = ?")
            .bind(id)
            .execute(&self.pool)
            .await?;

        if result.rows_affected() == 0 {
            return Err(AppError::NotFound("标签不存在".to_string()));
        }

        // 删除关联
        sqlx::query("DELETE FROM photo_tags WHERE tag_id = ?")
            .bind(id)
            .execute(&self.pool)
            .await?;

        Ok(())
    }

    /// 校验图片归属（防止跨用户读写他人图片的标签）
    async fn check_photo_ownership(&self, user_id: i64, photo_id: i64) -> AppResult<()> {
        let row: Option<(i64,)> = sqlx::query_as(
            "SELECT id FROM photos WHERE id = ? AND user_id = ? AND deleted_at IS NULL",
        )
        .bind(photo_id)
        .bind(user_id)
        .fetch_optional(&self.pool)
        .await?;

        if row.is_none() {
            return Err(AppError::NotFound("图片不存在".to_string()));
        }
        Ok(())
    }

    /// 绑定标签到图片（仅限自己的图片）
    pub async fn attach(&self, user_id: i64, tag_id: i64, photo_id: i64) -> AppResult<()> {
        self.check_photo_ownership(user_id, photo_id).await?;

        // 使用 INSERT OR IGNORE 避免重复
        let existing: Option<(i64, i64)> = sqlx::query_as(
            "SELECT photo_id, tag_id FROM photo_tags WHERE photo_id = ? AND tag_id = ?",
        )
        .bind(photo_id)
        .bind(tag_id)
        .fetch_optional(&self.pool)
        .await?;

        if existing.is_some() {
            return Ok(()); // 已存在，幂等
        }

        sqlx::query("INSERT INTO photo_tags (photo_id, tag_id) VALUES (?, ?)")
            .bind(photo_id)
            .bind(tag_id)
            .execute(&self.pool)
            .await?;

        Ok(())
    }

    /// 解绑标签（仅限自己的图片）
    pub async fn detach(&self, user_id: i64, tag_id: i64, photo_id: i64) -> AppResult<()> {
        self.check_photo_ownership(user_id, photo_id).await?;

        sqlx::query("DELETE FROM photo_tags WHERE tag_id = ? AND photo_id = ?")
            .bind(tag_id)
            .bind(photo_id)
            .execute(&self.pool)
            .await?;

        Ok(())
    }

    /// 按名称批量绑定（不存在则自动创建，仅限自己的图片）
    pub async fn attach_by_names(
        &self,
        user_id: i64,
        names: &[String],
        photo_id: i64,
    ) -> AppResult<Vec<Tag>> {
        self.check_photo_ownership(user_id, photo_id).await?;

        let mut result = Vec::new();
        let mut seen = std::collections::HashSet::new();

        for name in names {
            let name = name.trim();
            if name.is_empty() || seen.contains(name) {
                continue;
            }
            seen.insert(name.to_string());

            let tag = self.create(name).await?;
            self.attach(user_id, tag.id, photo_id).await?;
            result.push(tag);
        }

        Ok(result)
    }

    /// 查询图片绑定的标签（仅限自己的图片）
    pub async fn list_for_target(&self, user_id: i64, photo_id: i64) -> AppResult<Vec<Tag>> {
        self.check_photo_ownership(user_id, photo_id).await?;

        let rows = sqlx::query_as::<_, Tag>(
            r#"
            SELECT t.* FROM tags t
            JOIN photo_tags pt ON pt.tag_id = t.id
            WHERE pt.photo_id = ?
            ORDER BY t.id DESC
            "#,
        )
        .bind(photo_id)
        .fetch_all(&self.pool)
        .await?;

        Ok(rows)
    }
}
