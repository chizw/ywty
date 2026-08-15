//! 请求 / 响应 DTO
//!
//! 对应 Go 后端 service 包中的 Request/Response 结构体。
//! 验证用请求复用 `crate::models::*` 中已定义的结构体，
//! 响应 DTO 在本模块中定义（避免泄露内部字段如 password）。

pub mod album;
pub mod auth;
pub mod photo;
pub mod storage;
pub mod user;

use serde::Serialize;

/// 分页响应元数据（对齐 Go 的 `meta{current_page,per_page,total,last_page}`）
#[derive(Debug, Clone, Serialize)]
pub struct Meta {
    pub current_page: u64,
    pub per_page: u64,
    pub total: u64,
    pub last_page: u64,
}

/// 分页数据信封
#[derive(Debug, Clone, Serialize)]
pub struct PaginatedData<T: Serialize> {
    pub data: Vec<T>,
    pub meta: Meta,
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
