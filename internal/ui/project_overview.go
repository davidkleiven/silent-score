package ui

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/davidkleiven/silent-score/internal/db"
)

type ProjectOverviewMode int

const (
	browseMode ProjectOverviewMode = iota
	newProjectMode
	deleteConfirmationMode
)

type ProjectOverviewModel struct {
	store          db.ProjectStore
	projects       list.Model
	status         *Status
	newProjectName textinput.Model
	mode           ProjectOverviewMode
}

type projectItemDelegate struct{}

func (d projectItemDelegate) Height() int                             { return 1 }
func (d projectItemDelegate) Spacing() int                            { return 0 }
func (d projectItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d projectItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	project, ok := listItem.(*db.Project)

	if !ok {
		slog.Error("Could not cast list item to db.Project")
		return
	}

	s := fmt.Sprintf("%d. %s", index+1, project.Name)
	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("> " + strings.Join(s, " "))
		}
	}
	fmt.Fprint(w, fn(s))
}

//
// Initialization methods
//

func (p *ProjectOverviewModel) loadProjectsFromDb() {
	slog.Info("Loading projects from database")
	projects, err := p.store.Load()
	if err != nil {
		panic(err)
	}
	slog.Info(fmt.Sprintf("Loaded %d projects", len(projects)))

	var items []list.Item
	for _, project := range projects {
		items = append(items, &project)
	}
	p.projects = list.New(items, projectItemDelegate{}, 20, 14)
	p.projects.SetShowTitle(false)
	p.projects.Styles.PaginationStyle = paginationStyle
	p.projects.SetStatusBarItemName("project", "projects")
}

func (p *ProjectOverviewModel) Init() tea.Cmd {

	p.loadProjectsFromDb()
	if p.status == nil {
		p.status = NewStatus()
	}
	p.newProjectName = textinput.New()
	p.newProjectName.Placeholder = "Press ctrl+n to create new project"
	p.newProjectName.CharLimit = 128
	p.toBrowseMode("")
	return nil
}

//
// Mode changing actions
//

func (p *ProjectOverviewModel) toBrowseMode(msg string) {
	p.mode = browseMode
	p.newProjectName.Blur()
	p.newProjectName.Reset()
	p.status.Set(msg, nil)
}

func (p *ProjectOverviewModel) toTextInputMode(msg string) {
	p.mode = newProjectMode
	p.newProjectName.Focus()
	p.status.Set(msg, nil)
}

func (p *ProjectOverviewModel) toDeleteConfirmation() {
	p.mode = deleteConfirmationMode
	p.status.Set(fmt.Sprintf("Are you sure that %s should be deleted? (y/N)", p.projects.SelectedItem().FilterValue()), nil)
}

//
// Actions
//

func (p *ProjectOverviewModel) createNewProject() {
	if p.newProjectName.Value() == "" {
		p.status.Set("", errors.New("project name can not be empty"))
		return
	}
	err := p.store.Save(db.NewProject(db.WithName(p.newProjectName.Value())))
	if err != nil {
		p.status.Set("", err)
	} else {
		p.toBrowseMode("Successfully created new project")
	}
	p.loadProjectsFromDb()
}

func (p *ProjectOverviewModel) deleteChosenProject() {
	project, ok := p.projects.SelectedItem().(*db.Project)
	if !ok {
		slog.Info("Could not convert into Project")
		return
	}
	err := p.store.Delete(project.Id)
	p.status.Set(fmt.Sprintf("Successfully deleted %s", project.Name), err)
	p.loadProjectsFromDb()
}

//
// UI updates
//

func (p *ProjectOverviewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		p.projects.SetWidth(msg.Width)
		return p, nil
	}

	var cmd tea.Cmd
	switch p.mode {
	case newProjectMode:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "enter":
				p.createNewProject()
			case "esc":
				p.toBrowseMode("")
			}
		}
		p.newProjectName, cmd = p.newProjectName.Update(msg)
		return p, cmd
	case deleteConfirmationMode:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "y", "Y":
				p.deleteChosenProject()
				p.toBrowseMode("")
			default:
				p.toBrowseMode("")
			}
		}
		return p, nil
	default:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "ctrl+n":
				p.toTextInputMode("")
			case "delete":
				p.toDeleteConfirmation()
			case "enter":
				p.status.Set(fmt.Sprintf("Selected project %s", p.projects.SelectedItem().FilterValue()), nil)
				return p, func() tea.Msg {
					return toProjectWorkspace{
						project: p.projects.SelectedItem().(*db.Project),
					}
				}
			case "ctrl+l":
				return p, func() tea.Msg {
					return toLibraryList{}
				}
			case "ctrl+a":
				return p, func() tea.Msg {
					return toLibraryContent{}
				}
			}
		}
		p.projects, cmd = p.projects.Update(msg)
		return p, cmd
	}
}

func (p *ProjectOverviewModel) View() string {
	content := []string{
		p.projects.View(),
		p.newProjectName.View(),
		helpStyle.Render("ctrl+n: New project \u2022 delete: Delete project \u2022 enter: Open project \u2022 ctrl+l: List libraries \u2022 ctrl+a: List pieces \u2022 ctrl+c: Quit"),
		p.status.Render(modeDescription(p.mode)),
	}
	return lipgloss.JoinVertical(lipgloss.Left, content...)
}

func modeDescription(mode ProjectOverviewMode) string {
	switch mode {
	case browseMode:
		return "Browse mode"
	case deleteConfirmationMode:
		return "Delete confirm"
	default:
		return "Text enter mode (esc to leave)"
	}
}
