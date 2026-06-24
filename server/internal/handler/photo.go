package handler

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/ywty/server/internal/auth"
	bizerr "github.com/ywty/server/internal/errors"
	"github.com/ywty/server/internal/logger"
	"github.com/ywty/server/internal/response"
	"github.com/ywty/server/internal/service"
)

// PhotoHandler 图片相关接口
type PhotoHandler struct {
	svc *service.PhotoService
}

// NewPhotoHandler 创建图片处理器
func NewPhotoHandler(svc *service.PhotoService) *PhotoHandler {
	return &PhotoHandler{svc: svc}
}

// Upload POST /api/v1/photos
// 支持单文件 / 多文件（file 字段重复）
// 表单字段：
//   - file:    文件
//   - name:    可选别名
//   - intro:   可选介绍
//   - is_public: 可选，是否公开
func (h *PhotoHandler) Upload(c *gin.Context) {
	uid, ok := auth.CurrentUserID(c)
	if !ok {
		response.FailCode(c, bizerr.Unauthorized)
		return
	}

	form, err := c.MultipartForm()
	if err != nil {
		response.FailCode(c, bizerr.BadRequest.WithMessage("invalid multipart form: "+err.Error()))
		return
	}
	files := form.File["file"]
	if len(files) == 0 {
		// 兼容单文件上传
		if f := form.File["image"]; len(f) > 0 {
			files = f
		}
	}
	if len(files) == 0 {
		response.FailCode(c, bizerr.BadRequest.WithMessage("missing file field"))
		return
	}

	results := make([]*service.UploadResult, 0, len(files))
	ip := c.ClientIP()
	for _, fh := range files {
		f, err := fh.Open()
		if err != nil {
			logger.L.Warn("open uploaded file failed", zap.String("name", fh.Filename), zap.Error(err))
			response.FailCode(c, bizerr.BadRequest.WithMessage("open file: "+err.Error()))
			return
		}
		// 优先取表单里的 name / intro
		opts := service.UploadOptions{
			Name:     firstNonEmpty(form.Value["name"], fh.Filename),
			Intro:    firstString(form.Value["intro"]),
			IsPublic: firstBool(form.Value["is_public"]),
		}
		res, err := h.svc.Upload(c.Request.Context(), uid, fh.Filename, f, fh.Size, opts)
		_ = f.Close()
		if err != nil {
			logger.L.Warn("upload failed",
				zap.String("filename", fh.Filename),
				zap.String("ip", ip),
				zap.Error(err),
			)
			response.Fail(c, err)
			return
		}
		results = append(results, res)
	}

	if len(results) == 1 {
		response.Success(c, results[0])
		return
	}
	response.Success(c, gin.H{"items": results, "count": len(results)})
}

// List GET /api/v1/photos?page=1&per_page=20
func (h *PhotoHandler) List(c *gin.Context) {
	uid, ok := auth.CurrentUserID(c)
	if !ok {
		response.FailCode(c, bizerr.Unauthorized)
		return
	}
	page, _ := strconv.Atoi(c.Query("page"))
	perPage, _ := strconv.Atoi(c.Query("per_page"))
	photos, total, err := h.svc.List(c.Request.Context(), uid, page, perPage)
	if err != nil {
		response.Fail(c, err)
		return
	}
	meta := response.PageMeta{
		CurrentPage: page,
		PerPage:     perPage,
		Total:       total,
		LastPage:    calcLastPage(total, perPage),
	}
	if meta.PerPage == 0 {
		meta.PerPage = 20
	}
	if meta.CurrentPage == 0 {
		meta.CurrentPage = 1
	}
	response.Page(c, photos, meta)
}

// Get GET /api/v1/photos/:id
func (h *PhotoHandler) Get(c *gin.Context) {
	uid, ok := auth.CurrentUserID(c)
	if !ok {
		response.FailCode(c, bizerr.Unauthorized)
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.FailCode(c, bizerr.BadRequest.WithMessage("invalid id"))
		return
	}
	photo, err := h.svc.Get(c.Request.Context(), id, uid)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, photo)
}

// Delete DELETE /api/v1/photos/:id
func (h *PhotoHandler) Delete(c *gin.Context) {
	uid, ok := auth.CurrentUserID(c)
	if !ok {
		response.FailCode(c, bizerr.Unauthorized)
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.FailCode(c, bizerr.BadRequest.WithMessage("invalid id"))
		return
	}
	if err := h.svc.Delete(c.Request.Context(), id, uid); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, gin.H{"id": id, "deleted": true})
}

// Update PATCH /api/v1/photos/:id
func (h *PhotoHandler) Update(c *gin.Context) {
	uid, ok := auth.CurrentUserID(c)
	if !ok {
		response.FailCode(c, bizerr.Unauthorized)
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.FailCode(c, bizerr.BadRequest.WithMessage("invalid id"))
		return
	}
	var req service.UpdatePhotoReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, bizerr.BadRequest.WithMessage(err.Error()))
		return
	}
	p, err := h.svc.Update(c.Request.Context(), id, uid, req)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, p)
}

// BatchDelete POST /api/v1/photos/batch-delete
func (h *PhotoHandler) BatchDelete(c *gin.Context) {
	uid, ok := auth.CurrentUserID(c)
	if !ok {
		response.FailCode(c, bizerr.Unauthorized)
		return
	}
	var req struct {
		IDs []uint64 `json:"ids" binding:"required,min=1"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, bizerr.BadRequest.WithMessage(err.Error()))
		return
	}
	n, err := h.svc.BatchDelete(c.Request.Context(), req.IDs, uid)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, gin.H{"deleted": n})
}

// MoveToAlbum POST /api/v1/photos/:id/move-to-album
func (h *PhotoHandler) MoveToAlbum(c *gin.Context) {
	uid, ok := auth.CurrentUserID(c)
	if !ok {
		response.FailCode(c, bizerr.Unauthorized)
		return
	}
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req struct {
		AlbumID uint64 `json:"album_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, bizerr.BadRequest.WithMessage(err.Error()))
		return
	}
	if err := h.svc.MoveToAlbum(c.Request.Context(), id, req.AlbumID, uid); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, gin.H{"photo_id": id, "album_id": req.AlbumID})
}

// Copy POST /api/v1/photos/:id/copy
func (h *PhotoHandler) Copy(c *gin.Context) {
	uid, ok := auth.CurrentUserID(c)
	if !ok {
		response.FailCode(c, bizerr.Unauthorized)
		return
	}
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	dup, err := h.svc.Copy(c.Request.Context(), id, uid)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, dup)
}

// ImageRedirect GET /i/:id  公开图片重定向
func (h *PhotoHandler) ImageRedirect(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.FailCode(c, bizerr.BadRequest.WithMessage("invalid id"))
		return
	}
	url, _, err := h.svc.ServeRedirect(c.Request.Context(), id)
	if err != nil {
		response.Fail(c, err)
		return
	}
	c.Redirect(302, url)
}

// PhotoRedirect 提供给 router 使用的顶层 handler
func PhotoRedirect(svc *service.PhotoService) gin.HandlerFunc {
	h := &PhotoHandler{svc: svc}
	return h.ImageRedirect
}

// Mount 注册到路由
func (h *PhotoHandler) Mount(rg *gin.RouterGroup, mw gin.HandlerFunc) {
	g := rg.Group("/photos")
	g.GET("", mw, h.List)
	g.GET("/:id", mw, h.Get)
	g.POST("", mw, h.Upload)
	g.POST("/batch-delete", mw, h.BatchDelete)
	g.PATCH("/:id", mw, h.Update)
	g.DELETE("/:id", mw, h.Delete)
	g.POST("/:id/move-to-album", mw, h.MoveToAlbum)
	g.POST("/:id/copy", mw, h.Copy)
}

func firstNonEmpty(values []string, fallback string) string {
	if len(values) > 0 {
		if v := strings.TrimSpace(values[0]); v != "" {
			return v
		}
	}
	return fallback
}

func firstString(values []string) string {
	if len(values) == 0 {
		return ""
	}
	return strings.TrimSpace(values[0])
}

func firstBool(values []string) bool {
	if len(values) == 0 {
		return false
	}
	b, _ := strconv.ParseBool(values[0])
	return b
}

func calcLastPage(total int64, perPage int) int {
	if perPage <= 0 {
		perPage = 20
	}
	if total == 0 {
		return 1
	}
	n := int((total + int64(perPage) - 1) / int64(perPage))
	if n == 0 {
		return 1
	}
	return n
}
