// Package seed 初始化基础数据
package seed

import (
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/ywty/server/internal/model"
)

// 默认初始数据
const (
	DefaultAdminUsername = "admin"
	DefaultAdminEmail    = "admin@ywty.local"
	DefaultAdminPassword = "admin123456" // 首次登录后请立即修改

	DefaultGroupName    = "默认用户"
	DefaultGuestName    = "游客"
	DefaultGroupDefault = true
	DefaultGroupGuest   = true
)

// Run 执行全部种子（幂等）
func Run(db *gorm.DB) error {
	if err := seedDefaultGroups(db); err != nil {
		return fmt.Errorf("seed groups: %w", err)
	}
	if err := seedDefaultStorage(db); err != nil {
		return fmt.Errorf("seed storage: %w", err)
	}
	if err := seedAdminUser(db); err != nil {
		return fmt.Errorf("seed admin: %w", err)
	}
	return nil
}

// seedDefaultGroups 创建默认 + 游客角色组
func seedDefaultGroups(db *gorm.DB) error {
	defaults := []model.Group{
		{Name: DefaultGroupName, Intro: "系统默认角色组", IsDefault: true, Options: model.JSONMap{}},
		{Name: DefaultGuestName, Intro: "游客角色组（未登录）", IsGuest: true, Options: model.JSONMap{}},
	}
	for i := range defaults {
		g := defaults[i]
		var existing model.Group
		err := db.Where("name = ?", g.Name).First(&existing).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if err := db.Create(&g).Error; err != nil {
				return err
			}
		} else if err != nil {
			return err
		}
	}
	return nil
}

// seedDefaultStorage 创建本地存储
func seedDefaultStorage(db *gorm.DB) error {
	var existing model.Storage
	err := db.Where("name = ?", "本地存储").First(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		s := model.Storage{
			Name:     "本地存储",
			Intro:    "默认本地存储",
			Prefix:   "photos",
			Provider: model.StorageProviderLocal,
			Options:  model.JSONMap{},
		}
		return db.Create(&s).Error
	}
	return err
}

// seedAdminUser 创建默认管理员（密码：admin123456）
func seedAdminUser(db *gorm.DB) error {
	var existing model.User
	err := db.Where("username = ?", DefaultAdminUsername).First(&existing).Error
	if err == nil {
		return nil // 已存在
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(DefaultAdminPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	admin := model.User{
		Avatar:   "",
		Name:     "Administrator",
		Username: DefaultAdminUsername,
		Email:    DefaultAdminEmail,
		Password: string(hashed),
		IsAdmin:  true,
		Status:   model.UserStatusNormal,
		Options:  model.JSONMap{},
	}
	if err := db.Create(&admin).Error; err != nil {
		return err
	}

	// 同时把管理员绑定到默认角色组
	var defaultGroup model.Group
	if err := db.Where("is_default = ?", true).First(&defaultGroup).Error; err != nil {
		return err
	}
	ug := model.UserGroup{
		UserID:  admin.ID,
		GroupID: defaultGroup.ID,
		From:    model.GroupFromSystem,
	}
	return db.Create(&ug).Error
}
