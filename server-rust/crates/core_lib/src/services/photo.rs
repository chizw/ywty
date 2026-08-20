//! 图片服务

use chrono::Utc;
use sqlx::SqlitePool;
use uuid::Uuid;

use crate::error::{AppError, AppResult};

use crate::dto::photo::{
    BatchUpdateRequest, PhotoPublicResponse, PhotoResponse, UploadResponse,
};

#[derive(Clone)]
pub struct PhotoService {
    pool: SqlitePool,
    public_url_prefix: String,
    upload_root: String,
}

impl PhotoService {
    pub fn new(pool: SqlitePool, public_url_prefix: String, upload_root: String) -> Self {
        Self {
            pool,
            public_url_prefix,
            upload_root,
        }
    }

    /// 获取图片列表（分页）
    pub async fn list(
        &self,
        user_id: i64,
        page: u64,
        per_page: u64,
        album_id: Option<i64>,
    ) -> AppResult<(Vec<PhotoResponse>, i64)> {
        let offset = (page - 1) * per_page;

        // 构建查询条件
        let where_clause = match album_id {
            Some(_) => "WHERE user_id = ? AND album_id IS ? AND deleted_at IS NULL",
            None => "WHERE user_id = ? AND deleted_at IS NULL",
        };

        // 计数
        let count_query = format!("SELECT COUNT(*) as cnt FROM photos {}", where_clause);
        let total: i64 = if let Some(aid) = album_id {
            sqlx::query_scalar(&count_query)
                .bind(user_id)
                .bind(aid)
                .fetch_one(&self.pool)
                .await?
        } else {
            sqlx::query_scalar(&count_query)
                .bind(user_id)
                .fetch_one(&self.pool)
                .await?
        };

        // 查询列表（注意：不含敏感字段 path/md5/sha1，与 PhotoResponse 对齐）
        let list_query = format!(
            "SELECT id, uuid, user_id, album_id, filename, original_name, url, thumbnail_url, \
             size, width, height, mime_type, exif, is_public, views, likes, status, \
             expired_at, created_at, updated_at \
             FROM photos {} ORDER BY created_at DESC LIMIT ? OFFSET ?",
            where_clause
        );

        let rows: Vec<PhotoResponse> = if let Some(aid) = album_id {
            sqlx::query_as(&list_query)
                .bind(user_id)
                .bind(aid)
                .bind(per_page as i64)
                .bind(offset as i64)
                .fetch_all(&self.pool)
                .await?
        } else {
            sqlx::query_as(&list_query)
                .bind(user_id)
                .bind(per_page as i64)
                .bind(offset as i64)
                .fetch_all(&self.pool)
                .await?
        };

        Ok((rows, total))
    }

    /// 获取单张图片
    pub async fn get(&self, user_id: i64, photo_id: i64) -> AppResult<PhotoResponse> {
        sqlx::query_as::<_, PhotoResponse>(
            r#"
            SELECT id, uuid, user_id, album_id, filename, original_name, url, thumbnail_url,
                   size, width, height, mime_type, exif, is_public, views, likes, status,
                   expired_at, created_at, updated_at
            FROM photos WHERE id = ? AND user_id = ? AND deleted_at IS NULL
            "#,
        )
        .bind(photo_id)
        .bind(user_id)
        .fetch_optional(&self.pool)
        .await?
        .ok_or_else(|| AppError::NotFound("图片不存在".to_string()))
    }

    /// 创建图片记录（上传后由 handler 调用）
    pub async fn create(
        &self,
        user_id: i64,
        filename: &str,
        original_name: &str,
        path: &str,
        size: i64,
        mime_type: &str,
        width: Option<i32>,
        height: Option<i32>,
        md5: Option<&str>,
        album_id: Option<i64>,
    ) -> AppResult<UploadResponse> {
        self.create_with_thumbnail(user_id, filename, original_name, path, size, mime_type, width, height, md5, album_id, None).await
    }

    /// 创建图片记录（带缩略图）
    pub async fn create_with_thumbnail(
        &self,
        user_id: i64,
        filename: &str,
        original_name: &str,
        path: &str,
        size: i64,
        mime_type: &str,
        width: Option<i32>,
        height: Option<i32>,
        md5: Option<&str>,
        album_id: Option<i64>,
        thumbnail_url: Option<&str>,
    ) -> AppResult<UploadResponse> {
        let uuid = Uuid::new_v4().to_string();
        let url = format!("{}/{}", self.public_url_prefix.trim_end_matches('/'), path);
        let now = Utc::now().to_rfc3339();

        let result = sqlx::query(
            r#"
            INSERT INTO photos (uuid, user_id, album_id, filename, original_name, path, url,
                                thumbnail_url, size, width, height, mime_type, md5, is_public, status, created_at, updated_at)
            VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0, 1, ?, ?)
            "#,
        )
        .bind(&uuid)
        .bind(user_id)
        .bind(album_id)
        .bind(filename)
        .bind(original_name)
        .bind(path)
        .bind(&url)
        .bind(thumbnail_url)
        .bind(size)
        .bind(width)
        .bind(height)
        .bind(mime_type)
        .bind(md5)
        .bind(&now)
        .bind(&now)
        .execute(&self.pool)
        .await?;

        let id = result.last_insert_rowid();

        Ok(UploadResponse {
            id,
            uuid,
            url,
            thumbnail_url: thumbnail_url.map(String::from),
            size,
            width,
            height,
            filename: filename.to_string(),
        })
    }

    /// 更新图片
    pub async fn update(
        &self,
        user_id: i64,
        photo_id: i64,
        album_id: Option<i64>,
        is_public: Option<bool>,
    ) -> AppResult<PhotoResponse> {
        // 确认所有权
        self.check_ownership(user_id, photo_id).await?;

        let now = Utc::now().to_rfc3339();

        if let Some(aid) = album_id {
            sqlx::query("UPDATE photos SET album_id = ?, updated_at = ? WHERE id = ?")
                .bind(aid)
                .bind(&now)
                .bind(photo_id)
                .execute(&self.pool)
                .await?;
        }

        if let Some(ip) = is_public {
            sqlx::query("UPDATE photos SET is_public = ?, updated_at = ? WHERE id = ?")
                .bind(ip)
                .bind(&now)
                .bind(photo_id)
                .execute(&self.pool)
                .await?;
        }

        self.get(user_id, photo_id).await
    }

    /// 删除图片（软删除）
    pub async fn delete(&self, user_id: i64, photo_id: i64) -> AppResult<()> {
        self.check_ownership(user_id, photo_id).await?;

        let now = Utc::now().to_rfc3339();
        sqlx::query("UPDATE photos SET deleted_at = ? WHERE id = ?")
            .bind(&now)
            .bind(photo_id)
            .execute(&self.pool)
            .await?;

        Ok(())
    }

    /// 批量删除
    pub async fn batch_delete(&self, user_id: i64, ids: &[i64]) -> AppResult<u64> {
        let now = Utc::now().to_rfc3339();
        let mut affected = 0u64;

        for id in ids {
            let result = sqlx::query(
                "UPDATE photos SET deleted_at = ? WHERE id = ? AND user_id = ? AND deleted_at IS NULL",
            )
            .bind(&now)
            .bind(id)
            .bind(user_id)
            .execute(&self.pool)
            .await?;
            affected += result.rows_affected();
        }

        Ok(affected)
    }

    /// 批量更新
    pub async fn batch_update(
        &self,
        user_id: i64,
        req: &BatchUpdateRequest,
    ) -> AppResult<u64> {
        let now = Utc::now().to_rfc3339();
        let mut affected = 0u64;

        for id in &req.ids {
            if let Some(aid) = req.album_id {
                let result = sqlx::query(
                    "UPDATE photos SET album_id = ?, updated_at = ? WHERE id = ? AND user_id = ? AND deleted_at IS NULL",
                )
                .bind(aid)
                .bind(&now)
                .bind(id)
                .bind(user_id)
                .execute(&self.pool)
                .await?;
                affected += result.rows_affected();
            } else if let Some(ip) = req.is_public {
                let result = sqlx::query(
                    "UPDATE photos SET is_public = ?, updated_at = ? WHERE id = ? AND user_id = ? AND deleted_at IS NULL",
                )
                .bind(ip)
                .bind(&now)
                .bind(id)
                .bind(user_id)
                .execute(&self.pool)
                .await?;
                affected += result.rows_affected();
            }
        }

        Ok(affected)
    }

    /// 移动到相册
    pub async fn move_to_album(
        &self,
        user_id: i64,
        photo_id: i64,
        album_id: i64,
    ) -> AppResult<PhotoResponse> {
        self.check_ownership(user_id, photo_id).await?;

        // 确认相册所有权
        let album_exists: Option<(i64,)> = sqlx::query_as(
            "SELECT id FROM albums WHERE id = ? AND user_id = ? AND deleted_at IS NULL",
        )
        .bind(album_id)
        .bind(user_id)
        .fetch_optional(&self.pool)
        .await?;

        if album_exists.is_none() {
            return Err(AppError::NotFound("相册不存在".to_string()));
        }

        let now = Utc::now().to_rfc3339();
        sqlx::query("UPDATE photos SET album_id = ?, updated_at = ? WHERE id = ?")
            .bind(album_id)
            .bind(&now)
            .bind(photo_id)
            .execute(&self.pool)
            .await?;

        self.get(user_id, photo_id).await
    }

    /// 复制图片
    pub async fn copy(
        &self,
        user_id: i64,
        photo_id: i64,
        album_id: Option<i64>,
    ) -> AppResult<UploadResponse> {
        // 获取原图
        let original = self.get(user_id, photo_id).await?;

        // PhotoResponse 出于安全考虑不含 md5，此处单独查询用于复制
        let original_md5: Option<String> = sqlx::query_scalar(
            "SELECT md5 FROM photos WHERE id = ? AND user_id = ? AND deleted_at IS NULL",
        )
        .bind(photo_id)
        .bind(user_id)
        .fetch_optional(&self.pool)
        .await?
        .flatten();

        let uuid = Uuid::new_v4().to_string();
        let new_filename = format!("copy_{}", &original.filename);
        let path = format!("copies/{}", &new_filename);
        let url = format!("{}/{}", self.public_url_prefix.trim_end_matches('/'), &path);
        let now = Utc::now().to_rfc3339();

        let result = sqlx::query(
            r#"
            INSERT INTO photos (uuid, user_id, album_id, filename, original_name, path, url,
                                size, width, height, mime_type, md5, is_public, status, created_at, updated_at)
            VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0, 1, ?, ?)
            "#,
        )
        .bind(&uuid)
        .bind(user_id)
        .bind(album_id)
        .bind(&new_filename)
        .bind(&original.original_name)
        .bind(&path)
        .bind(&url)
        .bind(original.size)
        .bind(original.width)
        .bind(original.height)
        .bind(&original.mime_type)
        .bind(&original_md5)
        .bind(&now)
        .bind(&now)
        .execute(&self.pool)
        .await?;

        Ok(UploadResponse {
            id: result.last_insert_rowid(),
            uuid,
            url,
            thumbnail_url: original.thumbnail_url,
            size: original.size,
            width: original.width,
            height: original.height,
            filename: new_filename,
        })
    }

    /// 公开探索列表
    pub async fn list_public(&self, page: u64, per_page: u64) -> AppResult<(Vec<PhotoPublicResponse>, i64)> {
        let offset = (page - 1) * per_page;

        let total: i64 = sqlx::query_scalar(
            "SELECT COUNT(*) FROM photos WHERE is_public = 1 AND status = 1 AND deleted_at IS NULL",
        )
        .fetch_one(&self.pool)
        .await?;

        let rows: Vec<PhotoPublicResponse> = sqlx::query_as(
            r#"
            SELECT p.id, p.uuid, u.username, p.url, p.thumbnail_url,
                   p.width, p.height, p.size, p.views, p.likes, p.created_at
            FROM photos p
            JOIN users u ON p.user_id = u.id
            WHERE p.is_public = 1 AND p.status = 1 AND p.deleted_at IS NULL
            ORDER BY p.created_at DESC
            LIMIT ? OFFSET ?
            "#,
        )
        .bind(per_page as i64)
        .bind(offset as i64)
        .fetch_all(&self.pool)
        .await?;

        Ok((rows, total))
    }

    /// 检查图片所有权
    async fn check_ownership(&self, user_id: i64, photo_id: i64) -> AppResult<()> {
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
}
