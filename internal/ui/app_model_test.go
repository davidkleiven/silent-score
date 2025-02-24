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
		view  AppView
		check func(t *testing.T, a *AppModel)
		desc  string
	}{
		{
			msg: tea.WindowSizeMsg{Width: 32},
			check: func(t *testing.T, a *AppModel) {
				if a.currentWidth != 32 {
					t.Errorf("Current width should be 32 got %d", a.currentWidth)
				}
			},
			desc: "Check collecting window width on event",
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
