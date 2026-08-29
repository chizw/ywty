// Package captchax 图形验证码（math 模式），对齐原版 'math' 配置：
// 120x36、TTL 60 秒、响应 {sensitive, key, img(dataURI)}。
package captchax

import (
	"net/http"

	"github.com/chizw/ywty/server-go/internal/cache"
	"github.com/chizw/ywty/server-go/internal/support/r"
	base64Captcha "github.com/mojocn/base64Captcha"
)

type Service struct {
	store *cacheStore
	drv   *base64Captcha.DriverMath
}

func New(c *cache.Cache) *Service {
	return &Service{
		store: &cacheStore{cache: c},
		drv:   base64Captcha.NewDriverMath(36, 120, 0, base64Captcha.OptionShowHollowLine, nil, nil, []string{}),
	}
}

// Create 生成新验证码，返回与原版一致的响应体。
func (s *Service) Create() map[string]any {
	id, q, a := s.drv.GenerateIdQuestionAnswer()
	item, err := s.drv.DrawCaptcha(q)
	if err != nil {
		// 绘制失败极少发生；退化为文本题
		return map[string]any{"sensitive": false, "key": id, "img": ""}
	}
	s.store.Set(id, a)
	_ = err
	return map[string]any{
		"sensitive": false,
		"key":       id,
		"img":       item.EncodeB64string(),
	}
}

// Verify 校验并消费验证码（check_api 语义：一次有效）。
func (s *Service) Verify(key, value string) bool {
	if key == "" || value == "" {
		return false
	}
	return s.store.Get(key, true) == value
}

// Handler GET /api/v2/captcha。
func (s *Service) Handler(w http.ResponseWriter, _ *http.Request) {
	r.Success(w, s.Create())
}

// cacheStore 适配 base64Captcha.Store 到内部 cache。
type cacheStore struct {
	cache *cache.Cache
}

const prefix = "captcha:"

func (s *cacheStore) Set(id string, value string) error {
	s.cache.Put(prefix+id, value, 60)
	return nil
}

func (s *cacheStore) Get(id string, clear bool) string {
	if clear {
		v, _ := s.cache.Pull(prefix + id)
		return v
	}
	v, _ := s.cache.Get(prefix + id)
	return v
}

func (s *cacheStore) Verify(id, answer string, clear bool) bool {
	return s.Get(id, clear) == answer
}
