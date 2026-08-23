use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};
use sqlx::FromRow;
use uuid::Uuid;
use validator::Validate;

/// 图片实体
#[derive(Debug, Clone, FromRow, Serialize)]
pub struct Photo {
    pub id: i64,
    pub uuid: String,
    pub user_id: i64,
    pub album_id: Option<i64>,
    pub filename: String,
    pub original_name: String,
    pub path: String,
    pub url: String,
    pub thumbnail_url: Option<String>,
    pub size: i64,
    pub width: Option<i32>,
    pub height: Option<i32>,
    pub mime_type: String,
    pub md5: Option<String>,
    pub sha1: Option<String>,
    pub exif: Option<String>,
    pub is_public: bool,
    pub views: i64,
    pub likes: i64,
    pub status: i32,
    pub expired_at: Option<DateTime<Utc>>,
    pub created_at: DateTime<Utc>,
    pub updated_at: DateTime<Utc>,
    pub deleted_at: Option<DateTime<Utc>>,
}

/// 上传图片请求
#[derive(Debug, Clone, Deserialize, Default, utoipa::ToSchema)]
pub struct UploadPhotoRequest {
    pub album_id: Option<i64>,
    pub is_public: Option<bool>,
    pub expired_at: Option<DateTime<Utc>>,
}

/// 更新图片请求
#[derive(Debug, Clone, Deserialize, Default, utoipa::ToSchema)]
pub struct UpdatePhotoRequest {
    pub album_id: Option<i64>,
    pub is_public: Option<bool>,
    pub expired_at: Option<DateTime<Utc>>,
}

/// 图片公开信息
#[derive(Debug, Clone, Serialize)]
pub struct PhotoPublic {
    pub id: i64,
    pub uuid: String,
    pub username: String,
    pub filename: String,
    pub url: String,
    pub thumbnail_url: Option<String>,
    pub width: Option<i32>,
    pub height: Option<i32>,
    pub size: i64,
    pub views: i64,
    pub likes: i64,
    pub created_at: DateTime<Utc>,
}

/// 标签实体
#[derive(Debug, Clone, FromRow, Serialize, utoipa::ToSchema)]
pub struct Tag {
    pub id: i64,
    pub name: String,
    pub slug: String,
    pub photo_count: i64,
    pub created_at: DateTime<Utc>,
}

/// 图片-标签关联
#[derive(Debug, Clone, FromRow)]
pub struct PhotoTag {
    pub photo_id: i64,
    pub tag_id: i64,
}

/// 分享实体
///
/// `password` 存储访问密码的哈希，序列化时跳过；对外通过 `has_password` 表示是否加密。
#[derive(Debug, Clone, FromRow, Serialize, utoipa::ToSchema)]
pub struct Share {
    pub id: i64,
    pub user_id: i64,
    pub shareable_type: String,
    pub shareable_id: i64,
    pub slug: String,
    #[serde(skip)]
    pub password: Option<String>,
    #[sqlx(default)]
    pub has_password: bool,
    pub views: i64,
    pub expires_at: Option<DateTime<Utc>>,
    pub created_at: DateTime<Utc>,
    pub deleted_at: Option<DateTime<Utc>>,
}

/// 创建分享请求
#[derive(Debug, Clone, Deserialize, Validate, utoipa::ToSchema)]
pub struct CreateShareRequest {
    pub shareable_type: String,
    pub shareable_id: i64,
    pub password: Option<String>,
    pub expires_at: Option<DateTime<Utc>>,
}

/// 点赞实体
#[derive(Debug, Clone, FromRow)]
pub struct Like {
    pub id: i64,
    pub user_id: i64,
    pub likeable_type: String,
    pub likeable_id: i64,
    pub created_at: DateTime<Utc>,
}

/// 举报实体
#[derive(Debug, Clone, FromRow, Serialize)]
pub struct Report {
    pub id: i64,
    pub user_id: i64,
    pub reportable_type: String,
    pub reportable_id: i64,
    pub reason: String,
    pub status: i32,
    pub created_at: DateTime<Utc>,
    pub updated_at: DateTime<Utc>,
}

impl Photo {
    pub fn new(
        user_id: i64,
        filename: String,
        original_name: String,
        path: String,
        url: String,
        size: i64,
        mime_type: String,
    ) -> Self {
        Self {
            id: 0,
            uuid: Uuid::new_v4().to_string(),
            user_id,
            album_id: None,
            filename,
            original_name,
            path,
            url,
            thumbnail_url: None,
            size,
            width: None,
            height: None,
            mime_type,
            md5: None,
            sha1: None,
            exif: None,
            is_public: false,
            views: 0,
            likes: 0,
            status: 1,
            expired_at: None,
            created_at: Utc::now(),
            updated_at: Utc::now(),
            deleted_at: None,
        }
    }

    /// 检查图片是否已过期
    pub fn is_expired(&self) -> bool {
        match self.expired_at {
            Some(expired) => expired < Utc::now(),
            None => false,
        }
    }

    /// 检查图片是否可用
    pub fn is_available(&self) -> bool {
        self.status == 1 && !self.is_expired() && self.deleted_at.is_none()
    }
}
