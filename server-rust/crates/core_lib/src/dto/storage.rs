//! 存储域 DTO

use serde::Serialize;

/// 上传签名响应（用于前端直传）
#[derive(Debug, Clone, Serialize)]
pub struct StorageSignResponse {
    /// 上传 URL（本地驱动时为相对路径前缀）
    pub upload_url: String,
    /// 文件访问 URL 前缀
    pub public_url: String,
    /// 存储驱动
    pub driver: String,
    /// 单次上传限制（字节）
    pub max_size: i64,
    /// 允许的 MIME 类型
    pub allowed_types: Vec<String>,
}
