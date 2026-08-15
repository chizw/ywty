//! 用户服务

use chrono::Utc;
use sqlx::SqlitePool;

use crate::dto::user::UserProfile;
use crate::auth::password::hash_password;
use crate::error::{AppError, AppResult};

#[derive(Clone)]
pub struct UserService {
    pool: SqlitePool,
}

impl UserService {
    pub fn new(pool: SqlitePool) -> Self {
        Self { pool }
    }

    /// 获取用户资料
    pub async fn get_profile(&self, user_id: i64) -> AppResult<UserProfile> {
        sqlx::query_as::<_, UserProfile>(
            r#"
            SELECT id, uuid, username, email, avatar, bio, role,
                   capacity_used, capacity_max, plan_id, plan_expires_at,
                   email_verified_at, last_login_at, created_at
            FROM users WHERE id = ? AND deleted_at IS NULL
            "#,
        )
        .bind(user_id)
        .fetch_optional(&self.pool)
        .await?
        .ok_or_else(|| AppError::NotFound("用户不存在".to_string()))
    }

    /// 更新用户资料
    pub async fn update_profile(
        &self,
        user_id: i64,
        username: Option<&str>,
        avatar: Option<&str>,
        bio: Option<&str>,
    ) -> AppResult<UserProfile> {
        let now = Utc::now().to_rfc3339();

        // 如果修改用户名，检查唯一性
        if let Some(name) = username {
            let existing: Option<(i64,)> = sqlx::query_as(
                "SELECT id FROM users WHERE username = ? AND id != ? AND deleted_at IS NULL",
            )
            .bind(name)
            .bind(user_id)
            .fetch_optional(&self.pool)
            .await?;

            if existing.is_some() {
                return Err(AppError::Business("用户名已被占用".to_string()));
            }

            sqlx::query("UPDATE users SET username = ?, updated_at = ? WHERE id = ?")
                .bind(name)
                .bind(&now)
                .bind(user_id)
                .execute(&self.pool)
                .await?;
        }

        if let Some(a) = avatar {
            sqlx::query("UPDATE users SET avatar = ?, updated_at = ? WHERE id = ?")
                .bind(a)
                .bind(&now)
                .bind(user_id)
                .execute(&self.pool)
                .await?;
        }

        if let Some(b) = bio {
            sqlx::query("UPDATE users SET bio = ?, updated_at = ? WHERE id = ?")
                .bind(b)
                .bind(&now)
                .bind(user_id)
                .execute(&self.pool)
                .await?;
        }

        self.get_profile(user_id).await
    }

    /// 修改密码
    pub async fn change_password(
        &self,
        user_id: i64,
        old_password: &str,
        new_password: &str,
    ) -> AppResult<()> {
        // 查询当前密码
        let row: Option<(String,)> = sqlx::query_as("SELECT password FROM users WHERE id = ? AND deleted_at IS NULL")
            .bind(user_id)
            .fetch_optional(&self.pool)
            .await?;

        let current_hash =
            row.ok_or_else(|| AppError::NotFound("用户不存在".to_string()))?.0;

        // 验证旧密码
        if !crate::auth::password::verify_password(old_password, &current_hash)? {
            return Err(AppError::Business("原密码错误".to_string()));
        }

        // 更新密码
        let new_hash = hash_password(new_password)?;
        let now = Utc::now().to_rfc3339();
        sqlx::query("UPDATE users SET password = ?, updated_at = ? WHERE id = ?")
            .bind(&new_hash)
            .bind(&now)
            .bind(user_id)
            .execute(&self.pool)
            .await?;

        Ok(())
    }

    /// 修改邮箱（需验证码）
    pub async fn change_email(
        &self,
        user_id: i64,
        new_email: &str,
        verify_code: &str,
    ) -> AppResult<()> {
        // 验证验证码
        let row: Option<(i64, i64)> = sqlx::query_as(
            r#"
            SELECT id, expired_at FROM verify_codes
            WHERE account = ? AND event = 'change_email' AND code = ? AND used_at IS NULL
            ORDER BY id DESC LIMIT 1
            "#,
        )
        .bind(new_email)
        .bind(verify_code)
        .fetch_optional(&self.pool)
        .await?;

        let (code_id, expired_at) =
            row.ok_or_else(|| AppError::Business("验证码无效或已使用".to_string()))?;

        let now_ts = Utc::now().timestamp();
        if expired_at < now_ts {
            return Err(AppError::Business("验证码已过期".to_string()));
        }

        // 更新邮箱
        let now = Utc::now().to_rfc3339();
        sqlx::query("UPDATE users SET email = ?, email_verified_at = ?, updated_at = ? WHERE id = ?")
            .bind(new_email)
            .bind(&now)
            .bind(&now)
            .bind(user_id)
            .execute(&self.pool)
            .await?;

        // 标记验证码已使用
        sqlx::query("UPDATE verify_codes SET used_at = ? WHERE id = ?")
            .bind(now_ts)
            .bind(code_id)
            .execute(&self.pool)
            .await?;

        Ok(())
    }
}
