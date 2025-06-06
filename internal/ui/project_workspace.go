package ui

import (
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/davidkleiven/silent-score/internal/compose"
	"github.com/davidkleiven/silent-score/internal/db"
	"github.com/davidkleiven/silent-score/internal/musicxml"
	"github.com/davidkleiven/silent-score/internal/utils"
)

func confine(num, lower, upper int) int {
	if num < lower {
		return lower
	} else if num > upper {
		return upper
	}
	return num
}

const (
	tiScene = iota
	tiTempo
	tiKeywords
	tiTheme
	tiDuration
)
const rowPadding = 2

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
	row := []textinput.Model{sceneDescTi, tempoTi, keywordsTi, themeTi, startTi}

	for _, fn := range opts {
		fn(row)
	}
	row[tiScene].Focus()
	return row
}

func (row tiRow) SetWidth(width int) {
	remainingWidth := width - 2*rowPadding
	row[tiTempo].Width = confine(6, 0, remainingWidth)
	remainingWidth = remainingWidth - row[tiTempo].Width - 1

	row[tiTheme].Width = confine(6, 0, remainingWidth)
	remainingWidth = remainingWidth - row[tiTheme].Width - 1

	row[tiDuration].Width = confine(14, 0, remainingWidth)
	remainingWidth = remainingWidth - row[tiDuration].Width - 1

	row[tiKeywords].Width = confine(remainingWidth/2, 0, remainingWidth)
	remainingWidth = remainingWidth - row[tiKeywords].Width - 1
	row[tiScene].Width = confine(remainingWidth, 0, remainingWidth)
}

func NewTiRowFromRecord(record *db.ProjectContentRecord) tiRow {
	row := NewTiRow()
	if record.Tempo > 0 {
		row[tiTempo].SetValue(fmt.Sprintf("%d", record.Tempo))
	}

	if record.Theme > 0 {
		row[tiTheme].SetValue(fmt.Sprintf("%d", record.Theme))
	}

	if record.DurationSec > 0 {
		row[tiDuration].SetValue(fmt.Sprintf("%d", record.DurationSec))
	}

	row[tiScene].SetValue(record.SceneDesc)
	row[tiKeywords].SetValue(record.Keywords)
	return row
}

func (t tiRow) Blur() {
	for i := range t {
		t[i].Blur()
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
	return strings.Repeat(" ", rowPadding) + lipgloss.JoinHorizontal(lipgloss.Top, data...)
}

func (t tiRow) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	cmds := make([]tea.Cmd, len(t))
	for i, item := range t {
		t[i], cmd = item.Update(msg)
		cmds[i] = cmd
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		t.SetWidth(msg.Width)
	}
	return tea.Batch(cmds...)
}

func (t tiRow) Duration() string {
	return t[tiDuration].Value()
}

func (t tiRow) Tempo() string {
	return t[tiTempo].Value()
}

func (t tiRow) TempoOrDefault() (int, error) {
	return intOrDefault(t[tiTempo].Value(), 0)
}

func (t tiRow) ThemeOrDefault() (int, error) {
	return intOrDefault(t[tiTheme].Value(), 0)
}

func (t tiRow) DurationOrDefault() (int, error) {
	return intOrDefault(t[tiDuration].Value(), 0)
}

type tiOpt func(row tiRow)

func WithDuration(start string) tiOpt {
	return func(ti tiRow) {
		ti[tiDuration].SetValue(start)
	}
}

func WithTempo(tempo string) tiOpt {
	return func(ti tiRow) {
		ti[tiTempo].SetValue(tempo)
	}
}

func WithTheme(theme string) tiOpt {
	return func(ti tiRow) {
		ti[tiTheme].SetValue(theme)
	}
}

func WithWidth(width int) tiOpt {
	return func(ti tiRow) {
		ti.SetWidth(width)
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

	names := []string{"Scene desc", "Tempo", "Keywords", "Theme", "Duration (sec)"}
	header := make([]string, len(names))
	for i, name := range names {
		width := 20
		if len(it.iRows) > 0 {
			width = it.iRows[0][i].Width
		}
		header[i] = style.Width(width + 1).Render(name)
	}

	return strings.Repeat(" ", rowPadding) + lipgloss.JoinHorizontal(lipgloss.Top, header...)
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

func (it *InteractiveTable) deleteActiveRow() {
	if it.cursor < len(it.iRows)-1 {
		it.iRows = append(it.iRows[:it.cursor], it.iRows[it.cursor+1:]...)
	} else {
		it.iRows = it.iRows[:it.cursor]
	}

	it.cursor = confine(it.cursor-1, 0, len(it.iRows))
}

func (it *InteractiveTable) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:

		switch msg.String() {
		case "down", "enter":
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
		case "shift+up":
			it.interchangeRows(it.cursor, it.cursor-1)
		case "shift+down":
			it.interchangeRows(it.cursor, it.cursor+1)
		}

		// Ensure correct row is activated
		if r := it.activeTiRow(); r != nil {
			r.Update(msg)
		}
	case tea.WindowSizeMsg:
		for i := range it.iRows {
			it.iRows[i].SetWidth(msg.Width)
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
	if it.cursor == len(it.iRows)-1 || len(it.iRows) == 0 {
		it.createNewRow()
	}
}

func (it *InteractiveTable) interchangeRows(current, target int) {
	target = confine(target, 0, len(it.iRows)-1)
	it.blurActiveRow()
	it.iRows[current], it.iRows[target] = it.iRows[target], it.iRows[current]
	it.cursor = target
	it.activateCurrentRow()
}

func (it *InteractiveTable) createNewRow() {
	slog.Info("Creating new row")
	newRow := NewTiRow()
	it.iRows = append(it.iRows, newRow)
}

func (it *InteractiveTable) toRecords(projectId uint) ([]db.ProjectContentRecord, error) {
	rows := make([]db.ProjectContentRecord, len(it.iRows))
	for i, row := range it.iRows {
		var (
			duration int
			tempo    int
			theme    int
			ierr     error
		)

		err := utils.ReturnFirstError(
			func() error {
				duration, ierr = row.DurationOrDefault()
				return ierr
			},
			func() error {
				tempo, ierr = row.TempoOrDefault()
				return ierr
			},
			func() error {
				theme, ierr = row.ThemeOrDefault()
				return ierr
			},
		)
		if err != nil {
			return rows, err
		}

		rows[i] = db.ProjectContentRecord{
			ProjectID:   projectId,
			Scene:       uint(i),
			SceneDesc:   row[tiScene].Value(),
			DurationSec: duration,
			Keywords:    row[tiKeywords].Value(),
			Tempo:       uint(tempo),
			Theme:       uint(theme),
		}
	}
	return rows, nil
}

type ProjectWorkspace struct {
	store        db.ProjectStore
	project      *db.Project
	status       *Status
	iTable       *InteractiveTable
	library      compose.Library
	creator      musicxml.Creator
	initialWidth int
}

func (pw *ProjectWorkspace) Init() tea.Cmd {
	pw.status = NewStatus()
	pw.iTable = NewInteractiveTable()

	for _, record := range pw.project.Records {
		row := NewTiRowFromRecord(&record)
		row.SetWidth(pw.initialWidth)
		row.Blur()
		pw.iTable.iRows = append(pw.iTable.iRows, row)
	}
	pw.iTable.activateCurrentRow()
	return nil
}

func (pw *ProjectWorkspace) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch keyName := msg.String(); keyName {
		case "enter":
			pw.status.Set("OK", nil)
			if err := pw.validate(); err != nil {
				pw.status.Set("", err)
			}
		case "ctrl+s", "esc":
			err := pw.save()
			pw.status.Set("Successfully stored project", err)
			if keyName == "esc" {
				return pw, func() tea.Msg {
					return toProjectOverview{}
				}
			}
		case "delete":
			pw.iTable.deleteActiveRow()

			if err := pw.save(); err != nil {
				pw.status.Set("", err)
			} else {
				pw.status.Set(fmt.Sprintf("Successfully deleted schene %d", pw.iTable.cursor), err)
			}
		case "ctrl+g":
			if err := pw.save(); err != nil {
				pw.status.Set("", err)
				break
			}

			score := compose.CreateComposition(pw.library, pw.project)
			fname := musicxml.FileNameFromScore(score)
			err := musicxml.WriteScoreToFile(pw.creator, fname, score)
			pw.status.Set(fmt.Sprintf("Successfully stored compiled score to %s", fname), err)
		}
	}
	pw.iTable.Update(msg)
	return pw, nil
}

func (pw *ProjectWorkspace) View() string {

	helpString := helpStyle.Render("\u2191/\u2193 up/down \u2022 \u2190/\u2192 left/right \u2022 shift+(\u2191/\u2193) move row up/down \u2022 ctrl+g: generate score \u2022 ctrl+c: quit")
	return lipgloss.JoinVertical(lipgloss.Left, pw.iTable.View(), helpString, pw.status.Render("Edit"))
}

func (pw *ProjectWorkspace) validate() error {
	for _, item := range pw.iTable.iRows {
		err := utils.ReturnFirstError(
			func() error { return validateDuration(item.Duration()) },
			func() error { return validateTempo(item.Tempo()) },
		)

		if err != nil {
			return err
		}

	}
	return nil
}

func (pw *ProjectWorkspace) save() error {
	records, err := pw.iTable.toRecords(pw.project.Id)
	if err != nil {
		return err
	}
	pw.project.Records = records
	return pw.store.Save(pw.project)
}

func validateDuration(duration string) error {
	if duration == "" {
		return nil
	}
	_, err := strconv.Atoi(duration)
	if err != nil {
		return ErrDurationMustBeInteger
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

func intOrDefault(value string, defaultValue int) (int, error) {
	if value != "" {
		return strconv.Atoi(value)
	}
	return defaultValue, nil
}
