package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/davidkleiven/silent-score/internal/db"
)

func initProjectDb() *db.InMemoryStore {
	store := db.NewInMemoryStore()

	store.Save(db.NewProject(db.WithName("project1")))
	store.Save(db.NewProject(db.WithName("project2")))
	return store
}

func TestProjectOverviewInit(t *testing.T) {
	model := ProjectOverviewModel{store: initProjectDb()}
	model.Init()

	if model.mode != browseMode {
		t.Errorf("Expected mode to be %d got %d", browseMode, model.mode)
	}

	names := make(map[string]struct{})
	for _, item := range model.projects.Items() {
		names[item.FilterValue()] = struct{}{}
	}

	want := []string{"project1", "project2"}
	for _, name := range want {
		if _, ok := names[name]; !ok {
			t.Errorf("Wanted %v got %v\n", want, names)
			return
		}
	}

}

func TestEvents(t *testing.T) {
	// These tests can not be run in parallel because they alter model
	for _, test := range []struct {
		msg   tea.Msg
		check func(t *testing.T, model *ProjectOverviewModel)
		desc  string
		mode  ProjectOverviewMode
	}{
		{
			msg: tea.WindowSizeMsg{Width: 50},
			check: func(t *testing.T, model *ProjectOverviewModel) {
				if w := model.projects.Width(); w != 50 {
					t.Errorf("Wanted width 50 got %d", w)
				}
			},
			desc: "Check width propagated to underlying list",
		},
		{
			msg: tea.KeyMsg{Type: tea.KeyCtrlN},
			check: func(t *testing.T, model *ProjectOverviewModel) {
				if !model.newProjectName.Focused() {
					t.Errorf("Expected text input to be focused")
				}

				if model.mode != newProjectMode {
					t.Errorf("Expected mode to be %d got %d", newProjectMode, model.mode)
				}
			},
			desc: "Enter newproject mode on ctrl+n",
		},
		{
			msg: tea.KeyMsg{Type: tea.KeyDelete},
			check: func(t *testing.T, model *ProjectOverviewModel) {
				if model.mode != deleteConfirmationMode {
					t.Errorf("Expected to enter mode %d got %d", deleteConfirmationMode, model.mode)
				}
			},
			desc: "Enter delete confirmation mode on delete",
		},
		{
			msg: tea.KeyMsg{Type: tea.KeyEnter},
			check: func(t *testing.T, model *ProjectOverviewModel) {
				if model.mode != browseMode {
					t.Errorf("Expected to be in mode %d got %d", browseMode, model.mode)
				}

				if !strings.Contains(model.status.msg, "Selected project") {
					t.Errorf("Expected status field to contain \"Selected project\"")
				}
			},
			desc: "Check selected project is reported when enter is pressed",
		},
		{
			msg: tea.KeyMsg{Type: tea.KeyEsc},
			check: func(t *testing.T, model *ProjectOverviewModel) {
				if model.mode != browseMode {
					t.Errorf("Expected to be in mode %d got %d", browseMode, model.mode)
				}
			},
			desc: "Re-enters browse mode on esc",
			mode: newProjectMode,
		},
		{
			msg: tea.KeyMsg{Type: tea.KeyEnter},
			check: func(t *testing.T, model *ProjectOverviewModel) {
				if model.mode != newProjectMode {
					t.Errorf("Expected to be in mode %d got %d", newProjectMode, model.mode)
				}
				if !strings.Contains(model.status.msg, "name can not be empty") {
					t.Errorf("Expected empty error to be set since name is empty got %s", model.status.msg)
				}
			},
			desc: "Executes createNewProject on enter",
			mode: newProjectMode,
		},
		{
			msg: tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}},
			check: func(t *testing.T, model *ProjectOverviewModel) {
				if n := len(model.projects.Items()); n != 1 {
					t.Errorf("Expected one item to be deleted. Current number of items %d", n)
				}
			},
			desc: "Check that one item is deleted",
			mode: deleteConfirmationMode,
		},
		{
			msg: tea.KeyMsg{Type: tea.KeyEnter},
			check: func(t *testing.T, model *ProjectOverviewModel) {
				if model.mode != browseMode {
					t.Errorf("Expected to re-enter mode %d got %d", browseMode, model.mode)
				}

				if n := len(model.projects.Items()); n != 2 {
					t.Errorf("At least one project was deleted")
				}

			},
			desc: "Re-enter browse mode without deleetion on other any key press",
			mode: deleteConfirmationMode,
		},
	} {
		t.Run(test.desc, func(t *testing.T) {
			model := ProjectOverviewModel{store: initProjectDb()}
			model.Init()
			switch test.mode {
			case deleteConfirmationMode:
				model.toDeleteConfirmation()
			case newProjectMode:
				model.toTextInputMode("")
			}
			model.Update(test.msg)
			test.check(t, &model)
			model.View() // Just ensure we don't get any errors during View
		})
	}
}

func TestModeDescription(t *testing.T) {

	tokens := map[ProjectOverviewMode]string{
		browseMode:             "Browse",
		newProjectMode:         "Text",
		deleteConfirmationMode: "Delete",
	}
	for _, mode := range []ProjectOverviewMode{browseMode, deleteConfirmationMode, newProjectMode} {
		if desc := modeDescription(mode); !strings.Contains(desc, tokens[mode]) {
			t.Errorf("Expected %s tp be part of %s", tokens[mode], desc)
		}
	}
}

func TestNoErrorOnWhenDeletingOnEmptyList(t *testing.T) {
	view := &ProjectOverviewModel{}
	view.deleteChosenProject() // Used to crash
}

type filterString string

func (f filterString) FilterValue() string { return "" }

func TestProjectItemDelegatorNoCrashOnWrongType(t *testing.T) {
	model := ProjectOverviewModel{store: initProjectDb()}
	model.Init()
	model.projects.SetItem(0, filterString("whatever"))
	model.projects.View()
}

func TestSaveProjectSuccessfully(t *testing.T) {
	model := ProjectOverviewModel{store: initProjectDb()}
	model.Init()
	model.newProjectName.SetValue("awesomeProject")
	origNumProjects := len(model.projects.Items())
	model.createNewProject()

	if len(model.projects.Items()) != origNumProjects+1 {
		t.Errorf("Expected one extra project to be created")
	}

	if model.mode != browseMode {
		t.Errorf("Should return to mode %d is in %d", browseMode, model.mode)
	}
}

func TestSaveProjectWithRepeatedName(t *testing.T) {
	model := ProjectOverviewModel{store: initProjectDb()}
	model.Init()
	model.toTextInputMode("")
	model.newProjectName.SetValue(model.projects.SelectedItem().FilterValue())
	origNumProjects := len(model.projects.Items())
	model.createNewProject()

	if len(model.projects.Items()) != origNumProjects {
		t.Errorf("Expected one extra project to be created")
	}

	if model.mode != newProjectMode {
		t.Errorf("Should continue in mode %d is in %d", newProjectMode, model.mode)
	}

	if model.status.kind != errorStatus {
		t.Error("Status should be in the error state")
	}
}
