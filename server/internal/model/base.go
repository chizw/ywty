// Package model 提供全部业务实体的 GORM 模型定义
// 设计原则：与原 Laravel 数据库结构保持一致（便于平滑迁移）
package model

import (
	"time"

	"gorm.io/gorm"
)

// Base 通用字段：ID + 时间戳 + 软删除
// 所有业务表均嵌入此结构
type Base struct {
	ID        uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	CreatedAt time.Time      `gorm:"index" json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// SoftDelete 软删除（部分表无时间戳但有软删）
type SoftDelete struct {
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}
