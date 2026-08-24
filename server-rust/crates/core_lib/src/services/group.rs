//! 群组服务

use crate::db::DbPool;

use crate::error::{AppError, AppResult};
use crate::models::group::{CreateGroupRequest, Group, UpdateGroupRequest};

#[derive(Clone)]
pub struct GroupService {
    pool: DbPool,
}

impl GroupService {
    pub fn new(pool: DbPool) -> Self {
        Self { pool }
    }

    /// 列出群组
    pub async fn list(&self, page: i32, per_page: i32) -> AppResult<(Vec<Group>, i64)> {
        let page = if page < 1 { 1 } else { page };
        let per_page = if !(1..=100).contains(&per_page) {
            20
        } else {
            per_page
        };
        let offset = (page - 1) * per_page;

        let rows = sqlx::query_as(
            "SELECT * FROM `groups` WHERE deleted_at IS NULL ORDER BY id ASC LIMIT ? OFFSET ?",
        )
        .bind(per_page)
        .bind(offset)
        .fetch_all(&self.pool)
        .await?;

        let total: i64 =
            sqlx::query_scalar("SELECT COUNT(*) FROM `groups` WHERE deleted_at IS NULL")
                .fetch_one(&self.pool)
                .await?;

        Ok((rows, total))
    }

    /// 获取群组详情
    pub async fn get(&self, id: i64) -> AppResult<Group> {
        let group: Option<Group> =
            sqlx::query_as("SELECT * FROM `groups` WHERE id = ? AND deleted_at IS NULL")
                .bind(id)
                .fetch_optional(&self.pool)
                .await?;

        group.ok_or_else(|| AppError::NotFound("群组不存在".to_string()))
    }

    /// 创建群组
    pub async fn create(&self, req: &CreateGroupRequest) -> AppResult<Group> {
        let now = crate::db::now_str();
        let intro = req.intro.as_deref().unwrap_or("");
        let options = req
            .options
            .as_ref()
            .map(|v| serde_json::to_string(v).unwrap_or_default());
        let is_default = req.is_default.unwrap_or(0);
        let is_guest = req.is_guest.unwrap_or(0);

        let result = sqlx::query(
            r#"
            INSERT INTO `groups` (name, intro, options, is_default, is_guest, max_storage, created_at, updated_at)
            VALUES (?, ?, ?, ?, ?, ?, ?, ?)
            "#,
        )
        .bind(&req.name)
        .bind(intro)
        .bind(&options)
        .bind(is_default)
        .bind(is_guest)
        .bind(req.max_storage)
        .bind(&now)
        .bind(&now)
        .execute(&self.pool)
        .await?;

        let id = crate::db::last_id(&result);
        self.get(id).await
    }

    /// 更新群组
    pub async fn update(&self, id: i64, req: &UpdateGroupRequest) -> AppResult<Group> {
        let existing = self.get(id).await?;
        let now = crate::db::now_str();

        let name = req.name.clone().unwrap_or(existing.name);
        let intro = req.intro.clone().unwrap_or(existing.intro);
        let options = req
            .options
            .as_ref()
            .map(|v| serde_json::to_string(v).unwrap_or_default());
        let is_default = req.is_default.unwrap_or(existing.is_default);
        let is_guest = req.is_guest.unwrap_or(existing.is_guest);
        // 字段缺省 = 保持不变；显式 null = 清除配额（不限）
        let max_storage = req.max_storage.unwrap_or(existing.max_storage);

        sqlx::query(
            "UPDATE `groups` SET name = ?, intro = ?, options = ?, is_default = ?, is_guest = ?, max_storage = ?, updated_at = ? WHERE id = ?",
        )
        .bind(&name)
        .bind(&intro)
        .bind(&options)
        .bind(is_default)
        .bind(is_guest)
        .bind(max_storage)
        .bind(&now)
        .bind(id)
        .execute(&self.pool)
        .await?;

        self.get(id).await
    }

    /// 删除群组
    pub async fn delete(&self, id: i64) -> AppResult<()> {
        let now = crate::db::now_str();
        let result =
            sqlx::query("UPDATE `groups` SET deleted_at = ? WHERE id = ? AND deleted_at IS NULL")
                .bind(&now)
                .bind(id)
                .execute(&self.pool)
                .await?;

        if result.rows_affected() == 0 {
            return Err(AppError::NotFound("群组不存在".to_string()));
        }
        Ok(())
    }
}
