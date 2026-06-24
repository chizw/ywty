// Package handler 用户容量接口
package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/ywty/server/internal/auth"
	"github.com/ywty/server/internal/response"
	"github.com/ywty/server/internal/service"
)

// CapacityHandler 容量接口
type CapacityHandler struct {
	svc *service.CapacityService
}

func NewCapacityHandler(svc *service.CapacityService) *CapacityHandler {
	return &CapacityHandler{svc: svc}
}

// Get GET /api/v1/capacity
func (h *CapacityHandler) Get(c *gin.Context) {
	uid, ok := auth.CurrentUserID(c)
	if !ok {
		response.Success(c, gin.H{
			"used":         0,
			"capacity":     0,
			"max_image":    10 * 1024 * 1024,
			"unlimited":    true,
			"used_percent": 0,
			"remain":       0,
		})
		return
	}
	info, err := h.svc.GetUserCapacity(c.Request.Context(), uid)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, info)
}

func (h *CapacityHandler) Mount(rg *gin.RouterGroup, mw gin.HandlerFunc) {
	g := rg.Group("/capacity", mw)
	g.GET("", h.Get)
}
