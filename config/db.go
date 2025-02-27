package config

import (
	"MyBloge/global"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
	"time"
)

func InitDB() {
	dsn := AppConfig.Database.Dsn
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("database connection error: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("database error: %v", err)
	}
	sqlDB.SetMaxIdleConns(AppConfig.Database.MaxIdleConnes)
	sqlDB.SetMaxOpenConns(AppConfig.Database.MaxOpenConnes)
	sqlDB.SetConnMaxLifetime(time.Hour)
	global.SqlDb = db
}
