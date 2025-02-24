package db

import (
	"slices"
	"testing"
)

func TestCreateProject(t *testing.T) {
	db := InMemoryGormConnection()
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
	db := InMemoryGormConnection()
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
	db := InMemoryGormConnection()
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
	db := InMemoryGormConnection()
	AutoMigrate(db)
	project := NewProject("my-project")
	SaveProject(db, project)

	project2 := NewProject("my-project")
	err := SaveProject(db, project2)
	if err == nil {
		t.Errorf("Expected error because of duplicate name")
	}
}
