//! 多存储驱动（Storage Driver）抽象层
//!
//! 定义统一的存储 trait，支持 Local / S3 / OSS / COS / Qiniu 驱动。
//! 通过 `StorageDriver` 枚举在配置中指定，运行时创建对应的驱动实例。

use async_trait::async_trait;
use base64::Engine;
use bytes::Bytes;
use hmac::{Hmac, Mac};
use sha1::{Digest, Sha1};
use std::sync::Arc;

use crate::config::{CosConfig, OssConfig, S3Config};

type HmacSha1 = Hmac<Sha1>;

/// 计算 HMAC-SHA1（返回原始字节）
fn hmac_sha1(key: &[u8], data: &[u8]) -> Vec<u8> {
    let mut mac = HmacSha1::new_from_slice(key).expect("HMAC 可接受任意长度密钥");
    mac.update(data);
    mac.finalize().into_bytes().to_vec()
}

/// 对 object key 的各路径段做 URL 编码（保留 `/` 分隔符）
fn encode_key_path(key: &str) -> String {
    key.split('/')
        .map(urlencoding::encode)
        .collect::<Vec<_>>()
        .join("/")
}

/// 客户端直传参数
#[derive(Debug, Clone)]
pub struct DirectUploadInfo {
    /// 直传请求目标 URL
    pub upload_url: String,
    /// 直传后可公开访问的 URL 前缀（不含 key）
    pub public_base: String,
    /// 附加参数（随响应下发给前端，如 policy/signature/token 等）
    pub extra: serde_json::Value,
}

/// 上传结果
#[derive(Debug, Clone)]
pub struct UploadResult {
    /// 存储路径（key）
    pub path: String,
    /// 公开访问 URL
    pub url: String,
    /// 文件大小（字节）
    pub size: i64,
}

/// 存储驱动 trait
///
/// 所有存储驱动（Local / S3 / OSS / COS 等）实现此 trait。
#[async_trait]
pub trait StorageDriver: Send + Sync + 'static {
    /// 驱动名称
    fn name(&self) -> &'static str;

    /// 上传文件
    async fn upload(
        &self,
        key: &str,
        data: Bytes,
        content_type: &str,
    ) -> crate::AppResult<UploadResult>;

    /// 删除文件
    async fn delete(&self, key: &str) -> crate::AppResult<()>;

    /// 获取文件公开 URL
    fn url(&self, key: &str) -> String;

    /// 检查文件是否存在
    async fn exists(&self, key: &str) -> crate::AppResult<bool> {
        let _ = key;
        Ok(false)
    }

    /// 为指定 key 生成分片直传所需参数（表单类直传：OSS POST / 七牛 token）。
    ///
    /// 不支持此方式的驱动返回 `None`（如 Local、S3 走预签名 URL）。
    fn direct_upload_info(&self, _key: &str) -> Option<DirectUploadInfo> {
        None
    }

    /// 为指定 key 生成预签名 PUT URL（S3 兼容 / COS 查询签名等）。
    ///
    /// 不支持的驱动返回 `None`。
    async fn presign_put_url(
        &self,
        _key: &str,
        _expires_secs: i64,
    ) -> crate::AppResult<Option<String>> {
        Ok(None)
    }
}

// ─────────────────────────────────────────────
// Local 驱动
// ─────────────────────────────────────────────

/// 本地文件系统存储驱动
pub struct LocalDriver {
    /// 本地根目录
    pub root: String,
    /// 公开访问 URL 前缀
    pub public_url: String,
}

#[async_trait]
impl StorageDriver for LocalDriver {
    fn name(&self) -> &'static str {
        "local"
    }

    async fn upload(
        &self,
        key: &str,
        data: Bytes,
        _content_type: &str,
    ) -> crate::AppResult<UploadResult> {
        let path = std::path::Path::new(&self.root).join(key);

        // 创建父目录
        if let Some(parent) = path.parent() {
            tokio::fs::create_dir_all(parent).await?;
        }

        tokio::fs::write(&path, &data).await?;

        Ok(UploadResult {
            path: key.to_string(),
            url: self.url(key),
            size: data.len() as i64,
        })
    }

    async fn delete(&self, key: &str) -> crate::AppResult<()> {
        let path = std::path::Path::new(&self.root).join(key);
        if path.exists() {
            tokio::fs::remove_file(&path).await?;
        }
        Ok(())
    }

    fn url(&self, key: &str) -> String {
        format!("{}/{}", self.public_url.trim_end_matches('/'), key)
    }

    async fn exists(&self, key: &str) -> crate::AppResult<bool> {
        let path = std::path::Path::new(&self.root).join(key);
        Ok(path.exists())
    }
}

// ─────────────────────────────────────────────
// S3 兼容驱动（AWS S3 / MinIO / Cloudflare R2）
// ─────────────────────────────────────────────

/// S3 兼容存储驱动
///
/// 使用官方 `aws-sdk-s3` SDK 操作对象。
/// 支持 AWS S3、MinIO、Cloudflare R2、Backblaze B3 等 S3 兼容服务。
pub struct S3Driver {
    client: aws_sdk_s3::Client,
    bucket: String,
    public_url: String,
}

impl S3Driver {
    pub fn new(config: S3Config) -> Self {
        let region = aws_sdk_s3::config::Region::new(config.region.clone());
        let credentials = aws_credential_types::Credentials::new(
            &config.access_key,
            &config.secret_key,
            None,
            None,
            "static",
        );

        let mut sdk_config = aws_sdk_s3::config::Builder::new()
            .region(region)
            .credentials_provider(credentials);

        // 自定义 endpoint（MinIO、R2 等）
        if let Some(endpoint) = &config.endpoint {
            sdk_config = sdk_config.endpoint_url(endpoint.clone());
            // MinIO 和 R2 需要 path-style 寻址
            sdk_config = sdk_config.force_path_style(true);
        }

        let client = aws_sdk_s3::Client::from_conf(sdk_config.build());

        // 构造公开 URL
        let public_url = if let Some(endpoint) = &config.endpoint {
            // 自定义 endpoint：使用 path-style URL
            format!("{}/{}", endpoint.trim_end_matches('/'), config.bucket)
        } else {
            // AWS S3 标准 URL
            format!(
                "https://{}.s3.{}.amazonaws.com",
                config.bucket, config.region
            )
        };

        Self {
            client,
            bucket: config.bucket,
            public_url,
        }
    }
}

#[async_trait]
impl StorageDriver for S3Driver {
    fn name(&self) -> &'static str {
        "s3"
    }

    async fn upload(
        &self,
        key: &str,
        data: Bytes,
        content_type: &str,
    ) -> crate::AppResult<UploadResult> {
        let data_len = data.len() as i64;

        // 使用 ByteStream 流式上传
        let stream = aws_sdk_s3::primitives::ByteStream::from(data);

        self.client
            .put_object()
            .bucket(&self.bucket)
            .key(key)
            .body(stream)
            .content_type(content_type)
            .send()
            .await
            .map_err(|e| crate::error::AppError::External(format!("S3 上传失败: {}", e)))?;

        Ok(UploadResult {
            path: key.to_string(),
            url: self.url(key),
            size: data_len,
        })
    }

    async fn delete(&self, key: &str) -> crate::AppResult<()> {
        self.client
            .delete_object()
            .bucket(&self.bucket)
            .key(key)
            .send()
            .await
            .map_err(|e| {
                // NoSuchKey 视为成功（幂等删除）
                let err_str = format!("{}", e);
                if err_str.contains("NoSuchKey") || err_str.contains("NotFound") {
                    return crate::error::AppError::External("对象不存在（已删除）".to_string());
                }
                crate::error::AppError::External(format!("S3 删除失败: {}", e))
            })?;

        Ok(())
    }

    fn url(&self, key: &str) -> String {
        format!("{}/{}", self.public_url.trim_end_matches('/'), key)
    }

    async fn presign_put_url(
        &self,
        key: &str,
        expires_secs: i64,
    ) -> crate::AppResult<Option<String>> {
        use aws_sdk_s3::presigning::PresigningConfig;

        let config = PresigningConfig::expires_in(std::time::Duration::from_secs(
            expires_secs.max(1) as u64,
        ))
        .map_err(|e| crate::error::AppError::External(format!("S3 预签名配置失败: {}", e)))?;

        let presigned = self
            .client
            .put_object()
            .bucket(&self.bucket)
            .key(key)
            .presigned(config)
            .await
            .map_err(|e| crate::error::AppError::External(format!("S3 预签名失败: {}", e)))?;

        Ok(Some(presigned.uri().to_string()))
    }

    async fn exists(&self, key: &str) -> crate::AppResult<bool> {
        match self
            .client
            .head_object()
            .bucket(&self.bucket)
            .key(key)
            .send()
            .await
        {
            Ok(_) => Ok(true),
            Err(e) => {
                let err_str = format!("{}", e);
                if err_str.contains("NotFound") || err_str.contains("NoSuchKey") {
                    Ok(false)
                } else {
                    Err(crate::error::AppError::External(format!(
                        "S3 检查存在失败: {}",
                        e
                    )))
                }
            }
        }
    }
}

// ─────────────────────────────────────────────
// 阿里云 OSS 驱动
// ─────────────────────────────────────────────

/// 阿里云 OSS 驱动
pub struct OssDriver {
    pub config: OssConfig,
    client: reqwest::Client,
}

impl OssDriver {
    pub fn new(config: OssConfig) -> Self {
        Self {
            config,
            client: reqwest::Client::new(),
        }
    }

    fn object_url(&self, key: &str) -> String {
        format!(
            "https://{}.{}/{}",
            self.config.bucket,
            self.config.endpoint.trim_end_matches('/'),
            encode_key_path(key)
        )
    }

    fn host_base(&self) -> String {
        format!(
            "https://{}.{}",
            self.config.bucket,
            self.config.endpoint.trim_end_matches('/')
        )
    }

    /// 计算 OSS V1 签名
    ///
    /// StringToSign = VERB \\n Content-MD5 \\n Content-Type \\n Date \\n CanonicalizedOSSHeaders + CanonicalizedResource
    /// CanonicalizedResource = /{bucket}/{key}，key 不做 URL 编码。
    fn sign_request(
        &self,
        method: &str,
        key: &str,
        content_md5: &str,
        content_type: &str,
        date: &str,
    ) -> String {
        let resource = format!("/{}/{}", self.config.bucket, key);
        let string_to_sign = format!(
            "{}\n{}\n{}\n{}\n{}",
            method, content_md5, content_type, date, resource
        );

        let signature = hmac_sha1(self.config.secret_key.as_bytes(), string_to_sign.as_bytes());

        format!(
            "OSS {}:{}",
            self.config.access_key,
            base64::engine::general_purpose::STANDARD.encode(signature)
        )
    }

    fn http_date() -> String {
        chrono::Utc::now()
            .format("%a, %d %b %Y %H:%M:%S GMT")
            .to_string()
    }
}

#[async_trait]
impl StorageDriver for OssDriver {
    fn name(&self) -> &'static str {
        "oss"
    }

    async fn upload(
        &self,
        key: &str,
        data: Bytes,
        content_type: &str,
    ) -> crate::AppResult<UploadResult> {
        let url = self.object_url(key);
        let date = Self::http_date();

        // OSS 要求 Content-MD5 参与签名（base64 编码的 MD5 摘要）
        let digest = md5::compute(&data);
        let content_md5 = base64::engine::general_purpose::STANDARD.encode(digest.0);

        let auth = self.sign_request("PUT", key, &content_md5, content_type, &date);
        let data_len = data.len() as i64;

        let resp = self
            .client
            .put(&url)
            .header("Authorization", auth)
            .header("Content-Type", content_type)
            .header("Content-MD5", content_md5)
            .header("Date", date)
            .body(data)
            .send()
            .await
            .map_err(|e| crate::error::AppError::External(format!("OSS 上传失败: {}", e)))?;

        if !resp.status().is_success() {
            let status = resp.status();
            let body_text = resp.text().await.unwrap_or_default();
            return Err(crate::error::AppError::External(format!(
                "OSS 上传失败: HTTP {} - {}",
                status, body_text
            )));
        }

        Ok(UploadResult {
            path: key.to_string(),
            url: self.url(key),
            size: data_len,
        })
    }

    async fn delete(&self, key: &str) -> crate::AppResult<()> {
        let url = self.object_url(key);
        let date = Self::http_date();
        let auth = self.sign_request("DELETE", key, "", "", &date);

        let resp = self
            .client
            .delete(&url)
            .header("Authorization", auth)
            .header("Date", date)
            .send()
            .await
            .map_err(|e| crate::error::AppError::External(format!("OSS 删除失败: {}", e)))?;

        if !resp.status().is_success() && resp.status() != reqwest::StatusCode::NOT_FOUND {
            return Err(crate::error::AppError::External(format!(
                "OSS 删除失败: HTTP {}",
                resp.status()
            )));
        }

        Ok(())
    }

    fn url(&self, key: &str) -> String {
        self.object_url(key)
    }

    fn direct_upload_info(&self, _key: &str) -> Option<DirectUploadInfo> {
        // OSS PostObject：policy + HMAC-SHA1 签名，key 由前端上传时携带
        let expiration = (chrono::Utc::now() + chrono::Duration::hours(1)).to_rfc3339();
        let policy = serde_json::json!({
            "expiration": expiration,
            "conditions": [
                {"bucket": self.config.bucket},
                ["content-length-range", 1, 10485760],
                ["starts-with", "$key", ""],
            ]
        });
        let policy_b64 = base64::engine::general_purpose::STANDARD.encode(policy.to_string());
        let signature = base64::engine::general_purpose::STANDARD.encode(hmac_sha1(
            self.config.secret_key.as_bytes(),
            policy_b64.as_bytes(),
        ));

        Some(DirectUploadInfo {
            upload_url: format!("{}/", self.host_base()),
            public_base: self.host_base(),
            extra: serde_json::json!({
                "method": "POST",
                "access_key_id": self.config.access_key,
                "policy": policy_b64,
                "signature": signature,
            }),
        })
    }
}

// ─────────────────────────────────────────────
// 腾讯云 COS 驱动
// ─────────────────────────────────────────────

/// 腾讯云 COS 驱动
pub struct CosDriver {
    pub config: CosConfig,
    client: reqwest::Client,
}

impl CosDriver {
    pub fn new(config: CosConfig) -> Self {
        Self {
            config,
            client: reqwest::Client::new(),
        }
    }

    fn object_url(&self, key: &str) -> String {
        format!(
            "https://{}.cos.{}.myqcloud.com/{}",
            self.config.bucket,
            self.config.region,
            encode_key_path(key)
        )
    }
}

/// 计算 COS 签名核心（返回 q-signature）
///
/// SignKey = HMAC-SHA1(SecretKey, KeyTime) 的 hex；
/// StringToSign = "sha1\\n{KeyTime}\\n{Sha1(HttpString)}\\n"，
/// HttpString = "{Method}\\n{Uri}\\n{Params}\\n{Headers}\\n"。
fn cos_signature(secret_key: &str, method: &str, uri: &str, key_time: &str) -> String {
    let sign_key = hex::encode(hmac_sha1(secret_key.as_bytes(), key_time.as_bytes()));
    let http_string = format!("{}\n{}\n\n\n", method.to_lowercase(), uri);
    let sha1_http = hex::encode(Sha1::digest(http_string.as_bytes()));
    let string_to_sign = format!("sha1\n{}\n{}\n", key_time, sha1_http);
    hex::encode(hmac_sha1(sign_key.as_bytes(), string_to_sign.as_bytes()))
}

impl CosDriver {
    /// 生成 COS Authorization 头（PUT/DELETE 等服务端请求使用）
    fn auth_header(&self, method: &str, key: &str, expires_secs: i64) -> String {
        let now = chrono::Utc::now().timestamp();
        let key_time = format!("{};{}", now - 60, now + expires_secs);
        let uri = format!("/{}", encode_key_path(key));
        let signature = cos_signature(&self.config.secret_key, method, &uri, &key_time);

        format!(
            "q-sign-algorithm=sha1&q-ak={}&q-sign-time={}&q-key-time={}&q-header-list=&q-url-param-list=&q-signature={}",
            self.config.secret_id, key_time, key_time, signature
        )
    }

    /// 生成携带查询签名的 PUT URL（客户端直传使用）
    fn presigned_query_url(&self, key: &str, expires_secs: i64) -> String {
        format!(
            "{}?{}",
            self.object_url(key),
            self.auth_header("put", key, expires_secs)
        )
    }
}

#[async_trait]
impl StorageDriver for CosDriver {
    fn name(&self) -> &'static str {
        "cos"
    }

    async fn upload(
        &self,
        key: &str,
        data: Bytes,
        content_type: &str,
    ) -> crate::AppResult<UploadResult> {
        let url = self.object_url(key);
        let auth = self.auth_header("put", key, 3600);
        let data_len = data.len() as i64;

        let resp = self
            .client
            .put(&url)
            .header("Authorization", auth)
            .header("Content-Type", content_type)
            .body(data)
            .send()
            .await
            .map_err(|e| crate::error::AppError::External(format!("COS 上传失败: {}", e)))?;

        if !resp.status().is_success() {
            let status = resp.status();
            let body_text = resp.text().await.unwrap_or_default();
            return Err(crate::error::AppError::External(format!(
                "COS 上传失败: HTTP {} - {}",
                status, body_text
            )));
        }

        Ok(UploadResult {
            path: key.to_string(),
            url: self.url(key),
            size: data_len,
        })
    }

    async fn delete(&self, key: &str) -> crate::AppResult<()> {
        let url = self.object_url(key);
        let auth = self.auth_header("delete", key, 3600);

        let resp = self
            .client
            .delete(&url)
            .header("Authorization", auth)
            .send()
            .await
            .map_err(|e| crate::error::AppError::External(format!("COS 删除失败: {}", e)))?;

        if !resp.status().is_success() && resp.status() != reqwest::StatusCode::NOT_FOUND {
            return Err(crate::error::AppError::External(format!(
                "COS 删除失败: HTTP {}",
                resp.status()
            )));
        }

        Ok(())
    }

    fn url(&self, key: &str) -> String {
        self.object_url(key)
    }

    async fn presign_put_url(
        &self,
        key: &str,
        expires_secs: i64,
    ) -> crate::AppResult<Option<String>> {
        // COS 支持将签名放入查询串（与 Authorization 头同构）
        Ok(Some(self.presigned_query_url(key, expires_secs)))
    }
}

// ─────────────────────────────────────────────
// 七牛云驱动
// ─────────────────────────────────────────────

/// 七牛云存储驱动
///
/// 使用七牛云官方 API 进行上传、删除操作。
/// 参考文档：https://developer.qiniu.com/kodo/1239/python
pub struct QiniuDriver {
    config: crate::config::QiniuConfig,
    http: reqwest::Client,
    upload_endpoint: String,
    rs_host: String, // 管理 API 主机
}

impl QiniuDriver {
    pub fn new(config: crate::config::QiniuConfig) -> Self {
        // 根据 zone 选择上传域名
        let upload_endpoint = match config.zone.as_str() {
            "z1" => "https://upload-z1.qiniup.com".to_string(), // 华北
            "z2" => "https://upload-z2.qiniup.com".to_string(), // 华南
            "na0" => "https://upload-na0.qiniup.com".to_string(), // 北美
            "as0" => "https://upload-as0.qiniup.com".to_string(), // 东南亚
            "cn-east-2" => "https://upload-cn-east-2.qiniup.com".to_string(), // 华东-浙江2
            _ => "https://upload.qiniup.com".to_string(),       // 默认华东 z0
        };

        Self {
            config,
            http: reqwest::Client::new(),
            upload_endpoint,
            rs_host: "https://rs.qiniu.com".to_string(),
        }
    }

    /// 生成七牛上传 token（URL_SAFE_NO_PAD base64 签名的 put_policy）
    ///
    /// `scope` 为 `"bucket"` 或 `"bucket:key"`。
    fn upload_token_with_scope(&self, scope: &str, expires_in: i64) -> String {
        let deadline = chrono::Utc::now().timestamp() + expires_in;

        // put_policy（JSON）
        let put_policy = serde_json::json!({
            "scope": scope,
            "deadline": deadline,
        });

        let encoded_policy = base64::engine::general_purpose::URL_SAFE_NO_PAD
            .encode(put_policy.to_string().as_bytes());

        // HMAC-SHA1 签名
        let signature = hmac_sha1(self.config.secret_key.as_bytes(), encoded_policy.as_bytes());
        let encoded_signature = base64::engine::general_purpose::URL_SAFE_NO_PAD.encode(signature);

        format!(
            "{}:{}:{}",
            self.config.access_key, encoded_signature, encoded_policy
        )
    }

    /// 生成指定 key 的上传 token
    fn upload_token(&self, key: &str, expires_in: i64) -> String {
        self.upload_token_with_scope(&format!("{}:{}", self.config.bucket, key), expires_in)
    }

    /// 公开访问 URL 前缀
    fn domain_base(&self) -> String {
        format!("https://{}", self.config.domain.trim_end_matches('/'))
    }

    /// 生成管理 API 的 Authorization token
    fn manage_token(&self, path: &str, body: &str) -> String {
        let sign_str = format!("{}\n{}", path, body);
        let signature = hmac_sha1(self.config.secret_key.as_bytes(), sign_str.as_bytes());
        let encoded_signature = base64::engine::general_purpose::URL_SAFE_NO_PAD.encode(signature);

        format!("QBox {}:{}", self.config.access_key, encoded_signature)
    }
}

#[async_trait]
impl StorageDriver for QiniuDriver {
    fn name(&self) -> &'static str {
        "qiniu"
    }

    async fn upload(
        &self,
        key: &str,
        data: Bytes,
        content_type: &str,
    ) -> crate::AppResult<UploadResult> {
        let data_len = data.len() as i64;
        let token = self.upload_token(key, 3600); // 1小时有效期

        // multipart 表单上传
        let form = reqwest::multipart::Form::new()
            .text("key", key.to_string())
            .text("token", token)
            .part(
                "file",
                reqwest::multipart::Part::bytes(data.to_vec())
                    .file_name(key.to_string())
                    .mime_str(content_type)
                    .map_err(|e| {
                        crate::error::AppError::External(format!("七牛 MIME 错误: {}", e))
                    })?,
            );

        let resp = self
            .http
            .post(&self.upload_endpoint)
            .multipart(form)
            .send()
            .await
            .map_err(|e| crate::error::AppError::External(format!("七牛上传失败: {}", e)))?;

        if !resp.status().is_success() {
            let status = resp.status();
            let body_text = resp.text().await.unwrap_or_default();
            return Err(crate::error::AppError::External(format!(
                "七牛上传失败: HTTP {} - {}",
                status, body_text
            )));
        }

        Ok(UploadResult {
            path: key.to_string(),
            url: self.url(key),
            size: data_len,
        })
    }

    async fn delete(&self, key: &str) -> crate::AppResult<()> {
        // 七牛删除 API：POST /delete/<EncodedEntryURI>
        let entry = format!("{}:{}", self.config.bucket, key);
        let encoded_entry =
            base64::engine::general_purpose::URL_SAFE_NO_PAD.encode(entry.as_bytes());

        let path = format!("/delete/{}", encoded_entry);
        let token = self.manage_token(&path, "");

        let resp = self
            .http
            .post(format!("{}{}", self.rs_host, path))
            .header("Authorization", token)
            .header("Content-Type", "application/x-www-form-urlencoded")
            .send()
            .await
            .map_err(|e| crate::error::AppError::External(format!("七牛删除失败: {}", e)))?;

        if !resp.status().is_success() {
            let status = resp.status();
            let body_text = resp.text().await.unwrap_or_default();
            // 612 = 文件不存在，视为成功
            if body_text.contains("612") {
                return Ok(());
            }
            return Err(crate::error::AppError::External(format!(
                "七牛删除失败: HTTP {} - {}",
                status, body_text
            )));
        }

        Ok(())
    }

    fn url(&self, key: &str) -> String {
        // 七牛需要绑定自定义域名才能访问
        // 公开 URL 格式：https://<domain>/<key>
        format!("{}/{}", self.domain_base(), key)
    }

    fn direct_upload_info(&self, _key: &str) -> Option<DirectUploadInfo> {
        // scope 仅指定 bucket，允许客户端携带任意 key 直传
        let token = self.upload_token_with_scope(&self.config.bucket, 3600);
        Some(DirectUploadInfo {
            upload_url: self.upload_endpoint.clone(),
            public_base: self.domain_base(),
            extra: serde_json::json!({
                "method": "POST",
                "upload_token": token,
            }),
        })
    }
}

// ─────────────────────────────────────────────
// 工厂函数：根据配置创建驱动
// ─────────────────────────────────────────────

use crate::config::StorageConfig;

/// 根据存储配置创建对应的驱动实例
pub fn create_driver(config: &StorageConfig) -> crate::AppResult<Arc<dyn StorageDriver>> {
    let driver: Arc<dyn StorageDriver> = match config.driver {
        crate::config::StorageDriver::Local => {
            let root = config
                .root
                .clone()
                .unwrap_or_else(|| "./uploads".to_string());
            let public_url = config.url.clone().unwrap_or_default();
            Arc::new(LocalDriver { root, public_url })
        }
        crate::config::StorageDriver::S3 => {
            let cfg = config
                .s3
                .clone()
                .ok_or_else(|| crate::error::AppError::Config("缺少 S3 配置".to_string()))?;
            Arc::new(S3Driver::new(cfg))
        }
        crate::config::StorageDriver::Oss => {
            let cfg = config
                .oss
                .clone()
                .ok_or_else(|| crate::error::AppError::Config("缺少 OSS 配置".to_string()))?;
            Arc::new(OssDriver::new(cfg))
        }
        crate::config::StorageDriver::Cos => {
            let cfg = config
                .cos
                .clone()
                .ok_or_else(|| crate::error::AppError::Config("缺少 COS 配置".to_string()))?;
            Arc::new(CosDriver::new(cfg))
        }
        crate::config::StorageDriver::Qiniu => {
            let cfg = config
                .qiniu
                .clone()
                .ok_or_else(|| crate::error::AppError::Config("缺少七牛配置".to_string()))?;
            Arc::new(QiniuDriver::new(cfg))
        }
    };
    Ok(driver)
}

// ─────────────────────────────────────────────
// 辅助函数
// ─────────────────────────────────────────────
