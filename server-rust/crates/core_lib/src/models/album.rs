use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};
use sqlx::FromRow;
use uuid::Uuid;
use validator::Validate;

/// 相册实体
#[derive(Debug, Clone, FromRow, Serialize, utoipa::ToSchema)]
pub struct Album {
    pub id: i64,
    pub uuid: String,
    pub user_id: i64,
    pub name: String,
    pub description: Option<String>,
    pub cover_photo_id: Option<i64>,
    pub is_public: bool,
    pub photo_count: i64,
    pub views: i64,
    pub created_at: DateTime<Utc>,
    pub updated_at: DateTime<Utc>,
    pub deleted_at: Option<DateTime<Utc>>,
}

/// 创建相册请求
#[derive(Debug, Clone, Deserialize, Validate)]
pub struct CreateAlbumRequest {
    #[validate(length(min = 1, max = 100, message = "相册名称长度必须在 1-100 之间"))]
    pub name: String,
    pub description: Option<String>,
    pub is_public: Option<bool>,
}

/// 更新相册请求
#[derive(Debug, Clone, Deserialize, Validate, Default)]
pub struct UpdateAlbumRequest {
    #[validate(length(min = 1, max = 100, message = "相册名称长度必须在 1-100 之间"))]
    pub name: Option<String>,
    pub description: Option<String>,
    pub cover_photo_id: Option<i64>,
    pub is_public: Option<bool>,
}

/// 相册-图片关联
#[derive(Debug, Clone, FromRow)]
pub struct AlbumPhoto {
    pub album_id: i64,
    pub photo_id: i64,
    pub sort_order: i32,
    pub created_at: DateTime<Utc>,
}

impl Album {
    pub fn new(user_id: i64, name: String) -> Self {
        Self {
            id: 0,
            uuid: Uuid::new_v4().to_string(),
            user_id,
            name,
            description: None,
            cover_photo_id: None,
            is_public: false,
            photo_count: 0,
            views: 0,
            created_at: Utc::now(),
            updated_at: Utc::now(),
            deleted_at: None,
        }
    }
}
