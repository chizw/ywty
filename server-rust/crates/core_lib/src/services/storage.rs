//! 存储服务
//!
//! 为客户端直传提供签名/参数：
//! - s3：预签名 PUT URL（aws-sdk-s3 presigning）
//! - oss：PostObject 的 policy + signature
//! - cos：携带查询签名的 PUT URL
//! - qiniu：上传 token（bucket 作用域）
//! - local / 不支持直传：返回本地上传地址

use crate::config::StorageConfig;
use crate::error::AppResult;
use crate::services::storage_driver::{create_driver, DirectUploadInfo};

use crate::dto::storage::StorageSignResponse;

/// 直传 key 的默认生成规则：direct/{yyyy/MM/dd}/{uuid}
fn default_direct_key() -> String {
    let now = chrono::Utc::now().format("%Y/%m/%d");
    format!("direct/{}/{}", now, uuid::Uuid::new_v4())
}

/// 直传参数为空时的占位 JSON
fn empty_extra() -> serde_json::Value {
    serde_json::json!(null)
}

#[derive(Clone)]
pub struct StorageService {
    public_url_prefix: String,
}

impl StorageService {
    pub fn new(public_url_prefix: String) -> Self {
        Self { public_url_prefix }
    }

    /// 本地驱动的直传信息（走服务端上传接口）
    fn local_sign(&self) -> (StorageSignResponse, serde_json::Value) {
        let response = StorageSignResponse {
            upload_url: format!("{}/uploads", self.public_url_prefix.trim_end_matches('/')),
            public_url: self.public_url_prefix.clone(),
            driver: "local".to_string(),
            max_size: 10485760, // 10MB
            allowed_types: vec![
                "image/jpeg".to_string(),
                "image/png".to_string(),
                "image/gif".to_string(),
                "image/webp".to_string(),
                "image/svg+xml".to_string(),
            ],
        };
        (response, empty_extra())
    }

    /// 将驱动返回的直传信息组装为响应
    ///
    /// `driver` 同时用于兜底公开 URL 前缀（当 info.public_base 为空时）。
    fn build_remote_sign(
        &self,
        driver_name: &str,
        info: DirectUploadInfo,
    ) -> (StorageSignResponse, serde_json::Value) {
        let public_base = if info.public_base.is_empty() {
            self.public_url_prefix.clone()
        } else {
            info.public_base
        };

        (
            StorageSignResponse {
                upload_url: info.upload_url,
                public_url: public_base,
                driver: driver_name.to_string(),
                max_size: 10485760,
                allowed_types: vec![
                    "image/jpeg".to_string(),
                    "image/png".to_string(),
                    "image/gif".to_string(),
                    "image/webp".to_string(),
                ],
            },
            info.extra,
        )
    }

    /// 获取直传签名/参数
    ///
    /// 按当前生效的存储策略（config.storage 构建的驱动）返回对应的直传方式；
    /// 策略不可用或该策略不支持直传时，明确回退为本地上传地址。
    pub async fn sign(
        &self,
        storage_cfg: &StorageConfig,
        key: Option<&str>,
    ) -> AppResult<(StorageSignResponse, serde_json::Value)> {
        // 客户端可指定 object key；未指定时由服务端生成默认 key
        let direct_key = match key {
            Some(k) if !k.trim().is_empty() => k.trim().to_string(),
            _ => default_direct_key(),
        };

        let driver = match create_driver(storage_cfg) {
            Ok(d) => d,
            Err(e) => {
                tracing::warn!("存储策略不可用，回退本地上传地址: {}", e);
                return Ok(self.local_sign());
            }
        };
        let name = driver.name();

        if name == "local" {
            return Ok(self.local_sign());
        }

        // 表单类直传（OSS PostObject 的 policy/signature、七牛上传 token）
        if let Some(mut info) = driver.direct_upload_info(&direct_key) {
            if let Some(obj) = info.extra.as_object_mut() {
                obj.insert("key".to_string(), serde_json::json!(direct_key));
            } else {
                info.extra = serde_json::json!({ "key": direct_key });
            }
            return Ok(self.build_remote_sign(name, info));
        }

        // 预签名 PUT URL 类（S3 兼容预签名 / COS 查询签名）
        if let Some(presigned) = driver.presign_put_url(&direct_key, 3600).await? {
            let info = DirectUploadInfo {
                upload_url: presigned,
                public_base: driver.url("").trim_end_matches('/').to_string(),
                extra: serde_json::json!({ "method": "PUT", "key": direct_key }),
            };
            return Ok(self.build_remote_sign(name, info));
        }

        // 该策略不支持直传：明确回退本地，不造假数据
        tracing::warn!("存储策略 {} 不支持直传，回退本地上传地址", name);
        Ok(self.local_sign())
    }
}
