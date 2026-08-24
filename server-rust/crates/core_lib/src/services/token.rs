//! API Token 服务

use crate::db::DbPool;

use crate::error::{AppError, AppResult};
use crate::models::user::ApiToken;

#[derive(Clone)]
pub struct TokenService {
    pool: DbPool,
}

impl TokenService {
    pub fn new(pool: DbPool) -> Self {
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
        let now = crate::db::now_str();
        let expires_at_dt = ttl_days.map(|d| chrono::Utc::now() + chrono::Duration::days(d));
        let expires_at_sql = expires_at_dt
            .as_ref()
            .map(|t| t.format("%Y-%m-%d %H:%M:%S").to_string());

        let result = sqlx::query(
            r#"
            INSERT INTO personal_access_tokens (user_id, name, token, expires_at, created_at, updated_at)
            VALUES (?, ?, ?, ?, ?, ?)
            "#,
        )
        .bind(user_id)
        .bind(name)
        .bind(&token_hash)
        .bind(&expires_at_sql)
        .bind(&now)
        .bind(&now)
        .execute(&self.pool)
        .await?;

        let id = crate::db::last_id(&result);

        let info = ApiToken {
            id,
            user_id,
            name: name.to_string(),
            token: token_hash,
            scopes: None,
            last_used_at: None,
            expires_at: expires_at_dt,
            created_at: chrono::Utc::now(),
            deleted_at: None,
        };

        Ok((raw_token, info))
    }

    /// 撤销 token
    pub async fn revoke(&self, user_id: i64, id: i64) -> AppResult<()> {
        let now = crate::db::now_str();
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
