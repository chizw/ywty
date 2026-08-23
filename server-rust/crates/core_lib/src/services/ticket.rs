//! 工单服务

use sqlx::SqlitePool;

use crate::error::{AppError, AppResult};
use crate::models::ticket::{CreateTicketRequest, Ticket, TicketDetail, TicketReply};

#[derive(Clone)]
pub struct TicketService {
    pool: SqlitePool,
}

impl TicketService {
    pub fn new(pool: SqlitePool) -> Self {
        Self { pool }
    }

    /// 生成工单号
    fn generate_issue_no() -> String {
        use base64::{engine::general_purpose::URL_SAFE_NO_PAD, Engine as _};
        let bytes: Vec<u8> = (0..8).map(|_| rand::random::<u8>()).collect();
        format!("TK{}", URL_SAFE_NO_PAD.encode(&bytes))
    }

    /// 列出用户工单
    pub async fn list(
        &self,
        user_id: i64,
        page: i32,
        per_page: i32,
    ) -> AppResult<(Vec<Ticket>, i64)> {
        let page = if page < 1 { 1 } else { page };
        let per_page = if !(1..=100).contains(&per_page) {
            20
        } else {
            per_page
        };
        let offset = (page - 1) * per_page;

        let rows = sqlx::query_as(
            "SELECT * FROM tickets WHERE user_id = ? AND deleted_at IS NULL ORDER BY id DESC LIMIT ? OFFSET ?",
        )
        .bind(user_id)
        .bind(per_page)
        .bind(offset)
        .fetch_all(&self.pool)
        .await?;

        let total: i64 = sqlx::query_scalar(
            "SELECT COUNT(*) FROM tickets WHERE user_id = ? AND deleted_at IS NULL",
        )
        .bind(user_id)
        .fetch_one(&self.pool)
        .await?;

        Ok((rows, total))
    }

    /// 获取工单详情
    pub async fn get(&self, user_id: i64, id: i64) -> AppResult<TicketDetail> {
        let ticket: Option<Ticket> = sqlx::query_as(
            "SELECT * FROM tickets WHERE id = ? AND user_id = ? AND deleted_at IS NULL",
        )
        .bind(id)
        .bind(user_id)
        .fetch_optional(&self.pool)
        .await?;

        let ticket = ticket.ok_or_else(|| AppError::NotFound("工单不存在".to_string()))?;

        let replies = sqlx::query_as(
            "SELECT * FROM ticket_replies WHERE ticket_id = ? ORDER BY created_at ASC, id ASC",
        )
        .bind(id)
        .fetch_all(&self.pool)
        .await?;

        Ok(TicketDetail { ticket, replies })
    }

    /// 创建工单
    pub async fn create(&self, user_id: i64, req: &CreateTicketRequest) -> AppResult<Ticket> {
        let now = chrono::Utc::now().to_rfc3339();
        let issue_no = Self::generate_issue_no();
        let ticket_type = req.ticket_type.as_deref().unwrap_or("other");
        let level = req.level.as_deref().unwrap_or("low");

        let result = sqlx::query(
            r#"
            INSERT INTO tickets (user_id, issue_no, title, type, level, status, created_at, updated_at)
            VALUES (?, ?, ?, ?, ?, 'in_progress', ?, ?)
            "#,
        )
        .bind(user_id)
        .bind(&issue_no)
        .bind(&req.title)
        .bind(ticket_type)
        .bind(level)
        .bind(&now)
        .bind(&now)
        .execute(&self.pool)
        .await?;

        let ticket_id = result.last_insert_rowid();

        // 创建首条回复（工单内容）
        sqlx::query(
            "INSERT INTO ticket_replies (ticket_id, user_id, is_admin, content, is_notify, created_at, updated_at) VALUES (?, ?, 0, ?, 1, ?, ?)",
        )
        .bind(ticket_id)
        .bind(user_id)
        .bind(&req.content)
        .bind(&now)
        .bind(&now)
        .execute(&self.pool)
        .await?;

        self.get(user_id, ticket_id).await.map(|d| d.ticket)
    }

    /// 列出工单回复（时间正序）
    pub async fn list_replies(&self, ticket_id: i64) -> AppResult<Vec<TicketReply>> {
        let replies = sqlx::query_as(
            "SELECT * FROM ticket_replies WHERE ticket_id = ? ORDER BY created_at ASC, id ASC",
        )
        .bind(ticket_id)
        .fetch_all(&self.pool)
        .await?;
        Ok(replies)
    }

    /// 回复工单（用户，仅限自己的工单）
    pub async fn reply(
        &self,
        user_id: i64,
        ticket_id: i64,
        content: &str,
    ) -> AppResult<TicketReply> {
        // 校验工单归属：只能回复自己的工单
        let owned: Option<i64> = sqlx::query_scalar(
            "SELECT id FROM tickets WHERE id = ? AND user_id = ? AND deleted_at IS NULL",
        )
        .bind(ticket_id)
        .bind(user_id)
        .fetch_optional(&self.pool)
        .await?;
        if owned.is_none() {
            return Err(AppError::NotFound("工单不存在".to_string()));
        }
        self.insert_reply(user_id, ticket_id, content, false).await
    }

    /// 关闭工单
    pub async fn close(&self, user_id: i64, id: i64) -> AppResult<()> {
        let now = chrono::Utc::now().to_rfc3339();
        let result = sqlx::query(
            "UPDATE tickets SET status = 'closed', updated_at = ? WHERE id = ? AND user_id = ? AND deleted_at IS NULL",
        )
        .bind(&now)
        .bind(id)
        .bind(user_id)
        .execute(&self.pool)
        .await?;

        if result.rows_affected() == 0 {
            return Err(AppError::NotFound("工单不存在".to_string()));
        }
        Ok(())
    }

    /// 管理端列表
    pub async fn admin_list(
        &self,
        page: i32,
        per_page: i32,
        status: Option<&str>,
        level: Option<&str>,
    ) -> AppResult<(Vec<Ticket>, i64)> {
        let page = if page < 1 { 1 } else { page };
        let per_page = if !(1..=100).contains(&per_page) {
            20
        } else {
            per_page
        };
        let offset = (page - 1) * per_page;

        let mut where_clause = String::from("WHERE deleted_at IS NULL");
        if status.is_some() {
            where_clause.push_str(" AND status = ?");
        }
        if level.is_some() {
            where_clause.push_str(" AND level = ?");
        }

        let sql = format!(
            "SELECT * FROM tickets {} ORDER BY id DESC LIMIT ? OFFSET ?",
            where_clause
        );
        let mut query = sqlx::query_as::<_, Ticket>(&sql);
        if let Some(s) = status {
            query = query.bind(s);
        }
        if let Some(l) = level {
            query = query.bind(l);
        }
        let rows = query
            .bind(per_page)
            .bind(offset)
            .fetch_all(&self.pool)
            .await?;

        // 总数
        let mut count_sql = String::from("SELECT COUNT(*) FROM tickets WHERE deleted_at IS NULL");
        if status.is_some() {
            count_sql.push_str(" AND status = ?");
        }
        if level.is_some() {
            count_sql.push_str(" AND level = ?");
        }
        let mut count_query = sqlx::query_scalar::<_, i64>(&count_sql);
        if let Some(s) = status {
            count_query = count_query.bind(s);
        }
        if let Some(l) = level {
            count_query = count_query.bind(l);
        }
        let total = count_query.fetch_one(&self.pool).await?;

        Ok((rows, total))
    }

    /// 管理端获取详情
    pub async fn admin_get(&self, id: i64) -> AppResult<TicketDetail> {
        let ticket: Option<Ticket> =
            sqlx::query_as("SELECT * FROM tickets WHERE id = ? AND deleted_at IS NULL")
                .bind(id)
                .fetch_optional(&self.pool)
                .await?;

        let ticket = ticket.ok_or_else(|| AppError::NotFound("工单不存在".to_string()))?;

        let replies = sqlx::query_as(
            "SELECT * FROM ticket_replies WHERE ticket_id = ? ORDER BY created_at ASC, id ASC",
        )
        .bind(id)
        .fetch_all(&self.pool)
        .await?;

        Ok(TicketDetail { ticket, replies })
    }

    /// 管理端回复（任意工单）
    pub async fn admin_reply(
        &self,
        admin_user_id: i64,
        ticket_id: i64,
        content: &str,
    ) -> AppResult<TicketReply> {
        self.insert_reply(admin_user_id, ticket_id, content, true)
            .await
    }

    /// 插入回复的公共实现：校验工单存在、写入 is_admin 标记并刷新工单 updated_at
    async fn insert_reply(
        &self,
        user_id: i64,
        ticket_id: i64,
        content: &str,
        is_admin: bool,
    ) -> AppResult<TicketReply> {
        // 验证工单存在
        let exists: Option<i64> =
            sqlx::query_scalar("SELECT id FROM tickets WHERE id = ? AND deleted_at IS NULL")
                .bind(ticket_id)
                .fetch_optional(&self.pool)
                .await?;
        if exists.is_none() {
            return Err(AppError::NotFound("工单不存在".to_string()));
        }

        let now = chrono::Utc::now().to_rfc3339();
        let result = sqlx::query(
                r#"
                INSERT INTO ticket_replies (ticket_id, user_id, is_admin, content, is_notify, created_at, updated_at)
                VALUES (?, ?, ?, ?, 1, ?, ?)
                "#,
            )
            .bind(ticket_id)
            .bind(user_id)
            .bind(if is_admin { 1 } else { 0 })
            .bind(content)
            .bind(&now)
            .bind(&now)
            .execute(&self.pool)
            .await?;

        // 刷新工单更新时间
        sqlx::query("UPDATE tickets SET updated_at = ? WHERE id = ?")
            .bind(&now)
            .bind(ticket_id)
            .execute(&self.pool)
            .await?;

        let id = result.last_insert_rowid();
        let reply: TicketReply = sqlx::query_as("SELECT * FROM ticket_replies WHERE id = ?")
            .bind(id)
            .fetch_one(&self.pool)
            .await?;

        Ok(reply)
    }

    /// 管理端更新状态
    pub async fn admin_update_status(&self, id: i64, status: &str) -> AppResult<()> {
        let now = chrono::Utc::now().to_rfc3339();
        let result = sqlx::query(
            "UPDATE tickets SET status = ?, updated_at = ? WHERE id = ? AND deleted_at IS NULL",
        )
        .bind(status)
        .bind(&now)
        .bind(id)
        .execute(&self.pool)
        .await?;

        if result.rows_affected() == 0 {
            return Err(AppError::NotFound("工单不存在".to_string()));
        }
        Ok(())
    }
}
