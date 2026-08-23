//! OpenAPI / Swagger 文档
//!
//! 提供 `/api/v1/swagger-ui` 交互式 API 文档。

use utoipa::OpenApi;

/// OpenAPI 规范定义
#[derive(OpenApi)]
#[openapi(
    info(
        title = "云雾图驿 API",
        version = "0.1.0",
        description = "ywty 自托管图床/云相册平台 Rust 后端 API",
        contact(name = "ywty", url = "https://github.com/ywty")
    ),
    servers(
        (url = "/api/v1", description = "API v1")
    ),
    tags(
        (name = "认证", description = "用户注册、登录、Token 管理"),
        (name = "用户", description = "用户资料、密码、邮箱、手机"),
        (name = "图片", description = "图片上传、管理、批量操作"),
        (name = "相册", description = "相册 CRUD、图片归类"),
        (name = "分享", description = "分享链接管理"),
        (name = "标签", description = "标签管理、图片标签关联"),
        (name = "点赞/举报", description = "点赞、举报管理"),
        (name = "反馈/违规", description = "用户反馈、违规记录"),
        (name = "订单", description = "订单创建、支付、管理"),
        (name = "套餐", description = "套餐展示、管理"),
        (name = "优惠券", description = "优惠券校验、管理"),
        (name = "公告", description = "系统公告"),
        (name = "页面", description = "单页管理"),
        (name = "工单", description = "用户工单、客服回复"),
        (name = "管理后台", description = "管理员仪表盘、用户/图片管理"),
        (name = "存储", description = "存储签名、驱动管理"),
        (name = "OAuth", description = "第三方登录绑定"),
    ),
    // paths — 全部 101 条路由
    paths(
        // ── 认证 ──
        crate::handlers::auth::register,
        crate::handlers::auth::login,
        crate::handlers::auth::refresh,
        crate::handlers::auth::logout,
        crate::handlers::auth::reset_password,
        crate::handlers::auth::me,
        crate::handlers::auth::get_captcha,
        crate::handlers::auth::verify_captcha,
        crate::handlers::auth::send_verify_code,
        // ── 用户 ──
        crate::handlers::user::get_profile,
        crate::handlers::user::update_profile,
        crate::handlers::user::change_password,
        crate::handlers::user::change_email,
        crate::handlers::capacity::get,
        // ── 图片 ──
        crate::handlers::photo::list,
        crate::handlers::photo::upload,
        crate::handlers::photo::get,
        crate::handlers::photo::update,
        crate::handlers::photo::delete,
        crate::handlers::photo::batch_delete,
        crate::handlers::photo::batch_update,
        crate::handlers::photo::move_to_album,
        crate::handlers::photo::copy,
        crate::handlers::photo::list_public,
        // ── 相册 ──
        crate::handlers::album::list,
        crate::handlers::album::get,
        crate::handlers::album::create,
        crate::handlers::album::update,
        crate::handlers::album::delete,
        crate::handlers::album::list_photos,
        crate::handlers::album::add_photos,
        crate::handlers::album::remove_photo,
        // ── 分享 ──
        crate::handlers::share::list,
        crate::handlers::share::create,
        crate::handlers::share::update,
        crate::handlers::share::delete,
        // ── 标签 ──
        crate::handlers::tag::list,
        crate::handlers::tag::create,
        crate::handlers::tag::delete,
        crate::handlers::tag::attach,
        crate::handlers::tag::detach,
        // ── 点赞/举报 ──
        crate::handlers::like::toggle,
        crate::handlers::like::status,
        crate::handlers::like::create_report,
        crate::handlers::like::admin_list_reports,
        crate::handlers::like::admin_update_report_status,
        // ── 反馈/违规 ──
        crate::handlers::feedback::create_feedback,
        crate::handlers::feedback::create_violation,
        crate::handlers::feedback::admin_list_violations,
        crate::handlers::feedback::admin_update_violation_status,
        crate::handlers::feedback::admin_list_feedbacks,
        crate::handlers::feedback::admin_delete_feedback,
        // ── 订单 ──
        crate::handlers::order::list,
        crate::handlers::order::create,
        crate::handlers::order::get,
        crate::handlers::order::pay,
        crate::handlers::order::cancel,
        crate::handlers::order::notify,
        // ── 套餐 ──
        crate::handlers::plan::list_public,
        crate::handlers::plan::get_public,
        crate::handlers::plan::admin_list,
        crate::handlers::plan::admin_create,
        crate::handlers::plan::admin_get,
        crate::handlers::plan::admin_update,
        crate::handlers::plan::admin_delete,
        crate::handlers::plan::admin_toggle_up,
        // ── 优惠券 ──
        crate::handlers::coupon::validate,
        crate::handlers::coupon::admin_list,
        crate::handlers::coupon::admin_create,
        crate::handlers::coupon::admin_get,
        crate::handlers::coupon::admin_update,
        crate::handlers::coupon::admin_delete,
        // ── 公告 ──
        crate::handlers::notice::list_public,
        crate::handlers::notice::get_public,
        crate::handlers::notice::admin_list,
        crate::handlers::notice::admin_create,
        crate::handlers::notice::admin_update,
        crate::handlers::notice::admin_delete,
        // ── 页面 ──
        crate::handlers::page::list_public,
        crate::handlers::page::get_public,
        crate::handlers::page::admin_list,
        crate::handlers::page::admin_create,
        crate::handlers::page::admin_get,
        crate::handlers::page::admin_update,
        crate::handlers::page::admin_delete,
        // ── 工单 ──
        crate::handlers::ticket::list,
        crate::handlers::ticket::create,
        crate::handlers::ticket::get,
        crate::handlers::ticket::reply,
        crate::handlers::ticket::close,
        crate::handlers::ticket::admin_list,
        crate::handlers::ticket::admin_get,
        crate::handlers::ticket::admin_reply,
        crate::handlers::ticket::admin_update_status,
        // ── OAuth ──
        crate::handlers::oauth::authorize,
        crate::handlers::oauth::callback,
        crate::handlers::oauth::list,
        crate::handlers::oauth::bind,
        crate::handlers::oauth::unbind,
        crate::handlers::oauth::find_by_openid,
        // ── 管理后台 ──
        crate::handlers::admin::stats,
        crate::handlers::admin::list_users,
        crate::handlers::admin::update_user,
        crate::handlers::admin::list_all_photos,
        crate::handlers::admin::list_rbac_policies,
        crate::handlers::admin::add_rbac_policy,
        crate::handlers::admin::delete_rbac_policy,
        crate::handlers::admin::list_rbac_roles,
        crate::handlers::admin::assign_rbac_role,
        // ── 群组 ──
        crate::handlers::group::list,
        crate::handlers::group::create,
        crate::handlers::group::get,
        crate::handlers::group::update,
        crate::handlers::group::delete,
        // ── 令牌 ──
        crate::handlers::token::list,
        crate::handlers::token::create,
        crate::handlers::token::revoke,
        // ── 存储 ──
        crate::handlers::storage::sign,
        crate::handlers::storage_admin::list_drivers,
        crate::handlers::storage_admin::list_storages,
        crate::handlers::storage_admin::create_storage,
        crate::handlers::storage_admin::update_storage,
        crate::handlers::storage_admin::delete_storage,
        crate::handlers::storage_admin::copy,
        crate::handlers::drivers::list,
    ),
    components(
        schemas(
            // 分页
            crate::dto::Meta,
            // 认证
            crate::models::user::LoginRequest,
            crate::models::user::RegisterRequest,
            // 用户
            crate::models::user::User,
            crate::models::user::UserPublic,
            crate::models::user::ChangePasswordRequest,
            crate::models::user::UpdateProfileRequest,
            // 图片
            crate::dto::photo::PhotoResponse,
            crate::dto::photo::PhotoPublicResponse,
            crate::dto::photo::UploadResponse,
            crate::dto::photo::BatchIdsRequest,
            crate::dto::photo::BatchUpdateRequest,
            // 相册
            crate::models::album::Album,
            // 分享
            crate::models::photo::Share,
            crate::models::photo::CreateShareRequest,
            // 标签
            crate::models::photo::Tag,
            // 反馈
            crate::models::feedback::Feedback,
            crate::models::feedback::CreateFeedbackRequest,
            crate::models::feedback::Violation,
            crate::models::feedback::CreateViolationRequest,
            // 订单
            crate::models::order::Order,
            crate::models::order::CreateOrderRequest,
            // 套餐
            crate::models::plan::Plan,
            crate::models::plan::PlanDetail,
            crate::models::plan::PlanPrice,
            // 优惠券
            crate::models::coupon::Coupon,
            crate::models::coupon::ValidateCouponRequest,
            // 公告
            crate::models::notice::Notice,
            crate::models::notice::CreateNoticeRequest,
            // 页面
            crate::models::page::Page,
            crate::models::page::CreatePageRequest,
            // 工单
            crate::models::ticket::Ticket,
            crate::models::ticket::TicketReply,
            crate::models::ticket::CreateTicketRequest,
            // 群组
            crate::models::group::Group,
            crate::models::group::CreateGroupRequest,
        )
    )
)]
pub struct ApiDoc;

/// 创建 Swagger UI 路由
/// 注意：utoipa-swagger-ui 需要从 GitHub 下载资源，暂时返回空路由
pub fn swagger_routes() -> axum::Router<crate::app::AppState> {
    // use utoipa_swagger_ui::SwaggerUi;
    // SwaggerUi::new("/swagger-ui").url("/api-docs/openapi.json", ApiDoc::openapi()).into()
    axum::Router::new()
}
