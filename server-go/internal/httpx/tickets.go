package httpx

import (
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/chizw/ywty/server-go/internal/authx"
	"github.com/chizw/ywty/server-go/internal/model"
	"github.com/chizw/ywty/server-go/internal/pagination"
	"github.com/chizw/ywty/server-go/internal/support/r"
	"github.com/chizw/ywty/server-go/internal/validate"
	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

// ---------- 工单（auth） ----------

// GET /api/v2/user/tickets
func (d *deps) handleTicketsIndex(w http.ResponseWriter, req *http.Request) {
	u := authx.From(req).User
	p := pagination.FromRequest(req)
	where := "t.`user_id` = ? AND t.`deleted_at` IS NULL"
	args := []any{u.ID}
	order := "`created_at` DESC"
	if q := p.Q; q != "" {
		var keywords []string
		for _, part := range strings.Fields(q) {
			switch part {
			case "sort:level:ascend":
				order = "`level` ASC"
			case "sort:level:descend":
				order = "`level` DESC"
			case "sort:status:ascend":
				order = "`status` ASC"
			case "sort:status:descend":
				order = "`status` DESC"
			case "sort:created_at:ascend":
				order = "`created_at` ASC"
			case "sort:created_at:descend":
				order = "`created_at` DESC"
			default:
				keywords = append(keywords, part)
			}
		}
		if len(keywords) > 0 {
			like := "%" + strings.Join(keywords, " ") + "%"
			where += " AND (t.`issue_no` LIKE ? OR t.`title` LIKE ?)"
			args = append(args, like, like)
		}
	}
	var total int64
	d.gdb.Raw("SELECT count(*) FROM `tickets` t WHERE "+where, args...).Scan(&total)
	var rows []model.Ticket
	d.gdb.Raw("SELECT t.* FROM `tickets` t WHERE "+where+" ORDER BY "+order+" LIMIT ? OFFSET ?",
		append(append([]any{}, args...), p.PerPage, p.Offset())...).Scan(&rows)
	out := make([]map[string]any, 0, len(rows))
	for _, t := range rows {
		out = append(out, map[string]any{
			"id": t.ID, "issue_no": t.IssueNo, "title": t.Title,
			"level": t.Level, "status": t.Status, "created_at": timePtrJSON(t.CreatedAt),
		})
	}
	r.Success(w, pagination.New(out, total, p))
}

// POST /api/v2/user/tickets
func (d *deps) handleTicketsStore(w http.ResponseWriter, req *http.Request) {
	u := authx.From(req).User
	var body struct {
		Title   string `json:"title"`
		Level   string `json:"level"`
		Content string `json:"content"`
	}
	if err := readBody(req, &body); err != nil {
		r.Error(w, "请求体解析失败")
		return
	}
	v := validate.New()
	if !validate.Required(body.Title) || len(body.Title) > 200 {
		v.Add("title", "标题", "不能为空。")
	}
	if !validate.In(body.Level, "low", "medium", "high") {
		v.Add("level", "级别", "不存在。")
	}
	if !validate.Required(body.Content) || len(body.Content) > 2000 {
		v.Add("content", "内容", "不能为空。")
	}
	if v.Fail() {
		v.Respond(w)
		return
	}
	now := time.Now().UTC()
	ticket := model.Ticket{
		UserID: u.ID, IssueNo: d.generateIssueNo(), Title: body.Title,
		Level: body.Level, Status: "in_progress", CreatedAt: &now, UpdatedAt: &now,
	}
	if err := d.gdb.Create(&ticket).Error; err != nil {
		r.Error(w, "创建工单失败")
		return
	}
	reply := model.TicketReply{TicketID: ticket.ID, UserID: u.ID, Content: body.Content, IsNotify: true}
	if err := d.gdb.Create(&reply).Error; err != nil {
		r.Error(w, "创建工单失败")
		return
	}
	r.Created(w, map[string]any{"issue_no": ticket.IssueNo})
}

// loadUserTicket 按工单号取本人工单。
func (d *deps) loadUserTicket(u *model.User, issueNo string) (*model.Ticket, error) {
	var t model.Ticket
	err := d.gdb.Where("user_id = ? AND issue_no = ?", u.ID, issueNo).First(&t).Error
	if err != nil {
		return nil, gorm.ErrRecordNotFound
	}
	return &t, nil
}

// GET /api/v2/user/tickets/{issue_no}
func (d *deps) handleTicketShow(w http.ResponseWriter, req *http.Request) {
	u := authx.From(req).User
	t, err := d.loadUserTicket(u, chi.URLParam(req, "issue_no"))
	if err != nil {
		r.ErrorWithCode(w, http.StatusNotFound, "Not Found")
		return
	}
	r.Success(w, map[string]any{
		"id": t.ID, "issue_no": t.IssueNo, "title": t.Title,
		"level": t.Level, "status": t.Status, "created_at": timePtrJSON(t.CreatedAt),
	})
}

// GET /api/v2/user/tickets/{issue_no}/replies
func (d *deps) handleTicketReplies(w http.ResponseWriter, req *http.Request) {
	u := authx.From(req).User
	t, err := d.loadUserTicket(u, chi.URLParam(req, "issue_no"))
	if err != nil {
		r.ErrorWithCode(w, http.StatusNotFound, "Not Found")
		return
	}
	p := pagination.FromRequest(req)
	p.PerPage = clampPerPage(req.URL.Query().Get("per_page"), 40)
	// 对方回复标记已读
	d.gdb.Exec("UPDATE `ticket_replies` SET `read_at` = ? WHERE `ticket_id` = ? AND `user_id` <> ? AND `read_at` IS NULL",
		time.Now().UTC(), t.ID, u.ID)

	var total int64
	d.gdb.Raw("SELECT count(*) FROM `ticket_replies` WHERE `ticket_id` = ? AND `deleted_at` IS NULL", t.ID).Scan(&total)
	var rows []struct {
		ID        int64
		UserID    int64
		Content   string
		ReadAt    *time.Time
		CreatedAt *time.Time
		UserName  string
		UserAvatar string
	}
	d.gdb.Raw(
		"SELECT r.`id`, r.`user_id`, r.`content`, r.`read_at`, r.`created_at`, u.`name` AS user_name, u.`avatar` AS user_avatar "+
			"FROM `ticket_replies` r LEFT JOIN `users` u ON u.id = r.user_id "+
			"WHERE r.`ticket_id` = ? AND r.`deleted_at` IS NULL ORDER BY r.`created_at` ASC, r.`id` ASC LIMIT ? OFFSET ?",
		t.ID, p.PerPage, p.Offset(),
	).Scan(&rows)
	out := make([]map[string]any, 0, len(rows))
	for _, rr := range rows {
		out = append(out, map[string]any{
			"id": rr.ID, "content": rr.Content,
			"read_at": timePtrJSON(rr.ReadAt), "created_at": timePtrJSON(rr.CreatedAt),
			"user": map[string]any{
				"id": rr.UserID, "name": rr.UserName, "avatar_url": avatarURL(d.cfg, rr.UserAvatar),
			},
		})
	}
	r.Success(w, pagination.New(out, total, p))
}

// POST /api/v2/user/tickets/{issue_no}/reply
func (d *deps) handleTicketReply(w http.ResponseWriter, req *http.Request) {
	u := authx.From(req).User
	t, err := d.loadUserTicket(u, chi.URLParam(req, "issue_no"))
	if err != nil {
		r.ErrorWithCode(w, http.StatusNotFound, "Not Found")
		return
	}
	var body struct {
		Content  string `json:"content"`
		IsNotify *bool  `json:"is_notify"`
	}
	if err := readBody(req, &body); err != nil {
		r.Error(w, "请求体解析失败")
		return
	}
	v := validate.New()
	if !validate.Required(body.Content) || len(body.Content) > 2000 {
		v.Add("content", "内容", "不能为空。")
	}
	if body.IsNotify == nil {
		v.Add("is_notify", "是否通知", "不能为空。")
	}
	if v.Fail() {
		v.Respond(w)
		return
	}
	if t.Status == "completed" {
		r.Error(w, "工单已关闭，无法继续回复")
		return
	}
	isNotify := body.IsNotify != nil && *body.IsNotify
	now := time.Now().UTC()
	reply := model.TicketReply{TicketID: t.ID, UserID: u.ID, Content: body.Content, IsNotify: isNotify, CreatedAt: &now, UpdatedAt: &now}
	if err := d.gdb.Create(&reply).Error; err != nil {
		r.Error(w, "回复失败")
		return
	}
	r.Created(w, nil)
}

// PUT /api/v2/user/tickets/{issue_no}/close
func (d *deps) handleTicketClose(w http.ResponseWriter, req *http.Request) {
	u := authx.From(req).User
	issueNo := chi.URLParam(req, "issue_no")
	res := d.gdb.Model(&model.Ticket{}).
		Where("user_id = ? AND issue_no = ? AND deleted_at IS NULL", u.ID, issueNo).
		Update("status", "completed")
	if res.Error != nil || res.RowsAffected == 0 {
		r.ErrorWithCode(w, http.StatusNotFound, "Not Found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// DELETE /api/v2/user/tickets/{issue_no}
func (d *deps) handleTicketDestroy(w http.ResponseWriter, req *http.Request) {
	u := authx.From(req).User
	issueNo := chi.URLParam(req, "issue_no")
	res := d.gdb.Where("user_id = ? AND issue_no = ?", u.ID, issueNo).Delete(&model.Ticket{})
	if res.Error != nil || res.RowsAffected == 0 {
		r.ErrorWithCode(w, http.StatusNotFound, "Not Found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// generateIssueNo 工单号：YmdHis + 5 位随机（对齐 Ticket::generateIssueNo）。
func (d *deps) generateIssueNo() string {
	for {
		no := fmt.Sprintf("%s%05d", time.Now().Format("20060102150405"), 1+rand.Intn(99999))
		var n int64
		d.gdb.Raw("SELECT count(*) FROM `tickets` WHERE `issue_no` = ?", no).Scan(&n)
		if n == 0 {
			return no
		}
	}
}
