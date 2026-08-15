use axum::{
    response::{IntoResponse, Response},
    Json,
};
use http::StatusCode;
use serde::Serialize;
use serde_json::json;

/// 统一 API 响应
#[derive(Debug, Clone, Serialize)]
pub struct ApiResponse<T: Serialize> {
    pub code: String,
    pub data: Option<T>,
    pub message: Option<String>,
}

impl<T: Serialize> ApiResponse<T> {
    pub fn success(data: T) -> Self {
        Self {
            code: "SUCCESS".to_string(),
            data: Some(data),
            message: None,
        }
    }

    pub fn success_with_message(data: T, message: &str) -> Self {
        Self {
            code: "SUCCESS".to_string(),
            data: Some(data),
            message: Some(message.to_string()),
        }
    }

    pub fn error(code: &str, message: &str) -> Self {
        Self {
            code: code.to_string(),
            data: None,
            message: Some(message.to_string()),
        }
    }
}

impl<T: Serialize> IntoResponse for ApiResponse<T> {
    fn into_response(self) -> Response {
        let status = if self.code == "SUCCESS" {
            StatusCode::OK
        } else {
            StatusCode::BAD_REQUEST
        };
        let body = Json(json!({
            "code": self.code,
            "data": self.data,
            "message": self.message,
        }));
        (status, body).into_response()
    }
}
