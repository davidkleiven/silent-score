package ui

import (
	"fmt"
	"log/slog"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/davidkleiven/silent-score/internal/db"
)

func confine(num, lower, upper int) int {
	if num < lower {
		return lower
	} else if num > upper {
		return upper
	}
	return num
}

type tiRow []textinput.Model

func NewTiRow(opts ...tiOpt) tiRow {
	var (
		sceneTi     = textinput.New()
		sceneDescTi = textinput.New()
		tempoTi     = textinput.New()
		keywordsTi  = textinput.New()
		themeTi     = textinput.New()
		startTi     = textinput.New()
	)

	sceneTi.CharLimit = 4
	sceneDescTi.CharLimit = 64
	tempoTi.CharLimit = 3
	keywordsTi.Width = 120
	themeTi.CharLimit = 1
	startTi.CharLimit = 8

	row := []textinput.Model{sceneTi, sceneDescTi, tempoTi, keywordsTi, themeTi, startTi}

	for _, fn := range opts {
		fn(row)
	}
	row[0].Focus()
	return row
}

func (t tiRow) Blur() {
	for _, r := range t {
		r.Blur()
	}
}

func (t tiRow) Active() int {
	for i, item := range t {
		if item.Focused() {
			return i
		}
	}
	return 0
}

func (t tiRow) FocusRight() {
	active := t.Active()
	t[active].Blur()
	t[confine(active+1, 0, len(t)-1)].Focus()
}

func (t tiRow) FocusLeft() {
	active := t.Active()
	t[active].Blur()
	t[confine(active-1, 0, len(t)-1)].Focus()
}

func (t tiRow) Scene() *textinput.Model {
	return &t[0]
}

func (t tiRow) Content() table.Row {
	data := make([]string, len(t))
	for i, item := range t {
		data[i] = item.Value()
	}
	return data
}

type tiOpt func(row tiRow)

func WithScene(scene uint) tiOpt {
	return func(row tiRow) {
		row.Scene().SetValue(fmt.Sprintf("%d", scene))
	}
}

type InteractiveTable struct {
	iRows []tiRow
	table table.Model
}

func NewInteractiveTable() *InteractiveTable {
	columns := []table.Column{
		{
			Title: "Scene",
			Width: 5,
		},
		{
			Title: "Scene description",
			Width: 18,
		},
		{
			Title: "Tempo",
			Width: 5,
		},
		{
			Title: "Piece keywords",
			Width: 80,
		},
		{
			Title: "Theme",
			Width: 5,
		},
		{
			Title: "Start time",
			Width: 10,
		},
	}

	return &InteractiveTable{
		table: table.New(table.WithColumns(columns), table.WithFocused(true)),
	}
}

func (it *InteractiveTable) Init() tea.Cmd {
	return nil
}

func (it *InteractiveTable) View() string {
	return it.table.View()
}

func (it *InteractiveTable) activeTiRow() tiRow {
	if len(it.iRows) > 0 {
		return it.iRows[it.table.Cursor()]
	}
	return nil
}

func (it *InteractiveTable) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, it.table.KeyMap.LineDown):
			it.handleDown()
		}

		switch msg.String() {
		case "left":
			if r := it.activeTiRow(); r != nil {
				r.FocusLeft()
			}
		case "right":
			if r := it.activeTiRow(); r != nil {
				r.FocusRight()
			}
		}
	}

	// Ensure correct row is activated
	it.blurActiveRow()
	_, cmd := it.table.Update(msg)
	it.activateCurrentRow()
	return it, cmd
}

func (it *InteractiveTable) blurActiveRow() {
	if len(it.iRows) > 0 {
		it.iRows[it.table.Cursor()].Blur()
	}
}

func (it *InteractiveTable) activateCurrentRow() {
	if len(it.iRows) > 0 {
		it.iRows[it.table.Cursor()][0].Focus()
	}
}

func (it *InteractiveTable) handleDown() {
	if len(it.iRows) == 0 {
		it.createNewRow(0)
		return
	}
	if it.table.Cursor() == len(it.iRows)-1 {
		it.createNewRow(it.table.Cursor() + 1)
	}

}

func (it *InteractiveTable) createNewRow(scene int) {
	slog.Info("Creating new row")
	newRow := NewTiRow(WithScene(uint(scene)))
	it.iRows = append(it.iRows, newRow)
	it.table.SetRows(append(it.table.Rows(), newRow.Content()))
}

type ProjectWorkspace struct {
	database    db.CreateReadUpdateDeleter
	projectId   uint
	nextAppView AppView
	status      *Status
	iTable      *InteractiveTable
}

func (pw *ProjectWorkspace) NextView() AppView {
	return pw.nextAppView
}

func (pw *ProjectWorkspace) Init() tea.Cmd {
	pw.status = NewStatus()
	pw.iTable = NewInteractiveTable()
	return nil
}

func (pw *ProjectWorkspace) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		pw.iTable.table.SetWidth(msg.Width)
		return pw, nil
	}
	pw.iTable.Update(msg)
	return pw, nil
}

func (pw *ProjectWorkspace) View() string {
	return lipgloss.JoinVertical(lipgloss.Left, pw.iTable.View(), pw.status.msg)
}
