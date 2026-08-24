//! 认证服务

use crate::db::DbPool;
use chrono::Utc;
use uuid::Uuid;

use crate::auth::password::{hash_password, verify_password};
use crate::auth::JwtAuth;
use crate::dto::auth::{AuthResponse, UserBrief};
use crate::error::{AppError, AppResult};
use crate::services::{mail::MailService, settings};

#[derive(Clone)]
pub struct AuthService {
    pool: DbPool,
    jwt: JwtAuth,
    mail: MailService,
}

impl AuthService {
    pub fn new(pool: DbPool, jwt: JwtAuth, mail: MailService) -> Self {
        // 邮件服务绑定连接池：发送时优先读 settings 表中的 SMTP 配置，
        // 缺省回退到启动 config（mail 实例本身保持不可变）
        let mail = mail.with_pool(pool.clone());
        Self { pool, jwt, mail }
    }

    /// 用户注册
    pub async fn register(
        &self,
        username: &str,
        email: &str,
        password: &str,
    ) -> AppResult<AuthResponse> {
        // 注册开关
        if !settings::get_bool(&self.pool, settings::keys::SECURITY_ALLOW_REGISTER, true).await? {
            return Err(AppError::Validation("注册功能已关闭".to_string()));
        }

        // 检查邮箱是否已存在
        let existing: Option<(i64,)> =
            sqlx::query_as("SELECT id FROM users WHERE email = ? AND deleted_at IS NULL")
                .bind(email)
                .fetch_optional(&self.pool)
                .await?;

        if existing.is_some() {
            return Err(AppError::Business("邮箱已被注册".to_string()));
        }

        // 检查用户名是否已存在
        let existing: Option<(i64,)> =
            sqlx::query_as("SELECT id FROM users WHERE username = ? AND deleted_at IS NULL")
                .bind(username)
                .fetch_optional(&self.pool)
                .await?;

        if existing.is_some() {
            return Err(AppError::Business("用户名已被占用".to_string()));
        }

        // 哈希密码
        let password_hash = hash_password(password)?;

        let uuid = Uuid::new_v4().to_string();
        let now = crate::db::now_str();

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

        let user_id = crate::db::last_id(&result);

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
                is_super_admin: false,
                created_at: Utc::now(),
            },
        })
    }

    /// 用户登录（支持邮箱 / 用户名 / 手机号）
    /// 安全说明：登录失败统一返回"账号/密码错误或不存在"，避免泄露账号是否存在
    pub async fn login(&self, account: &str, password: &str) -> AppResult<AuthResponse> {
        const LOGIN_FAILED: &str = "账号/密码错误或不存在";

        // 根据账号类型查询用户（邮箱 / 用户名 / 手机号）
        let row: Option<(i64, String, String, String, String, String, bool)> = sqlx::query_as(
            r#"
            SELECT id, uuid, username, email, password, role, is_super_admin FROM users
            WHERE deleted_at IS NULL AND (
              email = ? OR username = ? OR phone = ? OR uuid = ?
            )
            LIMIT 1
            "#,
        )
        .bind(account)
        .bind(account)
        .bind(account)
        .bind(account)
        .fetch_optional(&self.pool)
        .await?;

        // 账号不存在时仍返回统一错误，不泄露注册状态
        let (user_id, uuid, username, email, password_hash, role, is_super_admin) =
            row.ok_or_else(|| AppError::Auth(LOGIN_FAILED.to_string()))?;

        // 检查用户状态 — 对外不暴露具体状态差异
        let status: Option<(i32,)> = sqlx::query_as("SELECT status FROM users WHERE id = ?")
            .bind(user_id)
            .fetch_optional(&self.pool)
            .await?;

        if let Some((s,)) = status {
            if s != 1 {
                // 禁用 / 未激活 / 异常状态统一返回登录失败，不提示"被禁用"等细节
                return Err(AppError::Auth(LOGIN_FAILED.to_string()));
            }
        }

        // 验证密码
        if !verify_password(password, &password_hash)? {
            return Err(AppError::Auth(LOGIN_FAILED.to_string()));
        }

        // 更新最后登录时间
        let now = crate::db::now_str();
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
                is_super_admin,
                created_at: created_at.map(|c| c.0).unwrap_or_else(Utc::now),
            },
        })
    }

    /// 刷新令牌
    pub async fn refresh(&self, refresh_token: &str) -> AppResult<AuthResponse> {
        let claims = self.jwt.verify_refresh_token(refresh_token)?;

        // 查询用户确认仍然存在
        let row: Option<(String, String, String, bool)> = sqlx::query_as(
            "SELECT username, email, role, is_super_admin FROM users WHERE id = ? AND deleted_at IS NULL",
        )
        .bind(claims.sub)
        .fetch_optional(&self.pool)
        .await?;

        let (username, email, role, is_super_admin) =
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

        let (uuid, created_at) =
            uuid_row.unwrap_or_else(|| (Uuid::new_v4().to_string(), Utc::now()));

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
                is_super_admin,
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
        // 找回密码开关
        if !settings::get_bool(
            &self.pool,
            settings::keys::SECURITY_ALLOW_PASSWORD_RESET,
            true,
        )
        .await?
        {
            return Err(AppError::Validation("找回密码功能已关闭".to_string()));
        }

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

        let (code_id, expired_at) =
            row.ok_or_else(|| AppError::Business("验证码无效或已使用".to_string()))?;

        // 检查过期（expired_at 存储为 unix 秒）
        let now_ts = Utc::now().timestamp();
        if expired_at < now_ts {
            return Err(AppError::Business("验证码已过期".to_string()));
        }

        // 更新密码
        let password_hash = hash_password(new_password)?;
        let now = crate::db::now_str();
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
        type UserBriefRow = (
            String,
            String,
            String,
            Option<String>,
            String,
            bool,
            chrono::DateTime<Utc>,
        );
        let row: Option<UserBriefRow> = sqlx::query_as(
                "SELECT uuid, username, email, avatar, role, is_super_admin, created_at FROM users WHERE id = ? AND deleted_at IS NULL",
            )
            .bind(user_id)
            .fetch_optional(&self.pool)
            .await?;

        let (uuid, username, email, avatar, role, is_super_admin, created_at) =
            row.ok_or_else(|| AppError::NotFound("用户不存在".to_string()))?;

        Ok(UserBrief {
            id: user_id,
            uuid,
            username,
            email,
            avatar,
            role,
            is_super_admin,
            created_at,
        })
    }

    /// 发送验证码（生成 6 位数字码并通过邮件发送）
    ///
    /// 验证码仅通过邮件下发，绝不返回给调用方，避免绕过邮箱验证。
    pub async fn send_verify_code(&self, email: &str, event: &str) -> AppResult<()> {
        // 功能开关：按事件类型校验对应能力是否开放
        match event {
            // 注册不需要邮箱验证时，不再下发注册验证码
            "register"
                if !settings::get_bool(
                    &self.pool,
                    settings::keys::SECURITY_REQUIRE_EMAIL_VERIFY,
                    true,
                )
                .await? =>
            {
                return Err(AppError::Validation("注册无需邮箱验证".to_string()));
            }
            // 找回密码关闭时，不下发重置密码验证码
            "reset_password"
                if !settings::get_bool(
                    &self.pool,
                    settings::keys::SECURITY_ALLOW_PASSWORD_RESET,
                    true,
                )
                .await? =>
            {
                return Err(AppError::Validation("找回密码功能已关闭".to_string()));
            }
            _ => {}
        }

        // 生成 6 位数字验证码
        let code = format!("{:06}", rand::random::<u32>() % 1_000_000);
        let now_ts = Utc::now().timestamp();
        let expired_at = now_ts + 300; // 5 分钟过期

        sqlx::query(
            r#"
            INSERT INTO verify_codes (channel, account, event, code, expired_at, created_at, updated_at)
            VALUES ('email', ?, ?, ?, ?, ?, ?)
            "#,
        )
        .bind(email)
        .bind(event)
        .bind(&code)
        .bind(expired_at)
        .bind(crate::db::now_str())
        .bind(crate::db::now_str())
        .execute(&self.pool)
        .await?;

        // 通过邮件发送验证码
        if let Err(e) = self.mail.send_verify_code(email, &code, event).await {
            tracing::warn!(email = %email, error = %e, "邮件发送失败，但验证码已生成");
            // 邮件发送失败不影响验证码生成，仅记录日志
        } else {
            tracing::info!(email = %email, event = %event, "验证码邮件已发送");
        }

        Ok(())
    }
}
