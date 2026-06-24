package model

// 图片状态
const (
	PhotoStatusNormal   = "normal"   // 正常
	PhotoStatusAuditing = "auditing" // 审核中
	PhotoStatusBlocked  = "blocked"  // 已屏蔽
	PhotoStatusDeleted  = "deleted"  // 已删除（软删）
)

// Photo 图片表
type Photo struct {
	Base
	UserID    uint64 `gorm:"index" json:"user_id"`                                       // 用户
	GroupID   uint64 `gorm:"index;default:0" json:"group_id"`                            // 角色组
	StorageID uint64 `gorm:"index;default:0" json:"storage_id"`                          // 储存
	Name      string `gorm:"size:255;not null" json:"name"`                              // 文件别名
	Intro     string `gorm:"size:2000;not null;default:''" json:"intro"`                 // 介绍
	Filename  string `gorm:"size:255;not null" json:"filename"`                          // 文件原始名称
	Pathname  string `gorm:"size:1024;not null" json:"pathname"`                         // 文件路径名称
	Mimetype  string `gorm:"size:64;not null;default:''" json:"mimetype"`                // 媒体类型
	Extension string `gorm:"size:32;not null;default:''" json:"extension"`               // 文件后缀
	MD5       string `gorm:"size:32;not null;default:'';index" json:"md5"`               // 文件 MD5
	SHA1      string `gorm:"size:64;not null;default:''" json:"sha1"`                    // 文件 SHA1
	Exif      JSONMap `gorm:"type:json" json:"exif,omitempty"`                           // EXIF 信息
	Size      int64   `gorm:"type:decimal(20,0);not null;default:0" json:"size"`         // 大小(字节)
	Width     uint    `gorm:"not null;default:0" json:"width"`                           // 宽度
	Height    uint    `gorm:"not null;default:0" json:"height"`                          // 高度
	ThumbPath string  `gorm:"size:1024;not null;default:''" json:"thumb_path,omitempty"` // 缩略图路径
	IsPublic  bool    `gorm:"not null;default:false;index" json:"is_public"`             // 是否公开
	Status    string  `gorm:"size:64;not null;default:'normal';index" json:"status"`     // 状态
	IPAddress string  `gorm:"size:45" json:"ip_address,omitempty"`                       // 上传 IP
	ExpiredAt int64   `gorm:"index;default:0" json:"expired_at"`                         // 到期时间（unix 时间戳）
}

// TableName 指定表名
func (Photo) TableName() string { return "photos" }
