// Package auth 提供 JWT 颁发、解析与上下文注入能力
package auth

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/ywty/server/internal/config"
)

// Claims 自定义 JWT 载荷
type Claims struct {
	UserID   uint64 `json:"uid"`
	Username string `json:"usr"`
	IsAdmin  bool   `json:"adm"`
	jwt.RegisteredClaims
}

// Issuer JWT 颁发者
type Issuer struct {
	secret  []byte
	issuer  string
	access  time.Duration
	refresh time.Duration
}

// NewIssuer 创建 JWT 颁发器
func NewIssuer(cfg config.AuthJWT) *Issuer {
	return &Issuer{
		secret:  []byte(cfg.Secret),
		issuer:  cfg.Issuer,
		access:  time.Duration(cfg.AccessExpire) * time.Second,
		refresh: time.Duration(cfg.RefreshExpire) * time.Second,
	}
}

// AccessExpire 返回 access 有效期
func (i *Issuer) AccessExpire() time.Duration { return i.access }

// RefreshExpire 返回 refresh 有效期
func (i *Issuer) RefreshExpire() time.Duration { return i.refresh }

// Issue 生成 access token
func (i *Issuer) Issue(userID uint64, username string, isAdmin bool) (string, time.Time, error) {
	exp := time.Now().Add(i.access)
	claims := Claims{
		UserID:   userID,
		Username: username,
		IsAdmin:  isAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    i.issuer,
			Subject:   fmt.Sprintf("%d", userID),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(exp),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(i.secret)
	if err != nil {
		return "", time.Time{}, err
	}
	return signed, exp, nil
}

// IssueRefresh 生成 refresh token（长有效期，subject 携带 userID）
func (i *Issuer) IssueRefresh(userID uint64) (string, time.Time, error) {
	exp := time.Now().Add(i.refresh)
	claims := jwt.RegisteredClaims{
		Issuer:    i.issuer,
		Subject:   fmt.Sprintf("%d", userID),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(exp),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(i.secret)
	if err != nil {
		return "", time.Time{}, err
	}
	return signed, exp, nil
}

// ParseRefresh 解析 refresh token（只验证签名和有效期，返回 userID）
func (i *Issuer) ParseRefresh(tokenStr string) (uint64, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &jwt.RegisteredClaims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return i.secret, nil
	})
	if err != nil {
		return 0, err
	}
	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok || !token.Valid {
		return 0, errors.New("invalid token")
	}
	// refresh token 的 userID 存在 Subject 中
	return strconv.ParseUint(claims.Subject, 10, 64)
}
// Parse 解析并校验 access token
func (i *Issuer) Parse(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return i.secret, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}
