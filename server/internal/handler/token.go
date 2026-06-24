// Package handler API Token HTTP 接口
package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/ywty/server/internal/auth"
	bizerr "github.com/ywty/server/internal/errors"
	"github.com/ywty/server/internal/response"
	"github.com/ywty/server/internal/service"
)

type TokenHandler struct {
	svc *service.TokenService
}

func NewTokenHandler(svc *service.TokenService) *TokenHandler { return &TokenHandler{svc: svc} }

// List GET /api/v1/tokens
func (h *TokenHandler) List(c *gin.Context) {
	uid, _ := auth.CurrentUserID(c)
	rows, err := h.svc.List(c.Request.Context(), uid)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, rows)
}

// Create POST /api/v1/tokens
func (h *TokenHandler) Create(c *gin.Context) {
	uid, _ := auth.CurrentUserID(c)
	var req struct {
		Name      string `json:"name" binding:"required"`
		Abilities []any  `json:"abilities"`
		TTLDays   int    `json:"ttl_days"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, bizerr.BadRequest.WithMessage(err.Error()))
		return
	}
	raw, info, err := h.svc.IssueToken(c.Request.Context(), uid, req.Name, req.Abilities, req.TTLDays)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, gin.H{
		"token":        raw, // 明文 token 仅此一次返回
		"access_token": raw,
		"info":         info,
	})
}

// Revoke DELETE /api/v1/tokens/:id
func (h *TokenHandler) Revoke(c *gin.Context) {
	uid, _ := auth.CurrentUserID(c)
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.svc.Revoke(c.Request.Context(), uid, id); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, gin.H{"id": id, "revoked": true})
}

func (h *TokenHandler) Mount(rg *gin.RouterGroup, mw gin.HandlerFunc) {
	g := rg.Group("/tokens")
	g.GET("", mw, h.List)
	g.POST("", mw, h.Create)
	g.DELETE("/:id", mw, h.Revoke)
}
