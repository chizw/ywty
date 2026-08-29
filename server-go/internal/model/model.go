// Package model 定义 GORM 模型。列名与 PHP 版数据库完全一致（"完全兼容"约束）。
package model

import (
	"time"

	"github.com/chizw/ywty/server-go/internal/db/types"
	"gorm.io/gorm"
)

// ---------- users ----------

type User struct {
	ID              int64          `gorm:"primaryKey;column:id"`
	Avatar          string         `gorm:"column:avatar"`
	Name            string         `gorm:"column:name"`
	Username        string         `gorm:"column:username"`
	Phone           *string        `gorm:"column:phone"`
	Email           *string        `gorm:"column:email"`
	Password        string         `gorm:"column:password"`
	Location        string         `gorm:"column:location"`
	URL             string         `gorm:"column:url"`
	Company         string         `gorm:"column:company"`
	CompanyTitle    string         `gorm:"column:company_title"`
	Tagline         string         `gorm:"column:tagline"`
	Bio             string         `gorm:"column:bio"`
	Interests       types.JSON     `gorm:"column:interests"`
	Socials         types.JSON     `gorm:"column:socials"`
	PhoneVerifiedAt *time.Time     `gorm:"column:phone_verified_at"`
	EmailVerifiedAt *time.Time     `gorm:"column:email_verified_at"`
	RememberToken   *string        `gorm:"column:remember_token"`
	IsAdmin         bool           `gorm:"column:is_admin"`
	Options         types.JSON     `gorm:"column:options"`
	LoginIP         *string        `gorm:"column:login_ip"`
	RegisterIP      *string        `gorm:"column:register_ip"`
	CountryCode     *string        `gorm:"column:country_code"`
	Status          string         `gorm:"column:status"`
	CreatedAt       *time.Time     `gorm:"column:created_at"`
	UpdatedAt       *time.Time     `gorm:"column:updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"column:deleted_at"`
}

func (User) TableName() string { return "users" }

// ---------- groups ----------

type Group struct {
	ID        int64          `gorm:"primaryKey;column:id"`
	Name      string         `gorm:"column:name"`
	Intro     string         `gorm:"column:intro"`
	Options   types.JSON     `gorm:"column:options"`
	IsDefault bool           `gorm:"column:is_default"`
	IsGuest   bool           `gorm:"column:is_guest"`
	CreatedAt *time.Time     `gorm:"column:created_at"`
	UpdatedAt *time.Time     `gorm:"column:updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at"`
}

func (Group) TableName() string { return "groups" }

// ---------- storages ----------

type Storage struct {
	ID        int64          `gorm:"primaryKey;column:id"`
	Name      string         `gorm:"column:name"`
	Intro     string         `gorm:"column:intro"`
	Prefix    string         `gorm:"column:prefix"`
	Provider  string         `gorm:"column:provider"`
	Options   types.JSON     `gorm:"column:options"`
	CreatedAt *time.Time     `gorm:"column:created_at"`
	UpdatedAt *time.Time     `gorm:"column:updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at"`
}

func (Storage) TableName() string { return "storages" }

// ---------- group_storage ----------

type GroupStorage struct {
	GroupID   int64 `gorm:"primaryKey;column:group_id"`
	StorageID int64 `gorm:"primaryKey;column:storage_id"`
	Sort      int32 `gorm:"column:sort"`
}

func (GroupStorage) TableName() string { return "group_storage" }

// ---------- pages ----------

type Page struct {
	ID          int64      `gorm:"primaryKey;column:id"`
	Type        string     `gorm:"column:type"`
	Name        string     `gorm:"column:name"`
	Icon        string     `gorm:"column:icon"`
	Title       string     `gorm:"column:title"`
	Content     *string    `gorm:"column:content"`
	Keywords    *string    `gorm:"column:keywords"`
	Description *string    `gorm:"column:description"`
	Slug        string     `gorm:"column:slug"`
	URL         string     `gorm:"column:url"`
	ViewCount   int64      `gorm:"column:view_count"`
	Sort        int32      `gorm:"column:sort"`
	IsShow      bool       `gorm:"column:is_show"`
	CreatedAt   *time.Time `gorm:"column:created_at"`
	UpdatedAt   *time.Time `gorm:"column:updated_at"`
}

func (Page) TableName() string { return "pages" }

// ---------- user_groups ----------

type UserGroup struct {
	ID        int64          `gorm:"primaryKey;column:id"`
	UserID    int64          `gorm:"column:user_id"`
	GroupID   int64          `gorm:"column:group_id"`
	OrderID   *int64         `gorm:"column:order_id"`
	From      string         `gorm:"column:from"`
	ExpiredAt *time.Time     `gorm:"column:expired_at"`
	CreatedAt *time.Time     `gorm:"column:created_at"`
	UpdatedAt *time.Time     `gorm:"column:updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at"`
}

func (UserGroup) TableName() string { return "user_groups" }

// ---------- user_capacities ----------

type UserCapacity struct {
	ID        int64          `gorm:"primaryKey;column:id"`
	UserID    int64          `gorm:"column:user_id"`
	OrderID   *int64         `gorm:"column:order_id"`
	Capacity  float64        `gorm:"column:capacity"`
	From      string         `gorm:"column:from"`
	ExpiredAt *time.Time     `gorm:"column:expired_at"`
	CreatedAt *time.Time     `gorm:"column:created_at"`
	UpdatedAt *time.Time     `gorm:"column:updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at"`
}

func (UserCapacity) TableName() string { return "user_capacities" }
