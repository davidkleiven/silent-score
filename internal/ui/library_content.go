package ui

import (
	"slices"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/davidkleiven/silent-score/internal/compose"
)

type LibraryContentView struct {
	lib     compose.Library
	content list.Model
	width   int
	height  int
}

func listHeight(componentHeight int) int {
	return confine(componentHeight-1, 0, componentHeight)
}

func (l *LibraryContentView) Init() tea.Cmd {
	var items []list.Item
	for _, item := range l.lib.Content() {
		items = append(items, &item)
	}
	slices.SortFunc(items, func(i, j list.Item) int {
		if i.FilterValue() < j.FilterValue() {
			return -1
		} else if i.FilterValue() > j.FilterValue() {
			return 1
		}
		return 0
	})

	l.content = list.New(items, list.NewDefaultDelegate(), l.width, listHeight(l.height))
	l.content.SetFilteringEnabled(true)
	l.content.SetShowFilter(true)
	l.content.SetShowHelp(true)
	l.content.SetShowTitle(false)
	return nil
}

func (l *LibraryContentView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		l.content.SetSize(msg.Width, listHeight(msg.Height))
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if l.content.FilterState() == list.Unfiltered {
				cmds = append(cmds, func() tea.Msg {
					return toProjectOverview{}
				})
			}
		}
	}
	var cmd tea.Cmd
	l.content, cmd = l.content.Update(msg)
	cmds = append(cmds, cmd)
	return l, tea.Batch(cmds...)
}

func (l *LibraryContentView) View() string {
	return l.content.View()
}
