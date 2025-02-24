package db

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func GormConnection(name string) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(name), &gorm.Config{})
	if err != nil {
		panic("Failed to connect to database")
	}
	return db
}

func InMemoryGormConnection() *gorm.DB {
	return GormConnection(":memory:")
}

func AutoMigrate(db *gorm.DB) {
	err := db.AutoMigrate(&Project{})
	if err != nil {
		panic("Can not auto-migrate schema because")
	}
}
