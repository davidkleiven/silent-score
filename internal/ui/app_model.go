package ui

import (
	"log/slog"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"gorm.io/gorm"
)

type AppView uint

const (
	AppViewProjectList AppView = iota
	AppViewProjectWorkspace
)

type AppModel struct {
	models       []MultiViewModel
	active       AppView
	currentWidth int
}

type MultiViewModel interface {
	tea.Model
	NextView() AppView
}

func NewAppModel(db *gorm.DB) *AppModel {
	return &AppModel{
		models: []MultiViewModel{
			&ProjectOverviewModel{db: db, nextProjectView: AppViewProjectList},
			&ProjectWorkspace{database: db, nextAppView: AppViewProjectWorkspace},
		},
	}
}

func (a *AppModel) Init() tea.Cmd {
	return a.models[a.active].Init()
}

func (a *AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.currentWidth = msg.Width
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return a, tea.Quit
		}
	}
	_, cmd := a.models[a.active].Update(msg)

	if nextView := a.models[a.active].NextView(); nextView != a.active {
		a.active = nextView
		a.Init()
	}
	return a, cmd
}

func (a *AppModel) View() string {
	slog.Info("Running view of main")
	return lipgloss.JoinVertical(lipgloss.Left,
		headerStyle.Width(a.currentWidth).Render("Silent Score"),
		a.models[a.active].View(),
	)
}
