//! 管理员服务

use crate::db::DbPool;

use crate::dto::user::AdminUserResponse;
use crate::error::{AppError, AppResult};

#[derive(Clone)]
pub struct AdminService {
    pool: DbPool,
}

impl AdminService {
    pub fn new(pool: DbPool) -> Self {
        Self { pool }
    }

    /// 仪表盘统计
    pub async fn stats(&self) -> AppResult<serde_json::Value> {
        let users: i64 = sqlx::query_scalar("SELECT COUNT(*) FROM users WHERE deleted_at IS NULL")
            .fetch_one(&self.pool)
            .await?;
        let photos: i64 =
            sqlx::query_scalar("SELECT COUNT(*) FROM photos WHERE deleted_at IS NULL")
                .fetch_one(&self.pool)
                .await?;
        let albums: i64 =
            sqlx::query_scalar("SELECT COUNT(*) FROM albums WHERE deleted_at IS NULL")
                .fetch_one(&self.pool)
                .await?;
        let shares: i64 =
            sqlx::query_scalar("SELECT COUNT(*) FROM shares WHERE deleted_at IS NULL")
                .fetch_one(&self.pool)
                .await?;
        let reports: i64 = sqlx::query_scalar("SELECT COUNT(*) FROM reports WHERE status = 0")
            .fetch_one(&self.pool)
            .await?;
        let orders: i64 = sqlx::query_scalar("SELECT COUNT(*) FROM orders")
            .fetch_one(&self.pool)
            .await?;
        let paid_orders: i64 =
            sqlx::query_scalar("SELECT COUNT(*) FROM orders WHERE status = 'paid'")
                .fetch_one(&self.pool)
                .await?;
        let total_income: i64 =
            sqlx::query_scalar("SELECT COALESCE(SUM(amount), 0) FROM orders WHERE status = 'paid'")
                .fetch_one(&self.pool)
                .await?;
        let notices: i64 =
            sqlx::query_scalar("SELECT COUNT(*) FROM notices WHERE deleted_at IS NULL")
                .fetch_one(&self.pool)
                .await?;
        let pages: i64 = sqlx::query_scalar("SELECT COUNT(*) FROM pages WHERE deleted_at IS NULL")
            .fetch_one(&self.pool)
            .await?;
        let tickets: i64 =
            sqlx::query_scalar("SELECT COUNT(*) FROM tickets WHERE deleted_at IS NULL")
                .fetch_one(&self.pool)
                .await?;

        Ok(serde_json::json!({
            "users": users,
            "photos": photos,
            "albums": albums,
            "shares": shares,
            "reports": reports,
            "pending_reports": reports,
            "orders": orders,
            "paid_orders": paid_orders,
            "total_income": total_income,
            "notices": notices,
            "pages": pages,
            "tickets": tickets,
        }))
    }

    /// 列出用户（管理端，不含密码等敏感字段）
    pub async fn list_users(
        &self,
        page: i32,
        per_page: i32,
        keyword: Option<&str>,
    ) -> AppResult<(Vec<AdminUserResponse>, i64)> {
        let page = if page < 1 { 1 } else { page };
        let per_page = if !(1..=100).contains(&per_page) {
            20
        } else {
            per_page
        };
        let offset = (page - 1) * per_page;

        const COLS: &str = "id, uuid, username, email, avatar, role, is_super_admin, status, capacity_used, capacity_max, created_at";

        let (rows, total) = if let Some(kw) = keyword {
            let like = format!("%{}%", kw);
            let rows = sqlx::query_as::<_, AdminUserResponse>(
                &format!(
                    "SELECT {} FROM users WHERE deleted_at IS NULL AND (username LIKE ? OR email LIKE ?) ORDER BY id DESC LIMIT ? OFFSET ?",
                    COLS
                ),
            )
            .bind(&like)
            .bind(&like)
            .bind(per_page)
            .bind(offset)
            .fetch_all(&self.pool)
            .await?;

            let total: i64 = sqlx::query_scalar(
                "SELECT COUNT(*) FROM users WHERE deleted_at IS NULL AND (username LIKE ? OR email LIKE ?)",
            )
            .bind(&like)
            .bind(&like)
            .fetch_one(&self.pool)
            .await?;

            (rows, total)
        } else {
            let rows = sqlx::query_as::<_, AdminUserResponse>(&format!(
                "SELECT {} FROM users WHERE deleted_at IS NULL ORDER BY id DESC LIMIT ? OFFSET ?",
                COLS
            ))
            .bind(per_page)
            .bind(offset)
            .fetch_all(&self.pool)
            .await?;

            let total: i64 =
                sqlx::query_scalar("SELECT COUNT(*) FROM users WHERE deleted_at IS NULL")
                    .fetch_one(&self.pool)
                    .await?;

            (rows, total)
        };

        Ok((rows, total))
    }

    /// 查询单个用户的角色与超管标记（供越权判断）
    pub async fn get_user_meta(&self, id: i64) -> AppResult<Option<(String, bool)>> {
        let row: Option<(String, bool)> = sqlx::query_as(
            "SELECT role, is_super_admin FROM users WHERE id = ? AND deleted_at IS NULL",
        )
        .bind(id)
        .fetch_optional(&self.pool)
        .await?;
        Ok(row)
    }

    /// 删除用户（软删用户及其图片/相册/分享）
    pub async fn delete_user(&self, id: i64) -> AppResult<()> {
        let now = crate::db::now_str();
        // 软删用户
        let result = sqlx::query(
            "UPDATE users SET deleted_at = ?, updated_at = ? WHERE id = ? AND deleted_at IS NULL",
        )
        .bind(&now)
        .bind(&now)
        .bind(id)
        .execute(&self.pool)
        .await?;
        if result.rows_affected() == 0 {
            return Err(AppError::NotFound("用户不存在".to_string()));
        }
        // 软删其内容
        sqlx::query("UPDATE photos SET deleted_at = ? WHERE user_id = ? AND deleted_at IS NULL")
            .bind(&now)
            .bind(id)
            .execute(&self.pool)
            .await?;
        sqlx::query("UPDATE albums SET deleted_at = ? WHERE user_id = ? AND deleted_at IS NULL")
            .bind(&now)
            .bind(id)
            .execute(&self.pool)
            .await?;
        sqlx::query("UPDATE shares SET deleted_at = ? WHERE user_id = ? AND deleted_at IS NULL")
            .bind(&now)
            .bind(id)
            .execute(&self.pool)
            .await?;
        Ok(())
    }

    /// 更新用户（参数化，安全；role/status/name 任一为 None 表示不改该字段）
    /// 权限判断由 handler 负责：谁有资格改什么，在这里只做受控的字段更新
    pub async fn update_user(
        &self,
        id: i64,
        role: Option<String>,
        status: Option<i32>,
        name: Option<String>,
    ) -> AppResult<()> {
        let mut sets: Vec<&str> = Vec::new();
        let mut role_pending = false;
        let mut status_pending = false;
        let mut name_pending = false;

        if role.is_some() {
            sets.push("role = ?");
            role_pending = true;
        }
        if status.is_some() {
            sets.push("status = ?");
            status_pending = true;
        }
        if name.is_some() {
            sets.push("name = ?");
            name_pending = true;
        }
        if sets.is_empty() {
            return Err(AppError::Validation("没有要更新的字段".to_string()));
        }

        let now = crate::db::now_str();
        sets.push("updated_at = ?");
        let sql = format!(
            "UPDATE users SET {} WHERE id = ? AND deleted_at IS NULL",
            sets.join(", ")
        );

        let mut q = sqlx::query(&sql);
        if role_pending {
            q = q.bind(role.unwrap_or_default());
        }
        if status_pending {
            q = q.bind(status.unwrap_or_default());
        }
        if name_pending {
            q = q.bind(name.unwrap_or_default());
        }
        q = q.bind(&now).bind(id);

        let result = q.execute(&self.pool).await?;
        if result.rows_affected() == 0 {
            return Err(AppError::NotFound("用户不存在".to_string()));
        }
        Ok(())
    }

    /// 列出所有图片（管理端）
    pub async fn list_all_photos(
        &self,
        page: i32,
        per_page: i32,
        keyword: Option<&str>,
    ) -> AppResult<(Vec<serde_json::Value>, i64)> {
        let page = if page < 1 { 1 } else { page };
        let per_page = if !(1..=100).contains(&per_page) {
            24
        } else {
            per_page
        };
        let offset = (page - 1) * per_page;

        let (rows, total) = if let Some(kw) = keyword {
            let like = format!("%{}%", kw);
            let rows = sqlx::query_as::<_, crate::models::photo::Photo>(
                "SELECT * FROM photos WHERE deleted_at IS NULL AND (name LIKE ? OR filename LIKE ?) ORDER BY id DESC LIMIT ? OFFSET ?",
            )
            .bind(&like)
            .bind(&like)
            .bind(per_page)
            .bind(offset)
            .fetch_all(&self.pool)
            .await?;

            let total: i64 = sqlx::query_scalar(
                "SELECT COUNT(*) FROM photos WHERE deleted_at IS NULL AND (name LIKE ? OR filename LIKE ?)",
            )
            .bind(&like)
            .bind(&like)
            .fetch_one(&self.pool)
            .await?;

            let rows_json = rows
                .into_iter()
                .map(|p| {
                    serde_json::json!({
                        "id": p.id,
                        "uuid": p.uuid,
                        "user_id": p.user_id,
                        "filename": p.filename,
                        "url": p.url,
                        "size": p.size,
                        "status": p.status,
                        "created_at": p.created_at,
                    })
                })
                .collect();
            (rows_json, total)
        } else {
            let rows = sqlx::query_as::<_, crate::models::photo::Photo>(
                "SELECT * FROM photos WHERE deleted_at IS NULL ORDER BY id DESC LIMIT ? OFFSET ?",
            )
            .bind(per_page)
            .bind(offset)
            .fetch_all(&self.pool)
            .await?;

            let total: i64 =
                sqlx::query_scalar("SELECT COUNT(*) FROM photos WHERE deleted_at IS NULL")
                    .fetch_one(&self.pool)
                    .await?;

            let rows_json = rows
                .into_iter()
                .map(|p| {
                    serde_json::json!({
                        "id": p.id,
                        "uuid": p.uuid,
                        "user_id": p.user_id,
                        "filename": p.filename,
                        "url": p.url,
                        "size": p.size,
                        "status": p.status,
                        "created_at": p.created_at,
                    })
                })
                .collect();
            (rows_json, total)
        };

        Ok((rows, total))
    }
}
