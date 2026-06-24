// Package migrate 注册全部 GORM 模型迁移
package migrate

import (
	"gorm.io/gorm"

	"github.com/ywty/server/internal/model"
)

// AllModels 返回所有需要迁移的模型（按依赖顺序）
func AllModels() []any {
	return []any{
		// 基础
		&model.User{},
		&model.Driver{},
		&model.Group{},
		&model.Storage{},

		// 中间表（多对多）
		&model.GroupStorage{},
		&model.GroupDriver{},
		&model.StorageDriver{},

		// 业务
		&model.Album{},
		&model.AlbumPhoto{},
		&model.Photo{},
		&model.Tag{},
		&model.Taggable{},
		&model.Share{},
		&model.Shareable{},
		&model.Like{},
		&model.Violation{},

		// 内容/运营
		&model.Notice{},
		&model.Page{},
		&model.Feedback{},

		// 工单
		&model.Ticket{},
		&model.TicketReply{},

		// 举报
		&model.Report{},

		// 套餐
		&model.Plan{},
		&model.PlanPrice{},
		&model.PlanGroup{},
		&model.PlanCapacity{},

		// 订单
		&model.Coupon{},
		&model.Order{},

		// 用户关联
		&model.UserGroup{},
		&model.UserCapacity{},

		// 三方
		&model.OAuth{},

		// P2 扩展
		&model.VerifyCode{},
		&model.PersonalAccessToken{},
		&model.AdminSession{},
	}
}

// AutoMigrate 执行全部模型自动迁移
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(AllModels()...)
}
