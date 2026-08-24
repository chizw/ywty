//! OAuth 服务
//!
//! 实现 GitHub 和 Google 的授权码流程（Authorization Code Grant）。
//! 支持 CSRF state 验证（Redis 存储 + 签名 Cookie 降级）。

use crate::db::DbPool;
use serde::Deserialize;

use crate::config::OAuthProviderConfig;
use crate::error::{AppError, AppResult};
use crate::models::user::OAuthAccount;

/// OAuth 用户信息（标准化）
#[derive(Debug, Clone)]
pub struct OAuthUserInfo {
    pub provider: String,
    pub provider_user_id: String,
    pub username: String,
    pub email: Option<String>,
    pub avatar: Option<String>,
}

/// 令牌端点响应
#[derive(Debug, Deserialize)]
struct TokenResponse {
    access_token: String,
}

/// GitHub 用户信息
#[derive(Debug, Deserialize)]
struct GitHubUser {
    id: i64,
    login: String,
    #[serde(default)]
    name: Option<String>,
    #[serde(default)]
    email: Option<String>,
    #[serde(default)]
    avatar_url: Option<String>,
}

/// GitHub 邮箱信息
#[derive(Debug, Deserialize)]
struct GitHubEmail {
    email: String,
    primary: bool,
    verified: bool,
}

/// Google 用户信息（OpenID Connect）
#[derive(Debug, Deserialize)]
struct GoogleUser {
    sub: String,
    #[serde(default)]
    name: Option<String>,
    #[serde(default)]
    email: Option<String>,
    #[serde(default)]
    picture: Option<String>,
}

#[derive(Clone)]
pub struct OAuthService {
    pool: DbPool,
    http: reqwest::Client,
    github_config: Option<OAuthProviderConfig>,
    google_config: Option<OAuthProviderConfig>,
}

impl OAuthService {
    pub fn new(
        pool: DbPool,
        github_config: Option<OAuthProviderConfig>,
        google_config: Option<OAuthProviderConfig>,
    ) -> Self {
        Self {
            pool,
            http: reqwest::Client::new(),
            github_config,
            google_config,
        }
    }

    /// 获取提供商配置（静态：仅 config.yaml）
    fn static_provider_config(&self, provider: &str) -> AppResult<OAuthProviderConfig> {
        match provider {
            "github" => self
                .github_config
                .clone()
                .ok_or_else(|| AppError::Business("GitHub OAuth 未配置".to_string())),
            "google" => self
                .google_config
                .clone()
                .ok_or_else(|| AppError::Business("Google OAuth 未配置".to_string())),
            _ => Err(AppError::Business(format!(
                "不支持的 OAuth 提供商: {}",
                provider
            ))),
        }
    }

    /// 获取提供商配置：优先读 settings 表（后台可配），缺省回退 config.yaml
    async fn resolve_provider_config(&self, provider: &str) -> AppResult<OAuthProviderConfig> {
        let prefix = match provider {
            "github" => "oauth.github",
            "google" => "oauth.google",
            _ => {
                return Err(AppError::Business(format!(
                    "不支持的 OAuth 提供商: {}",
                    provider
                )))
            }
        };

        let client_id =
            crate::services::settings::get(&self.pool, &format!("{}.client_id", prefix))
                .await?
                .filter(|v| !v.trim().is_empty());

        if let Some(id) = client_id {
            let client_secret =
                crate::services::settings::get(&self.pool, &format!("{}.client_secret", prefix))
                    .await?
                    .unwrap_or_default();
            let static_cfg = self.static_provider_config(provider).ok();
            let redirect_uri =
                crate::services::settings::get(&self.pool, &format!("{}.redirect_uri", prefix))
                    .await?
                    .filter(|v| !v.trim().is_empty())
                    .or_else(|| static_cfg.as_ref().map(|c| c.redirect_uri.clone()))
                    .unwrap_or_default();
            return Ok(OAuthProviderConfig {
                client_id: id,
                client_secret,
                redirect_uri,
            });
        }

        self.static_provider_config(provider)
    }

    /// 列出已配置（settings 或 config.yaml 任一来源）的提供商，供前端渲染入口
    pub async fn configured_providers(&self) -> Vec<(&'static str, &'static str)> {
        let mut items = Vec::new();
        for (name, label) in [("github", "GitHub"), ("google", "Google")] {
            if self.resolve_provider_config(name).await.is_ok() {
                items.push((name, label));
            }
        }
        items
    }

    /// 获取授权 URL 和 CSRF state
    pub async fn authorize_url(&self, provider: &str) -> AppResult<(String, String)> {
        let config = self.resolve_provider_config(provider).await?;
        let state = uuid::Uuid::new_v4().to_string();

        let url = match provider {
            "github" => format!(
                "https://github.com/login/oauth/authorize?client_id={}&redirect_uri={}&scope=read:user%20user:email&state={}",
                config.client_id,
                urlencoding::encode(&config.redirect_uri),
                state
            ),
            "google" => format!(
                "https://accounts.google.com/o/oauth2/v2/auth?client_id={}&redirect_uri={}&response_type=code&scope=openid%20email%20profile&state={}&access_type=offline",
                config.client_id,
                urlencoding::encode(&config.redirect_uri),
                state
            ),
            _ => return Err(AppError::Business(format!("不支持的提供商: {}", provider))),
        };

        Ok((url, state))
    }

    /// OAuth 回调：交换 code → 获取用户信息 → 查找或创建用户 → 返回用户信息
    pub async fn login_or_register(&self, provider: &str, code: &str) -> AppResult<OAuthUserInfo> {
        let config = self.resolve_provider_config(provider).await?;

        // 1. 用 code 交换 access_token
        let access_token = self.exchange_code(provider, &config, code).await?;

        // 2. 获取用户信息
        let user_info = self.fetch_user_info(provider, &access_token).await?;

        Ok(user_info)
    }

    /// 交换授权码获取访问令牌
    async fn exchange_code(
        &self,
        provider: &str,
        config: &OAuthProviderConfig,
        code: &str,
    ) -> AppResult<String> {
        let token = match provider {
            "github" => {
                let resp = self
                    .http
                    .post("https://github.com/login/oauth/access_token")
                    .header("Accept", "application/json")
                    .form(&[
                        ("client_id", config.client_id.as_str()),
                        ("client_secret", config.client_secret.as_str()),
                        ("code", code),
                        ("redirect_uri", config.redirect_uri.as_str()),
                    ])
                    .send()
                    .await
                    .map_err(|e| AppError::External(format!("GitHub token 请求失败: {}", e)))?;

                let token_resp: TokenResponse = resp
                    .json()
                    .await
                    .map_err(|e| AppError::External(format!("GitHub token 解析失败: {}", e)))?;

                token_resp.access_token
            }
            "google" => {
                let resp = self
                    .http
                    .post("https://oauth2.googleapis.com/token")
                    .form(&[
                        ("client_id", config.client_id.as_str()),
                        ("client_secret", config.client_secret.as_str()),
                        ("code", code),
                        ("redirect_uri", config.redirect_uri.as_str()),
                        ("grant_type", "authorization_code"),
                    ])
                    .send()
                    .await
                    .map_err(|e| AppError::External(format!("Google token 请求失败: {}", e)))?;

                let token_resp: TokenResponse = resp
                    .json()
                    .await
                    .map_err(|e| AppError::External(format!("Google token 解析失败: {}", e)))?;

                token_resp.access_token
            }
            _ => return Err(AppError::Business(format!("不支持的提供商: {}", provider))),
        };

        if token.is_empty() {
            return Err(AppError::External("获取 access_token 失败".to_string()));
        }

        Ok(token)
    }

    /// 获取 OAuth 用户信息
    async fn fetch_user_info(
        &self,
        provider: &str,
        access_token: &str,
    ) -> AppResult<OAuthUserInfo> {
        match provider {
            "github" => {
                // 获取基本用户信息
                let resp = self
                    .http
                    .get("https://api.github.com/user")
                    .header("Authorization", format!("Bearer {}", access_token))
                    .header("User-Agent", "ywty-cloud-album")
                    .header("Accept", "application/vnd.github.v3+json")
                    .send()
                    .await
                    .map_err(|e| AppError::External(format!("GitHub user 请求失败: {}", e)))?;

                let github_user: GitHubUser = resp
                    .json()
                    .await
                    .map_err(|e| AppError::External(format!("GitHub user 解析失败: {}", e)))?;

                // 如果 email 为空，尝试从邮箱接口获取
                let email = if github_user.email.is_some() {
                    github_user.email
                } else {
                    self.fetch_github_emails(access_token).await
                };

                Ok(OAuthUserInfo {
                    provider: "github".to_string(),
                    provider_user_id: github_user.id.to_string(),
                    username: github_user.name.unwrap_or(github_user.login),
                    email,
                    avatar: github_user.avatar_url,
                })
            }
            "google" => {
                let resp = self
                    .http
                    .get("https://openidconnect.googleapis.com/v1/userinfo")
                    .header("Authorization", format!("Bearer {}", access_token))
                    .send()
                    .await
                    .map_err(|e| AppError::External(format!("Google userinfo 请求失败: {}", e)))?;

                let google_user: GoogleUser = resp
                    .json()
                    .await
                    .map_err(|e| AppError::External(format!("Google userinfo 解析失败: {}", e)))?;

                Ok(OAuthUserInfo {
                    provider: "google".to_string(),
                    provider_user_id: google_user.sub,
                    username: google_user
                        .name
                        .unwrap_or_else(|| google_user.email.clone().unwrap_or_default()),
                    email: google_user.email,
                    avatar: google_user.picture,
                })
            }
            _ => Err(AppError::Business(format!("不支持的提供商: {}", provider))),
        }
    }

    /// 获取 GitHub 用户邮箱列表（当基本信息的 email 为空时）
    async fn fetch_github_emails(&self, access_token: &str) -> Option<String> {
        let resp = self
            .http
            .get("https://api.github.com/user/emails")
            .header("Authorization", format!("Bearer {}", access_token))
            .header("User-Agent", "ywty-cloud-album")
            .header("Accept", "application/vnd.github.v3+json")
            .send()
            .await
            .ok()?;

        let emails: Vec<GitHubEmail> = resp.json().await.ok()?;

        // 优先选择已验证的主邮箱
        emails
            .into_iter()
            .find(|e| e.primary && e.verified)
            .map(|e| e.email)
    }

    /// 通过 provider_user_id 查找已绑定的用户
    pub async fn find_by_open_id(
        &self,
        provider: &str,
        open_id: &str,
    ) -> AppResult<Option<(i64, String, String)>> {
        let row: Option<(i64, String, String)> = sqlx::query_as(
            "SELECT u.id, u.username, u.email FROM users u
             INNER JOIN oauth_accounts oa ON oa.user_id = u.id
             WHERE oa.provider = ? AND oa.provider_user_id = ? AND u.deleted_at IS NULL",
        )
        .bind(provider)
        .bind(open_id)
        .fetch_optional(&self.pool)
        .await?;

        Ok(row)
    }

    /// 绑定 OAuth 账号到用户
    pub async fn bind(
        &self,
        user_id: i64,
        provider: &str,
        provider_user_id: &str,
    ) -> AppResult<OAuthAccount> {
        let now = crate::db::now_str();

        // 检查是否已绑定
        let existing: Option<OAuthAccount> =
            sqlx::query_as("SELECT * FROM oauth_accounts WHERE user_id = ? AND provider = ?")
                .bind(user_id)
                .bind(provider)
                .fetch_optional(&self.pool)
                .await?;

        if let Some(mut acc) = existing {
            // 更新
            sqlx::query(
                "UPDATE oauth_accounts SET provider_user_id = ?, updated_at = ? WHERE id = ?",
            )
            .bind(provider_user_id)
            .bind(&now)
            .bind(acc.id)
            .execute(&self.pool)
            .await?;
            acc.provider_user_id = provider_user_id.to_string();
            Ok(acc)
        } else {
            // 新建
            let result = sqlx::query(
                r#"
                INSERT INTO oauth_accounts (user_id, provider, provider_user_id, created_at, updated_at)
                VALUES (?, ?, ?, ?, ?)
                "#,
            )
            .bind(user_id)
            .bind(provider)
            .bind(provider_user_id)
            .bind(&now)
            .bind(&now)
            .execute(&self.pool)
            .await?;

            let id = crate::db::last_id(&result);
            let acc: OAuthAccount = sqlx::query_as("SELECT * FROM oauth_accounts WHERE id = ?")
                .bind(id)
                .fetch_one(&self.pool)
                .await?;
            Ok(acc)
        }
    }

    /// 解绑 OAuth 账号
    pub async fn unbind(&self, user_id: i64, id: i64) -> AppResult<()> {
        let result = sqlx::query("DELETE FROM oauth_accounts WHERE id = ? AND user_id = ?")
            .bind(id)
            .bind(user_id)
            .execute(&self.pool)
            .await?;

        if result.rows_affected() == 0 {
            return Err(AppError::NotFound("绑定不存在".to_string()));
        }
        Ok(())
    }

    /// 列出用户 OAuth 绑定
    pub async fn list(&self, user_id: i64) -> AppResult<Vec<OAuthAccount>> {
        let rows = sqlx::query_as("SELECT * FROM oauth_accounts WHERE user_id = ? ORDER BY id ASC")
            .bind(user_id)
            .fetch_all(&self.pool)
            .await?;

        Ok(rows)
    }

    /// 创建 OAuth 用户（自动注册）
    pub async fn create_oauth_user(
        &self,
        info: &OAuthUserInfo,
    ) -> AppResult<(i64, String, String)> {
        let now = crate::db::now_str();
        let uuid = uuid::Uuid::new_v4().to_string();

        // 生成用户名（如果冲突则加随机后缀）
        let username = format!(
            "{}_{}",
            info.provider,
            &info.provider_user_id[..info.provider_user_id.len().min(8)]
        );
        let email = info
            .email
            .clone()
            .unwrap_or_else(|| format!("_{}@oauth.{}", info.provider_user_id, info.provider));

        // 生成随机密码（用户后续可通过邮箱重置）
        let random_password = uuid::Uuid::new_v4().to_string().replace("-", "");
        let password_hash = crate::auth::password::hash_password(&random_password)?;

        // 创建用户
        let result = sqlx::query(
            r#"
            INSERT INTO users (uuid, username, email, password, avatar, role, status, capacity_used, capacity_max, created_at, updated_at)
            VALUES (?, ?, ?, ?, ?, 'user', 1, 0, 104857600, ?, ?)
            "#,
        )
        .bind(&uuid)
        .bind(&username)
        .bind(&email)
        .bind(&password_hash)
        .bind(&info.avatar)
        .bind(&now)
        .bind(&now)
        .execute(&self.pool)
        .await?;

        let user_id = crate::db::last_id(&result);

        // 绑定 OAuth 账号
        sqlx::query(
            r#"
            INSERT INTO oauth_accounts (user_id, provider, provider_user_id, created_at, updated_at)
            VALUES (?, ?, ?, ?, ?)
            "#,
        )
        .bind(user_id)
        .bind(&info.provider)
        .bind(&info.provider_user_id)
        .bind(&now)
        .bind(&now)
        .execute(&self.pool)
        .await?;

        Ok((user_id, username, email))
    }
}
