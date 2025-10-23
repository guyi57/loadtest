package main

import (
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	_ "modernc.org/sqlite"
)

var db *gorm.DB

func initDatabase() error {
	var err error
	db, err = gorm.Open(sqlite.Dialector{
		DriverName: "sqlite",
		DSN:        "loadtest.db",
	}, &gorm.Config{})
	if err != nil {
		return err
	}

	// 自动迁移数据表
	err = db.AutoMigrate(&ScheduledTask{}, &TestLog{})
	if err != nil {
		return err
	}

	log.Println("✅ 数据库初始化成功")
	return nil
}
