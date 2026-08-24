//! 分享服务

use crate::db::DbPool;
use chrono::Utc;

use crate::error::{AppError, AppResult};
use crate::models::photo::{CreateShareRequest, Share};

#[derive(Clone)]
pub struct ShareService {
    pool: DbPool,
}

impl ShareService {
    pub fn new(pool: DbPool) -> Self {
        Self { pool }
    }

    /// 生成随机 slug（URL-safe base64，16 字节 ≈ 22 字符）
    fn generate_slug() -> String {
        use base64::{engine::general_purpose::URL_SAFE_NO_PAD, Engine as _};
        let bytes: Vec<u8> = (0..16).map(|_| rand::random::<u8>()).collect();
        URL_SAFE_NO_PAD.encode(&bytes)
    }

    /// 创建分享
    pub async fn create(&self, user_id: i64, req: &CreateShareRequest) -> AppResult<Share> {
        let slug = Self::generate_slug();
        let now = crate::db::now_str();

        // 访问密码以哈希形式入库，绝不存明文
        let password_hash = match &req.password {
            Some(pwd) => Some(crate::auth::password::hash_password(pwd)?),
            None => None,
        };

        let result = sqlx::query(
            r#"
            INSERT INTO shares (user_id, shareable_type, shareable_id, slug, password, expires_at, created_at)
            VALUES (?, ?, ?, ?, ?, ?, ?)
            "#,
        )
        .bind(user_id)
        .bind(&req.shareable_type)
        .bind(req.shareable_id)
        .bind(&slug)
        .bind(&password_hash)
        .bind(req.expires_at)
        .bind(&now)
        .execute(&self.pool)
        .await?;

        let id = crate::db::last_id(&result);

        Ok(Share {
            id,
            user_id,
            shareable_type: req.shareable_type.clone(),
            shareable_id: req.shareable_id,
            slug,
            password: password_hash,
            has_password: req.password.is_some(),
            views: 0,
            expires_at: req.expires_at,
            created_at: Utc::now(),
            deleted_at: None,
        })
    }

    /// 列出我的分享
    pub async fn list(&self, user_id: i64) -> AppResult<Vec<Share>> {
        let mut rows = sqlx::query_as::<_, Share>(
            "SELECT * FROM shares WHERE user_id = ? AND deleted_at IS NULL ORDER BY id DESC",
        )
        .bind(user_id)
        .fetch_all(&self.pool)
        .await?;

        for s in &mut rows {
            s.has_password = s.password.is_some();
        }

        Ok(rows)
    }

    /// 更新分享（仅所有者）
    ///
    /// - `password`: `None` = 不修改, `Some(None)` = 清除密码, `Some(Some("pwd"))` = 设置密码
    /// - `expires_at`: `None` = 不修改, `Some(None)` = 取消过期, `Some(Some(dt))` = 设置过期
    pub async fn update(
        &self,
        user_id: i64,
        id: i64,
        password: Option<Option<&str>>,
        expires_at: Option<Option<chrono::DateTime<Utc>>>,
    ) -> AppResult<Share> {
        // 确认所有权
        let share: Option<Share> = sqlx::query_as(
            "SELECT * FROM shares WHERE id = ? AND user_id = ? AND deleted_at IS NULL",
        )
        .bind(id)
        .bind(user_id)
        .fetch_optional(&self.pool)
        .await?;

        if share.is_none() {
            return Err(AppError::NotFound("分享不存在".to_string()));
        }

        // 更新密码（Some 表示需要修改；存哈希）
        if let Some(pwd_opt) = password {
            let hash = match pwd_opt {
                Some(pwd) => Some(crate::auth::password::hash_password(pwd)?),
                None => None,
            };
            sqlx::query("UPDATE shares SET password = ? WHERE id = ?")
                .bind(&hash)
                .bind(id)
                .execute(&self.pool)
                .await?;
        }

        // 更新过期时间（Some 表示需要修改）
        if let Some(exp) = expires_at {
            sqlx::query("UPDATE shares SET expires_at = ? WHERE id = ?")
                .bind(exp)
                .bind(id)
                .execute(&self.pool)
                .await?;
        }

        // 返回更新后的分享
        let mut updated: Share = sqlx::query_as("SELECT * FROM shares WHERE id = ?")
            .bind(id)
            .fetch_one(&self.pool)
            .await?;
        updated.has_password = updated.password.is_some();

        Ok(updated)
    }

    /// 删除分享（仅所有者，软删除）
    pub async fn delete(&self, user_id: i64, id: i64) -> AppResult<()> {
        let now = crate::db::now_str();
        let result = sqlx::query(
            "UPDATE shares SET deleted_at = ? WHERE id = ? AND user_id = ? AND deleted_at IS NULL",
        )
        .bind(&now)
        .bind(id)
        .bind(user_id)
        .execute(&self.pool)
        .await?;

        if result.rows_affected() == 0 {
            return Err(AppError::NotFound("分享不存在".to_string()));
        }

        Ok(())
    }
}
