package model

// 组与储存中间表（无时间戳/无自增主键）
type GroupStorage struct {
	GroupID   uint64 `gorm:"primaryKey" json:"group_id"`     // 角色组
	StorageID uint64 `gorm:"primaryKey" json:"storage_id"`   // 储存
	Sort      int    `gorm:"not null;default:0" json:"sort"` // 排序值
}

// TableName 指定表名
func (GroupStorage) TableName() string { return "group_storage" }

// GroupDriver 组与驱动中间表
type GroupDriver struct {
	Type     string `gorm:"size:32;primaryKey" json:"type"` // 驱动类型
	GroupID  uint64 `gorm:"primaryKey" json:"group_id"`     // 角色组
	DriverID uint64 `gorm:"primaryKey" json:"driver_id"`    // 驱动
	Sort     int    `gorm:"not null;default:0" json:"sort"` // 排序值
}

// TableName 指定表名
func (GroupDriver) TableName() string { return "group_driver" }

// StorageDriver 储存与驱动中间表
type StorageDriver struct {
	Type      string `gorm:"size:32;primaryKey" json:"type"` // 驱动类型
	StorageID uint64 `gorm:"primaryKey" json:"storage_id"`   // 储存
	DriverID  uint64 `gorm:"primaryKey" json:"driver_id"`    // 驱动
	Sort      int    `gorm:"not null;default:0" json:"sort"` // 排序值
}

// TableName 指定表名
func (StorageDriver) TableName() string { return "storage_driver" }
