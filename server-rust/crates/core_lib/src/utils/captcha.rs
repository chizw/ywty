//! 图片验证码生成工具
//!
//! 使用 `captcha` crate 生成扭曲PNG图片，返回 base64 编码的图片数据。

use base64::Engine;
use captcha::filters::{Dots, Noise, Wave};
use captcha::{Captcha, Difficulty};
use uuid::Uuid;

use crate::error::{AppError, AppResult};

/// 验证码结果
pub struct CaptchaResult {
    /// 验证码唯一 ID（用于后续校验）
    pub id: String,
    /// base64 编码的 PNG 图片（data URI 格式）
    pub image_base64: String,
    /// 验证码正确答案（应由调用方存储到缓存/DB）
    pub code: String,
    /// 过期时间（秒）
    pub expires_in: u64,
}

/// 生成图片验证码
///
/// 生成 4 位字符（大小写字母+数字）的扭曲 PNG 图片。
pub fn generate_captcha() -> AppResult<CaptchaResult> {
    let id = Uuid::new_v4().to_string();

    // 使用 captcha crate 生成验证码图片和文本
    let captcha = captcha::gen(Difficulty::Medium);
    let code = captcha.chars_as_string();

    let png_data = captcha
        .as_png()
        .ok_or_else(|| AppError::Internal("验证码图片生成失败".to_string()))?;

    // 编码为 base64 data URI
    let image_base64 = format!(
        "data:image/png;base64,{}",
        base64::engine::general_purpose::STANDARD.encode(&png_data)
    );

    Ok(CaptchaResult {
        id,
        image_base64,
        code,
        expires_in: 300, // 5 分钟
    })
}

/// 生成纯数字验证码（用于邮件验证码场景）
pub fn generate_numeric_captcha(length: usize) -> AppResult<CaptchaResult> {
    let id = Uuid::new_v4().to_string();

    let nums: String = (0..length)
        .map(|_| {
            let n: u32 = rand::random::<u32>() % 10;
            char::from_digit(n, 10).unwrap()
        })
        .collect();

    let png_data = Captcha::new()
        .add_chars(length as u32)
        .apply_filter(Noise::new(0.15))
        .apply_filter(Wave::new(1.5, 15.0).horizontal())
        .apply_filter(Dots::new(4))
        .view(200, 70)
        .as_png()
        .ok_or_else(|| AppError::Internal("验证码图片生成失败".to_string()))?;

    let image_base64 = format!(
        "data:image/png;base64,{}",
        base64::engine::general_purpose::STANDARD.encode(&png_data)
    );

    Ok(CaptchaResult {
        id,
        image_base64,
        code: nums,
        expires_in: 300,
    })
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_generate_captcha_returns_base64_png() {
        // Run multiple times to cover random length variations
        for _ in 0..20 {
            let result = generate_captcha().unwrap();
            assert!(result.image_base64.starts_with("data:image/png;base64,"));
            assert!(
                result.code.len() >= 3 && result.code.len() <= 8,
                "unexpected code length: {}",
                result.code.len()
            );
            assert_eq!(result.expires_in, 300);
            assert!(!result.id.is_empty());
        }
    }

    #[test]
    fn test_generate_numeric_captcha() {
        let result = generate_numeric_captcha(6).unwrap();
        assert!(result.image_base64.starts_with("data:image/png;base64,"));
        assert_eq!(result.code.len(), 6);
        assert!(result.code.chars().all(|c| c.is_ascii_digit()));
    }
}
