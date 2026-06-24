// Package handler 管理后台接口（仅管理员）
package handler

import (
	"fmt"
	"strconv"

	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/ywty/server/internal/auth"
	bizerr "github.com/ywty/server/internal/errors"
	"github.com/ywty/server/internal/model"
	"github.com/ywty/server/internal/response"
	"github.com/ywty/server/internal/service"
)

// AdminHandler 管理后台接口
type AdminHandler struct {
	db       *gorm.DB
	user     *service.UserService
	authSvc  *service.AuthService
	enforcer *casbin.Enforcer
}

func NewAdminHandler(db *gorm.DB, user *service.UserService, authSvc *service.AuthService, enforcer *casbin.Enforcer) *AdminHandler {
	return &AdminHandler{db: db, user: user, authSvc: authSvc, enforcer: enforcer}
}

// ===== 后台登录 =====

// Login POST /api/v1/admin/login
// 后台登录入口（无需鉴权 / 无需 AdminOnly）：校验账号密码并要求 IsAdmin=true
func (h *AdminHandler) Login(c *gin.Context) {
	var req service.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, bizerr.BadRequest.WithMessage(err.Error()))
		return
	}
	res, err := h.authSvc.Login(c.Request.Context(), req)
	if err != nil {
		response.Fail(c, service.ToAPIError(err))
		return
	}
	if !res.User.IsAdmin {
		response.FailCode(c, bizerr.Forbidden.WithMessage("需要管理员权限"))
		return
	}
	response.Success(c, res)
}

// ===== 仪表盘 =====

// Stats GET /api/v1/admin/stats
func (h *AdminHandler) Stats(c *gin.Context) {
	var (
		users       int64
		photos      int64
		albums      int64
		shares      int64
		reports     int64
		pendingRpts int64

		orders      int64
		paidOrders  int64
		totalIncome int64
		notices     int64
		pages       int64
		tickets     int64
	)
	_ = h.db.Model(&model.User{}).Count(&users)
	_ = h.db.Model(&model.Photo{}).Count(&photos)
	_ = h.db.Model(&model.Album{}).Count(&albums)
	_ = h.db.Model(&model.Share{}).Count(&shares)
	_ = h.db.Model(&model.Report{}).Count(&reports)
	_ = h.db.Model(&model.Report{}).Where("status = ?", model.ReportStatusUnhandled).Count(&pendingRpts)

	_ = h.db.Model(&model.Order{}).Count(&orders)
	_ = h.db.Model(&model.Order{}).Where("status = ?", model.OrderStatusPaid).Count(&paidOrders)
	row := h.db.Model(&model.Order{}).Where("status = ?", model.OrderStatusPaid).Select("COALESCE(SUM(amount),0)").Row()
	_ = row.Scan(&totalIncome)
	_ = h.db.Model(&model.Notice{}).Count(&notices)
	_ = h.db.Model(&model.Page{}).Count(&pages)
	_ = h.db.Model(&model.Ticket{}).Count(&tickets)

	response.Success(c, gin.H{
		"users":           users,
		"photos":          photos,
		"albums":          albums,
		"shares":          shares,
		"reports":         reports,
		"pending_reports": pendingRpts,
		"orders":          orders,
		"paid_orders":     paidOrders,
		"total_income":    totalIncome,
		"notices":         notices,
		"pages":           pages,
		"tickets":         tickets,
	})
}

// ===== 用户管理 =====

// ListUsers GET /api/v1/admin/users
func (h *AdminHandler) ListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	keyword := c.Query("keyword")
	if page <= 0 {
		page = 1
	}
	if perPage <= 0 || perPage > 100 {
		perPage = 20
	}
	q := h.db.Model(&model.User{})
	if keyword != "" {
		like := "%" + keyword + "%"
		q = q.Where("username LIKE ? OR email LIKE ? OR name LIKE ?", like, like, like)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		response.Fail(c, err)
		return
	}
	var rows []model.User
	if err := q.Order("id DESC").Offset((page - 1) * perPage).Limit(perPage).Find(&rows).Error; err != nil {
		response.Fail(c, err)
		return
	}
	// 不返回密码字段
	for i := range rows {
		rows[i].Password = ""
	}
	response.Page(c, rows, response.PageMeta{
		CurrentPage: page,
		PerPage:     perPage,
		Total:       total,
		LastPage:    int((total + int64(perPage) - 1) / int64(perPage)),
	})
}

// UpdateUser PATCH /api/v1/admin/users/:id
func (h *AdminHandler) UpdateUser(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req struct {
		IsAdmin  *bool   `json:"is_admin"`
		Status   *string `json:"status"`
		GroupID  *uint64 `json:"group_id"`
		Nickname *string `json:"nickname"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, bizerr.BadRequest.WithMessage(err.Error()))
		return
	}
	updates := map[string]any{}
	if req.IsAdmin != nil {
		updates["is_admin"] = *req.IsAdmin
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.Nickname != nil {
		updates["name"] = *req.Nickname
	}
	if len(updates) == 0 {
		response.FailCode(c, bizerr.BadRequest.WithMessage("no fields to update"))
		return
	}
	res := h.db.Model(&model.User{}).Where("id = ?", id).Updates(updates)
	if res.Error != nil {
		response.Fail(c, res.Error)
		return
	}
	if res.RowsAffected == 0 {
		response.FailCode(c, bizerr.UserNotFound)
		return
	}
	// 如果修改了角色组，更新 user_group
	if req.GroupID != nil {
		_ = h.db.Where("user_id = ?", id).Delete(&model.UserGroup{}).Error
		_ = h.db.Create(&model.UserGroup{
			UserID:  id,
			GroupID: *req.GroupID,
			From:    model.GroupFromAdmin,
		}).Error
	}
	response.Success(c, gin.H{"id": id, "updated": true})
}

// ===== 角色组管理 =====

// ListGroups GET /api/v1/admin/groups
func (h *AdminHandler) ListGroups(c *gin.Context) {
	var rows []model.Group
	if err := h.db.Order("id ASC").Find(&rows).Error; err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, rows)
}

// CreateGroup POST /api/v1/admin/groups
func (h *AdminHandler) CreateGroup(c *gin.Context) {
	var req struct {
		Name      string `json:"name" binding:"required"`
		Intro     string `json:"intro"`
		IsDefault bool   `json:"is_default"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, bizerr.BadRequest.WithMessage(err.Error()))
		return
	}
	g := &model.Group{
		Name:      req.Name,
		Intro:     req.Intro,
		IsDefault: req.IsDefault,
	}
	if err := h.db.Create(g).Error; err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, g)
}

// UpdateGroup PATCH /api/v1/admin/groups/:id
func (h *AdminHandler) UpdateGroup(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req struct {
		Name      *string `json:"name"`
		Intro     *string `json:"intro"`
		IsDefault *bool   `json:"is_default"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, bizerr.BadRequest.WithMessage(err.Error()))
		return
	}
	updates := map[string]any{}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Intro != nil {
		updates["intro"] = *req.Intro
	}
	if req.IsDefault != nil {
		updates["is_default"] = *req.IsDefault
	}
	if len(updates) == 0 {
		response.FailCode(c, bizerr.BadRequest.WithMessage("no fields to update"))
		return
	}
	res := h.db.Model(&model.Group{}).Where("id = ?", id).Updates(updates)
	if res.Error != nil {
		response.Fail(c, res.Error)
		return
	}
	if res.RowsAffected == 0 {
		response.FailCode(c, bizerr.ResourceNotFound)
		return
	}
	response.Success(c, gin.H{"id": id, "updated": true})
}

// DeleteGroup DELETE /api/v1/admin/groups/:id
func (h *AdminHandler) DeleteGroup(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	res := h.db.Delete(&model.Group{}, id)
	if res.Error != nil {
		response.Fail(c, res.Error)
		return
	}
	if res.RowsAffected == 0 {
		response.FailCode(c, bizerr.ResourceNotFound)
		return
	}
	_ = h.db.Where("group_id = ?", id).Delete(&model.UserGroup{}).Error
	response.Success(c, gin.H{"id": id, "deleted": true})
}

// ===== 图片管理 =====

// ListAllPhotos GET /api/v1/admin/photos
func (h *AdminHandler) ListAllPhotos(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "24"))
	keyword := c.Query("keyword")
	if page <= 0 {
		page = 1
	}
	if perPage <= 0 || perPage > 100 {
		perPage = 24
	}
	q := h.db.Model(&model.Photo{})
	if keyword != "" {
		q = q.Where("name LIKE ? OR filename LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		response.Fail(c, err)
		return
	}
	var rows []model.Photo
	if err := q.Order("id DESC").Offset((page - 1) * perPage).Limit(perPage).Find(&rows).Error; err != nil {
		response.Fail(c, err)
		return
	}
	response.Page(c, rows, response.PageMeta{
		CurrentPage: page,
		PerPage:     perPage,
		Total:       total,
		LastPage:    int((total + int64(perPage) - 1) / int64(perPage)),
	})
}

// ===== RBAC 管理 =====

// ListRBACPolicies GET /api/v1/admin/rbac/policies
func (h *AdminHandler) ListRBACPolicies(c *gin.Context) {
	if h.enforcer == nil {
		response.FailCode(c, bizerr.Internal.WithMessage("casbin enforcer not initialized"))
		return
	}
	policies, err := h.enforcer.GetPolicy()
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, policies)
}

// AddRBACPolicy POST /api/v1/admin/rbac/policies
func (h *AdminHandler) AddRBACPolicy(c *gin.Context) {
	if h.enforcer == nil {
		response.FailCode(c, bizerr.Internal.WithMessage("casbin enforcer not initialized"))
		return
	}
	var req struct {
		Sub string `json:"sub" binding:"required"`
		Obj string `json:"obj" binding:"required"`
		Act string `json:"act" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, bizerr.BadRequest.WithMessage(err.Error()))
		return
	}
	ok, err := h.enforcer.AddPolicy(req.Sub, req.Obj, req.Act)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, gin.H{"added": ok, "sub": req.Sub, "obj": req.Obj, "act": req.Act})
}

// DeleteRBACPolicy DELETE /api/v1/admin/rbac/policies
func (h *AdminHandler) DeleteRBACPolicy(c *gin.Context) {
	if h.enforcer == nil {
		response.FailCode(c, bizerr.Internal.WithMessage("casbin enforcer not initialized"))
		return
	}
	var req struct {
		Sub string `json:"sub" binding:"required"`
		Obj string `json:"obj" binding:"required"`
		Act string `json:"act" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, bizerr.BadRequest.WithMessage(err.Error()))
		return
	}
	ok, err := h.enforcer.RemovePolicy(req.Sub, req.Obj, req.Act)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, gin.H{"deleted": ok, "sub": req.Sub, "obj": req.Obj, "act": req.Act})
}

// ListRBACRoles GET /api/v1/admin/rbac/roles
func (h *AdminHandler) ListRBACRoles(c *gin.Context) {
	if h.enforcer == nil {
		response.FailCode(c, bizerr.Internal.WithMessage("casbin enforcer not initialized"))
		return
	}
	// 合并策略中的主体（角色）与 g 表中的角色
	roleSet := make(map[string]struct{})
	policies, err := h.enforcer.GetPolicy()
	if err != nil {
		response.Fail(c, err)
		return
	}
	for _, p := range policies {
		if len(p) > 0 {
			roleSet[p[0]] = struct{}{}
		}
	}
	allRoles, err := h.enforcer.GetAllRoles()
	if err != nil {
		response.Fail(c, err)
		return
	}
	for _, r := range allRoles {
		roleSet[r] = struct{}{}
	}
	roles := make([]string, 0, len(roleSet))
	for r := range roleSet {
		roles = append(roles, r)
	}
	response.Success(c, roles)
}

// AssignRBACRole POST /api/v1/admin/rbac/roles
func (h *AdminHandler) AssignRBACRole(c *gin.Context) {
	if h.enforcer == nil {
		response.FailCode(c, bizerr.Internal.WithMessage("casbin enforcer not initialized"))
		return
	}
	var req struct {
		UserID uint64 `json:"user_id" binding:"required"`
		Role   string `json:"role" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, bizerr.BadRequest.WithMessage(err.Error()))
		return
	}
	subject := fmt.Sprintf("user:%d", req.UserID)
	ok, err := h.enforcer.AddRoleForUser(subject, req.Role)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, gin.H{"assigned": ok, "subject": subject, "role": req.Role})
}

// Mount 挂载路由
// /admin/login 为公开登录入口（不挂 AdminOnly）；其余 /admin/* 需鉴权 + AdminOnly
func (h *AdminHandler) Mount(rg *gin.RouterGroup, mw gin.HandlerFunc) {
	// 后台登录入口：无需鉴权、无需 AdminOnly
	rg.POST("/admin/login", h.Login)

	adminOnly := auth.AdminOnly()
	g := rg.Group("/admin", mw, adminOnly)
	g.GET("/stats", h.Stats)
	g.GET("/users", h.ListUsers)
	g.PATCH("/users/:id", h.UpdateUser)
	g.GET("/groups", h.ListGroups)
	g.POST("/groups", h.CreateGroup)
	g.PATCH("/groups/:id", h.UpdateGroup)
	g.DELETE("/groups/:id", h.DeleteGroup)
	g.GET("/photos", h.ListAllPhotos)

	// RBAC 管理
	rbacG := g.Group("/rbac")
	rbacG.GET("/policies", h.ListRBACPolicies)
	rbacG.POST("/policies", h.AddRBACPolicy)
	rbacG.DELETE("/policies", h.DeleteRBACPolicy)
	rbacG.GET("/roles", h.ListRBACRoles)
	rbacG.POST("/roles", h.AssignRBACRole)
}
