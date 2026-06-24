// Package handler 驱动管理接口（管理后台）
package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/ywty/server/internal/auth"
	"github.com/ywty/server/internal/notify"
	"github.com/ywty/server/internal/process"
	"github.com/ywty/server/internal/response"
	"github.com/ywty/server/internal/scan"
	"github.com/ywty/server/internal/social"
	"github.com/ywty/server/internal/storage"
)

// DriversHandler 驱动管理
type DriversHandler struct{}

func NewDriversHandler() *DriversHandler { return &DriversHandler{} }

// List GET /api/v1/admin/drivers
// 返回所有可用驱动：storage / sms / mail / social / scan / process
func (h *DriversHandler) List(c *gin.Context) {
	storageDrivers := make([]string, 0)
	for _, n := range storage.Drivers() {
		storageDrivers = append(storageDrivers, n)
	}
	response.Success(c, gin.H{
		"storage": storageDrivers,
		"sms":     notify.ListSMSDrivers(),
		"mail":    notify.ListMailDrivers(),
		"social":  social.Drivers(),
		"scan":    scan.Drivers(),
		"process": process.Drivers(),
	})
}

func (h *DriversHandler) Mount(rg *gin.RouterGroup, mw gin.HandlerFunc) {
	adminOnly := auth.AdminOnly()
	g := rg.Group("/admin/drivers", mw, adminOnly)
	g.GET("", h.List)
}
