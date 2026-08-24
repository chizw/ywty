//! 相册服务

use crate::db::DbPool;
use uuid::Uuid;

use crate::error::{AppError, AppResult};

use crate::dto::album::{AlbumResponse, CreateAlbumRequest, UpdateAlbumRequest};
use crate::dto::photo::PhotoResponse;

#[derive(Clone)]
pub struct AlbumService {
    pool: DbPool,
}

impl AlbumService {
    pub fn new(pool: DbPool) -> Self {
        Self { pool }
    }

    /// 获取相册列表
    pub async fn list(
        &self,
        user_id: i64,
        page: u64,
        per_page: u64,
    ) -> AppResult<(Vec<AlbumResponse>, i64)> {
        let offset = (page - 1) * per_page;

        let total: i64 = sqlx::query_scalar(
            "SELECT COUNT(*) FROM albums WHERE user_id = ? AND deleted_at IS NULL",
        )
        .bind(user_id)
        .fetch_one(&self.pool)
        .await?;

        let rows: Vec<AlbumResponse> = sqlx::query_as(
            r#"
            SELECT id, uuid, user_id, name, description, cover_photo_id, is_public,
                   photo_count, views, created_at, updated_at
            FROM albums WHERE user_id = ? AND deleted_at IS NULL
            ORDER BY created_at DESC LIMIT ? OFFSET ?
            "#,
        )
        .bind(user_id)
        .bind(per_page as i64)
        .bind(offset as i64)
        .fetch_all(&self.pool)
        .await?;

        Ok((rows, total))
    }

    /// 获取相册详情
    pub async fn get(&self, user_id: i64, album_id: i64) -> AppResult<AlbumResponse> {
        sqlx::query_as::<_, AlbumResponse>(
            r#"
            SELECT id, uuid, user_id, name, description, cover_photo_id, is_public,
                   photo_count, views, created_at, updated_at
            FROM albums WHERE id = ? AND user_id = ? AND deleted_at IS NULL
            "#,
        )
        .bind(album_id)
        .bind(user_id)
        .fetch_optional(&self.pool)
        .await?
        .ok_or_else(|| AppError::NotFound("相册不存在".to_string()))
    }

    /// 创建相册
    pub async fn create(&self, user_id: i64, req: &CreateAlbumRequest) -> AppResult<AlbumResponse> {
        let uuid = Uuid::new_v4().to_string();
        let is_public = req.is_public.unwrap_or(false);
        let now = crate::db::now_str();

        let result = sqlx::query(
            r#"
            INSERT INTO albums (uuid, user_id, name, description, is_public, photo_count, views, created_at, updated_at)
            VALUES (?, ?, ?, ?, ?, 0, 0, ?, ?)
            "#,
        )
        .bind(&uuid)
        .bind(user_id)
        .bind(&req.name)
        .bind(&req.description)
        .bind(is_public)
        .bind(&now)
        .bind(&now)
        .execute(&self.pool)
        .await?;

        let id = crate::db::last_id(&result);

        self.get(user_id, id).await
    }

    /// 更新相册
    pub async fn update(
        &self,
        user_id: i64,
        album_id: i64,
        req: &UpdateAlbumRequest,
    ) -> AppResult<AlbumResponse> {
        self.check_ownership(user_id, album_id).await?;

        let now = crate::db::now_str();

        if let Some(name) = &req.name {
            sqlx::query("UPDATE albums SET name = ?, updated_at = ? WHERE id = ?")
                .bind(name)
                .bind(&now)
                .bind(album_id)
                .execute(&self.pool)
                .await?;
        }

        if let Some(desc) = &req.description {
            sqlx::query("UPDATE albums SET description = ?, updated_at = ? WHERE id = ?")
                .bind(desc)
                .bind(&now)
                .bind(album_id)
                .execute(&self.pool)
                .await?;
        }

        if let Some(cover) = req.cover_photo_id {
            sqlx::query("UPDATE albums SET cover_photo_id = ?, updated_at = ? WHERE id = ?")
                .bind(cover)
                .bind(&now)
                .bind(album_id)
                .execute(&self.pool)
                .await?;
        }

        if let Some(ip) = req.is_public {
            sqlx::query("UPDATE albums SET is_public = ?, updated_at = ? WHERE id = ?")
                .bind(ip)
                .bind(&now)
                .bind(album_id)
                .execute(&self.pool)
                .await?;
        }

        self.get(user_id, album_id).await
    }

    /// 删除相册
    pub async fn delete(&self, user_id: i64, album_id: i64) -> AppResult<()> {
        self.check_ownership(user_id, album_id).await?;

        let now = crate::db::now_str();

        // 软删除相册
        sqlx::query("UPDATE albums SET deleted_at = ? WHERE id = ?")
            .bind(&now)
            .bind(album_id)
            .execute(&self.pool)
            .await?;

        // 解除相册内图片的关联
        sqlx::query("UPDATE photos SET album_id = NULL WHERE album_id = ?")
            .bind(album_id)
            .execute(&self.pool)
            .await?;

        // 删除相册-图片关联
        sqlx::query("DELETE FROM album_photos WHERE album_id = ?")
            .bind(album_id)
            .execute(&self.pool)
            .await?;

        Ok(())
    }

    /// 获取相册内的图片
    pub async fn list_photos(
        &self,
        user_id: i64,
        album_id: i64,
        page: u64,
        per_page: u64,
    ) -> AppResult<(Vec<PhotoResponse>, i64)> {
        self.check_ownership(user_id, album_id).await?;

        let offset = (page - 1) * per_page;

        let total: i64 = sqlx::query_scalar(
            "SELECT COUNT(*) FROM photos WHERE album_id = ? AND deleted_at IS NULL",
        )
        .bind(album_id)
        .fetch_one(&self.pool)
        .await?;

        let rows: Vec<PhotoResponse> = sqlx::query_as(
            r#"
            SELECT id, uuid, user_id, album_id, filename, original_name, url, thumbnail_url,
                   size, width, height, mime_type, exif, is_public, views, likes, status,
                   expired_at, created_at, updated_at
            FROM photos WHERE album_id = ? AND deleted_at IS NULL
            ORDER BY created_at DESC LIMIT ? OFFSET ?
            "#,
        )
        .bind(album_id)
        .bind(per_page as i64)
        .bind(offset as i64)
        .fetch_all(&self.pool)
        .await?;

        Ok((rows, total))
    }

    /// 添加图片到相册
    pub async fn add_photos(
        &self,
        user_id: i64,
        album_id: i64,
        photo_ids: &[i64],
    ) -> AppResult<u64> {
        self.check_ownership(user_id, album_id).await?;

        let now = crate::db::now_str();
        let mut added = 0u64;

        for photo_id in photo_ids {
            // 确认图片所有权
            let photo_exists: Option<(i64,)> = sqlx::query_as(
                "SELECT id FROM photos WHERE id = ? AND user_id = ? AND deleted_at IS NULL",
            )
            .bind(photo_id)
            .bind(user_id)
            .fetch_optional(&self.pool)
            .await?;

            if photo_exists.is_none() {
                continue;
            }

            // 插入关联（忽略已存在）
            let sql = crate::db::sql_insert_ignore(
                "INSERT OR IGNORE INTO album_photos (album_id, photo_id, sort_order, created_at) VALUES (?, ?, 0, ?)",
            );
            let result = sqlx::query(&sql)
                .bind(album_id)
                .bind(photo_id)
                .bind(&now)
                .execute(&self.pool)
                .await?;

            if result.rows_affected() > 0 {
                added += 1;
                // 更新图片的 album_id
                sqlx::query("UPDATE photos SET album_id = ? WHERE id = ?")
                    .bind(album_id)
                    .bind(photo_id)
                    .execute(&self.pool)
                    .await?;
            }
        }

        // 更新相册图片计数
        if added > 0 {
            sqlx::query("UPDATE albums SET photo_count = photo_count + ? WHERE id = ?")
                .bind(added as i64)
                .bind(album_id)
                .execute(&self.pool)
                .await?;
        }

        Ok(added)
    }

    /// 从相册移除图片
    pub async fn remove_photo(&self, user_id: i64, album_id: i64, photo_id: i64) -> AppResult<()> {
        self.check_ownership(user_id, album_id).await?;

        // 删除关联
        let result = sqlx::query("DELETE FROM album_photos WHERE album_id = ? AND photo_id = ?")
            .bind(album_id)
            .bind(photo_id)
            .execute(&self.pool)
            .await?;

        if result.rows_affected() > 0 {
            // 清除图片的 album_id
            sqlx::query("UPDATE photos SET album_id = NULL WHERE id = ? AND album_id = ?")
                .bind(photo_id)
                .bind(album_id)
                .execute(&self.pool)
                .await?;

            // 更新计数
            sqlx::query("UPDATE albums SET photo_count = MAX(0, photo_count - 1) WHERE id = ?")
                .bind(album_id)
                .execute(&self.pool)
                .await?;
        }

        Ok(())
    }

    /// 检查相册所有权
    async fn check_ownership(&self, user_id: i64, album_id: i64) -> AppResult<()> {
        let row: Option<(i64,)> = sqlx::query_as(
            "SELECT id FROM albums WHERE id = ? AND user_id = ? AND deleted_at IS NULL",
        )
        .bind(album_id)
        .bind(user_id)
        .fetch_optional(&self.pool)
        .await?;

        if row.is_none() {
            return Err(AppError::NotFound("相册不存在".to_string()));
        }

        Ok(())
    }
}
