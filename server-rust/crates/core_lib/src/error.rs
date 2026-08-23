//! 统一错误类型与响应
//!
//! 成功响应使用 `crate::utils::response::ApiResponse`，
//! 错误响应使用本文件的 `AppError`（两者信封格式一致）。
//!
//! 统一响应信封格式（前端依赖此结构）：
//! - 成功：`{"code": 0, "message": "ok", "data": ...}`
//! - 分页：`{"code": 0, "message": "ok", "data": [...], "meta": {...}}`
//! - 错误：`{"code": <int>, "message": "..."}`

use axum::{
    response::{IntoResponse, Response},
    Json,
};
use http::StatusCode;
use thiserror::Error;

pub type AppResult<T> = Result<T, AppError>;

/// 业务错误码
#[derive(Error, Debug)]
pub enum AppError {
    #[error("数据库错误: {0}")]
    Database(#[from] sqlx::Error),

    #[error("Redis 错误: {0}")]
    Redis(#[from] redis::RedisError),

    #[error("验证错误: {0}")]
    Validation(String),

    #[error("认证失败: {0}")]
    Auth(String),

    #[error("未授权访问")]
    Forbidden,

    #[error("资源未找到: {0}")]
    NotFound(String),

    #[error("业务错误: {0}")]
    Business(String),

    #[error("内部服务器错误: {0}")]
    Internal(String),

    #[error("外部服务错误: {0}")]
    External(String),

    #[error("配置错误: {0}")]
    Config(String),

    #[error("IO 错误: {0}")]
    Io(#[from] std::io::Error),

    #[error("功能未实现: {0}")]
    NotImplemented(String),
}

impl AppError {
    /// HTTP 状态码
    /// 遵循标准 HTTP 语义：每种错误映射到对应的状态码，
    /// 同时 JSON body 中保留业务错误码（code）供前端细粒度判断。
    pub fn status_code(&self) -> StatusCode {
        match self {
            AppError::Database(_) => StatusCode::INTERNAL_SERVER_ERROR,
            AppError::Redis(_) => StatusCode::INTERNAL_SERVER_ERROR,
            AppError::Internal(_) => StatusCode::INTERNAL_SERVER_ERROR,
            AppError::External(_) => StatusCode::BAD_GATEWAY,
            AppError::Config(_) => StatusCode::INTERNAL_SERVER_ERROR,
            AppError::Io(_) => StatusCode::INTERNAL_SERVER_ERROR,
            AppError::NotImplemented(_) => StatusCode::NOT_IMPLEMENTED,
            AppError::Auth(_) => StatusCode::UNAUTHORIZED,
            AppError::Forbidden => StatusCode::FORBIDDEN,
            AppError::NotFound(_) => StatusCode::NOT_FOUND,
            AppError::Validation(_) => StatusCode::BAD_REQUEST,
            AppError::Business(_) => StatusCode::BAD_REQUEST,
        }
    }

    /// 业务错误码
    pub fn error_code(&self) -> i32 {
        match self {
            AppError::Database(_) => 50000,
            AppError::Redis(_) => 50000,
            AppError::Validation(_) => 40000,
            AppError::Auth(_) => 40100,
            AppError::Forbidden => 40300,
            AppError::NotFound(_) => 40400,
            AppError::Business(_) => 40000,
            AppError::Internal(_) => 50000,
            AppError::External(_) => 50200,
            AppError::Config(_) => 50000,
            AppError::Io(_) => 50000,
            AppError::NotImplemented(_) => 50100,
        }
    }
}

impl IntoResponse for AppError {
    fn into_response(self) -> Response {
        let status = self.status_code();
        let code = self.error_code();

        let message = match &self {
            AppError::Database(e) => {
                tracing::error!(error = %e, "数据库错误（已脱敏返回）");
                "服务器内部错误".to_string()
            }
            AppError::Redis(e) => {
                tracing::error!(error = %e, "Redis 错误（已脱敏返回）");
                "服务器内部错误".to_string()
            }
            AppError::Internal(e) => {
                tracing::error!(error = %e, "内部错误（已脱敏返回）");
                "服务器内部错误".to_string()
            }
            AppError::External(e) => {
                tracing::error!(error = %e, "外部服务错误（已脱敏返回）");
                "服务器内部错误".to_string()
            }
            AppError::Config(e) => {
                tracing::error!(error = %e, "配置错误（已脱敏返回）");
                "服务器内部错误".to_string()
            }
            AppError::Io(e) => {
                tracing::error!(error = %e, "IO 错误（已脱敏返回）");
                "服务器内部错误".to_string()
            }
            other => other.to_string(),
        };

        let body = crate::utils::response::ApiResponse::<serde_json::Value> {
            code,
            message,
            data: None,
        };
        (status, Json(body)).into_response()
    }
}
