package db

import "time"

type Project struct {
	Id        uint `gorm:"primarykey,autoincrement,unique"`
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}
