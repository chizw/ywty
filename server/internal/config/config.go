// Package config 提供基于 Viper 的配置加载能力
// 支持 YAML 配置文件 + 环境变量覆盖（YWTY_* 前缀）
package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type App struct {
	Name     string `mapstructure:"name"`
	Env      string `mapstructure:"env"`
	Version  string `mapstructure:"version"`
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Timezone string `mapstructure:"timezone"`
	BaseURL  string `mapstructure:"base_url"`
}

type Database struct {
	Driver          string `mapstructure:"driver"`
	Host            string `mapstructure:"host"`
	Port            int    `mapstructure:"port"`
	Username        string `mapstructure:"username"`
	Password        string `mapstructure:"password"`
	DBName          string `mapstructure:"dbname"`
	Charset         string `mapstructure:"charset"`
	ParseTime       bool   `mapstructure:"parse_time"`
	Loc             string `mapstructure:"loc"`
	Path            string `mapstructure:"path"`
	MaxOpenConns    int    `mapstructure:"max_open_conns"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`
	ConnMaxLifetime int    `mapstructure:"conn_max_lifetime"`
	ConnMaxIdleTime int    `mapstructure:"conn_max_idle_time"`
}

type Redis struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type Queue struct {
	Concurrency int      `mapstructure:"concurrency"`
	Queues      []string `mapstructure:"queues"`
}

type Log struct {
	Level      string `mapstructure:"level"`
	Format     string `mapstructure:"format"`
	Output     string `mapstructure:"output"`
	FilePath   string `mapstructure:"file_path"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxAge     int    `mapstructure:"max_age"`
	MaxBackups int    `mapstructure:"max_backups"`
	Compress   bool   `mapstructure:"compress"`
}

type CORS struct {
	AllowOrigins     []string `mapstructure:"allow_origins"`
	AllowMethods     []string `mapstructure:"allow_methods"`
	AllowHeaders     []string `mapstructure:"allow_headers"`
	ExposeHeaders    []string `mapstructure:"expose_headers"`
	AllowCredentials bool     `mapstructure:"allow_credentials"`
	MaxAge           int      `mapstructure:"max_age"`
}

type AuthJWT struct {
	Secret        string `mapstructure:"secret"`
	Issuer        string `mapstructure:"issuer"`
	AccessExpire  int    `mapstructure:"access_expire"`
	RefreshExpire int    `mapstructure:"refresh_expire"`
}

type AuthAdminSession struct {
	Name   string `mapstructure:"name"`
	Secret string `mapstructure:"secret"`
	Expire int    `mapstructure:"expire"`
}

type AuthCaptcha struct {
	Enable bool `mapstructure:"enable"`
	Expire int  `mapstructure:"expire"`
}

type Auth struct {
	JWT          AuthJWT          `mapstructure:"jwt"`
	AdminSession AuthAdminSession `mapstructure:"admin_session"`
	Captcha      AuthCaptcha      `mapstructure:"captcha"`
}

type RateLimit struct {
	Enable          bool `mapstructure:"enable"`
	UploadPerMinute int  `mapstructure:"upload_per_minute"`
	APIPerMinute    int  `mapstructure:"api_per_minute"`
}

// StorageDriver 存储驱动配置（与 model.Storage 字段一致）
type StorageDriver struct {
	Provider   string                 `mapstructure:"provider"`   // local / s3 / oss / ...
	Root       string                 `mapstructure:"root"`       // 本地根目录
	PublicURL  string                 `mapstructure:"public_url"` // 对外前缀
	Visibility string                 `mapstructure:"visibility"` // public | private
	Extra      map[string]interface{} `mapstructure:"extra"`      // 其他驱动自定义参数
}

// Storage 默认存储配置（其余储存策略从 DB 加载）
type Storage struct {
	Driver StorageDriver `mapstructure:"driver"`
}

// Mail 邮件配置
type Mail struct {
	Provider string `mapstructure:"provider"` // smtp | log
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	From     string `mapstructure:"from"`
	FromName string `mapstructure:"from_name"`
}

// SMS 短信配置
type SMS struct {
	Provider  string `mapstructure:"provider"` // aliyun / tencent / log ...
	AccessKey string `mapstructure:"access_key"`
	SecretKey string `mapstructure:"secret_key"`
	Sign      string `mapstructure:"sign"`
	Region    string `mapstructure:"region"`
}

// PaymentDriver 单个支付驱动配置
type PaymentDriver struct {
	Enabled bool                   `mapstructure:"enabled"`
	Options map[string]interface{} `mapstructure:"options"`
}

// Payment 支付配置
type Payment struct {
	Default string                   `mapstructure:"default"`
	Drivers map[string]PaymentDriver `mapstructure:"drivers"`
}

type Config struct {
	App       App       `mapstructure:"app"`
	Database  Database  `mapstructure:"database"`
	Redis     Redis     `mapstructure:"redis"`
	Queue     Queue     `mapstructure:"queue"`
	Log       Log       `mapstructure:"log"`
	CORS      CORS      `mapstructure:"cors"`
	Auth      Auth      `mapstructure:"auth"`
	RateLimit RateLimit `mapstructure:"ratelimit"`
	Storage   Storage   `mapstructure:"storage"`
	Mail      Mail      `mapstructure:"mail"`
	SMS       SMS       `mapstructure:"sms"`
	Payment   Payment   `mapstructure:"payment"`
}

// DSN 构造 MySQL DSN（兼容 MariaDB 10.6.20+）
func (d Database) DSN() string {
	switch d.Driver {
	case "mysql":
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=%t&loc=%s&timeout=10s&readTimeout=30s&writeTimeout=30s&interpolateParams=true",
			d.Username, d.Password, d.Host, d.Port, d.DBName, d.Charset, d.ParseTime, d.Loc)
	default:
		return ""
	}
}

func (d Database) ConnMaxLifetimeDuration() time.Duration {
	return time.Duration(d.ConnMaxLifetime) * time.Second
}

func (d Database) ConnMaxIdleTimeDuration() time.Duration {
	return time.Duration(d.ConnMaxIdleTime) * time.Second
}

// Load 加载配置（YAML + 环境变量）
func Load(path ...string) (*Config, error) {
	v := viper.New()
	v.SetConfigType("yaml")
	v.SetEnvPrefix("YWTY")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if len(path) > 0 && path[0] != "" {
		v.SetConfigFile(path[0])
	} else {
		v.SetConfigName("config")
		v.AddConfigPath("./configs")
		v.AddConfigPath(".")
	}

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	return &cfg, nil
}
