// Package handler License 管理接口
package handler

import (
	"github.com/gin-gonic/gin"

	bizerr "github.com/ywty/server/internal/errors"
	"github.com/ywty/server/internal/response"
	"github.com/ywty/server/internal/service"
)

// LicenseHandler License 管理接口
type LicenseHandler struct {
	svc *service.LicenseService
}

// NewLicenseHandler 创建 License 管理接口
func NewLicenseHandler(svc *service.LicenseService) *LicenseHandler {
	return &LicenseHandler{svc: svc}
}

// Mount 挂载路由
func (h *LicenseHandler) Mount(rg *gin.RouterGroup, mw gin.HandlerFunc) {
	g := rg.Group("/admin/license")
	if mw != nil {
		g.Use(mw)
	}
	g.GET("", h.Get)
	g.POST("/activate", h.Activate)
	g.GET("/check", h.Check)
}

// Get 获取当前 License
// GET /api/v1/admin/license
func (h *LicenseHandler) Get(c *gin.Context) {
	license, err := h.svc.GetLicense(c.Request.Context())
	if err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, license)
}

// Activate 激活 License
// POST /api/v1/admin/license/activate
func (h *LicenseHandler) Activate(c *gin.Context) {
	var req struct {
		Key string `json:"key" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, bizerr.BadRequest.WithMessage(err.Error()))
		return
	}

	license, err := h.svc.ActivateLicense(c.Request.Context(), req.Key)
	if err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, license)
}

// Check 检查 License 状态
// GET /api/v1/admin/license/check
func (h *LicenseHandler) Check(c *gin.Context) {
	valid, reason, err := h.svc.CheckLicense(c.Request.Context())
	if err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, gin.H{
		"valid":  valid,
		"reason": reason,
	})
}
