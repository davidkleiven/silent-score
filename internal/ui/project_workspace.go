package ui

import (
	"fmt"
	"log/slog"
	"regexp"
	"strconv"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/davidkleiven/silent-score/internal/db"
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
	tiStart
)

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
	row[tiScene].Focus()
	return row
}

func NewTiRowFromRecord(record *db.ProjectContentRecord) tiRow {
	row := NewTiRow()
	if record.Tempo > 0 {
		row[tiTempo].SetValue(fmt.Sprintf("%d", record.Tempo))
	}

	if record.Theme > 0 {
		row[tiTheme].SetValue(fmt.Sprintf("%d", record.Theme))
	}

	row[tiScene].SetValue(record.SceneDesc)
	row[tiKeywords].SetValue(record.Keywords)
	row[tiStart].SetValue(record.Start.Format(time.TimeOnly))
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
	return lipgloss.JoinHorizontal(lipgloss.Top, data...)
}

func (t tiRow) Update(msg tea.Msg) {
	for i, item := range t {
		t[i], _ = item.Update(msg)
	}
}

func (t tiRow) Time() string {
	return t[tiStart].Value()
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

type tiOpt func(row tiRow)

func WithTime(start string) tiOpt {
	return func(ti tiRow) {
		ti[tiStart].SetValue(start)
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
	it.iRows[current][tiStart].SetValue(it.iRows[target][tiStart].Value())
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
			startTime time.Time
			tempo     int
			theme     int
			ierr      error
		)

		err := utils.ReturnFirstError(
			func() error {
				startTime, ierr = time.Parse(time.TimeOnly, row[tiStart].Value())
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
			ProjectID: projectId,
			Scene:     uint(i),
			SceneDesc: row[tiScene].Value(),
			Start:     startTime,
			Keywords:  row[tiKeywords].Value(),
			Tempo:     uint(tempo),
			Theme:     uint(theme),
		}
	}
	return rows, nil
}

type ProjectWorkspace struct {
	database  db.CreateReadUpdateDeleter
	projectId uint
	status    *Status
	iTable    *InteractiveTable
}

func (pw *ProjectWorkspace) Init() tea.Cmd {
	pw.status = NewStatus()
	pw.iTable = NewInteractiveTable()

	var records []db.ProjectContentRecord
	tx := pw.database.Find(&records, "project_id = ?", pw.projectId)
	if tx.Error != nil {
		panic(tx.Error)
	}
	for _, record := range records {
		row := NewTiRowFromRecord(&record)
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
				return &ProjectOverviewModel{db: pw.database}, nil
			}
		case "delete":
			if err := pw.save(); err != nil {
				pw.status.Set("", err)
			} else {
				err := db.DeleteRecords(pw.database, pw.projectId, uint(pw.iTable.cursor))
				pw.status.Set(fmt.Sprintf("Successfully deleted schene %d", pw.iTable.cursor), err)
				pw.Init()
			}
		}
	}
	pw.iTable.Update(msg)
	return pw, nil
}

func (pw *ProjectWorkspace) View() string {

	helpString := helpStyle.Render("\u2191/\u2193 up/down \u2022 \u2190/\u2192 left/right \u2022 shift+(\u2191/\u2193) move row up/down \u2022 ctrl+c: quit")
	return lipgloss.JoinVertical(lipgloss.Left, pw.iTable.View(), helpString, pw.status.Render("Edit"))
}

func (pw *ProjectWorkspace) validate() error {
	for _, item := range pw.iTable.iRows {
		err := utils.ReturnFirstError(
			func() error { return validateTime(item.Time()) },
			func() error { return validateTempo(item.Tempo()) },
		)

		if err != nil {
			return err
		}

	}
	return nil
}

func (pw *ProjectWorkspace) save() error {
	records, err := pw.iTable.toRecords(pw.projectId)
	if err != nil {
		return err
	}
	return db.SaveProjectRecords(pw.database, records)
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

func intOrDefault(value string, defaultValue int) (int, error) {
	if value != "" {
		return strconv.Atoi(value)
	}
	return defaultValue, nil
}
