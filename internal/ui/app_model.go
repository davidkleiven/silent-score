package ui

import (
	"log/slog"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"gorm.io/gorm"
)

type AppView uint

const (
	ProjectOverview = iota
)

type AppModel struct {
	models       []tea.Model
	active       AppView
	currentWidth int
}

func NewAppModel(db *gorm.DB) *AppModel {
	return &AppModel{
		models: []tea.Model{
			&ProjectOverviewModel{db: db},
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
	}
	_, cmd := a.models[a.active].Update(msg)
	return a, cmd
}

func (a *AppModel) View() string {
	slog.Info("Running view of main")
	return lipgloss.JoinVertical(lipgloss.Left,
		headerStyle.Width(a.currentWidth).Render("Silent Score"),
		a.models[a.active].View(),
	)
}
