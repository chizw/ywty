//! 订单处理器

use axum::body::Bytes;
use axum::extract::{Path, Query, State};
use axum::http::HeaderMap;
use axum::Json;

use crate::error::{AppError, AppResult};
use crate::handlers::{validate_req, CurrentUser};
use crate::models::order::CreateOrderRequest;
use crate::services::order::PaidOutcome;
use crate::services::payment::{MockDriver, PaymentDriver, SIGNATURE_HEADER};
use crate::utils::response::ApiResponse;
use crate::AppState;

#[derive(Debug, serde::Deserialize)]
pub struct OrderListQuery {
    pub page: Option<i32>,
    pub per_page: Option<i32>,
}

#[utoipa::path(
    post,
    path = "/api/v1/orders",
    request_body = CreateOrderRequest,
    responses(
        (status = 200, description = "成功", body = serde_json::Value),
    ),
    tag = "订单"
)]
/// 创建订单
pub async fn create(
    State(state): State<AppState>,
    CurrentUser { user_id, .. }: CurrentUser,
    Json(req): Json<CreateOrderRequest>,
) -> AppResult<Json<ApiResponse<serde_json::Value>>> {
    validate_req(&req)?;
    let order = state.order_svc.create(user_id, &req).await?;
    Ok(Json(ApiResponse::success(
        serde_json::json!({ "data": order }),
    )))
}

#[utoipa::path(
    get,
    path = "/api/v1/orders",
    params(
        ("page" = Option<i32>, Query, description = "页码"),
        ("per_page" = Option<i32>, Query, description = "每页数量"),
    ),
    responses(
        (status = 200, description = "成功", body = serde_json::Value),
    ),
    tag = "订单"
)]
/// 列出我的订单
pub async fn list(
    State(state): State<AppState>,
    CurrentUser { user_id, .. }: CurrentUser,
    Query(query): Query<OrderListQuery>,
) -> AppResult<Json<ApiResponse<serde_json::Value>>> {
    let orders = state
        .order_svc
        .list(
            user_id,
            query.page.unwrap_or(1),
            query.per_page.unwrap_or(20),
        )
        .await?;
    Ok(Json(ApiResponse::success(
        serde_json::json!({ "data": orders }),
    )))
}

#[utoipa::path(
    get,
    path = "/api/v1/orders/:id",
    params(
        ("id" = i64, Path, description = "订单ID"),
    ),
    responses(
        (status = 200, description = "成功", body = serde_json::Value),
    ),
    tag = "订单"
)]
/// 获取订单详情
pub async fn get(
    State(state): State<AppState>,
    CurrentUser { user_id, .. }: CurrentUser,
    Path(id): Path<i64>,
) -> AppResult<Json<ApiResponse<serde_json::Value>>> {
    let order = state.order_svc.get(user_id, id).await?;
    Ok(Json(ApiResponse::success(
        serde_json::json!({ "data": order }),
    )))
}

#[utoipa::path(
    post,
    path = "/api/v1/orders/:id/cancel",
    params(
        ("id" = i64, Path, description = "订单ID"),
    ),
    responses(
        (status = 200, description = "成功", body = serde_json::Value),
    ),
    tag = "订单"
)]
/// 取消订单
pub async fn cancel(
    State(state): State<AppState>,
    CurrentUser { user_id, .. }: CurrentUser,
    Path(id): Path<i64>,
) -> AppResult<Json<ApiResponse<serde_json::Value>>> {
    state.order_svc.cancel(user_id, id).await?;
    Ok(Json(ApiResponse::success(
        serde_json::json!({ "message": "订单已取消" }),
    )))
}

#[utoipa::path(
    post,
    path = "/api/v1/orders/:id/pay",
    params(
        ("id" = i64, Path, description = "订单ID"),
    ),
    responses(
        (status = 200, description = "成功", body = serde_json::Value),
    ),
    tag = "订单"
)]
/// 发起支付（返回收银台地址等支付参数，Mock 模式附带预签名回调）
pub async fn pay(
    State(state): State<AppState>,
    CurrentUser { user_id, .. }: CurrentUser,
    Path(id): Path<i64>,
) -> AppResult<Json<ApiResponse<serde_json::Value>>> {
    let driver = payment_driver(&state);
    let (order, params) = state.order_svc.pay_initiate(user_id, id, &driver).await?;
    Ok(Json(ApiResponse::success(
        serde_json::json!({ "data": { "order": order, "payment": params } }),
    )))
}

#[utoipa::path(
    post,
    path = "/api/v1/orders/notify",
    request_body = serde_json::Value,
    responses(
        (status = 200, description = "成功", body = serde_json::Value),
    ),
    tag = "订单"
)]
/// 支付回调（公开，需 HMAC-SHA256 验签）
///
/// - 请求头 `X-Signature`：`hex(HMAC_SHA256(secret, "{trade_no}|{amount}|{timestamp}"))`
/// - 请求体：`{"trade_no":"...","amount":123,"timestamp":unix秒}`
/// - 仅 `unpaid -> paid` 允许转换；已支付重复通知幂等返回成功；金额必须一致。
pub async fn notify(
    State(state): State<AppState>,
    headers: HeaderMap,
    body: Bytes,
) -> AppResult<Json<ApiResponse<serde_json::Value>>> {
    let signature = headers
        .get(SIGNATURE_HEADER)
        .and_then(|v| v.to_str().ok())
        .unwrap_or_default();
    if signature.is_empty() {
        return Err(AppError::Validation("缺少 X-Signature 头".to_string()));
    }

    let driver = payment_driver(&state);
    let payload = driver.verify_notify(&body, signature)?;

    match state
        .order_svc
        .handle_notify(&payload, driver.name())
        .await?
    {
        PaidOutcome::AlreadyPaid(order) => Ok(Json(ApiResponse::success(
            serde_json::json!({ "message": "订单已支付", "data": order }),
        ))),
        PaidOutcome::Paid(order) => Ok(Json(ApiResponse::success(
            serde_json::json!({ "message": "支付成功", "data": order }),
        ))),
    }
}

/// 构建支付驱动
///
/// 支付密钥暂取 `config.auth.jwt.secret`（config 暂无独立 payment_secret 配置项），
/// 后续新增配置后仅需替换此处。
fn payment_driver(state: &AppState) -> MockDriver {
    MockDriver::new(state.config.auth.jwt.secret.clone())
}
