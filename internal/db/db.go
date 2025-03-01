package db

import (
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
)

func GormConnection(name string) (*gorm.DB, error) {
	return gorm.Open(sqlite.Open(name), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
}

func InMemoryGormConnection() (*gorm.DB, error) {
	return GormConnection(":memory:")
}

func AutoMigrate(con *gorm.DB) error {
	return con.AutoMigrate(&Project{})
}

type Project struct {
	Id        uint   `gorm:"primarykey,autoincrement,unique"`
	Name      string `gorm:"unique"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Satisfy bubble.Item interface
func (p *Project) FilterValue() string {
	return p.Name
}

func NewProject(name string) *Project {
	return &Project{
		Name:      name,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func SaveProject(db *gorm.DB, p *Project) error {
	p.UpdatedAt = time.Now()
	tx := db.Clauses(
		clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}},
			DoUpdates: clause.AssignmentColumns([]string{"name", "updated_at"}),
		},
	).Create(p)
	return tx.Error
}

func DeleteProject(db *gorm.DB, id int) error {
	var project Project
	tx := db.Delete(&project, id)
	return tx.Error
}
