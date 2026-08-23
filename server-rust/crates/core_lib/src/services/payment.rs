//! 支付驱动抽象与 Mock 实现
//!
//! - [`PaymentDriver`]：支付驱动统一抽象（发起支付 / 校验异步回调）。
//! - [`MockDriver`]：模拟支付驱动，用于本地开发与测试。
//!
//! 回调验签协议（Mock 与未来真实驱动保持一致的信封）：
//! - 通知体为 JSON：`{"trade_no":"T..","amount":123,"timestamp":1700000000}`
//! - 请求头携带 `X-Signature = hex(HMAC_SHA256(secret, "{trade_no}|{amount}|{timestamp}"))`
//! - 时间戳偏离服务器时间超过 5 分钟视为过期（防重放）

use std::collections::BTreeMap;

use async_trait::async_trait;
use hmac::{Hmac, Mac};
use sha2::Sha256;

use crate::error::{AppError, AppResult};
use crate::models::order::Order;

type HmacSha256 = Hmac<Sha256>;

/// 支付通知时间戳允许的最大偏移（秒），防重放
const NOTIFY_MAX_AGE_SECS: i64 = 300;

/// 支付回调签名头
pub const SIGNATURE_HEADER: &str = "x-signature";

/// 发起支付后返回给前端的支付参数
#[derive(Debug, Clone, serde::Serialize)]
pub struct PaymentParams {
    /// 支付驱动标识（mock / alipay / wechat ...）
    pub driver: &'static str,
    /// 收银台跳转地址（Mock 指向前端订单结果页并携带订单号）
    pub pay_url: String,
    /// Mock 专用：预签名的模拟回调请求（真实驱动不返回）
    #[serde(skip_serializing_if = "Option::is_none")]
    pub mock_notify: Option<MockNotifyRequest>,
}

/// 模拟收银台"确认支付"时，前端应原样发送的回调 HTTP 请求
///
/// 签名由服务端用支付密钥生成，前端只透传，
/// 因此伪造回调仍需持有服务端密钥。
#[derive(Debug, Clone, serde::Serialize)]
pub struct MockNotifyRequest {
    pub method: String,
    pub url: String,
    pub headers: BTreeMap<String, String>,
    pub body: String,
}

/// 验签通过后的回调载荷
#[derive(Debug, Clone)]
pub struct NotifyPayload {
    pub trade_no: String,
    pub amount: i64,
}

/// 支付驱动抽象
#[async_trait]
pub trait PaymentDriver: Send + Sync {
    /// 驱动标识
    fn name(&self) -> &'static str;

    /// 创建支付，返回收银台地址等支付参数
    async fn create_payment(&self, order: &Order) -> AppResult<PaymentParams>;

    /// 校验异步回调：验签 + 防重放，失败返回错误
    fn verify_notify(&self, raw_body: &[u8], signature: &str) -> AppResult<NotifyPayload>;
}

/// 模拟支付驱动
pub struct MockDriver {
    secret: String,
}

impl MockDriver {
    pub fn new(secret: impl Into<String>) -> Self {
        Self {
            secret: secret.into(),
        }
    }

    /// HMAC-SHA256 签名（hex 编码）
    fn sign(&self, data: &[u8]) -> String {
        let mut mac =
            HmacSha256::new_from_slice(self.secret.as_bytes()).expect("HMAC key length is valid");
        mac.update(data);
        hex::encode(mac.finalize().into_bytes())
    }

    /// 参与签名的规范化字符串
    fn canonical(trade_no: &str, amount: i64, timestamp: i64) -> String {
        format!("{trade_no}|{amount}|{timestamp}")
    }
}

#[async_trait]
impl PaymentDriver for MockDriver {
    fn name(&self) -> &'static str {
        "mock"
    }

    async fn create_payment(&self, order: &Order) -> AppResult<PaymentParams> {
        let ts = chrono::Utc::now().timestamp();
        let body = serde_json::json!({
            "trade_no": order.trade_no,
            "amount": order.amount,
            "timestamp": ts,
        });
        let signature = self.sign(Self::canonical(&order.trade_no, order.amount, ts).as_bytes());

        let mut headers = BTreeMap::new();
        headers.insert("X-Signature".to_string(), signature);

        Ok(PaymentParams {
            driver: self.name(),
            pay_url: format!("/dashboard/orders?trade_no={}", order.trade_no),
            mock_notify: Some(MockNotifyRequest {
                method: "POST".to_string(),
                url: "/api/v1/orders/notify".to_string(),
                headers,
                body: body.to_string(),
            }),
        })
    }

    fn verify_notify(&self, raw_body: &[u8], signature: &str) -> AppResult<NotifyPayload> {
        let value: serde_json::Value = serde_json::from_slice(raw_body)
            .map_err(|_| AppError::Validation("通知体不是合法 JSON".to_string()))?;

        let trade_no = value
            .get("trade_no")
            .and_then(|v| v.as_str())
            .unwrap_or_default()
            .to_string();
        let amount = value.get("amount").and_then(|v| v.as_i64());
        let timestamp = value.get("timestamp").and_then(|v| v.as_i64());

        if trade_no.is_empty() || amount.is_none() || timestamp.is_none() {
            return Err(AppError::Validation(
                "通知缺少 trade_no / amount / timestamp 字段".to_string(),
            ));
        }
        let (amount, timestamp) = (amount.unwrap(), timestamp.unwrap());

        // 防重放：时间窗校验
        let now = chrono::Utc::now().timestamp();
        if (now - timestamp).abs() > NOTIFY_MAX_AGE_SECS {
            return Err(AppError::Business("支付通知已过期".to_string()));
        }

        // HMAC-SHA256 验签（常数时间比较）
        let sig_bytes = hex::decode(signature.trim())
            .map_err(|_| AppError::Validation("签名格式错误".to_string()))?;
        let mut mac =
            HmacSha256::new_from_slice(self.secret.as_bytes()).expect("HMAC key length is valid");
        mac.update(Self::canonical(&trade_no, amount, timestamp).as_bytes());
        mac.verify_slice(&sig_bytes)
            .map_err(|_| AppError::Business("支付通知验签失败".to_string()))?;

        Ok(NotifyPayload { trade_no, amount })
    }
}
