// Package config 解析环境变量与 .env 文件。
// 环境变量命名与 docker-compose.yml / PHP 版 .env 保持兼容。
package config

import (
	"bufio"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	Host    string // HOST
	Port    string // PORT
	AppURL  string // APP_URL
	AppName string // APP_NAME

	DataDir    string // DATA_DIR，默认 data（SQLite 数据库、app_key、installed.lock 存放处）
	UploadsDir string // UPLOADS_DIR，本地存储根目录兜底值
	StaticDir  string // STATIC_DIR，前端主题静态资源目录

	DBDriver      string // DB_DRIVER: sqlite | mysql
	DBPath        string // DB_PATH，SQLite 文件路径，默认 {DATA_DIR}/ywty.db
	DBHost        string
	DBPort        string
	DBUser        string
	DBPassword    string
	DBName        string
	RedisAddr     string
	RedisPassword string
	RedisDB       int
	AppKey        string // APP_KEY，base64: 开头的 32 字节密钥，未提供时自动生成并持久化
	JWTSecret     string // 保留接收，暂不使用（token 采用 Sanctum 机制）
	LicenseKey    string // APP_LICENSE_KEY，docker 一键安装用
	AdminUsername string // ADMIN_USERNAME，docker 一键安装用
	AdminEmail    string // ADMIN_EMAIL
	AdminPassword string // ADMIN_PASSWORD
}

func Load() (*Config, error) {
	_ = loadDotEnv(".env")

	cfg := &Config{
		Host:          getEnv("HOST", "127.0.0.1"),
		Port:          getEnv("PORT", "3000"),
		AppURL:        strings.TrimRight(getEnv("APP_URL", "http://localhost"), "/"),
		AppName:       getEnv("APP_NAME", "ywty"),
		DataDir:       getEnv("DATA_DIR", "data"),
		UploadsDir:    getEnv("UPLOADS_DIR", "uploads"),
		StaticDir:     getEnv("STATIC_DIR", "public"),
		DBDriver:      strings.ToLower(getEnv("DB_DRIVER", "sqlite")),
		DBHost:        getEnv("DB_HOST", "127.0.0.1"),
		DBPort:        getEnv("DB_PORT", "3306"),
		DBUser:        getEnv("DB_USER", getEnv("DB_USERNAME", "root")),
		DBPassword:    getEnv("DB_PASSWORD", ""),
		DBName:        getEnv("DB_NAME", getEnv("DB_DATABASE", "ywty")),
		RedisAddr:     getEnv("REDIS_ADDR", ""),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		AppKey:        getEnv("APP_KEY", ""),
		JWTSecret:     getEnv("JWT_SECRET", getEnv("JWT_SECRET_FILE", "")),
		LicenseKey:    getEnv("APP_LICENSE_KEY", "LICENSE_KEY", ""),
		AdminUsername: getEnv("ADMIN_USERNAME", ""),
		AdminEmail:    getEnv("ADMIN_EMAIL", ""),
		AdminPassword: getEnv("ADMIN_PASSWORD", ""),
	}

	if cfg.DBDriver != "sqlite" && cfg.DBDriver != "mysql" {
		return nil, errors.New("config: DB_DRIVER 仅支持 sqlite 或 mysql")
	}
	if cfg.JWTSecret == "" {
		// 兼容 docker-compose 将 JWT_SECRET 持久化到数据卷的约定
		if b, err := os.ReadFile(filepath.Join(cfg.DataDir, ".jwt_secret")); err == nil {
			cfg.JWTSecret = strings.TrimSpace(string(b))
		}
	}
	cfg.DBPath = getEnv("DB_PATH", filepath.Join(cfg.DataDir, "ywty.db"))

	key, err := resolveAppKey(cfg.AppKey, cfg.DataDir)
	if err != nil {
		return nil, err
	}
	cfg.AppKey = key

	return cfg, nil
}

func (c *Config) Addr() string {
	return c.Host + ":" + c.Port
}

// resolveAppKey 与 Laravel APP_KEY 兼容：接受 base64:xxx 环境变量；
// 未提供时从 DATA_DIR/app_key 读取；仍无则生成新密钥并写入文件。
func resolveAppKey(envKey, dataDir string) (string, error) {
	if envKey != "" {
		if err := validateAppKey(envKey); err != nil {
			return "", err
		}
		return envKey, nil
	}

	path := filepath.Join(dataDir, "app_key")
	if b, err := os.ReadFile(path); err == nil {
		key := strings.TrimSpace(string(b))
		if key != "" {
			if err := validateAppKey(key); err != nil {
				return "", err
			}
			return key, nil
		}
	}

	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	key := "base64:" + base64.StdEncoding.EncodeToString(raw)
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return "", err
	}
	if err := os.WriteFile(path, []byte(key), 0o600); err != nil {
		return "", err
	}
	return key, nil
}

func validateAppKey(key string) error {
	v, ok := strings.CutPrefix(key, "base64:")
	if !ok {
		return errors.New("config: APP_KEY 必须为 base64: 开头")
	}
	decoded, err := base64.StdEncoding.DecodeString(v)
	if err != nil || (len(decoded) != 32 && len(decoded) != 16) {
		return errors.New("config: APP_KEY 解码后必须为 32 或 16 字节")
	}
	return nil
}

// getEnv 依次查找环境变量（可传多个备选名），最后一个参数为兜底默认值。
func getEnv(keys ...string) string {
	for _, k := range keys[:len(keys)-1] {
		if v, ok := os.LookupEnv(k); ok && strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return keys[len(keys)-1]
}

// loadDotEnv 极简 .env 加载：KEY=VALUE，支持 # 注释与成对引号，不覆盖已有环境变量。
func loadDotEnv(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err //nolint:nilerr // .env 不存在是常态
	}
	defer func() { _ = f.Close() }()

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if len(value) >= 2 {
			if (value[0] == '"' && value[len(value)-1] == '"') || (value[0] == '\'' && value[len(value)-1] == '\'') {
				value = value[1 : len(value)-1]
			}
		}
		if _, exists := os.LookupEnv(key); !exists {
			_ = os.Setenv(key, value)
		}
	}
	return nil
}
