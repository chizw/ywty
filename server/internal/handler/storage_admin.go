// Package handler 存储管理 + 跨存储复制 + 直传签名
package handler

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/ywty/server/internal/auth"
	bizerr "github.com/ywty/server/internal/errors"
	"github.com/ywty/server/internal/model"
	"github.com/ywty/server/internal/response"
	"github.com/ywty/server/internal/storage"
)

// StorageHandler 存储管理接口
type StorageHandler struct {
	db *gorm.DB
}

func NewStorageHandler(db *gorm.DB) *StorageHandler { return &StorageHandler{db: db} }

// ListDrivers GET /api/v1/admin/storage/drivers
func (h *StorageHandler) ListDrivers(c *gin.Context) {
	names := storage.Drivers()
	out := make([]gin.H, 0, len(names))
	for _, n := range names {
		out = append(out, gin.H{
			"name": n,
			"id":   n,
		})
	}
	response.Success(c, gin.H{"drivers": out, "count": len(out)})
}

// ListStorages GET /api/v1/admin/storages
func (h *StorageHandler) ListStorages(c *gin.Context) {
	var rows []model.Storage
	if err := h.db.Order("id ASC").Find(&rows).Error; err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, rows)
}

// CreateStorage POST /api/v1/admin/storages
func (h *StorageHandler) CreateStorage(c *gin.Context) {
	var req struct {
		Name     string `json:"name" binding:"required"`
		Provider string `json:"provider" binding:"required"`
		Intro    string `json:"intro"`
		Prefix   string `json:"prefix"`
		Options  gin.H  `json:"options"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, bizerr.BadRequest.WithMessage(err.Error()))
		return
	}
	// 校验驱动名是否已注册（不实际实例化，避免无 options 报错）
	supported := false
	for _, n := range storage.Drivers() {
		if n == req.Provider {
			supported = true
			break
		}
	}
	if !supported {
		response.FailCode(c, bizerr.BadRequest.WithMessage("unsupported driver: "+req.Provider))
		return
	}
	opts := model.JSONMap{}
	if req.Options != nil {
		for k, v := range req.Options {
			opts[k] = v
		}
	}
	rec := &model.Storage{
		Name:     req.Name,
		Provider: req.Provider,
		Intro:    req.Intro,
		Prefix:   req.Prefix,
		Options:  opts,
	}
	if err := h.db.Create(rec).Error; err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, rec)
}

// UpdateStorage PATCH /api/v1/admin/storages/:id
func (h *StorageHandler) UpdateStorage(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Name    *string `json:"name"`
		Intro   *string `json:"intro"`
		Prefix  *string `json:"prefix"`
		Options gin.H   `json:"options"`
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
	if req.Prefix != nil {
		updates["prefix"] = *req.Prefix
	}
	if req.Options != nil {
		cfg := model.JSONMap{}
		for k, v := range req.Options {
			cfg[k] = v
		}
		updates["options"] = cfg
	}
	res := h.db.Model(&model.Storage{}).Where("id = ?", id).Updates(updates)
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

// DeleteStorage DELETE /api/v1/admin/storages/:id
func (h *StorageHandler) DeleteStorage(c *gin.Context) {
	id := c.Param("id")
	res := h.db.Delete(&model.Storage{}, id)
	if res.Error != nil {
		response.Fail(c, res.Error)
		return
	}
	if res.RowsAffected == 0 {
		response.FailCode(c, bizerr.ResourceNotFound)
		return
	}
	response.Success(c, gin.H{"id": id, "deleted": true})
}

// SignURL GET /api/v1/storage/sign?key=...&storage_id=...
func (h *StorageHandler) SignURL(c *gin.Context) {
	key := c.Query("key")
	if key == "" {
		response.FailCode(c, bizerr.BadRequest.WithMessage("key required"))
		return
	}
	storageID := c.Query("storage_id")
	var s model.Storage
	if storageID != "" {
		if err := h.db.First(&s, storageID).Error; err != nil {
			response.Fail(c, err)
			return
		}
	} else {
		// 没有 storage 表默认存储配置 → 退化为 local
		drv, err := storage.Get(storage.DriverNameLocal, storage.Options{
			"root":       "storage/uploads",
			"public_url": "",
		})
		if err != nil {
			response.Fail(c, err)
			return
		}
		if url, signErr := drv.SignURL(c.Request.Context(), key, 3600); signErr == nil {
			response.Success(c, gin.H{"url": url})
			return
		}
		// 本地驱动无签名时退化到公开 URL
		response.Success(c, gin.H{"url": drv.PublicURL(key)})
		return
	}
	drv, err := storage.Get(s.Provider, storage.Options(s.Options))
	if err != nil {
		response.Fail(c, fmt.Errorf("init driver %s: %w", s.Provider, err))
		return
	}
	if url, signErr := drv.SignURL(c.Request.Context(), key, 3600); signErr == nil {
		response.Success(c, gin.H{"url": url})
		return
	}
	// 不支持签名时退化到公开 URL
	response.Success(c, gin.H{"url": drv.PublicURL(key)})
}

// Copy 跨存储复制 POST /api/v1/admin/storage/copy
func (h *StorageHandler) Copy(c *gin.Context) {
	var req struct {
		FromStorageID uint64 `json:"from_storage_id"`
		ToStorageID   uint64 `json:"to_storage_id" binding:"required"`
		Key           string `json:"key" binding:"required"`
		DestKey       string `json:"dest_key"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, bizerr.BadRequest.WithMessage(err.Error()))
		return
	}
	// 源默认用 local（若未指定 FromStorageID）
	var src, dst storage.Driver
	var err error
	if req.FromStorageID == 0 {
		src, err = storage.Get(storage.DriverNameLocal, storage.Options{
			"root":       "storage/uploads",
			"public_url": "",
		})
	} else {
		var sm model.Storage
		if e := h.db.First(&sm, req.FromStorageID).Error; e != nil {
			response.Fail(c, e)
			return
		}
		src, err = storage.Get(sm.Provider, storage.Options(sm.Options))
	}
	if err != nil {
		response.Fail(c, err)
		return
	}
	var tm model.Storage
	if e := h.db.First(&tm, req.ToStorageID).Error; e != nil {
		response.Fail(c, e)
		return
	}
	dst, err = storage.Get(tm.Provider, storage.Options(tm.Options))
	if err != nil {
		response.Fail(c, err)
		return
	}
	if req.DestKey == "" {
		req.DestKey = req.Key
	}
	// 读 → 写
	rd, err := src.Get(c.Request.Context(), req.Key)
	if err != nil {
		response.Fail(c, err)
		return
	}
	defer rd.Close()
	_, err = dst.Put(c.Request.Context(), req.DestKey, rd, 0, "")
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, gin.H{
		"from":    req.FromStorageID,
		"to":      req.ToStorageID,
		"key":     req.Key,
		"dest":    req.DestKey,
		"success": true,
	})
}

func (h *StorageHandler) Mount(rg *gin.RouterGroup, mw gin.HandlerFunc) {
	adminOnly := auth.AdminOnly()
	g := rg.Group("/admin/storage", mw, adminOnly)
	g.GET("/drivers", h.ListDrivers)
	g.GET("/list", h.ListStorages)
	g.POST("/create", h.CreateStorage)
	g.PATCH("/update/:id", h.UpdateStorage)
	g.DELETE("/delete/:id", h.DeleteStorage)
	g.POST("/copy", h.Copy)

	// 公开直传签名（仅登录后）
	rg.Group("/storage", mw).GET("/sign", h.SignURL)
}
