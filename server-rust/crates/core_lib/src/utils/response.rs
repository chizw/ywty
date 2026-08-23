//! 统一 API 响应信封
//!
//! 前端依赖此结构，成功与错误响应使用统一格式：
//!
//! - 成功：`{"code": 0, "message": "ok", "data": ...}`
//! - 分页：`{"code": 0, "message": "ok", "data": [...], "meta": {...}}`
//! - 错误：`{"code": <int>, "message": "...", "data": null}`

use serde::Serialize;
use utoipa::ToSchema;

/// 成功业务码
pub const CODE_OK: i32 = 0;
/// 成功消息
pub const MSG_OK: &str = "ok";

/// 标准 API 响应信封
#[derive(Debug, Clone, Serialize, ToSchema)]
pub struct ApiResponse<T: Serialize> {
    pub code: i32,
    pub message: String,
    pub data: Option<T>,
}

impl<T: Serialize> ApiResponse<T> {
    /// 成功响应
    pub fn success(data: T) -> Self {
        Self {
            code: CODE_OK,
            message: MSG_OK.to_string(),
            data: Some(data),
        }
    }

    /// 成功响应（带自定义消息）
    pub fn success_with_message(data: T, message: &str) -> Self {
        Self {
            code: CODE_OK,
            message: message.to_string(),
            data: Some(data),
        }
    }
}

/// 分页数据信封
#[derive(Debug, Clone, Serialize)]
pub struct PageMeta {
    pub current_page: i64,
    pub per_page: i64,
    pub total: i64,
    pub last_page: i64,
}

/// 分页响应信封
#[derive(Debug, Clone, Serialize)]
pub struct PaginatedResponse<T: Serialize> {
    pub code: i32,
    pub message: String,
    pub data: Vec<T>,
    pub meta: PageMeta,
}

impl<T: Serialize> PaginatedResponse<T> {
    pub fn new(data: Vec<T>, total: i64, page: i64, per_page: i64) -> Self {
        let last_page = if per_page == 0 {
            0
        } else {
            (total as f64 / per_page as f64).ceil() as i64
        };
        Self {
            code: CODE_OK,
            message: MSG_OK.to_string(),
            data,
            meta: PageMeta {
                current_page: page,
                per_page,
                total,
                last_page,
            },
        }
    }
}

/// 便捷函数：构造成功响应
pub fn ok<T: Serialize>(data: T) -> ApiResponse<T> {
    ApiResponse::success(data)
}

/// 便捷函数：构造成功响应（带消息）
pub fn ok_msg<T: Serialize>(data: T, msg: &str) -> ApiResponse<T> {
    ApiResponse::success_with_message(data, msg)
}

/// 从旧版 `PaginatedData` 迁移的兼容类型
#[derive(Debug, Clone, Serialize)]
pub struct PaginatedData<T: Serialize> {
    pub data: Vec<T>,
    pub meta: Meta,
}

#[derive(Debug, Clone, Serialize)]
pub struct Meta {
    pub current_page: u64,
    pub per_page: u64,
    pub total: u64,
    pub last_page: u64,
}

impl<T: Serialize> PaginatedData<T> {
    pub fn new(data: Vec<T>, total: u64, page: u64, per_page: u64) -> Self {
        let last_page = if per_page == 0 {
            0
        } else {
            (total as f64 / per_page as f64).ceil() as u64
        };
        Self {
            data,
            meta: Meta {
                current_page: page,
                per_page,
                total,
                last_page,
            },
        }
    }
}
