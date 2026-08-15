//! 统一错误类型与响应
//!
//! 成功响应使用 `crate::utils::response::ApiResponse`，
//! 错误响应使用本文件的 `AppError`（两者信封格式一致）。

use axum::{
    response::{IntoResponse, Response},
    Json,
};
use http::StatusCode;
use serde_json::json;
use thiserror::Error;

pub type AppResult<T> = Result<T, AppError>;

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
}

impl AppError {
    pub fn status_code(&self) -> StatusCode {
        match self {
            AppError::Database(_) => StatusCode::INTERNAL_SERVER_ERROR,
            AppError::Redis(_) => StatusCode::INTERNAL_SERVER_ERROR,
            AppError::Validation(_) => StatusCode::BAD_REQUEST,
            AppError::Auth(_) => StatusCode::UNAUTHORIZED,
            AppError::Forbidden => StatusCode::FORBIDDEN,
            AppError::NotFound(_) => StatusCode::NOT_FOUND,
            AppError::Business(_) => StatusCode::BAD_REQUEST,
            AppError::Internal(_) => StatusCode::INTERNAL_SERVER_ERROR,
            AppError::External(_) => StatusCode::BAD_GATEWAY,
            AppError::Config(_) => StatusCode::INTERNAL_SERVER_ERROR,
            AppError::Io(_) => StatusCode::INTERNAL_SERVER_ERROR,
        }
    }

    pub fn error_code(&self) -> &'static str {
        match self {
            AppError::Database(_) => "DATABASE_ERROR",
            AppError::Redis(_) => "REDIS_ERROR",
            AppError::Validation(_) => "VALIDATION_ERROR",
            AppError::Auth(_) => "AUTH_ERROR",
            AppError::Forbidden => "FORBIDDEN",
            AppError::NotFound(_) => "NOT_FOUND",
            AppError::Business(_) => "BUSINESS_ERROR",
            AppError::Internal(_) => "INTERNAL_ERROR",
            AppError::External(_) => "EXTERNAL_ERROR",
            AppError::Config(_) => "CONFIG_ERROR",
            AppError::Io(_) => "IO_ERROR",
        }
    }
}

impl IntoResponse for AppError {
    fn into_response(self) -> Response {
        let status = self.status_code();
        let body = Json(json!({
            "code": self.error_code(),
            "message": self.to_string(),
        }));

        (status, body).into_response()
    }
}
