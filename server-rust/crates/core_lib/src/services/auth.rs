//! 认证服务

use chrono::Utc;
use sqlx::SqlitePool;
use uuid::Uuid;

use crate::dto::auth::{AuthResponse, UserBrief};
use crate::auth::password::{hash_password, verify_password};
use crate::auth::JwtAuth;
use crate::error::{AppError, AppResult};

#[derive(Clone)]
pub struct AuthService {
    pool: SqlitePool,
    jwt: JwtAuth,
}

impl AuthService {
    pub fn new(pool: SqlitePool, jwt: JwtAuth) -> Self {
        Self { pool, jwt }
    }

    /// 用户注册
    pub async fn register(
        &self,
        username: &str,
        email: &str,
        password: &str,
    ) -> AppResult<AuthResponse> {
        // 检查邮箱是否已存在
        let existing: Option<(i64,)> = sqlx::query_as("SELECT id FROM users WHERE email = ? AND deleted_at IS NULL")
            .bind(email)
            .fetch_optional(&self.pool)
            .await?;

        if existing.is_some() {
            return Err(AppError::Business("邮箱已被注册".to_string()));
        }

        // 检查用户名是否已存在
        let existing: Option<(i64,)> = sqlx::query_as("SELECT id FROM users WHERE username = ? AND deleted_at IS NULL")
            .bind(username)
            .fetch_optional(&self.pool)
            .await?;

        if existing.is_some() {
            return Err(AppError::Business("用户名已被占用".to_string()));
        }

        // 哈希密码
        let password_hash = hash_password(password)?;

        let uuid = Uuid::new_v4().to_string();
        let now = Utc::now().to_rfc3339();

        // 插入用户
        let result = sqlx::query(
            r#"
            INSERT INTO users (uuid, username, email, password, role, status, capacity_used, capacity_max, created_at, updated_at)
            VALUES (?, ?, ?, ?, 'user', 1, 0, 104857600, ?, ?)
            "#,
        )
        .bind(&uuid)
        .bind(username)
        .bind(email)
        .bind(&password_hash)
        .bind(&now)
        .bind(&now)
        .execute(&self.pool)
        .await?;

        let user_id = result.last_insert_rowid();

        // 签发令牌
        let token_pair = self.jwt.generate_token_pair(user_id, username, "user")?;

        Ok(AuthResponse {
            access_token: token_pair.access_token,
            refresh_token: token_pair.refresh_token,
            token_type: token_pair.token_type,
            expires_in: token_pair.expires_in,
            user: UserBrief {
                id: user_id,
                uuid,
                username: username.to_string(),
                email: email.to_string(),
                avatar: None,
                role: "user".to_string(),
                created_at: Utc::now(),
            },
        })
    }

    /// 用户登录
    pub async fn login(&self, email: &str, password: &str) -> AppResult<AuthResponse> {
        // 查询用户
        let row: Option<(i64, String, String, String, String, String)> = sqlx::query_as(
            "SELECT id, uuid, username, email, password, role FROM users WHERE email = ? AND deleted_at IS NULL",
        )
        .bind(email)
        .fetch_optional(&self.pool)
        .await?;

        let (user_id, uuid, username, email, password_hash, role) =
            row.ok_or_else(|| AppError::Auth("邮箱或密码错误".to_string()))?;

        // 检查用户状态
        let status: Option<(i32,)> = sqlx::query_as("SELECT status FROM users WHERE id = ?")
            .bind(user_id)
            .fetch_optional(&self.pool)
            .await?;

        if let Some((s,)) = status {
            if s != 1 {
                return Err(AppError::Auth("账号已被禁用".to_string()));
            }
        }

        // 验证密码
        if !verify_password(password, &password_hash)? {
            return Err(AppError::Auth("邮箱或密码错误".to_string()));
        }

        // 更新最后登录时间
        let now = Utc::now().to_rfc3339();
        let _ = sqlx::query("UPDATE users SET last_login_at = ?, updated_at = ? WHERE id = ?")
            .bind(&now)
            .bind(&now)
            .bind(user_id)
            .execute(&self.pool)
            .await;

        // 签发令牌
        let token_pair = self.jwt.generate_token_pair(user_id, &username, &role)?;

        let avatar: Option<(Option<String>,)> =
            sqlx::query_as("SELECT avatar FROM users WHERE id = ?")
                .bind(user_id)
                .fetch_optional(&self.pool)
                .await?;

        let created_at: Option<(chrono::DateTime<Utc>,)> =
            sqlx::query_as("SELECT created_at FROM users WHERE id = ?")
                .bind(user_id)
                .fetch_optional(&self.pool)
                .await?;

        Ok(AuthResponse {
            access_token: token_pair.access_token,
            refresh_token: token_pair.refresh_token,
            token_type: token_pair.token_type,
            expires_in: token_pair.expires_in,
            user: UserBrief {
                id: user_id,
                uuid,
                username: username.clone(),
                email,
                avatar: avatar.and_then(|a| a.0),
                role,
                created_at: created_at.map(|c| c.0).unwrap_or_else(Utc::now),
            },
        })
    }

    /// 刷新令牌
    pub async fn refresh(&self, refresh_token: &str) -> AppResult<AuthResponse> {
        let claims = self.jwt.verify_refresh_token(refresh_token)?;

        // 查询用户确认仍然存在
        let row: Option<(String, String, String)> = sqlx::query_as(
            "SELECT username, email, role FROM users WHERE id = ? AND deleted_at IS NULL",
        )
        .bind(claims.sub)
        .fetch_optional(&self.pool)
        .await?;

        let (username, email, role) =
            row.ok_or_else(|| AppError::Auth("用户不存在".to_string()))?;

        // 签发新令牌对
        let token_pair = self.jwt.generate_token_pair(claims.sub, &username, &role)?;

        let avatar: Option<(Option<String>,)> =
            sqlx::query_as("SELECT avatar FROM users WHERE id = ?")
                .bind(claims.sub)
                .fetch_optional(&self.pool)
                .await?;

        let uuid_row: Option<(String, chrono::DateTime<Utc>)> =
            sqlx::query_as("SELECT uuid, created_at FROM users WHERE id = ?")
                .bind(claims.sub)
                .fetch_optional(&self.pool)
                .await?;

        let (uuid, created_at) = uuid_row.unwrap_or_else(|| {
            (Uuid::new_v4().to_string(), Utc::now())
        });

        Ok(AuthResponse {
            access_token: token_pair.access_token,
            refresh_token: token_pair.refresh_token,
            token_type: token_pair.token_type,
            expires_in: token_pair.expires_in,
            user: UserBrief {
                id: claims.sub,
                uuid,
                username: username.clone(),
                email,
                avatar: avatar.and_then(|a| a.0),
                role,
                created_at,
            },
        })
    }

    /// 通过验证码重置密码
    pub async fn reset_password(
        &self,
        email: &str,
        new_password: &str,
        verify_code: &str,
    ) -> AppResult<()> {
        // 验证验证码
        let row: Option<(i64, i64)> = sqlx::query_as(
            r#"
            SELECT id, expired_at FROM verify_codes
            WHERE account = ? AND event = 'reset_password' AND code = ? AND used_at IS NULL
            ORDER BY id DESC LIMIT 1
            "#,
        )
        .bind(email)
        .bind(verify_code)
        .fetch_optional(&self.pool)
        .await?;

        let (code_id, expired_at) = row.ok_or_else(|| AppError::Business("验证码无效或已使用".to_string()))?;

        // 检查过期（expired_at 存储为 unix 秒）
        let now_ts = Utc::now().timestamp();
        if expired_at < now_ts {
            return Err(AppError::Business("验证码已过期".to_string()));
        }

        // 更新密码
        let password_hash = hash_password(new_password)?;
        let now = Utc::now().to_rfc3339();
        sqlx::query("UPDATE users SET password = ?, updated_at = ? WHERE email = ?")
            .bind(&password_hash)
            .bind(&now)
            .bind(email)
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

    /// 获取当前用户信息
    pub async fn get_me(&self, user_id: i64) -> AppResult<UserBrief> {
        let row: Option<(String, String, String, Option<String>, String, chrono::DateTime<Utc>)> =
            sqlx::query_as(
                "SELECT uuid, username, email, avatar, role, created_at FROM users WHERE id = ? AND deleted_at IS NULL",
            )
            .bind(user_id)
            .fetch_optional(&self.pool)
            .await?;

        let (uuid, username, email, avatar, role, created_at) =
            row.ok_or_else(|| AppError::NotFound("用户不存在".to_string()))?;

        Ok(UserBrief {
            id: user_id,
            uuid,
            username,
            email,
            avatar,
            role,
            created_at,
        })
    }

    /// 发送验证码（简化实现：生成 6 位数字码并存储）
    pub async fn send_verify_code(&self, email: &str, event: &str) -> AppResult<String> {
        // 简单数字验证码
        let code = format!("{:06}", rand::random::<u32>() % 1_000_000);
        let now_ts = Utc::now().timestamp();
        let expired_at = now_ts + 300; // 5 分钟过期

        sqlx::query(
            r#"
            INSERT INTO verify_codes (channel, account, event, code, expired_at, created_at, updated_at)
            VALUES ('email', ?, ?, ?, ?, datetime('now'), datetime('now'))
            "#,
        )
        .bind(email)
        .bind(event)
        .bind(&code)
        .bind(expired_at)
        .execute(&self.pool)
        .await?;

        // TODO: 实际发送邮件（对接 mail 驱动）
        tracing::info!(email = %email, event = %event, "验证码已生成（未实际发送）: {}", code);

        Ok(code) // 实际生产环境不应返回验证码，此处便于开发测试
    }
}
