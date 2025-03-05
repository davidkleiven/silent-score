package ui

import (
	"log/slog"
	"regexp"
	"strconv"

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
		sceneDescTi = textinput.New()
		tempoTi     = textinput.New()
		keywordsTi  = textinput.New()
		themeTi     = textinput.New()
		startTi     = textinput.New()
	)

	sceneDescTi.Width = 64
	sceneDescTi.Prompt = ""

	tempoTi.Width = 6
	tempoTi.Prompt = ""

	keywordsTi.Width = 120
	keywordsTi.Prompt = ""

	themeTi.Width = 6
	themeTi.Prompt = ""

	startTi.Width = 8
	startTi.Prompt = ""
	startTi.Placeholder = "HH:MM:SS"

	row := []textinput.Model{sceneDescTi, tempoTi, keywordsTi, themeTi, startTi}

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

func (t tiRow) View() string {
	data := make([]string, len(t))
	for i, item := range t {
		data[i] = item.View()
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, data...)
}

func (t tiRow) Update(msg tea.Msg) {
	for i, item := range t {
		t[i], _ = item.Update(msg)
	}
}

func (t tiRow) Time() string {
	return t[4].Value()
}

func (t tiRow) Tempo() string {
	return t[1].Value()
}

type tiOpt func(row tiRow)

func WithTime(start string) tiOpt {
	return func(ti tiRow) {
		ti[4].SetValue(start)
	}
}

type InteractiveTable struct {
	iRows  []tiRow
	cursor int
}

func NewInteractiveTable() *InteractiveTable {
	return &InteractiveTable{}
}

func (it *InteractiveTable) Header() string {
	style := lipgloss.NewStyle()

	names := []string{"Scene desc", "Tempo", "Keywords", "Theme", "Time"}
	header := make([]string, len(names))
	for i, name := range names {
		width := 20
		if len(it.iRows) > 0 {
			width = it.iRows[0][i].Width + it.iRows[0][i].TextStyle.GetPaddingLeft() + it.iRows[0][i].TextStyle.GetPaddingRight()
		}
		header[i] = style.Width(width + 1).Render(name)
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, header...)
}

func (it *InteractiveTable) Init() tea.Cmd {
	return nil
}

func (it *InteractiveTable) View() string {
	rows := make([]string, len(it.iRows))
	for i, r := range it.iRows {
		rows[i] = r.View()
	}
	rows = append([]string{it.Header()}, rows...)
	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

func (it *InteractiveTable) activeTiRow() tiRow {
	if len(it.iRows) > 0 {
		return it.iRows[it.cursor]
	}
	return nil
}

func (it *InteractiveTable) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:

		switch msg.String() {
		case "down":
			it.handleDown()
			it.blurActiveRow()
			it.cursor = confine(it.cursor+1, 0, len(it.iRows)-1)
			it.activateCurrentRow()
		case "up":
			it.blurActiveRow()
			it.cursor = confine(it.cursor-1, 0, len(it.iRows)-1)
			it.activateCurrentRow()
		case "left":
			if r := it.activeTiRow(); r != nil {
				slog.Info("Switching active row left")
				r.FocusLeft()
			}
		case "right", "tab":
			if r := it.activeTiRow(); r != nil {
				slog.Info("Switching active row right")
				r.FocusRight()
			}
		}

		// Ensure correct row is activated
		if r := it.activeTiRow(); r != nil {
			r.Update(msg)
		}
	}
	return it, nil
}

func (it *InteractiveTable) blurActiveRow() {
	if len(it.iRows) > 0 {
		it.iRows[it.cursor].Blur()
	}
}

func (it *InteractiveTable) activateCurrentRow() {
	if len(it.iRows) > 0 {
		it.iRows[it.cursor][0].Focus()
	}
}

func (it *InteractiveTable) handleDown() {
	if len(it.iRows) == 0 {
		it.createNewRow(0)
		return
	}

	if it.cursor == len(it.iRows)-1 {
		it.createNewRow(it.cursor + 1)
	}
}

func (it *InteractiveTable) createNewRow(scene int) {
	slog.Info("Creating new row")
	newRow := NewTiRow()
	it.iRows = append(it.iRows, newRow)
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
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			pw.status.Set("OK", nil)
			if err := pw.validate(); err != nil {
				pw.status.Set("", err)
			}
		}
	}
	pw.iTable.Update(msg)
	return pw, nil
}

func (pw *ProjectWorkspace) View() string {

	helpString := helpStyle.Render("\u2191/\u2193 up/down \u2022 \u2190/\u2192 left/right \u2022 ctrl+(\u2191/\u2193) move row up/down \u2022 ctrl+c: quit")
	return lipgloss.JoinVertical(lipgloss.Left, pw.iTable.View(), helpString, pw.status.Render("Edit"))
}

func (pw *ProjectWorkspace) validate() error {
	for _, item := range pw.iTable.iRows {
		if err := validateTime(item.Time()); err != nil {
			return err
		} else if err := validateTempo(item.Tempo()); err != nil {
			return err
		}
	}
	return nil
}

func validateTime(initTime string) error {
	if initTime == "" {
		return nil
	}
	hoursMinSec := regexp.MustCompile(`\d{2}:\d{2}:\d{2}`)

	if !hoursMinSec.MatchString(initTime) {
		return ErrWrongTimeFormat
	}
	return nil
}

func validateTempo(tempo string) error {
	if tempo == "" {
		return nil
	}
	_, err := strconv.Atoi(tempo)
	if err != nil {
		return ErrTempoMustBeInteger
	}
	return nil
}
