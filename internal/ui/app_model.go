package ui

import (
	"log/slog"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/davidkleiven/silent-score/internal/compose"
	"github.com/davidkleiven/silent-score/internal/db"
	"github.com/davidkleiven/silent-score/internal/musicxml"
)

type AppModel struct {
	view    viewport.Model
	current tea.Model
	store   db.Store
}

func NewAppModel(store db.Store) *AppModel {
	vp := viewport.New(120, 32)

	a := AppModel{view: vp, store: store, current: &ProjectOverviewModel{store: store}}
	return &a
}

func (a *AppModel) Init() tea.Cmd {
	return a.current.Init()
}

func (a *AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var nextModel tea.Model
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.view.Width = msg.Width
		a.view.Height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return a, tea.Quit
		}
	case toProjectOverview:
		nextModel = &ProjectOverviewModel{store: a.store}
	case toProjectWorkspace:
		libs := libraries(a.store)
		libs = append(libs, compose.NewStandardLibrary())
		nextModel = &ProjectWorkspace{
			store:   a.store,
			project: msg.project,
			library: compose.NewMultiSourceLibrary(libs...),
			creator: &musicxml.FileCreator{},
		}
	case toLibraryList:
		nextModel = &LibraryModel{store: a.store}
	}

	if nextModel != nil && nextModel != a.current {
		nextModel.Init()
		a.current = nextModel
	}
	_, cmd := a.current.Update(msg)
	return a, cmd
}

func (a *AppModel) View() string {
	content := lipgloss.JoinVertical(lipgloss.Left,
		headerStyle.Width(a.view.Width).Render("Silent Score"),
		a.current.View(),
	)
	a.view.SetContent(content)
	return a.view.View()
}

func libraries(store db.LibraryList) []compose.Library {
	libraries, err := store.ListLibraries()
	var result []compose.Library
	if err != nil {
		slog.Error("Could not list libraries", "err", err)
		return result
	}

	for _, item := range libraries {
		result = append(result, compose.NewLocalLibrary(item.Path))
	}
	return result
}
