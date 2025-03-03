package db

import (
	"cmp"
	"errors"
	"slices"
	"testing"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func inMemoryDbPanicOnError() *gorm.DB {
	db, err := InMemoryGormConnection()
	if err != nil {
		panic(err)
	}
	return db
}

func TestCreateProject(t *testing.T) {
	db := inMemoryDbPanicOnError()
	AutoMigrate(db)

	project1 := NewProject("project1")
	project2 := NewProject("project2")
	SaveProject(db, project1)
	SaveProject(db, project2)

	var projects []Project
	db.Find(&projects)

	if len(projects) != 2 {
		t.Errorf("Wanted 2 projects got %d", len(projects))
	}

	ids := make([]int, 2)
	for i, p := range projects {
		ids[i] = int(p.Id)
	}

	expect := []int{1, 2}
	if slices.Compare(ids, expect) != 0 {
		t.Errorf("Wanted %v got %v", expect, ids)
	}
}

func TestUpdateExistingProject(t *testing.T) {
	db := inMemoryDbPanicOnError()
	AutoMigrate(db)
	project := NewProject("my-project")
	SaveProject(db, project)

	otherProject := Project{
		Id:        1,
		Name:      "my-project2",
		CreatedAt: project.CreatedAt,
	}
	SaveProject(db, &otherProject)

	var projects []Project
	db.Find(&projects)

	if len(projects) != 1 {
		t.Errorf("Expected 1 project god %d", len(projects))
	}

	if projects[0].Name != "my-project2" || !projects[0].UpdatedAt.After(project.UpdatedAt) {
		t.Errorf("Expect name my-project2 got %s and updatedAt (%v) to be after %v", projects[0].Name, projects[0].UpdatedAt, project.UpdatedAt)
	}
}

func TestDeleteProject(t *testing.T) {
	db := inMemoryDbPanicOnError()
	AutoMigrate(db)
	project := NewProject("my-project")
	SaveProject(db, project)

	var projects []Project
	db.Find(&projects)
	if len(projects) != 1 {
		t.Errorf("Expected one project stored got %d", len(projects))
	}
	DeleteProject(db, int(project.Id))
	db.Find(&projects)
	if len(projects) != 0 {
		t.Errorf("Expected 0 projetts got %d", len(projects))
	}
}

func TestErrorOnDuplicateName(t *testing.T) {
	db := inMemoryDbPanicOnError()
	AutoMigrate(db)
	project := NewProject("my-project")
	SaveProject(db, project)

	project2 := NewProject("my-project")
	err := SaveProject(db, project2)
	if err == nil {
		t.Errorf("Expected error because of duplicate name")
	}
}

func TestFilterValue(t *testing.T) {
	p := Project{Name: "my-name"}
	if p.FilterValue() != "my-name" {
		t.Errorf("FilterValue should be equal to name")
	}
}

func TestInsertProjectRecords(t *testing.T) {
	database := inMemoryDbPanicOnError()
	AutoMigrate(database)
	project := NewProject("project1")

	project2 := NewProject("project2")

	record1 := ProjectContentRecord{
		ProjectID: project.Id,
		Project:   project,
		Scene:     0,
		Start:     time.Now(),
	}
	record2 := ProjectContentRecord{
		Project: project2,
		Scene:   1,
		Start:   time.Now(),
	}

	if err := SaveProjectRecords(database, []ProjectContentRecord{record1}); err != nil {
		t.Error(err)
		return
	}
	var projects []Project
	database.Find(&projects)
	if len(projects) != 1 {
		t.Errorf("Expected no new project to be created. Have %d", len(projects))
		return
	}

	if err := SaveProjectRecords(database, []ProjectContentRecord{record2}); err != nil {
		t.Error(err)
		return
	}
	var records []ProjectContentRecord
	database.Find(&records)
	database.Find(&projects)
	if len(projects) != 2 {
		t.Errorf("Expected 2 projects got %d", len(projects))
	}
}

func TestProjectRecordRoundtrip(t *testing.T) {
	database := inMemoryDbPanicOnError()
	AutoMigrate(database)
	project := NewProject("project1")
	records := []ProjectContentRecord{
		{
			Project:  project,
			Keywords: "my title",
			Tempo:    82,
			Theme:    1,
			Scene:    1,
		},
		{
			Project:  project,
			Keywords: "agitato",
			Tempo:    106,
			Theme:    2,
			Scene:    2,
		},
	}

	SaveProjectRecords(database, records)

	var readRecords []ProjectContentRecord
	database.Find(&readRecords)

	if len(records) != len(readRecords) {
		t.Errorf("Expected 2 records, got %d", len(readRecords))
		return
	}

	for i := range records {
		r1, r2 := records[i], readRecords[i]
		if (r1.Keywords != r2.Keywords) || (r1.Tempo != r2.Tempo) || (r1.Theme != r2.Theme) || (r1.Scene != r2.Scene) ||
			(r1.ProjectID != r2.ProjectID) {
			t.Errorf("Wanted %v got %v\n", r1, r2)
		}
	}
}

func TestUpdateOnRecordConflict(t *testing.T) {
	database := inMemoryDbPanicOnError()
	AutoMigrate(database)
	project := NewProject("project1")
	records := []ProjectContentRecord{
		{
			Project:  project,
			Keywords: "my title",
			Tempo:    82,
			Theme:    1,
			Scene:    1,
		},
	}
	SaveProjectRecords(database, records)
	records[0].Tempo = 87
	SaveProjectRecords(database, records)

	var readRecords []ProjectContentRecord
	database.Find(&readRecords)
	if len(readRecords) != 1 {
		t.Errorf("Extra records was added")
	}

	if readRecords[0].Tempo != 87 {
		t.Errorf("Record was not updated on conflict")
	}
}

func TestFailureOnMissingProject(t *testing.T) {
	database := inMemoryDbPanicOnError()
	AutoMigrate(database)
	records := []ProjectContentRecord{
		{
			Keywords: "my title",
			Tempo:    82,
			Theme:    1,
			Scene:    1,
		},
	}
	if err := SaveProjectRecords(database, records); err == nil {
		t.Errorf("Should fail because project is missing")
	}
}

func TestUniqueProjectIds(t *testing.T) {
	records := []ProjectContentRecord{{ProjectID: 1}, {ProjectID: 1}}

	id, err := uniqueProjectId(records)
	if id != 1 || err != nil {
		t.Errorf("Wanted (0, nil) got (%d, %v)", id, err)
	}
}

func TestUniqueProjectDifferentIds(t *testing.T) {
	records := []ProjectContentRecord{{ProjectID: 1}, {ProjectID: 2}}

	_, err := uniqueProjectId(records)
	if err == nil {
		t.Errorf("Should fail because ids are duplicated")
	}
}

func TestUpdateRecords(t *testing.T) {
	oldRecords := []ProjectContentRecord{
		{
			Scene:     0,
			SceneDesc: "Scene A",
		},
		{
			Scene:     2,
			SceneDesc: "Scene B",
		},
		{
			Scene:     1,
			SceneDesc: "Middle scene",
		},
	}

	newRecords := []ProjectContentRecord{
		{
			Scene:     0,
			SceneDesc: "Pre scene A",
		},
		{
			Scene:     2,
			SceneDesc: "Pre ending",
		},
	}

	got := updateRecords(newRecords, oldRecords)

	want := []ProjectContentRecord{
		{
			Scene:     0,
			SceneDesc: "Pre scene A",
		},
		{
			Scene:     1,
			SceneDesc: "Scene A",
		},
		{
			Scene:     2,
			SceneDesc: "Middle scene",
		},
		{
			Scene:     3,
			SceneDesc: "Pre ending",
		},
		{
			Scene:     4,
			SceneDesc: "Scene B",
		},
	}

	for i, p := range want {
		r := got[i]

		if p.Scene != r.Scene || p.SceneDesc != r.SceneDesc {
			t.Errorf("Wanted %d: %v got %v", i, p, r)
		}
	}
}

func TestInsertNewRecords(t *testing.T) {
	database := inMemoryDbPanicOnError()
	AutoMigrate(database)

	project := NewProject("project1")
	project2 := NewProject("project2")

	records := []ProjectContentRecord{
		{
			Project:   project,
			Scene:     1,
			SceneDesc: "Scene 1",
		},
		{
			Project:   project,
			Scene:     2,
			SceneDesc: "Scene 2",
		},
		{
			Project:   project2,
			Scene:     2,
			SceneDesc: "Scene 2",
		},
	}

	if err := SaveProjectRecords(database, records); err != nil {
		t.Error(err)
		return
	}

	newRecord := []ProjectContentRecord{
		{
			ProjectID: project.Id,
			Project:   project,
			Scene:     2,
			SceneDesc: "Alternative scene 2",
		},
	}
	if err := InsertRecords(database, newRecord); err != nil {
		t.Error(err)
		return
	}

	var allRecords []ProjectContentRecord
	if tx := database.Find(&allRecords); tx.Error != nil {
		t.Error(tx.Error)
		return
	}

	want := []ProjectContentRecord{
		{
			ProjectID: project.Id,
			Scene:     1,
			SceneDesc: "Scene 1",
		},
		{
			ProjectID: project.Id,
			Scene:     2,
			SceneDesc: "Alternative scene 2",
		},
		{
			ProjectID: project.Id,
			Scene:     3,
			SceneDesc: "Scene 2",
		},
		{
			ProjectID: project2.Id,
			Scene:     2,
			SceneDesc: "Scene 2",
		},
	}

	if len(allRecords) != len(want) {
		t.Errorf("Expected %d records got %d", len(want), len(allRecords))
		return
	}

	fn := func(a, b ProjectContentRecord) int {
		projectsEqual := cmp.Compare(a.ProjectID, b.ProjectID)
		if projectsEqual == 0 {
			return cmp.Compare(a.Scene, b.Scene)
		}
		return projectsEqual
	}

	slices.SortFunc(want, fn)
	slices.SortFunc(allRecords, fn)

	for i, record := range want {
		if record.Scene != allRecords[i].Scene || record.SceneDesc != allRecords[i].SceneDesc {
			t.Errorf("Pos %d: wanted %v got%v\n", i, record, allRecords[i])
		}
	}
}

func TestInsertErrorOnNonUniqueProjectIDs(t *testing.T) {
	database := inMemoryDbPanicOnError()
	records := []ProjectContentRecord{
		{
			ProjectID: 1,
		},
		{
			ProjectID: 2,
		},
	}
	if err := InsertRecords(database, records); !errors.Is(err, ErrProjectIdsNotUnique) {
		t.Errorf("Got %v wanted %v", err, ErrProjectIdsNotUnique)
	}
}

var errFind = errors.New("find error")

type failingDatastore struct{}

func (f failingDatastore) Clauses(clauses ...clause.Expression) *gorm.DB { return &gorm.DB{} }
func (f failingDatastore) Find(dest interface{}, conds ...interface{}) *gorm.DB {
	return &gorm.DB{Error: errFind}
}

func TestTransactionError(t *testing.T) {
	store := failingDatastore{}
	records := []ProjectContentRecord{}
	err := InsertRecords(&store, records)

	if !errors.Is(err, errFind) {
		t.Errorf("Wanted %v got %v", err, errFind)
	}
}
