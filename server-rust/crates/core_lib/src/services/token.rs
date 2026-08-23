//! API Token 服务

use sqlx::SqlitePool;

use crate::error::{AppError, AppResult};
use crate::models::user::ApiToken;

#[derive(Clone)]
pub struct TokenService {
    pool: SqlitePool,
}

impl TokenService {
    pub fn new(pool: SqlitePool) -> Self {
        Self { pool }
    }

    /// 生成随机 token
    fn generate_token() -> String {
        use base64::{engine::general_purpose::URL_SAFE_NO_PAD, Engine as _};
        let bytes: Vec<u8> = (0..32).map(|_| rand::random::<u8>()).collect();
        format!("ywty_{}", URL_SAFE_NO_PAD.encode(&bytes))
    }

    /// 列出用户 tokens
    pub async fn list(&self, user_id: i64) -> AppResult<Vec<ApiToken>> {
        let rows = sqlx::query_as(
            "SELECT * FROM personal_access_tokens WHERE user_id = ? AND deleted_at IS NULL ORDER BY id DESC",
        )
        .bind(user_id)
        .fetch_all(&self.pool)
        .await?;

        Ok(rows)
    }

    /// 创建 token
    pub async fn create(
        &self,
        user_id: i64,
        name: &str,
        ttl_days: Option<i64>,
    ) -> AppResult<(String, ApiToken)> {
        let raw_token = Self::generate_token();
        let token_hash = {
            use sha2::{Digest, Sha256};
            format!("{:x}", Sha256::digest(raw_token.as_bytes()))
        };
        let now = chrono::Utc::now().to_rfc3339();
        let expires_at =
            ttl_days.map(|d| (chrono::Utc::now() + chrono::Duration::days(d)).to_rfc3339());

        let result = sqlx::query(
            r#"
            INSERT INTO personal_access_tokens (user_id, name, token, expires_at, created_at, updated_at)
            VALUES (?, ?, ?, ?, ?, ?)
            "#,
        )
        .bind(user_id)
        .bind(name)
        .bind(&token_hash)
        .bind(&expires_at)
        .bind(&now)
        .bind(&now)
        .execute(&self.pool)
        .await?;

        let id = result.last_insert_rowid();

        let info = ApiToken {
            id,
            user_id,
            name: name.to_string(),
            token: token_hash,
            scopes: None,
            last_used_at: None,
            expires_at: expires_at
                .as_deref()
                .and_then(|s| chrono::DateTime::parse_from_rfc3339(s).ok())
                .map(|dt| dt.with_timezone(&chrono::Utc)),
            created_at: chrono::Utc::now(),
            deleted_at: None,
        };

        Ok((raw_token, info))
    }

    /// 撤销 token
    pub async fn revoke(&self, user_id: i64, id: i64) -> AppResult<()> {
        let now = chrono::Utc::now().to_rfc3339();
        let result = sqlx::query(
            "UPDATE personal_access_tokens SET deleted_at = ? WHERE id = ? AND user_id = ? AND deleted_at IS NULL",
        )
        .bind(&now)
        .bind(id)
        .bind(user_id)
        .execute(&self.pool)
        .await?;

        if result.rows_affected() == 0 {
            return Err(AppError::NotFound("Token 不存在".to_string()));
        }
        Ok(())
    }
}
