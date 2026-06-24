package main

import (
	"fmt"
	"os"

	"github.com/ywty/server/internal/config"
	"github.com/ywty/server/internal/database"
	"github.com/ywty/server/internal/model"
	"gorm.io/gorm"
)

func main() {
	cfgPath := "configs/config.yaml"
	if len(os.Args) > 1 {
		cfgPath = os.Args[1]
	}
	cfg, err := config.Load(cfgPath)
	if err != nil {
		fmt.Println("config load err:", err)
		return
	}
	db, err := database.New(cfg.Database)
	if err != nil {
		fmt.Println("db open err:", err)
		return
	}

	var users []model.User
	db.Find(&users)
	fmt.Println("=== users ===")
	for _, u := range users {
		fmt.Printf("  id=%d username=%s email=%s\n", u.ID, u.Username, u.Email)
	}

	var photoCount int64
	db.Model(&model.Photo{}).Count(&photoCount)
	fmt.Println("=== photos count ===", photoCount)

	var albumCount int64
	db.Model(&model.Album{}).Count(&albumCount)
	fmt.Println("=== albums count ===", albumCount)

	// 原始 SQL 看真实存储
	type rawPhoto struct {
		ID     int64
		UserID int64
		Name   string
		Size   float64
		Pub    bool
	}
	var raws []rawPhoto
	db.Raw("SELECT id, user_id, name, size, is_public FROM photos ORDER BY id DESC LIMIT 10").Scan(&raws)
	fmt.Println("=== raw photos ===")
	for _, r := range raws {
		fmt.Printf("  id=%d user_id=%d name=%s size=%v pub=%v\n", r.ID, r.UserID, r.Name, r.Size, r.Pub)
	}

	// schema（兼容 SQLite / MySQL）
	switch cfg.Database.Driver {
	case "sqlite":
		type colInfo struct {
			Cid       int
			Name      string
			Type      string
			Notnull   int
			DfltValue any
			Pk        int
		}
		var cols []colInfo
		db.Raw("PRAGMA table_info(photos)").Scan(&cols)
		fmt.Println("=== photos schema (sqlite) ===")
		for _, c := range cols {
			fmt.Printf("  %d: %s %s notnull=%d default=%v pk=%d\n", c.Cid, c.Name, c.Type, c.Notnull, c.DfltValue, c.Pk)
		}
	case "mysql":
		type colInfo struct {
			Field   string
			Type    string
			Null    string
			Key     string
			Default any
			Extra   string
		}
		var cols []colInfo
		db.Raw("SHOW COLUMNS FROM photos").Scan(&cols)
		fmt.Println("=== photos schema (mysql) ===")
		for _, c := range cols {
			fmt.Printf("  %s %s null=%s key=%s default=%v extra=%s\n", c.Field, c.Type, c.Null, c.Key, c.Default, c.Extra)
		}
	}

	var photos []model.Photo
	if photoCount > 0 {
		db.Order("id DESC").Limit(5).Find(&photos)
		for _, p := range photos {
			fmt.Printf("  photo id=%d user_id=%d name=%s size=%d public=%v\n", p.ID, p.UserID, p.Name, p.Size, p.IsPublic)
		}
	}

	// 容量计算
	type row struct{ Total int64 }
	var sizeRow row
	db.Model(&model.Photo{}).Select("COALESCE(SUM(size), 0) as total").Scan(&sizeRow)
	fmt.Println("=== total size (KB) ===", sizeRow.Total)

	// 同时输出当前 cfg 里的 DB 路径
	fmt.Println("=== db path ===", cfg.Database.Path)

	// 检查 store 中是否有问题
	var gormDB *gorm.DB = db
	_ = gormDB
}
