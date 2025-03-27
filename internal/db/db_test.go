package db

import (
	"os"
	"slices"
	"testing"
	"time"
)

func namedGormStore(name string) *GormStore {
	db, err := GormConnection(name)
	if err != nil {
		panic(err)
	}
	if err := AutoMigrate(db); err != nil {
		panic(err)
	}
	return &GormStore{Database: db}
}

func TestCreateProject(t *testing.T) {
	tests := storeTests(t.Name())
	defer os.Remove(t.Name())

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			project1 := NewProject(WithName("project1"))
			project2 := NewProject(WithName("project2"))
			if err := test.store.Save(project1); err != nil {
				t.Error(err)
			}
			if err := test.store.Save(project2); err != nil {
				t.Error(err)
			}

			projects, err := test.store.Load()
			if err != nil {
				t.Error(err)
			}

			if len(projects) != 2 {
				t.Errorf("Wanted 2 projects got %d", len(projects))
			}

			ids := make([]int, 2)
			for i, p := range projects {
				ids[i] = int(p.Id)
			}
			slices.Sort(ids)

			expect := []int{1, 2}
			if slices.Compare(ids, expect) != 0 {
				t.Errorf("Wanted %v got %v", expect, ids)
			}
		})
	}

}

func TestUpdateExistingProject(t *testing.T) {
	store := namedGormStore(t.Name())
	defer os.Remove(t.Name())

	project := NewProject(WithName("my-project"))
	store.Save(project)

	otherProject := Project{
		Id:        1,
		Name:      "my-project2",
		CreatedAt: project.CreatedAt,
	}
	store.Save(&otherProject)

	projects, err := store.Load()
	if err != nil {
		t.Error(err)
	}

	if len(projects) != 1 {
		t.Errorf("Expected 1 project god %d", len(projects))
	}

	if projects[0].Name != "my-project2" || !projects[0].UpdatedAt.After(project.UpdatedAt) {
		t.Errorf("Expect name my-project2 got %s and updatedAt (%v) to be after %v", projects[0].Name, projects[0].UpdatedAt, project.UpdatedAt)
	}
}

func TestDeleteProject(t *testing.T) {
	tests := storeTests(t.Name())
	defer os.Remove(t.Name())
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			project := NewProject(WithName("my-project"))
			if err := test.store.Save(project); err != nil {
				t.Error(err)
			}

			projects, err := test.store.Load()
			if err != nil {
				t.Error(err)
			}
			if len(projects) != 1 {
				t.Errorf("Expected one project stored got %d", len(projects))
			}

			test.store.Delete(project.Id)
			projects, err = test.store.Load()
			if err != nil {
				t.Error(err)
			}
			if len(projects) != 0 {
				t.Errorf("Expected 0 projetts got %d", len(projects))
			}

		})
	}
}

func TestErrorOnDuplicateName(t *testing.T) {
	tests := storeTests(t.Name())
	defer os.Remove(t.Name())
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			project := NewProject(WithName("my-project"))
			if err := test.store.Save(project); err != nil {
				t.Error(err)
			}

			project2 := NewProject(WithName("my-project"))
			err := test.store.Save(project2)
			if err == nil {
				t.Errorf("Expected error because of duplicate name")
			}
		})
	}
}

func TestProjectWithRecordsRoundTrip(t *testing.T) {
	tests := storeTests(t.Name())
	defer os.Remove(t.Name())
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {

			records := []ProjectContentRecord{
				{
					Scene: 1,
					Start: time.Now(),
				},
				{
					Scene: 2,
					Start: time.Now(),
				},
			}

			project := NewProject(WithName("my-project"), WithRecords(records))
			if err := test.store.Save(project); err != nil {
				t.Error(err)
			}

			loadedProject, err := test.store.Load()
			if err != nil {
				t.Error(err)
			}

			if len(loadedProject) != 1 {
				t.Errorf("Wanted 1 project got %d", len(loadedProject))
			}

			if len(loadedProject[0].Records) != len(records) {
				t.Errorf("Wanted %d associated records got %d", len(records), len(loadedProject[0].Records))
			}
		})
	}
}

func TestUpdateRecords(t *testing.T) {
	tests := storeTests(t.Name())
	defer os.Remove(t.Name())
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			records := []ProjectContentRecord{
				{
					Scene: 1,
					Start: time.Now(),
				},
				{
					Scene: 2,
					Start: time.Now(),
				},
			}
			project := NewProject(WithName("my-project"), WithRecords(records))
			if err := test.store.Save(project); err != nil {
				t.Error(err)
			}

			loadedProject, err := test.store.Load()
			if err != nil {
				t.Error(err)
			}

			if len(loadedProject) != 1 {
				t.Errorf("Expected only one project. Got %d", len(loadedProject))
			}

			if len(loadedProject[0].Records) != len(records) {
				t.Errorf("Expected %d records got %d", len(records), len(loadedProject[0].Records))
			}

			project.Records = []ProjectContentRecord{records[1]}

			if err := test.store.Save(project); err != nil {
				t.Error(err)
			}

			loadedProject, err = test.store.Load()
			if err != nil {
				t.Error(err)
			}
			if len(loadedProject) != 1 {
				t.Errorf("Expected 1 project got %d", len(loadedProject))
			}
			if len(loadedProject[0].Records) != 1 {
				t.Errorf("Expected only one associated record got %d", len(loadedProject[0].Records))
			}

		})
	}
}

func TestFilterValue(t *testing.T) {
	p := Project{Name: "my-name"}
	if p.FilterValue() != "my-name" {
		t.Errorf("FilterValue should be equal to name")
	}
}

func TestGormCascadeDelete(t *testing.T) {
	records := []ProjectContentRecord{
		{
			Start: time.Now(),
			Scene: 1,
		},
	}
	p := NewProject(WithName("my-project"), WithRecords(records))
	store := namedGormStore(t.Name())
	defer os.Remove(t.Name())
	if err := store.Save(p); err != nil {
		t.Error(err)
	}
	if err := store.Delete(p.Id); err != nil {
		t.Error(err)
	}

	readProjects, err := store.Load()
	if err != nil {
		t.Error(err)
	}

	if len(readProjects) != 0 {
		t.Errorf("Should be no projects in the database got %d", len(readProjects))
	}

	var remainingRecords []ProjectContentRecord
	if tx := store.Database.Find(&remainingRecords); tx.Error != nil {
		t.Error(err)
	}
	if len(remainingRecords) != 0 {
		t.Errorf("There should be no remaining records left. Got %d", len(remainingRecords))
	}

}

type storeTest struct {
	store ProjectStore
	desc  string
}

func storeTests(dbName string) []storeTest {
	return []storeTest{
		{
			store: namedGormStore(dbName),
			desc:  "gorm store",
		},
		{
			store: NewInMemoryStore(),
			desc:  "in memory store",
		},
	}
}
