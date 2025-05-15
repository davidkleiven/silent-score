package ui

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/davidkleiven/silent-score/internal/db"
	"golang.org/x/exp/slog"
)

type ListEntry struct {
	Path     string
	NumFiles int
	Id       uint
}

func (l *ListEntry) FilterValue() string {
	return l.Path
}

type listProjectItemDelegate struct{}

func (d listProjectItemDelegate) Height() int                             { return 1 }
func (d listProjectItemDelegate) Spacing() int                            { return 0 }
func (d listProjectItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d listProjectItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	project, ok := listItem.(*ListEntry)

	if !ok {
		slog.Error("Could not cast list item to db.Project", nil)
		return
	}

	fn := itemStyle.Render
	s := fmt.Sprintf("%-60s %10d", project.Path, project.NumFiles)
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("> " + strings.Join(s, " "))
		}
	}
	fmt.Fprint(w, fn(s))
}

type LibraryModel struct {
	store            db.LibraryList
	currentLibraries list.Model
	inputField       textinput.Model
	currentBestGuess string
}

func (l *LibraryModel) Init() tea.Cmd {
	var items []list.Item
	l.currentLibraries = list.New(items, listProjectItemDelegate{}, 20, 14)
	cmd := l.loadFromDb()
	l.currentLibraries.Title = fmt.Sprintf("%-60s %10s", "File path", "#musicxml")
	l.currentLibraries.Styles.Title = helpStyle
	l.currentLibraries.SetWidth(120)
	l.inputField = textinput.New()
	l.inputField.Placeholder = "Enter path to library"
	l.inputField.Width = 120
	return tea.Batch(cmd, l.inputField.Focus())
}

func (l *LibraryModel) loadFromDb() tea.Cmd {
	var items []list.Item

	libaries, err := l.store.ListLibraries()
	if err != nil {
		slog.Error("Could not list libraries", err)
	}

	for _, item := range libaries {
		items = append(items, &ListEntry{Path: item.Path, NumFiles: numberOfScores(item.Path), Id: item.ID})
	}
	return l.currentLibraries.SetItems(items)
}

func (l *LibraryModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var inpCmd tea.Cmd
	_, libCmd := l.currentLibraries.Update(msg)
	l.inputField, inpCmd = l.inputField.Update(msg)
	l.currentBestGuess = existingFolder(l.inputField.Value())

	cmds = append(cmds, libCmd)
	cmds = append(cmds, inpCmd)
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return l, func() tea.Msg {
				return toProjectOverview{}
			}
		case "tab":
			l.inputField.SetValue(l.currentBestGuess)
			l.inputField.CursorEnd()

			// Trigger a new slash at the end of the input
			cmds = append(cmds, func() tea.Msg {
				return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}}
			})
		case "enter":
			if err := l.store.AddLibrary(l.inputField.Value()); err != nil {
				slog.Error("Could not add library", err)
				return l, nil
			}
			cmd := l.loadFromDb()
			cmds = append(cmds, cmd)
			l.inputField.SetValue("")
			l.inputField.CursorStart()
		case "down":
			l.currentLibraries.CursorDown()
		case "up":
			l.currentLibraries.CursorUp()
		case "delete":
			selected, ok := l.currentLibraries.SelectedItem().(*ListEntry)
			if !ok {
				slog.Error("Could not convert into ListEntry", nil)
				return l, nil
			}
			if err := l.store.RemoveLibrary(selected.Id); err != nil {
				slog.Error("Could not remove library", err)
				return l, nil
			}
			cmd := l.loadFromDb()
			cmds = append(cmds, cmd)
			l.currentLibraries.CursorUp()
		}

	}
	return l, tea.Batch(cmds...)
}

func (l *LibraryModel) View() string {
	content := []string{
		l.currentLibraries.View(),
		fmt.Sprintf("Add library: %s", l.currentBestGuess),
		l.inputField.View(),
		helpStyle.Render("esc: to project overview \u2022 enter: add library \u2022 delete: remove library"),
	}
	return lipgloss.JoinVertical(lipgloss.Left, content...)
}

func existingFolder(path string) string {
	matches, err := filepath.Glob(path + "*")
	if err != nil {
		panic(err)
	}

	if len(matches) > 0 {
		return matches[0]
	}
	return path
}

func numberOfScores(path string) int {
	matches, err := filepath.Glob(path + "*.musicxml")
	if err != nil {
		panic(err)
	}
	return len(matches)
}
