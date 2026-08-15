//! 存储服务

use sqlx::SqlitePool;

use crate::error::AppResult;

use crate::dto::storage::StorageSignResponse;

#[derive(Clone)]
pub struct StorageService {
    pool: SqlitePool,
    public_url_prefix: String,
    upload_root: String,
}

impl StorageService {
    pub fn new(pool: SqlitePool, public_url_prefix: String, upload_root: String) -> Self {
        Self {
            pool,
            public_url_prefix,
            upload_root,
        }
    }

    /// 获取上传签名（简化实现：返回本地存储的配置信息）
    pub async fn sign(&self) -> AppResult<StorageSignResponse> {
        // 从数据库读取默认存储配置（如有）
        let _row: Option<(String,)> = sqlx::query_as("SELECT provider FROM storages LIMIT 1")
            .fetch_optional(&self.pool)
            .await?;

        Ok(StorageSignResponse {
            upload_url: format!("{}/uploads", self.public_url_prefix.trim_end_matches('/')),
            public_url: self.public_url_prefix.clone(),
            driver: "local".to_string(),
            max_size: 10485760, // 10MB
            allowed_types: vec![
                "image/jpeg".to_string(),
                "image/png".to_string(),
                "image/gif".to_string(),
                "image/webp".to_string(),
                "image/svg+xml".to_string(),
            ],
        })
    }
}
