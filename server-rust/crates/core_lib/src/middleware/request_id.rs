use axum::{extract::Request, middleware::Next, response::Response};
use http::HeaderValue;
use uuid::Uuid;

const REQUEST_ID_HEADER: &str = "x-request-id";

/// 请求 ID 中间件 - 为每个请求生成唯一 ID
pub async fn request_id_middleware(mut request: Request, next: Next) -> Response {
    let request_id = Uuid::new_v4().to_string();

    // 添加到 request extensions
    request.extensions_mut().insert(request_id.clone());

    let mut response = next.run(request).await;

    // 添加到 response headers
    if let Ok(header_value) = HeaderValue::from_str(&request_id) {
        response
            .headers_mut()
            .insert(REQUEST_ID_HEADER, header_value);
    }

    response
}
