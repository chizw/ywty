//! 邮件发送服务
//!
//! 基于 lettre crate，支持 SMTP 发送邮件。
//! 发送时优先读取 settings 表中的 SMTP 配置（host/port/username/password/from/ssl），
//! 缺省回退到启动 config.notify.mail；两者皆无时视为禁用（跳过发送）。

use lettre::message::{header::ContentType, Mailbox, Message};
use lettre::transport::smtp::authentication::Credentials;
use lettre::{AsyncSmtpTransport, AsyncTransport, Tokio1Executor};
use sqlx::SqlitePool;

use crate::config::MailConfig;
use crate::error::{AppError, AppResult};
use crate::services::{self, settings};

/// 解析后的有效 SMTP 配置
struct SmtpSettings {
    host: String,
    port: u16,
    username: String,
    password: String,
    from: String,
    ssl: bool,
}

impl SmtpSettings {
    /// 从 settings 表解析（host 非空视为已配置）
    async fn from_settings(pool: &SqlitePool) -> AppResult<Option<Self>> {
        let host = services::settings::get(pool, settings::keys::MAIL_SMTP_HOST)
            .await?
            .unwrap_or_default();
        if host.trim().is_empty() {
            return Ok(None);
        }
        let port = settings::get_i64(pool, settings::keys::MAIL_SMTP_PORT, 465).await? as u16;
        let ssl = settings::get_bool(pool, settings::keys::MAIL_SMTP_SSL, true).await?;
        Ok(Some(Self {
            host,
            port,
            username: settings::get(pool, settings::keys::MAIL_SMTP_USERNAME)
                .await?
                .unwrap_or_default(),
            password: settings::get(pool, settings::keys::MAIL_SMTP_PASSWORD)
                .await?
                .unwrap_or_default(),
            from: settings::get(pool, settings::keys::MAIL_SMTP_FROM)
                .await?
                .unwrap_or_default(),
            ssl,
        }))
    }

    /// 从启动 config 回退构建
    fn from_config(mail_config: &MailConfig) -> Self {
        Self {
            host: mail_config.host.clone(),
            port: mail_config.port,
            username: mail_config.username.clone(),
            password: mail_config.password.clone(),
            from: mail_config.from.clone(),
            ssl: true,
        }
    }
}

#[derive(Clone, Default)]
pub struct MailService {
    fallback: Option<MailConfig>,
    pool: Option<SqlitePool>,
}

impl MailService {
    /// 从配置创建邮件服务（配置作为回退，发送时优先读 settings 表）
    pub fn new(mail_config: &MailConfig) -> Self {
        Self {
            fallback: Some(mail_config.clone()),
            pool: None,
        }
    }

    /// 创建禁用邮件的服务（用于测试或无邮件配置时）
    ///
    /// 注意：若后续通过 [`MailService::with_pool`] 绑定数据库且管理员在
    /// 系统设置中配置了 SMTP，则仍会按 settings 表配置发送。
    pub fn disabled() -> Self {
        Self::default()
    }

    /// 绑定数据库连接池（发送时从 settings 表读取 SMTP 配置）
    pub fn with_pool(mut self, pool: SqlitePool) -> Self {
        self.pool = Some(pool);
        self
    }

    /// 解析本次发送使用的 SMTP 配置：settings 表优先，回退启动 config
    async fn resolve_settings(&self) -> AppResult<Option<SmtpSettings>> {
        if let Some(pool) = &self.pool {
            if let Some(s) = SmtpSettings::from_settings(pool).await? {
                return Ok(Some(s));
            }
        }
        Ok(self.fallback.as_ref().map(SmtpSettings::from_config))
    }

    /// 按当前配置惰性构建传输层
    fn build_transport(smtp: &SmtpSettings) -> AsyncSmtpTransport<Tokio1Executor> {
        let builder = if smtp.ssl {
            AsyncSmtpTransport::<Tokio1Executor>::relay(&smtp.host).unwrap_or_else(|_| {
                // 回退到 builder 模式
                AsyncSmtpTransport::<Tokio1Executor>::builder_dangerous(&smtp.host)
            })
        } else {
            AsyncSmtpTransport::<Tokio1Executor>::builder_dangerous(&smtp.host)
        };
        builder
            .port(smtp.port)
            .credentials(Credentials::new(
                smtp.username.clone(),
                smtp.password.clone(),
            ))
            .build()
    }

    /// 发送邮件
    pub async fn send(&self, to: &str, subject: &str, body: &str) -> AppResult<()> {
        let Some(smtp) = self.resolve_settings().await? else {
            tracing::debug!("未配置邮件服务（settings 与 config 均无 SMTP），跳过发送");
            return Ok(());
        };

        let sender = smtp
            .from
            .parse::<Mailbox>()
            .unwrap_or_else(|_| Mailbox::new(None, "noreply@example.com".parse().unwrap()));

        let to_addr = to
            .parse::<Mailbox>()
            .map_err(|e| AppError::Validation(format!("收件人地址无效 '{}': {}", to, e)))?;

        let email = Message::builder()
            .from(sender)
            .to(to_addr)
            .subject(subject)
            .header(ContentType::TEXT_HTML)
            .body(body.to_string())
            .map_err(|e| AppError::Internal(format!("构建邮件失败: {}", e)))?;

        Self::build_transport(&smtp)
            .send(email)
            .await
            .map_err(|e| AppError::External(format!("发送邮件失败: {}", e)))?;

        tracing::info!(recipient = %to, subject = %subject, "邮件发送成功");
        Ok(())
    }

    /// 发送验证码邮件
    pub async fn send_verify_code(&self, to: &str, code: &str, event: &str) -> AppResult<()> {
        let (subject, content) = match event {
            "register" => (
                "【云雾图驿】注册验证码",
                format!(
                    r#"<div style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif; max-width: 480px; margin: 0 auto; padding: 32px;">
                        <h2 style="color: #1a1a1a; margin-bottom: 24px;">欢迎注册 云雾图驿</h2>
                        <p style="color: #666; font-size: 14px;">您的注册验证码为：</p>
                        <div style="background: #f5f5f5; padding: 16px 24px; text-align: center; margin: 24px 0; border-radius: 8px;">
                            <span style="font-size: 32px; font-weight: bold; letter-spacing: 8px; color: #1a1a1a;">{}</span>
                        </div>
                        <p style="color: #999; font-size: 12px;">验证码 5 分钟内有效，请勿泄露给他人。</p>
                    </div>"#,
                    code
                ),
            ),
            "reset_password" => (
                "【云雾图驿】重置密码验证码",
                format!(
                    r#"<div style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif; max-width: 480px; margin: 0 auto; padding: 32px;">
                        <h2 style="color: #1a1a1a; margin-bottom: 24px;">重置密码</h2>
                        <p style="color: #666; font-size: 14px;">您正在重置密码，验证码为：</p>
                        <div style="background: #f5f5f5; padding: 16px 24px; text-align: center; margin: 24px 0; border-radius: 8px;">
                            <span style="font-size: 32px; font-weight: bold; letter-spacing: 8px; color: #1a1a1a;">{}</span>
                        </div>
                        <p style="color: #999; font-size: 12px;">验证码 5 分钟内有效，请勿泄露给他人。如非本人操作，请忽略此邮件。</p>
                    </div>"#,
                    code
                ),
            ),
            _ => (
                "【云雾图驿】验证码",
                format!(
                    r#"<div style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif; max-width: 480px; margin: 0 auto; padding: 32px;">
                        <h2 style="color: #1a1a1a; margin-bottom: 24px;">验证码</h2>
                        <div style="background: #f5f5f5; padding: 16px 24px; text-align: center; margin: 24px 0; border-radius: 8px;">
                            <span style="font-size: 32px; font-weight: bold; letter-spacing: 8px; color: #1a1a1a;">{}</span>
                        </div>
                        <p style="color: #999; font-size: 12px;">验证码 5 分钟内有效。</p>
                    </div>"#,
                    code
                ),
            ),
        };

        self.send(to, subject, &content).await
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[tokio::test]
    async fn test_disabled_mail_service() {
        let mail = MailService::disabled();
        // 未配置任何 SMTP 时发送不会真正发送，直接返回成功
        let result = mail.send("test@example.com", "Test", "body").await;
        assert!(result.is_ok());
    }

    #[test]
    fn test_new_service_keeps_fallback_config() {
        let cfg = MailConfig {
            driver: "smtp".to_string(),
            host: "smtp.example.com".to_string(),
            port: 465,
            username: "u".to_string(),
            password: "p".to_string(),
            from: "noreply@example.com".to_string(),
        };
        let mail = MailService::new(&cfg);
        assert!(mail.fallback.is_some());
        assert!(mail.pool.is_none());
    }
}
