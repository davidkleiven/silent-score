package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
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

	app.Update(tea.KeyMsg{Type: tea.KeyEnter})

	switch app.current.(type) {
	case *ProjectWorkspace:
	default:
		t.Error("Wanted project list")
	}
}
