package db

import (
	"time"

	"github.com/davidkleiven/silent-score/internal/utils"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
)

func GormConnection(name string) (*gorm.DB, error) {
	return gorm.Open(sqlite.Open(name), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
}

func AutoMigrate(con *gorm.DB) error {
	return utils.ReturnFirstError(
		func() error { return con.Exec("PRAGMA foreign_keys = ON", nil).Error },
		func() error { return con.AutoMigrate(&Project{}, &ProjectContentRecord{}) },
	)
}

type Project struct {
	Id        uint   `gorm:"primarykey,autoincrement,unique"`
	Name      string `gorm:"unique"`
	CreatedAt time.Time
	UpdatedAt time.Time
	Records   []ProjectContentRecord `gorm:"constraint:OnDelete:CASCADE"`
}

// Satisfy bubble.Item interface
func (p *Project) FilterValue() string {
	return p.Name
}

type ProjectOpts func(p *Project)

func WithName(name string) ProjectOpts {
	return func(p *Project) {
		p.Name = name
	}
}

func WithRecords(records []ProjectContentRecord) ProjectOpts {
	return func(p *Project) {
		p.Records = records
	}
}

func NewProject(opts ...ProjectOpts) *Project {
	p := Project{
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	for _, opt := range opts {
		opt(&p)
	}
	return &p
}

type ProjectContentRecord struct {
	ProjectID   uint
	Scene       uint
	SceneDesc   string `gorm:"default:''"`
	DurationSec int    `gorm:"default:0"`
	Keywords    string `gorm:"default:''"`
	Tempo       uint   `gorm:"default:0"`
	Theme       uint   `gorm:"default:0"`
}

type GormStore struct {
	Database *gorm.DB
}

func (g *GormStore) Save(p *Project) error {
	p.UpdatedAt = time.Now()
	return g.Database.Transaction(func(tx *gorm.DB) error {
		var deleteRecords []ProjectContentRecord
		if err := tx.Delete(&deleteRecords, "project_id = ?", p.Id).Error; err != nil {
			return err
		}

		return tx.Clauses(
			clause.OnConflict{
				Columns:   []clause.Column{{Name: "id"}},
				DoUpdates: clause.AssignmentColumns([]string{"name", "updated_at"}),
			},
		).Create(p).Error
	})

}

func (g *GormStore) Delete(id uint) error {
	var project Project
	tx := g.Database.Delete(&project, id)
	return tx.Error
}

func (g *GormStore) Load() ([]Project, error) {
	var projects []Project
	tx := g.Database.Model(&Project{}).Preload("Records").Find(&projects)
	return projects, tx.Error
}
