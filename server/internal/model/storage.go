package model

// 存储提供者
const (
	StorageProviderLocal = "local" // 本地
	StorageProviderS3    = "s3"    // AWS S3 兼容
	StorageProviderOSS   = "oss"   // 阿里云 OSS
	StorageProviderCOS   = "cos"   // 腾讯云 COS
	StorageProviderQiniu = "qiniu" // 七牛云
	StorageProviderUpyun = "upyun" // 又拍云
	StorageProviderFTP   = "ftp"   // FTP
	StorageProviderSFTP  = "sftp"  // SFTP
)

// Storage 储存表
type Storage struct {
	Base
	Name     string  `gorm:"size:255;not null" json:"name"`               // 名称
	Intro    string  `gorm:"size:2000;not null;default:''" json:"intro"`  // 描述
	Prefix   string  `gorm:"size:1000;not null;default:''" json:"prefix"` // 储存前缀
	Provider string  `gorm:"size:64;not null;index" json:"provider"`      // 储存提供者
	Options  JSONMap `gorm:"type:json" json:"options,omitempty"`          // 储存配置
}

// TableName 指定表名
func (Storage) TableName() string { return "storages" }
