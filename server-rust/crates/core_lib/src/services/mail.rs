//! 邮件发送服务
//!
//! 基于 lettre crate，支持 SMTP 发送邮件。
//! 配置从 config.notify.mail 读取。

use lettre::message::{header::ContentType, Mailbox, Message};
use lettre::transport::smtp::authentication::Credentials;
use lettre::{AsyncSmtpTransport, AsyncTransport, Tokio1Executor};

use crate::config::MailConfig;
use crate::error::{AppError, AppResult};

#[derive(Clone)]
pub struct MailService {
    sender: Mailbox,
    transport: AsyncSmtpTransport<Tokio1Executor>,
    enabled: bool,
}

impl MailService {
    /// 从配置创建邮件服务
    pub fn new(mail_config: &MailConfig) -> Self {
        let sender = mail_config
            .from
            .parse::<Mailbox>()
            .unwrap_or_else(|_| Mailbox::new(None, "noreply@example.com".parse().unwrap()));

        // 构建 SMTP 传输
        let transport = AsyncSmtpTransport::<Tokio1Executor>::relay(&mail_config.host)
            .unwrap_or_else(|_| {
                // 回退到 builder 模式
                AsyncSmtpTransport::<Tokio1Executor>::builder_dangerous(&mail_config.host)
            })
            .port(mail_config.port)
            .credentials(Credentials::new(
                mail_config.username.clone(),
                mail_config.password.clone(),
            ))
            .build();

        Self {
            sender,
            transport,
            enabled: true,
        }
    }

    /// 创建禁用邮件的服务（用于测试或无邮件配置时）
    pub fn disabled() -> Self {
        // 使用安全的默认值，避免 panic
        let addr: lettre::Address = "noreply@example.com".parse().unwrap_or_else(|_| {
            // 回退到简单地址
            "localhost@localhost".parse().unwrap_or_else(|_| {
                // 最后的回退（理论上不会发生）
                lettre::Address::new("localhost", "localhost").expect("valid fallback address")
            })
        });
        let sender = Mailbox::new(None, addr);
        let transport = AsyncSmtpTransport::<Tokio1Executor>::builder_dangerous("localhost")
            .port(25)
            .build();

        Self {
            sender,
            transport,
            enabled: false,
        }
    }

    /// 发送邮件
    pub async fn send(
        &self,
        to: &str,
        subject: &str,
        body: &str,
    ) -> AppResult<()> {
        if !self.enabled {
            tracing::debug!("邮件服务已禁用，跳过发送");
            return Ok(());
        }

        let to_addr = to.parse::<Mailbox>().map_err(|e| {
            AppError::Validation(format!("收件人地址无效 '{}': {}", to, e))
        })?;

        let email = Message::builder()
            .from(self.sender.clone())
            .to(to_addr)
            .subject(subject)
            .header(ContentType::TEXT_HTML)
            .body(body.to_string())
            .map_err(|e| AppError::Internal(format!("构建邮件失败: {}", e)))?;

        self.transport.send(email).await.map_err(|e| {
            AppError::External(format!("发送邮件失败: {}", e))
        })?;

        tracing::info!(recipient = %to, subject = %subject, "邮件发送成功");
        Ok(())
    }

    /// 发送验证码邮件
    pub async fn send_verify_code(
        &self,
        to: &str,
        code: &str,
        event: &str,
    ) -> AppResult<()> {
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
        assert!(!mail.enabled);
        // 测试禁用状态下发送不会真正发送
        let result = mail.send("test@example.com", "Test", "body").await;
        assert!(result.is_ok());
    }
}
