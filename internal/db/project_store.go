package db

import "fmt"

type ProjectStore interface {
	Save(project *Project) error
	Delete(id uint) error
	Load() ([]Project, error)
}

type InMemoryProjectStore struct {
	projects map[uint]Project
}

func NewInMemoryStore() *InMemoryProjectStore {
	return &InMemoryProjectStore{
		projects: make(map[uint]Project),
	}
}

func (im *InMemoryProjectStore) autoIncrementProjectId(project *Project) {
	if project.Id == 0 {
		project.Id = uint(len(im.projects) + 1)
	}
}

func (im *InMemoryProjectStore) Save(project *Project) error {
	im.autoIncrementProjectId(project)

	for id, p := range im.projects {
		if id != project.Id && p.Name == project.Name {
			return fmt.Errorf("name already exists")
		}
	}

	im.projects[project.Id] = *project
	return nil
}

func (im *InMemoryProjectStore) Delete(id uint) error {
	delete(im.projects, id)
	return nil
}

func (im *InMemoryProjectStore) Load() ([]Project, error) {
	projList := make([]Project, 0, len(im.projects))
	for _, project := range im.projects {
		projList = append(projList, project)
	}
	return projList, nil
}
