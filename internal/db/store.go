package db

type Store interface {
	ProjectStore
	LibraryList
}

type InMemoryStore struct {
	InMemoryProjectStore
	InMemoryLibraryList
}

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		InMemoryProjectStore: *NewInMemoryProjectStore(),
		InMemoryLibraryList:  *NewInMemoryLibraryList(),
	}
}
