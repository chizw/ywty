//! 优惠券处理器

use axum::extract::{Path, Query, State};
use axum::Json;

use crate::error::AppResult;
use crate::handlers::{validate_req, CurrentUser};
use crate::models::coupon::{AdminCouponRequest, ValidateCouponRequest};
use crate::utils::response::ApiResponse;
use crate::AppState;

#[derive(Debug, serde::Deserialize)]
pub struct PaginationQuery {
    pub page: Option<i32>,
    pub per_page: Option<i32>,
}

#[utoipa::path(
    post,
    path = "/api/v1/coupons/validate",
    request_body = ValidateCouponRequest,
    responses(
        (status = 200, description = "成功", body = serde_json::Value),
    ),
    tag = "优惠券"
)]
/// 用户侧：校验优惠券
pub async fn validate(
    State(state): State<AppState>,
    Json(req): Json<ValidateCouponRequest>,
) -> AppResult<Json<ApiResponse<serde_json::Value>>> {
    validate_req(&req)?;
    let result = state.coupon_svc.validate(&req).await?;
    Ok(Json(ApiResponse::success(
        serde_json::json!({ "data": result }),
    )))
}

#[utoipa::path(
    get,
    path = "/api/v1/admin/coupons",
    params(
        ("page" = Option<i32>, Query, description = "页码"),
        ("per_page" = Option<i32>, Query, description = "每页数量"),
    ),
    responses(
        (status = 200, description = "成功", body = serde_json::Value),
    ),
    tag = "优惠券"
)]
/// 管理端：列出优惠券
pub async fn admin_list(
    State(state): State<AppState>,
    Query(query): Query<PaginationQuery>,
) -> AppResult<Json<serde_json::Value>> {
    let (coupons, total) = state
        .coupon_svc
        .list(query.page.unwrap_or(1), query.per_page.unwrap_or(20))
        .await?;
    Ok(Json(serde_json::json!({
        "data": coupons,
        "meta": {
            "current_page": query.page.unwrap_or(1),
            "per_page": query.per_page.unwrap_or(20),
            "total": total,
            "last_page": (total + query.per_page.unwrap_or(20) as i64 - 1) / query.per_page.unwrap_or(20) as i64,
        }
    })))
}

#[utoipa::path(
    get,
    path = "/api/v1/admin/coupons/:id",
    params(
        ("id" = i64, Path, description = "优惠券ID"),
    ),
    responses(
        (status = 200, description = "成功", body = serde_json::Value),
    ),
    tag = "优惠券"
)]
/// 管理端：获取优惠券详情
pub async fn admin_get(
    State(state): State<AppState>,
    Path(id): Path<i64>,
) -> AppResult<Json<ApiResponse<serde_json::Value>>> {
    let coupon = state.coupon_svc.get(id).await?;
    Ok(Json(ApiResponse::success(
        serde_json::json!({ "data": coupon }),
    )))
}

#[utoipa::path(
    post,
    path = "/api/v1/admin/coupons",
    request_body = AdminCouponRequest,
    responses(
        (status = 200, description = "成功", body = serde_json::Value),
    ),
    tag = "优惠券"
)]
/// 管理端：创建优惠券
pub async fn admin_create(
    State(state): State<AppState>,
    CurrentUser { .. }: CurrentUser,
    Json(req): Json<AdminCouponRequest>,
) -> AppResult<Json<ApiResponse<serde_json::Value>>> {
    validate_req(&req)?;
    let coupon = state.coupon_svc.create(&req).await?;
    Ok(Json(ApiResponse::success(
        serde_json::json!({ "data": coupon }),
    )))
}

#[utoipa::path(
    patch,
    path = "/api/v1/admin/coupons/:id",
    params(
        ("id" = i64, Path, description = "优惠券ID"),
    ),
    request_body = AdminCouponRequest,
    responses(
        (status = 200, description = "成功", body = serde_json::Value),
    ),
    tag = "优惠券"
)]
/// 管理端：更新优惠券
pub async fn admin_update(
    State(state): State<AppState>,
    Path(id): Path<i64>,
    Json(req): Json<AdminCouponRequest>,
) -> AppResult<Json<ApiResponse<serde_json::Value>>> {
    validate_req(&req)?;
    let coupon = state.coupon_svc.update(id, &req).await?;
    Ok(Json(ApiResponse::success(
        serde_json::json!({ "data": coupon }),
    )))
}

#[utoipa::path(
    delete,
    path = "/api/v1/admin/coupons/:id",
    params(
        ("id" = i64, Path, description = "优惠券ID"),
    ),
    responses(
        (status = 200, description = "成功", body = serde_json::Value),
    ),
    tag = "优惠券"
)]
/// 管理端：删除优惠券
pub async fn admin_delete(
    State(state): State<AppState>,
    Path(id): Path<i64>,
) -> AppResult<Json<ApiResponse<serde_json::Value>>> {
    state.coupon_svc.delete(id).await?;
    Ok(Json(ApiResponse::success_with_message(
        serde_json::json!({}),
        "删除成功",
    )))
}
