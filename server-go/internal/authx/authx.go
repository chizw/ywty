// Package authx 移植 Sanctum 令牌认证、会话认证、CheckTokenPermission
// 与 Initialize（角色组解析）的完整语义。
package authx

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/chizw/ywty/server-go/internal/config"
	"github.com/chizw/ywty/server-go/internal/model"
	"github.com/chizw/ywty/server-go/internal/support/r"
	"gorm.io/gorm"
)

const tokenableType = "App\\Models\\User"

// TokenableType Sanctum tokenable 类型（与 PHP Eloquent 类名一致）。
const TokenableType = "App\\Models\\User"

// Token personal_access_tokens 行（仅认证所需字段）。
type Token struct {
	ID         int64
	UserID     int64
	Name       string
	Abilities  []string
	ExpiresAt  *time.Time
	LastUsedAt *time.Time
}

// Ctx 请求级认证上下文。
type Ctx struct {
	User  *model.User
	Token *Token       // Bearer 令牌认证时有值；会话认证为 nil
	Group *model.Group // Initialize 解析出的角色组（含游客组）
	File  *UploadCtx   // 上传相关（M2 使用）
}

// UploadCtx 预留：上传频率限制等上下文数据。
type UploadCtx struct{}

type ctxKey struct{}

// From 取请求上下文；未认证时 User 为 nil。
func From(req *http.Request) *Ctx {
	if v, ok := req.Context().Value(ctxKey{}).(*Ctx); ok {
		return v
	}
	return &Ctx{}
}

// With 把上下文写入 request。
func With(req *http.Request, c *Ctx) context.Context {
	return context.WithValue(req.Context(), ctxKey{}, c)
}

// ---------- 令牌 ----------

// HashToken 等价 Sanctum：sha256(plainTextToken) 的 hex。
func HashToken(plain string) string {
	sum := sha256.Sum256([]byte(plain))
	return hex.EncodeToString(sum[:])
}

// patRow personal_access_tokens 插入载体（借助 gorm.Create 回填自增 ID）。
type patRow struct {
	ID            int64      `gorm:"primaryKey;column:id"`
	TokenableType string     `gorm:"column:tokenable_type"`
	TokenableID   int64      `gorm:"column:tokenable_id"`
	Name          string     `gorm:"column:name"`
	Token         string     `gorm:"column:token"`
	Abilities     string     `gorm:"column:abilities"`
	ExpiresAt     *time.Time `gorm:"column:expires_at"`
}

func (patRow) TableName() string { return "personal_access_tokens" }

// CreateToken 等价 $user->createToken()：明文 "{id}|{40随机}" 返回一次，库存 sha256。
func CreateToken(gdb *gorm.DB, userID int64, name string, abilities []string, expiresAt *time.Time) (plain string, err error) {
	random := make([]byte, 20)
	if _, err = rand.Read(random); err != nil {
		return "", err
	}
	plainPart := hex.EncodeToString(random) // 40 字符
	if abilities == nil {
		abilities = []string{"*"}
	}
	abilitiesJSON := "[]"
	if len(abilities) > 0 {
		abilitiesJSON = `["` + strings.Join(abilities, `","`) + `"]`
	}
	row := patRow{
		TokenableType: tokenableType,
		TokenableID:   userID,
		Name:          name,
		Token:         "pending",
		Abilities:     abilitiesJSON,
		ExpiresAt:     expiresAt,
	}
	if err = gdb.Create(&row).Error; err != nil {
		return "", err
	}
	id := row.ID
	plain = fmt.Sprintf("%d|%s", id, plainPart)
	hashed := HashToken(plain)
	if err = gdb.Exec("UPDATE `personal_access_tokens` SET `token` = ? WHERE `id` = ?", hashed, id).Error; err != nil {
		return "", err
	}
	return plain, nil
}

// findToken 按 "{id}|{plain}" 查找有效令牌。
func findToken(gdb *gorm.DB, plain string) (*Token, error) {
	idPart, _, ok := strings.Cut(plain, "|")
	if !ok {
		return nil, nil
	}
	var id int64
	if _, err := fmt.Sscanf(idPart, "%d", &id); err != nil {
		return nil, nil
	}
	var row struct {
		ID         int64      `gorm:"column:id"`
		UserID     int64      `gorm:"column:tokenable_id"`
		Name       string     `gorm:"column:name"`
		Abilities  *string    `gorm:"column:abilities"`
		LastUsedAt *time.Time `gorm:"column:last_used_at"`
		ExpiresAt  *time.Time `gorm:"column:expires_at"`
	}
	err := gdb.Raw(
		"SELECT `id`, `tokenable_id`, `name`, `abilities`, `last_used_at`, `expires_at` FROM `personal_access_tokens` "+
			"WHERE `id` = ? AND `token` = ? AND `tokenable_type` = ?", id, HashToken(plain), tokenableType,
	).Scan(&row).Error
	if err != nil {
		return nil, err
	}
	if row.ID == 0 {
		return nil, nil
	}
	t := &Token{ID: row.ID, UserID: row.UserID, Name: row.Name, Abilities: []string{"*"}, LastUsedAt: row.LastUsedAt, ExpiresAt: row.ExpiresAt}
	if row.Abilities != nil {
		_ = parseAbilities(*row.Abilities, &t.Abilities)
	}
	if t.ExpiresAt != nil && !t.ExpiresAt.IsZero() && t.ExpiresAt.Before(time.Now()) {
		return nil, nil // 已过期
	}
	return t, nil
}

// ---------- 会话 ----------

// SessionCookieName 对齐 Laravel：SESSION_COOKIE 或 slug(APP_NAME)+"_session"。
func SessionCookieName(cfg *config.Config) string {
	if cfg != nil {
		if v := envOr("SESSION_COOKIE", ""); v != "" {
			return v
		}
		return slugify(cfg.AppName) + "_session"
	}
	return "laravel_session"
}

func envOr(k, def string) string {
	if v, ok := os.LookupEnv(k); ok && strings.TrimSpace(v) != "" {
		return strings.TrimSpace(v)
	}
	return def
}

// CreateSession 写入 sessions 表并返回会话 ID。
func CreateSession(gdb *gorm.DB, userID int64, ip, agent string) (string, error) {
	raw := make([]byte, 20)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	sid := hex.EncodeToString(raw)
	if err := gdb.Exec(
		"INSERT INTO `sessions` (`id`, `user_id`, `ip_address`, `user_agent`, `payload`, `last_activity`) VALUES (?, ?, ?, ?, ?, ?)",
		sid, userID, ip, agent, "{}", time.Now().Unix(),
	).Error; err != nil {
		return "", err
	}
	return sid, nil
}

func DestroySession(gdb *gorm.DB, sid string) {
	_ = gdb.Exec("DELETE FROM `sessions` WHERE `id` = ?", sid).Error
}

// ---------- 用户 ----------

func userByID(gdb *gorm.DB, id int64) (*model.User, error) {
	var u model.User
	err := gdb.Where("id = ?", id).First(&u).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// ResolveUser 尝试 Bearer 令牌 → 会话，返回 (user, token, cookieName/sid)。
func ResolveUser(gdb *gorm.DB, cfg *config.Config, req *http.Request) (*model.User, *Token) {
	if h := req.Header.Get("Authorization"); h != "" {
		if plain, ok := strings.CutPrefix(h, "Bearer "); ok {
			plain = strings.TrimSpace(plain)
			if tok, err := findToken(gdb, plain); err == nil && tok != nil {
				if u, err := userByID(gdb, tok.UserID); err == nil {
					// last_used_at 节流更新：距上次超过 60 秒才写库
					if tok.LastUsedAt == nil || time.Since(*tok.LastUsedAt) > time.Minute {
						_ = gdb.Exec("UPDATE `personal_access_tokens` SET `last_used_at` = CURRENT_TIMESTAMP WHERE `id` = ?", tok.ID).Error
					}
					return u, tok
				}
			}
		}
	}
	if c, err := req.Cookie(SessionCookieName(cfg)); err == nil && c.Value != "" {
		var userID *int64
		if err := gdb.Raw("SELECT `user_id` FROM `sessions` WHERE `id` = ?", c.Value).Scan(&userID).Error; err == nil && userID != nil {
			if u, err := userByID(gdb, *userID); err == nil {
				_ = gdb.Exec("UPDATE `sessions` SET `last_activity` = ? WHERE `id` = ?", time.Now().Unix(), c.Value).Error
				return u, nil
			}
		}
	}
	return nil, nil
}

// ---------- 中间件 ----------

// Auth 认证中间件：未认证返回 401 envelope。
func Auth(gdb *gorm.DB, cfg *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			u, tok := ResolveUser(gdb, cfg, req)
			if u == nil {
				r.ErrorWithCode(w, http.StatusUnauthorized, "Unauthenticated.")
				return
			}
			ctx := From(req)
			ctx.User = u
			ctx.Token = tok
			next.ServeHTTP(w, req.WithContext(With(req, ctx)))
		})
	}
}

// OptionalAuth 尽力解析认证信息但不强制（对齐 Sanctum 对公共路由的行为）。
func OptionalAuth(gdb *gorm.DB, cfg *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			u, tok := ResolveUser(gdb, cfg, req)
			ctx := From(req)
			ctx.User = u
			ctx.Token = tok
			next.ServeHTTP(w, req.WithContext(With(req, ctx)))
		})
	}
}

// Initialize 解析当前请求的角色组：登录用户取有效期内的最新用户组，否则游客组。
// 必须先于其它需要 group 的中间件执行。
func Initialize(gdb *gorm.DB, cfg *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			ctx := From(req)
			if ctx.User == nil {
				u, tok := ResolveUser(gdb, cfg, req)
				ctx.User, ctx.Token = u, tok
			}
			ctx.Group = ResolveGroup(gdb, ctx.User)
			next.ServeHTTP(w, req.WithContext(With(req, ctx)))
		})
	}
}

// ResolveGroup 用户有效组 → 游客组 → nil。
func ResolveGroup(gdb *gorm.DB, u *model.User) *model.Group {
	if u != nil {
		var g model.Group
		err := gdb.Raw(
			"SELECT g.* FROM `groups` g INNER JOIN `user_groups` ug ON ug.group_id = g.id "+
				"WHERE ug.user_id = ? AND ug.deleted_at IS NULL AND g.deleted_at IS NULL "+
				"AND (ug.expired_at > CURRENT_TIMESTAMP OR ug.expired_at IS NULL) "+
				"ORDER BY ug.created_at DESC LIMIT 1", u.ID,
		).Scan(&g).Error
		if err == nil && g.ID != 0 {
			return &g
		}
	}
	var guest model.Group
	err := gdb.Where("is_guest = 1").First(&guest).Error
	if err != nil {
		return nil
	}
	return &guest
}

// ---------- API 权限映射（ApiPermission.php 移植） ----------

type Permission string

const (
	PermUserProfileRead   Permission = "user:profile:read"
	PermUserProfileWrite  Permission = "user:profile:write"
	PermUserTokenRead     Permission = "user:token:read"
	PermUserTokenWrite    Permission = "user:token:write"
	PermUserGroupRead     Permission = "user:group:read"
	PermUserGroupWrite    Permission = "user:group:write"
	PermUserCapacityRead  Permission = "user:capacity:read"
	PermUserCapacityWrite Permission = "user:capacity:write"
	PermUserAlbumRead     Permission = "user:album:read"
	PermUserAlbumWrite    Permission = "user:album:write"
	PermUserPhotoRead     Permission = "user:photo:read"
	PermUserPhotoWrite    Permission = "user:photo:write"
	PermUserShareRead     Permission = "user:share:read"
	PermUserShareWrite    Permission = "user:share:write"
	PermUserTicketRead    Permission = "user:ticket:read"
	PermUserTicketWrite   Permission = "user:ticket:write"
	PermUserOrderRead     Permission = "user:order:read"
	PermUserOrderWrite    Permission = "user:order:write"
	PermOAuthRead         Permission = "oauth:read"
	PermOAuthWrite        Permission = "oauth:write"
	PermExploreRead       Permission = "explore:read"
	PermExploreWrite      Permission = "explore:write"
	PermUploadWrite       Permission = "upload:write"
	PermBasic             Permission = "basic"
)

type permMap map[string]map[string]Permission // path -> METHOD -> permission

// routePermissionMap 完整移植 ApiPermission::getRoutePermissionMap()。
func routePermissionMap() permMap {
	m := permMap{}
	add := func(path string, rules map[string]Permission) { m[path] = rules }
	add("api/v2/user/profile", map[string]Permission{"GET": PermUserProfileRead, "POST": PermUserProfileWrite})
	add("api/v2/user/setting", map[string]Permission{"POST": PermUserProfileWrite})
	add("api/v2/user/bind_phone", map[string]Permission{"POST": PermUserProfileWrite})
	add("api/v2/user/bind_email", map[string]Permission{"POST": PermUserProfileWrite})
	add("api/v2/user/tokens", map[string]Permission{"GET": PermUserTokenRead, "POST": PermUserTokenWrite, "DELETE": PermUserTokenWrite})
	add("api/v2/user/tokens/user-permissions", map[string]Permission{"GET": PermUserTokenRead})
	add("api/v2/user/groups", map[string]Permission{"GET": PermUserGroupRead, "DELETE": PermUserGroupWrite})
	add("api/v2/user/capacities", map[string]Permission{"GET": PermUserCapacityRead, "DELETE": PermUserCapacityWrite})
	add("api/v2/user/albums", map[string]Permission{"GET": PermUserAlbumRead, "POST": PermUserAlbumWrite, "PUT": PermUserAlbumWrite, "DELETE": PermUserAlbumWrite})
	add("api/v2/user/albums/*", map[string]Permission{"GET": PermUserAlbumRead, "POST": PermUserAlbumWrite, "PUT": PermUserAlbumWrite, "DELETE": PermUserAlbumWrite})
	add("api/v2/user/photos", map[string]Permission{"GET": PermUserPhotoRead, "DELETE": PermUserPhotoWrite})
	add("api/v2/user/photos/*", map[string]Permission{"GET": PermUserPhotoRead, "POST": PermUserPhotoWrite, "PUT": PermUserPhotoWrite, "DELETE": PermUserPhotoWrite})
	add("api/v2/user/photos/timeline", map[string]Permission{"GET": PermUserPhotoRead})
	add("api/v2/user/shares", map[string]Permission{"GET": PermUserShareRead, "POST": PermUserShareWrite, "PUT": PermUserShareWrite, "DELETE": PermUserShareWrite})
	add("api/v2/user/tickets", map[string]Permission{"GET": PermUserTicketRead, "POST": PermUserTicketWrite, "DELETE": PermUserTicketWrite})
	add("api/v2/user/tickets/*", map[string]Permission{"GET": PermUserTicketRead, "POST": PermUserTicketWrite, "PUT": PermUserTicketWrite})
	add("api/v2/user/orders", map[string]Permission{"GET": PermUserOrderRead, "POST": PermUserOrderWrite})
	add("api/v2/user/orders/*", map[string]Permission{"GET": PermUserOrderRead, "POST": PermUserOrderWrite, "PUT": PermUserOrderWrite})
	add("api/v2/user/orders/preview", map[string]Permission{"POST": PermUserOrderRead})
	add("api/v2/oauth/binds", map[string]Permission{"GET": PermOAuthRead})
	add("api/v2/oauth/*", map[string]Permission{"POST": PermOAuthWrite, "DELETE": PermOAuthWrite})
	add("api/v2/explore/*", map[string]Permission{"GET": PermExploreRead, "POST": PermExploreWrite, "DELETE": PermExploreWrite})
	add("api/v2/shares/*", map[string]Permission{"GET": PermExploreRead, "POST": PermExploreWrite, "DELETE": PermExploreWrite})
	add("api/v2/upload", map[string]Permission{"POST": PermUploadWrite})
	add("api/v1/upload", map[string]Permission{"POST": PermUploadWrite})
	add("api/v1/images/tokens", map[string]Permission{"POST": PermUploadWrite})
	add("api/v2/configs", map[string]Permission{"GET": PermBasic})
	add("api/v2/group", map[string]Permission{"GET": PermBasic})
	add("api/v2/notices/*", map[string]Permission{"GET": PermBasic})
	add("api/v2/pages/*", map[string]Permission{"GET": PermBasic})
	add("api/v2/plans/*", map[string]Permission{"GET": PermBasic})
	return m
}

var cachedPermMap = routePermissionMap()

// RoutePermission 精确匹配 → 通配符前缀匹配；未命中返回 false。
func RoutePermission(path, method string) (Permission, bool) {
	if rule, ok := cachedPermMap[path]; ok {
		if p, ok := rule[method]; ok {
			return p, true
		}
		return "", false
	}
	for route, rule := range cachedPermMap {
		if strings.HasSuffix(route, "*") {
			prefix := strings.TrimSuffix(route, "*")
			if strings.HasPrefix(path, prefix) {
				if p, ok := rule[method]; ok {
					return p, true
				}
			}
		}
	}
	return "", false
}

// CheckTokenPermission 对齐 PHP 中间件：会话认证跳过；'*' 全通过；BASIC 强制附加。
func CheckTokenPermission(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		ctx := From(req)
		if ctx.Token == nil {
			next.ServeHTTP(w, req)
			return
		}
		abilities := ctx.Token.Abilities
		has := func(p Permission) bool {
			for _, a := range abilities {
				if a == string(p) {
					return true
				}
			}
			return false
		}
		if has("*") {
			next.ServeHTTP(w, req)
			return
		}
		// 等价 PHP：BASIC 总是被强制附加到能力列表
		hasBasic := has(PermBasic)
		allowed := func(p Permission) bool { return has(p) || (hasBasic && p == PermBasic) }
		perm, ok := RoutePermission(strings.TrimPrefix(req.URL.Path, "/"), req.Method)
		if ok && !allowed(perm) {
			r.ErrorWithCode(w, http.StatusForbidden, "您的令牌没有权限访问此API")
			return
		}
		next.ServeHTTP(w, req)
	})
}

// ---------- 杂项 ----------

func ClientIP(req *http.Request) string {
	if xff := req.Header.Get("X-Forwarded-For"); xff != "" {
		if host, _, err := net.SplitHostPort(xff); err == nil {
			return host
		}
		return strings.TrimSpace(strings.Split(xff, ",")[0])
	}
	if xr := req.Header.Get("X-Real-IP"); xr != "" {
		return xr
	}
	host, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		return req.RemoteAddr
	}
	return host
}

func parseAbilities(jsonStr string, out *[]string) error {
	s := strings.TrimSpace(jsonStr)
	if s == "" || s == "null" {
		*out = []string{"*"}
		return nil
	}
	var arr []string
	if err := json.Unmarshal([]byte(s), &arr); err != nil {
		return err
	}
	*out = arr
	return nil
}

func slugify(s string) string {
	s = strings.ToLower(s)
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		} else {
			b.WriteByte('_')
		}
	}
	return b.String()
}
