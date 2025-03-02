package ui

import (
	"log/slog"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"gorm.io/gorm"
)

type AppModel struct {
	view    viewport.Model
	current tea.Model
}

func NewAppModel(db *gorm.DB) *AppModel {
	vp := viewport.New(120, 32)
	return &AppModel{
		current: &ProjectOverviewModel{db: db},
		view:    vp,
	}
}

func (a *AppModel) Init() tea.Cmd {
	return a.current.Init()
}

func (a *AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.view.Width = msg.Width
		a.view.Height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return a, tea.Quit
		}
	}
	nextModel, cmd := a.current.Update(msg)

	if nextModel != a.current {
		nextModel.Init()
		a.current = nextModel
	}
	return a, cmd
}

func (a *AppModel) View() string {
	slog.Info("Running view of main")
	content := lipgloss.JoinVertical(lipgloss.Left,
		headerStyle.Width(a.view.Width).Render("Silent Score"),
		a.current.View(),
	)
	a.view.SetContent(content)
	return a.view.View()
}
