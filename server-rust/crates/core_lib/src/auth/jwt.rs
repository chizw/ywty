use chrono::{Duration, Utc};
use jsonwebtoken::{decode, encode, DecodingKey, EncodingKey, Header, Validation};
use serde::{Deserialize, Serialize};
use uuid::Uuid;

use crate::config::JwtConfig;
use crate::error::AppError;

/// JWT Claims
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Claims {
    /// 用户 ID
    pub sub: i64,
    /// 用户名
    pub username: String,
    /// 角色
    pub role: String,
    /// 签发时间
    pub iat: i64,
    /// 过期时间
    pub exp: i64,
    /// Token ID (用于刷新令牌)
    pub jti: String,
}

/// JWT 认证工具
#[derive(Clone)]
pub struct JwtAuth {
    encoding_key: EncodingKey,
    decoding_key: DecodingKey,
    access_expire: Duration,
    refresh_expire: Duration,
}

impl JwtAuth {
    pub fn new(config: &JwtConfig) -> Self {
        Self {
            encoding_key: EncodingKey::from_secret(config.secret.as_bytes()),
            decoding_key: DecodingKey::from_secret(config.secret.as_bytes()),
            access_expire: Duration::seconds(config.access_expire),
            refresh_expire: Duration::seconds(config.refresh_expire),
        }
    }

    /// 生成访问令牌
    pub fn generate_access_token(
        &self,
        user_id: i64,
        username: &str,
        role: &str,
    ) -> crate::AppResult<String> {
        let now = Utc::now();
        let claims = Claims {
            sub: user_id,
            username: username.to_string(),
            role: role.to_string(),
            iat: now.timestamp(),
            exp: (now + self.access_expire).timestamp(),
            jti: Uuid::new_v4().to_string(),
        };

        encode(&Header::default(), &claims, &self.encoding_key)
            .map_err(|e| AppError::Auth(format!("生成令牌失败: {}", e)))
    }

    /// 生成刷新令牌
    pub fn generate_refresh_token(&self, user_id: i64) -> crate::AppResult<String> {
        let now = Utc::now();
        let claims = Claims {
            sub: user_id,
            username: String::new(),
            role: String::new(),
            iat: now.timestamp(),
            exp: (now + self.refresh_expire).timestamp(),
            jti: Uuid::new_v4().to_string(),
        };

        encode(&Header::default(), &claims, &self.encoding_key)
            .map_err(|e| AppError::Auth(format!("生成刷新令牌失败: {}", e)))
    }

    /// 生成令牌对
    pub fn generate_token_pair(
        &self,
        user_id: i64,
        username: &str,
        role: &str,
    ) -> crate::AppResult<TokenPair> {
        let access_token = self.generate_access_token(user_id, username, role)?;
        let refresh_token = self.generate_refresh_token(user_id)?;

        Ok(TokenPair {
            access_token,
            refresh_token,
            token_type: "Bearer".to_string(),
            expires_in: self.access_expire.num_seconds() as u64,
        })
    }

    /// 验证并解析令牌
    pub fn verify_token(&self, token: &str) -> crate::AppResult<Claims> {
        decode::<Claims>(token, &self.decoding_key, &Validation::default())
            .map(|data| data.claims)
            .map_err(|e| AppError::Auth(format!("令牌验证失败: {}", e)))
    }

    /// 验证刷新令牌
    pub fn verify_refresh_token(&self, token: &str) -> crate::AppResult<Claims> {
        self.verify_token(token)
    }
}

/// 令牌对响应
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TokenPair {
    pub access_token: String,
    pub refresh_token: String,
    pub token_type: String,
    pub expires_in: u64,
}
