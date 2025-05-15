package ui

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/davidkleiven/silent-score/internal/db"
	"golang.org/x/exp/slices"
)

func TestTextInput(t *testing.T) {
	model := LibraryModel{
		store: db.NewInMemoryLibraryList(),
	}

	model.Init()

	message := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/et")}
	model.Update(message)

	if model.inputField.Value() != "/et" {
		t.Errorf("Wanted /et got %s", model.inputField.Value())
	}
	if model.currentBestGuess != "/etc" {
		t.Errorf("Wanted /etc got %s", model.currentBestGuess)
	}

	message = tea.KeyMsg{Type: tea.KeyTab}
	_, cmd := model.Update(message)
	if model.inputField.Value() != "/etc" {
		t.Errorf("Wanted /etc got %s", model.inputField.Value())
	}

	msg := cmd().(tea.KeyMsg)
	if msg.Type != tea.KeyRunes {
		t.Errorf("Wanted to be a key runes message got %v", msg)
	}

	if msg.Runes[0] != '/' {
		t.Errorf("Wanted to be a key runes message got %v", msg)
	}
}

func TestListIsPopulatedOnInit(t *testing.T) {
	model := LibraryModel{
		store: db.NewInMemoryLibraryList(),
	}
	if err := model.store.AddLibrary("/MyLibrary"); err != nil {
		t.Errorf("Failed to add library: %v", err)
	}

	model.Init()

	if len(model.currentLibraries.Items()) != 1 {
		t.Error("Wanted libraries to be populated")
	}
}

func TestSavedOnEnter(t *testing.T) {
	model := LibraryModel{
		store: db.NewInMemoryLibraryList(),
	}
	model.Init()
	model.inputField.SetValue("/MyLibrary")
	model.Update(tea.KeyMsg{Type: tea.KeyEnter})

	storedLibs, err := model.store.ListLibraries()
	if err != nil {
		t.Error(err)
	}
	if len(storedLibs) != 1 {
		t.Error("Wanted libraries to be populated")
	}
}

func TestLibraryInView(t *testing.T) {
	model := LibraryModel{
		store: db.NewInMemoryLibraryList(),
	}
	if err := model.store.AddLibrary("/MyLibrary"); err != nil {
		t.Errorf("Failed to add library: %v", err)
	}
	model.Init()
	if !strings.Contains(model.View(), "/MyLibrary") {
		t.Error("Wanted library to be in view")
	}
}

func TestToProjectOverviewOnEsc(t *testing.T) {
	model := LibraryModel{
		store: db.NewInMemoryLibraryList(),
	}
	model.Init()
	_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	result := cmd()

	switch result.(type) {
	case toProjectOverview:
	default:
		t.Error("Wanted project overview")
	}
}

func TestTableTraversal(t *testing.T) {
	model := LibraryModel{
		store: db.NewInMemoryLibraryList(),
	}
	model.store.AddLibrary("/MyLibrary")
	model.store.AddLibrary("/MyLibrary2")

	model.Init()

	if model.currentLibraries.Index() != 0 {
		t.Errorf("Wanted index to be 0 got %d", model.currentLibraries.Index())
	}
	model.Update(tea.KeyMsg{Type: tea.KeyDown})
	if model.currentLibraries.Index() != 1 {
		t.Errorf("Wanted index to be 1 got %d", model.currentLibraries.Index())
	}
	model.Update(tea.KeyMsg{Type: tea.KeyUp})
	if model.currentLibraries.Index() != 0 {
		t.Errorf("Wanted index to be 0 got %d", model.currentLibraries.Index())
	}
}

func TestDelete(t *testing.T) {
	model := LibraryModel{
		store: db.NewInMemoryLibraryList(),
	}
	model.store.AddLibrary("/MyLibrary")
	model.store.AddLibrary("/MyLibrary2")

	model.Init()

	model.Update(tea.KeyMsg{Type: tea.KeyDelete})
	if len(model.currentLibraries.Items()) != 1 {
		t.Errorf("Wanted one item to be deleted")
	}

	paths, err := model.store.ListLibraries()
	if err != nil {
		t.Errorf("Failed to list libraries: %v", err)
	}

	extractedPaths := make([]string, len(paths))
	for i, path := range paths {
		extractedPaths[i] = path.Path
	}

	if slices.Compare(extractedPaths, []string{"/MyLibrary2"}) != 0 {
		t.Errorf("Wanted to delete /MyLibrary got %s", extractedPaths)
	}

}

func TestFilterValue(t *testing.T) {
	item := ListEntry{
		Path: "/MyLibrary",
	}

	if item.FilterValue() != "/MyLibrary" {
		t.Errorf("Wanted /MyLibrary got %s", item.FilterValue())
	}
}

type failingLibraryStore struct{}

func (f *failingLibraryStore) AddLibrary(name string) error {
	return errors.New("failed to add library")
}

func (f *failingLibraryStore) RemoveLibrary(id uint) error {
	return errors.New("failed to remove library")
}
func (f *failingLibraryStore) ListLibraries() ([]db.ConfiguredLibraries, error) {
	return nil, errors.New("failed to list libraries")
}

func TestEmptyListOnListLibraryError(t *testing.T) {
	model := LibraryModel{
		store: &failingLibraryStore{},
	}
	model.Init()

	if len(model.currentLibraries.Items()) != 0 {
		t.Errorf("Wanted no items in list got %d", len(model.currentLibraries.Items()))
	}
}

func TestAddLibraryFailure(t *testing.T) {
	model := LibraryModel{
		store: &failingLibraryStore{},
	}
	model.Init()
	model.inputField.SetValue("/MyLibrary")
	model.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if len(model.currentLibraries.Items()) != 0 {
		t.Errorf("Wanted no items in list got %d", len(model.currentLibraries.Items()))
	}
}

func TestRemoveLibraryFailure(t *testing.T) {
	model := LibraryModel{
		store: &failingLibraryStore{},
	}
	model.Init()
	model.inputField.SetValue("/MyLibrary")
	model.Update(tea.KeyMsg{Type: tea.KeyDelete})
}

func TestRemoveLibraryFailureWithPopulatedList(t *testing.T) {
	model := LibraryModel{
		store: &failingLibraryStore{},
	}
	model.store.AddLibrary("/MyLibrary")
	model.Init()

	model.currentLibraries.InsertItem(0, &ListEntry{Path: "/MyLibrary", NumFiles: 0, Id: 1})
	model.Update(tea.KeyMsg{Type: tea.KeyDelete})
	if len(model.currentLibraries.Items()) != 1 {
		t.Errorf("Wanted one item in list got %d", len(model.currentLibraries.Items()))
	}
}

func TestRenderWithWrongType(t *testing.T) {
	model := LibraryModel{
		store: db.NewInMemoryLibraryList(),
	}
	model.Init()

	delegate := listProjectItemDelegate{}
	writer := bytes.NewBufferString("")

	delegate.Render(writer, model.currentLibraries, 0, nil)
	if writer.String() != "" {
		t.Errorf("Wanted empty string got %s", writer.String())
	}
}
