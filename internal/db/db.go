package db

import (
	"cmp"
	"slices"
	"time"

	"github.com/davidkleiven/silent-score/internal/utils"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
)

type Finder interface {
	Find(dest interface{}, conds ...interface{}) *gorm.DB
}

type Storer interface {
	Clauses(conds ...clause.Expression) *gorm.DB
}

type Deleter interface {
	Delete(dest interface{}, conds ...interface{}) *gorm.DB
}

type CreateReadUpdater interface {
	Finder
	Storer
}

type CreateReadUpdateDeleter interface {
	CreateReadUpdater
	Deleter
}

func GormConnection(name string) (*gorm.DB, error) {
	return gorm.Open(sqlite.Open(name), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
}

func InMemoryGormConnection() (*gorm.DB, error) {
	return GormConnection(":memory:")
}

func AutoMigrate(con *gorm.DB) error {
	return utils.ReturnFirstError(
		func() error { return con.Exec("PRAGMA foreign_keys = ON", nil).Error },
		func() error { return con.AutoMigrate(&Project{}) },
		func() error { return con.AutoMigrate(&ProjectContentRecord{}) },
	)
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

func SaveProject(db Storer, p *Project) error {
	p.UpdatedAt = time.Now()
	tx := db.Clauses(
		clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}},
			DoUpdates: clause.AssignmentColumns([]string{"name", "updated_at"}),
		},
	).Create(p)
	return tx.Error
}

func DeleteProject(db Deleter, id int) error {
	var project Project
	tx := db.Delete(&project, id)
	return tx.Error
}

func SaveProjectRecords(db Storer, records []ProjectContentRecord) error {
	if len(records) == 0 {
		return nil
	}
	tx := db.Clauses(
		clause.OnConflict{
			Columns:   []clause.Column{{Name: "project_id"}, {Name: "scene"}},
			UpdateAll: true,
		},
	).Create(records)
	return tx.Error
}

func uniqueProjectId(records []ProjectContentRecord) (uint, error) {
	projId := uint(0)
	isFirst := true
	for _, record := range records {
		if isFirst {
			projId = record.ProjectID
			isFirst = false
		} else if record.ProjectID != projId {
			return projId, ErrProjectIdsNotUnique
		}
	}
	return projId, nil
}

func updateRecords(newRecords []ProjectContentRecord, oldRecords []ProjectContentRecord) []ProjectContentRecord {
	var records []ProjectContentRecord
	records = append(records, newRecords...)
	records = append(records, oldRecords...)

	// Sort stable such that equal elements are preserved
	slices.SortStableFunc(records, func(a, b ProjectContentRecord) int {
		return cmp.Compare(a.Scene, b.Scene)
	})

	// Update scene number
	diff := uint(0)
	for i := 1; i < len(records); i++ {
		records[i].Scene += diff
		if records[i].Scene == records[i-1].Scene {
			diff += 1
			records[i].Scene += 1
		}
	}
	return records
}

func InsertRecords(db CreateReadUpdater, newRecords []ProjectContentRecord) error {
	projId, err := uniqueProjectId(newRecords)
	if err != nil {
		return err
	}

	var existingRecords []ProjectContentRecord
	tx := db.Find(&existingRecords, "project_id = ?", projId)
	if tx.Error != nil {
		return tx.Error
	}
	toUpdate := updateRecords(newRecords, existingRecords)
	return SaveProjectRecords(db, toUpdate)
}

func DeleteRecords(db Deleter, projectId uint, scene uint) error {
	var record ProjectContentRecord
	tx := db.Delete(&record, "project_id = ? AND scene = ?", projectId, scene)
	return tx.Error
}

type ProjectContentRecord struct {
	Project   *Project `gorm:"not null;default:null"`
	ProjectID uint     `gorm:"primaryKey;autoIncrement:false"`
	Scene     uint     `gorm:"primaryKey;autoIncrement:false"`
	SceneDesc string   `gorm:"default:''"`
	Start     time.Time
	Keywords  string `gorm:"default:''"`
	Tempo     uint   `gorm:"default:0"`
	Theme     uint   `gorm:"default:0"`
}
