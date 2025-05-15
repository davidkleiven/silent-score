package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/davidkleiven/silent-score/internal/db"
)

func TestAppIsViewable(t *testing.T) {
	app := NewAppModel(initProjectDb())
	app.Init()
	app.View()
}

func TestAppEvent(t *testing.T) {
	for _, test := range []struct {
		msg   tea.Msg
		check func(t *testing.T, a *AppModel)
		desc  string
	}{
		{
			msg: tea.WindowSizeMsg{Width: 32, Height: 120},
			check: func(t *testing.T, a *AppModel) {
				if a.view.Width != 32 || a.view.Height != 120 {
					t.Errorf("Wanted w/h 32/120 got %d/%d", a.view.Width, a.view.Height)
				}
			},
			desc: "Check collecting window width on event",
		},
		{
			msg:   tea.KeyMsg{Type: tea.KeyCtrlC},
			check: func(t *testing.T, a *AppModel) {},
			desc:  "Quit on ctl+c",
		},
	} {

		t.Run(test.desc, func(t *testing.T) {
			a := NewAppModel(initProjectDb())
			a.Init()
			a.Update(test.msg)
			test.check(t, a)
		})
	}
}

func TestTransitionFromProjectOverviewToWorkspace(t *testing.T) {
	app := NewAppModel(initProjectDb())
	app.Init()

	switch app.current.(type) {
	case *ProjectOverviewModel:
	default:
		t.Error("Wanted project overview")

	}

	_, cmd := app.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if cmd == nil {
		t.Error("Wanted command to be not nil")
	}

	result := cmd()

	switch result.(type) {
	case toProjectWorkspace:
	default:
		t.Error("Wanted project list")
	}
}

func TestTransitionToLibraryList(t *testing.T) {
	app := NewAppModel(initProjectDb())
	app.Init()
	_, cmd := app.Update(tea.KeyMsg{Type: tea.KeyCtrlL})

	if cmd == nil {
		t.Error("Wanted command to be not nil")
	}

	result := cmd()

	switch result.(type) {
	case toLibraryList:
	default:
		t.Error("Wanted library list")
	}
}

func TestTransitionToProjectOverview(t *testing.T) {
	app := NewAppModel(initProjectDb())
	app.Init()
	app.Update(toProjectOverview{})
	switch app.current.(type) {
	case *ProjectOverviewModel:
	default:
		t.Error("Wanted project overview")
	}
}

func TestTransitionToLibraryListInApp(t *testing.T) {
	app := NewAppModel(initProjectDb())
	app.Init()
	app.Update(toLibraryList{})

	switch app.current.(type) {
	case *LibraryModel:
	default:
		t.Error("Wanted library list")
	}
}

func TestTransitionToProjectWorkspace(t *testing.T) {
	app := NewAppModel(initProjectDb())
	app.Init()
	app.Update(toProjectWorkspace{project: &db.Project{}})

	switch app.current.(type) {
	case *ProjectWorkspace:
	default:
		t.Error("Wanted project workspace")
	}
}

func TestLibraries(t *testing.T) {
	store := db.NewInMemoryLibraryList()
	store.AddLibrary("test")
	store.AddLibrary("test2")
	libs := libraries(store)
	if len(libs) != 2 {
		t.Errorf("Wanted 2 libraries, got %d", len(libs))
	}
}

func TestLibrariesFailingStore(t *testing.T) {
	store := failingLibraryStore{}
	libs := libraries(&store)
	if len(libs) != 0 {
		t.Errorf("Wanted 0 libraries, got %d", len(libs))
	}
}
