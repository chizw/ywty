//! 存储管理服务

use crate::db::DbPool;

use crate::error::{AppError, AppResult};

#[derive(Clone)]
pub struct StorageAdminService {
    pool: DbPool,
}

/// 存储记录（对应 storages 表的查询行）
#[derive(Debug, Clone, serde::Serialize, sqlx::FromRow)]
pub struct StorageRecord {
    pub id: i64,
    pub name: String,
    pub provider: String,
    pub intro: Option<String>,
    pub prefix: Option<String>,
    pub options: Option<String>,
    pub created_at: String,
    pub updated_at: String,
}

impl StorageAdminService {
    pub fn new(pool: DbPool) -> Self {
        Self { pool }
    }

    /// 列出存储
    pub async fn list_storages(&self) -> AppResult<Vec<StorageRecord>> {
        let rows = sqlx::query_as::<_, StorageRecord>(
            "SELECT id, name, provider, intro, prefix, options, created_at, updated_at FROM storages ORDER BY id ASC",
        )
        .fetch_all(&self.pool)
        .await?;

        Ok(rows)
    }

    /// 创建存储
    pub async fn create_storage(
        &self,
        name: &str,
        provider: &str,
        intro: Option<&str>,
        prefix: Option<&str>,
        options: Option<&str>,
    ) -> AppResult<StorageRecord> {
        let now = crate::db::now_str();

        let result = sqlx::query(
            r#"
            INSERT INTO storages (name, provider, intro, prefix, options, created_at, updated_at)
            VALUES (?, ?, ?, ?, ?, ?, ?)
            "#,
        )
        .bind(name)
        .bind(provider)
        .bind(intro)
        .bind(prefix)
        .bind(options)
        .bind(&now)
        .bind(&now)
        .execute(&self.pool)
        .await?;

        let id = crate::db::last_id(&result);
        self.get_storage(id).await
    }

    /// 获取存储详情
    pub async fn get_storage(&self, id: i64) -> AppResult<StorageRecord> {
        let row: Option<StorageRecord> = sqlx::query_as(
            "SELECT id, name, provider, intro, prefix, options, created_at, updated_at FROM storages WHERE id = ?",
        )
        .bind(id)
        .fetch_optional(&self.pool)
        .await?;

        row.ok_or_else(|| AppError::NotFound("存储不存在".to_string()))
    }

    /// 更新存储
    pub async fn update_storage(
        &self,
        id: i64,
        name: Option<&str>,
        intro: Option<&str>,
        prefix: Option<&str>,
        options: Option<&str>,
    ) -> AppResult<()> {
        let mut updates: Vec<&str> = Vec::new();
        if name.is_some() {
            updates.push("name = ?");
        }
        if intro.is_some() {
            updates.push("intro = ?");
        }
        if prefix.is_some() {
            updates.push("prefix = ?");
        }
        if options.is_some() {
            updates.push("options = ?");
        }

        if updates.is_empty() {
            return Err(AppError::Validation("没有要更新的字段".to_string()));
        }

        let now = crate::db::now_str();
        let sql = format!(
            "UPDATE storages SET {}, updated_at = ? WHERE id = ?",
            updates.join(", ")
        );

        let mut q = sqlx::query(&sql);
        if let Some(v) = name {
            q = q.bind(v);
        }
        if let Some(v) = intro {
            q = q.bind(v);
        }
        if let Some(v) = prefix {
            q = q.bind(v);
        }
        if let Some(v) = options {
            q = q.bind(v);
        }
        let result = q.bind(&now).bind(id).execute(&self.pool).await?;

        if result.rows_affected() == 0 {
            return Err(AppError::NotFound("存储不存在".to_string()));
        }
        Ok(())
    }

    /// 删除存储
    pub async fn delete_storage(&self, id: i64) -> AppResult<()> {
        let result = sqlx::query("DELETE FROM storages WHERE id = ?")
            .bind(id)
            .execute(&self.pool)
            .await?;

        if result.rows_affected() == 0 {
            return Err(AppError::NotFound("存储不存在".to_string()));
        }
        Ok(())
    }
}
