use serde::{Deserialize, Serialize};

/// 分页参数
#[derive(Debug, Clone, Deserialize)]
pub struct Pagination {
    #[serde(default = "default_page")]
    pub page: u64,
    #[serde(default = "default_per_page")]
    pub per_page: u64,
    /// 可选筛选：按相册 ID 过滤（图片列表）
    pub album_id: Option<i64>,
}

fn default_page() -> u64 {
    1
}

fn default_per_page() -> u64 {
    20
}

impl Pagination {
    pub fn new(page: u64, per_page: u64) -> Self {
        Self {
            page: page.max(1),
            per_page: per_page.clamp(1, 100),
            album_id: None,
        }
    }

    pub fn offset(&self) -> u64 {
        (self.page - 1) * self.per_page
    }

    pub fn limit(&self) -> u64 {
        self.per_page
    }
}

impl Default for Pagination {
    fn default() -> Self {
        Self {
            page: 1,
            per_page: 20,
            album_id: None,
        }
    }
}

/// 分页响应
#[derive(Debug, Clone, Serialize)]
pub struct PaginatedResponse<T> {
    pub items: Vec<T>,
    pub total: u64,
    pub page: u64,
    pub per_page: u64,
    pub total_pages: u64,
}

impl<T> PaginatedResponse<T> {
    pub fn new(items: Vec<T>, total: u64, pagination: &Pagination) -> Self {
        let total_pages = (total as f64 / pagination.per_page as f64).ceil() as u64;
        Self {
            items,
            total,
            page: pagination.page,
            per_page: pagination.per_page,
            total_pages,
        }
    }
}
