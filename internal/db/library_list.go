package db

import "time"

type LibraryList interface {
	AddLibrary(name string) error
	RemoveLibrary(id uint) error
	ListLibraries() ([]ConfiguredLibraries, error)
}

type InMemoryLibraryList struct {
	libraries map[uint]ConfiguredLibraries
}

func NewInMemoryLibraryList() *InMemoryLibraryList {
	return &InMemoryLibraryList{
		libraries: make(map[uint]ConfiguredLibraries),
	}
}

func (im *InMemoryLibraryList) AddLibrary(name string) error {
	for _, lib := range im.libraries {
		if lib.Path == name {
			return nil
		}
	}
	id := uint(len(im.libraries) + 1)
	im.libraries[id] = ConfiguredLibraries{
		ID:        id,
		CreatedAt: time.Now(),
		Path:      name,
	}
	return nil
}

func (im *InMemoryLibraryList) RemoveLibrary(id uint) error {
	delete(im.libraries, id)
	return nil
}

func (im *InMemoryLibraryList) ListLibraries() ([]ConfiguredLibraries, error) {
	libraries := make([]ConfiguredLibraries, 0, len(im.libraries))
	for _, lib := range im.libraries {
		libraries = append(libraries, lib)
	}
	return libraries, nil
}
