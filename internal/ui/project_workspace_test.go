package ui

import (
	"errors"
	"os"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/davidkleiven/silent-score/internal/compose"
	"github.com/davidkleiven/silent-score/internal/db"
	"github.com/davidkleiven/silent-score/internal/musicxml"
	"github.com/davidkleiven/silent-score/test"
	"pgregory.net/rapid"
)

func initializedPw() ProjectWorkspace {
	pw := ProjectWorkspace{store: db.NewInMemoryStore(), project: db.NewProject(db.WithName("my-project"))}
	pw.Init()
	return pw
}

func numFocused(row tiRow) int {
	num := 0
	for _, item := range row {
		if item.Focused() {
			num += 1
		}
	}
	return num
}

func TestInteravtoveTableEvent(t *testing.T) {
	for _, test := range []struct {
		messages []tea.Msg
		check    func(*testing.T, *InteractiveTable)
		desc     string
	}{
		{
			messages: []tea.Msg{tea.KeyMsg{Type: tea.KeyDown}},
			check: func(t *testing.T, it *InteractiveTable) {
				if it.cursor != 0 {
					t.Errorf("Cursor should be 0")
				}

				if len(it.iRows) != 1 {
					t.Errorf("One row should have been created")
				}
			},
			desc: "Press arrow down from empty states",
		},
		{
			messages: []tea.Msg{tea.KeyMsg{Type: tea.KeyDown}, tea.KeyMsg{Type: tea.KeyDown}},
			check: func(t *testing.T, it *InteractiveTable) {
				if it.cursor != 1 {
					t.Errorf("Cursor should be 1")
				}

				if len(it.iRows) != 2 {
					t.Errorf("One row should have been created")
				}

				// No element in row 0 should be focused
				for i, item := range it.iRows[0] {
					if item.Focused() {
						t.Errorf("%d-th item in row 0 is focused", i)
					}
				}

				if num := numFocused(it.iRows[1]); num != 1 {
					t.Errorf("Wanted 1 element to be focused got %d", num)
				}
			},
			desc: "Press two arrows down",
		},
		{
			messages: []tea.Msg{tea.KeyMsg{Type: tea.KeyDown}, tea.KeyMsg{Type: tea.KeyDown}, tea.KeyMsg{Type: tea.KeyUp}},
			check: func(t *testing.T, it *InteractiveTable) {
				if it.cursor != 0 {
					t.Errorf("Row 0 should be active. Got %d", it.cursor)
				}
			},
			desc: "Two down,one up",
		},
		{
			messages: []tea.Msg{tea.KeyMsg{Type: tea.KeyLeft}, tea.KeyMsg{Type: tea.KeyRight}},
			check: func(t *testing.T, it *InteractiveTable) {
				if it.cursor != 0 {
					t.Errorf("Row 0 should be active. Got %d", it.cursor)
				}
			},
			desc: "No error when pressing left/right with no rows",
		},
		{
			messages: []tea.Msg{tea.KeyMsg{Type: tea.KeyDown}, tea.KeyMsg{Type: tea.KeyRight}},
			check: func(t *testing.T, it *InteractiveTable) {
				if it.iRows[it.cursor].Active() != 1 {
					t.Errorf("Active field should have been shifted one to the left")
				}
			},
			desc: "Shift active field when pressing arrow left",
		},
		{
			messages: []tea.Msg{tea.KeyMsg{Type: tea.KeyDown}, tea.KeyMsg{Type: tea.KeyRight}, tea.KeyMsg{Type: tea.KeyRight}, tea.KeyMsg{Type: tea.KeyLeft}},
			check: func(t *testing.T, it *InteractiveTable) {
				if it.iRows[it.cursor].Active() != 1 {
					t.Errorf("Active field should have been shifted one to the left")
				}
			},
			desc: "Shift active field back to 1 when pressing 2 rights and one left",
		},
	} {
		it := NewInteractiveTable()
		it.Init()
		for _, msg := range test.messages {
			it.Update(msg)
		}
		t.Run(test.desc, func(t *testing.T) { test.check(t, it) })
		it.View() // Confirm no errors when calling view
	}
}

func TestValidateTime(t *testing.T) {
	for _, test := range []struct {
		timestamp string
		err       error
		desc      string
	}{
		{
			timestamp: "0",
			err:       nil,
			desc:      "Valid timestamp only minutes",
		},
		{
			timestamp: "",
			err:       nil,
			desc:      "Empty",
		},
		{
			timestamp: "02:02",
			err:       ErrDurationMustBeInteger,
			desc:      "It is a timestamp and not int",
		},
	} {
		t.Run(test.desc, func(t *testing.T) {
			if err := validateDuration(test.timestamp); err != test.err {
				t.Errorf("Wanted %v got %v", test.err, err)
			}
		})
	}
}

func TestValidateTempo(t *testing.T) {
	for _, test := range []struct {
		tempo string
		err   error
		desc  string
	}{
		{
			tempo: "88",
			err:   nil,
			desc:  "Valid tempo",
		},
		{
			tempo: "8a",
			err:   ErrTempoMustBeInteger,
			desc:  "Invalid tempo",
		},
		{
			tempo: "",
			err:   nil,
			desc:  "Empty tempo",
		},
	} {
		t.Run(test.desc, func(t *testing.T) {
			if err := validateTempo(test.tempo); err != test.err {
				t.Errorf("Wanted %v got %v", test.err, err)
			}
		})
	}
}

func TestBlur(t *testing.T) {
	row := NewTiRow()
	row[1].Focus()

	if n := numFocused(row); n != 2 {
		t.Errorf("2 item should be focused got %d", n)
	}
	row.Blur()

	if n := numFocused(row); n != 0 {
		t.Errorf("0 items should be focused got %d", n)
	}

}

func TestConfine(t *testing.T) {
	for i, test := range []struct {
		num  int
		want int
	}{
		{
			num:  2,
			want: 2,
		},
		{
			num:  0,
			want: 1,
		},
		{
			num:  6,
			want: 4,
		},
	} {
		if n := confine(test.num, 1, 4); n != test.want {
			t.Errorf("Test %d: wanted %d got %d", i, test.want, n)
		}
	}
}

func TestActive(t *testing.T) {
	row := NewTiRow()
	row.Blur()

	// Active should be zero when there are no active
	if row.Active() != 0 {
		t.Errorf("Active should be zero when no row is active")
	}

	row[2].Focus()
	if a := row.Active(); a != 2 {
		t.Errorf("Wanted 2 got %d", a)
	}
}

func TestFocusRL(t *testing.T) {
	r := NewTiRow()
	for i, test := range []struct {
		direction string
		want      int
		desc      string
	}{
		{
			direction: "left",
			want:      0,
			desc:      "Left from position 0",
		},
		{
			direction: "right",
			want:      1,
			desc:      "Right",
		},
	} {
		switch test.direction {
		case "left":
			r.FocusLeft()
		case "right":
			r.FocusRight()
		default:
			t.Errorf("Unknown option")
		}

		if a := r.Active(); a != test.want {
			t.Errorf("Step %d: wanted %d got %d", i, test.want, a)
		}
	}
}

func TestTimeConstency(t *testing.T) {
	row := NewTiRow(WithDuration("20:00:00"))
	if timefield := row.Duration(); timefield != "20:00:00" {
		t.Errorf("Wanted 20:00:00 got %s", timefield)
	}
}

func TestProjectWorkspaceEnterOk(t *testing.T) {
	pw := initializedPw()
	pw.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if pw.status.msg != "OK" {
		t.Errorf("Status should be OK got %s", pw.status.msg)
	}
	pw.View()
}

func TestProjectWorkspaceEnterErr(t *testing.T) {
	pw := initializedPw()
	pw.iTable.createNewRow()
	pw.iTable.iRows[0][tiDuration].SetValue("01:00")
	pw.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if pw.status.msg != ErrDurationMustBeInteger.Error() {
		t.Errorf("Wanted %s got %s", ErrDurationMustBeInteger.Error(), pw.status.msg)
	}
}

func TestValidate(t *testing.T) {
	for i, test := range []struct {
		row tiRow
		err error
	}{
		{
			row: NewTiRow(),
			err: nil,
		},
		{
			row: NewTiRow(WithDuration("00:01")),
			err: ErrDurationMustBeInteger,
		},
		{
			row: NewTiRow(WithTempo("andante")),
			err: ErrTempoMustBeInteger,
		},
	} {
		pw := initializedPw()
		pw.iTable.iRows = append(pw.iTable.iRows, test.row)
		if err := pw.validate(); err != test.err {
			t.Errorf("%d: wanted %v got %v", i, test.err, err)
		}
	}
}

func TestInterchange(t *testing.T) {

	for _, test := range []struct {
		want    []string
		current int
		target  int
		desc    string
	}{
		{
			want:    []string{"2", "1"},
			current: 1,
			target:  0,
			desc:    "Move last row up",
		},
		{
			want:    []string{"2", "1"},
			current: 0,
			target:  1,
			desc:    "Move last row down",
		},
	} {
		t.Run(test.desc, func(t *testing.T) {
			pw := initializedPw()
			pw.iTable.iRows = append(pw.iTable.iRows, NewTiRow(WithDuration("1")))
			pw.iTable.iRows = append(pw.iTable.iRows, NewTiRow(WithDuration("2")))
			pw.iTable.interchangeRows(test.current, test.target)

			for i, row := range pw.iTable.iRows {
				if v := row[tiDuration].Value(); v != test.want[i] {
					t.Errorf("Wanted %s got %s", test.want[i], v)
				}
			}
		})
	}
}

func TestShiftKeyUpDown(t *testing.T) {
	pw := initializedPw()
	pw.iTable.iRows = append(pw.iTable.iRows, NewTiRow(WithTempo("10")))
	pw.iTable.iRows = append(pw.iTable.iRows, NewTiRow(WithTempo("12")))

	pw.iTable.Update(tea.KeyMsg{Type: tea.KeyShiftDown})
	want := []string{"12", "10"}
	for i, row := range pw.iTable.iRows {
		if v := row[tiTempo].Value(); v != want[i] {
			t.Errorf("Wanted %s got %s", want[i], v)
		}
	}

	if pw.iTable.cursor != 1 {
		t.Errorf("Cursor should have been on last row")
	}

	// Reverse operation
	pw.iTable.Update(tea.KeyMsg{Type: tea.KeyShiftUp})
	want = []string{"10", "12"}
	for i, row := range pw.iTable.iRows {
		if v := row[tiTempo].Value(); v != want[i] {
			t.Errorf("Wanted %s got %s", want[i], v)
		}
	}

	if pw.iTable.cursor != 0 {
		t.Error("Cursor should have been back at zero")
	}

}

func TestIntOrDefault(t *testing.T) {
	for _, test := range []struct {
		value        string
		defaultValue int
		shouldFail   bool
		want         int
	}{
		{
			value:        "71",
			defaultValue: 0,
			shouldFail:   false,
			want:         71,
		},
		{
			value:      "89d",
			shouldFail: true,
		},
		{
			value:        "",
			defaultValue: 1,
			shouldFail:   false,
			want:         1,
		},
	} {
		{
			t.Run(test.value, func(t *testing.T) {
				result, err := intOrDefault(test.value, test.defaultValue)
				if !test.shouldFail {
					if result != test.want {
						t.Errorf("Wanted %d got %d", test.want, result)
					}
				} else {
					if err == nil {
						t.Errorf("Should result in an error got %d %v", result, err)
					}
				}
			})
		}
	}
}

func TestToRecords(t *testing.T) {

	for _, test := range []struct {
		rows       []tiRow
		shouldFail bool
		desc       string
		want       []db.ProjectContentRecord
	}{
		{
			rows:       []tiRow{NewTiRow(WithDuration("12:00"))},
			shouldFail: true,
			desc:       "Wrong time",
		},
		{
			rows:       []tiRow{NewTiRow(WithTempo("Andante"))},
			shouldFail: true,
			desc:       "Wrong tempo (should be integer)",
		},
		{
			rows:       []tiRow{NewTiRow(WithTheme("A"))},
			shouldFail: true,
			desc:       "Wrong tempo (should be integer)",
		},
		{
			rows:       []tiRow{NewTiRow(WithDuration("2"), WithTempo("88"), WithTheme("0"))},
			shouldFail: false,
			desc:       "Valid row",
		},
	} {
		table := NewInteractiveTable()
		table.Init()
		t.Run(test.desc, func(t *testing.T) {
			table.iRows = append(table.iRows, test.rows...)
			_, err := table.toRecords(8)
			if test.shouldFail && err == nil {
				t.Errorf("Should result in error")
			}

			if !test.shouldFail && err != nil {
				t.Errorf("Error was not nil even though rows are valid got %v", err)
			}
		})
	}
}

func TestCtrlS(t *testing.T) {
	for _, test := range []struct {
		row            tiRow
		desc           string
		wantRows       int
		wantStatusKind StatusKind
	}{
		{
			row:            NewTiRow(WithDuration("15")),
			desc:           "One valid row",
			wantRows:       1,
			wantStatusKind: okStatus,
		},
		{
			row:            NewTiRow(WithDuration("00:15")),
			desc:           "One row with wrong time format",
			wantRows:       0,
			wantStatusKind: errorStatus,
		},
	} {
		t.Run(test.desc, func(t *testing.T) {
			pw := ProjectWorkspace{
				store:   db.NewInMemoryStore(),
				project: db.NewProject(db.WithName("my-project")),
			}
			pw.Init()

			if err := pw.save(); err != nil {
				t.Error(err)
			}
			pw.iTable.iRows = append(pw.iTable.iRows, test.row)
			pw.Update(tea.KeyMsg{Type: tea.KeyCtrlS})

			items, err := pw.store.Load()
			if err != nil {
				t.Error(err)
			}
			if len(items[0].Records) != test.wantRows {
				t.Errorf("Expected one record to be stored in the database")
			}

			if pw.status.kind != test.wantStatusKind {
				t.Errorf("Wanted %d got %d (%s)", test.wantStatusKind, pw.status.kind, pw.status.msg)
			}
		})
	}
}

func TestEsc(t *testing.T) {
	pw := initializedPw()
	pw.store = db.NewInMemoryStore()
	model, _ := pw.Update(tea.KeyMsg{Type: tea.KeyEsc})

	switch model.(type) {
	case *ProjectOverviewModel:
	default:
		t.Error("Wanted project overview model")
	}
}

func TestToFromProjectRecordRoundTrip(t *testing.T) {
	gen := rapid.Custom(func(t *rapid.T) db.ProjectContentRecord {
		return test.GenerateProjectContentRecord(t, 1)
	})

	projectId := uint(1)
	rapid.Check(t, func(t *rapid.T) {
		records := rapid.SliceOfN(gen, 0, 3).Draw(t, "rec")
		tiRows := make([]tiRow, len(records))
		for i := range records {
			records[i].Scene = uint(i)
			records[i].ProjectID = projectId
			tiRow := NewTiRowFromRecord(&records[i])
			tiRows[i] = tiRow
		}

		table := NewInteractiveTable()
		table.iRows = append(table.iRows, tiRows...)
		generatedRecords, err := table.toRecords(projectId)
		if err != nil {
			t.Error(err)
			return
		}

		for i := range records {
			r, g := records[i], generatedRecords[i]
			if (r.Keywords != g.Keywords) ||
				(r.Scene != g.Scene) ||
				(r.SceneDesc != g.SceneDesc) ||
				(r.Tempo != g.Tempo) ||
				(r.Theme != g.Theme) {
				t.Errorf("Wanted\n%+v\ngot\n%+v", r, g)
				return
			}

		}

	})
}

func TestInitializedWithCorrectRows(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		projects := test.GenerateProjects(t)
		pw, err := initProjectFromWithRecords(projects)
		if err != nil {
			t.Error(err)
		}
		numRows := len(projects[0].Records)

		if len(pw.iTable.iRows) != numRows {
			t.Errorf("Wanted %d rows got %d", numRows, len(pw.iTable.iRows))
		}
	})
}

func TestDeleteRecord(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		projects := test.GenerateProjects(t)
		pw, err := initProjectFromWithRecords(projects)
		if err != nil {
			t.Error(err)
			return
		}

		origNumRows := len(pw.iTable.iRows)
		maxCursor := 0
		if origNumRows > 0 {
			maxCursor = origNumRows - 1
		}

		toDelete := rapid.IntRange(0, maxCursor).Draw(t, "to-draw")
		pw.iTable.cursor = toDelete
		pw.Update(tea.KeyMsg{Type: tea.KeyDelete})
		finalNumRows := len(pw.iTable.iRows)

		if len(projects[0].Records) > 0 && finalNumRows != origNumRows-1 {
			t.Errorf("Wanted number of rows to be %d got %d", origNumRows-1, finalNumRows)
		}
	})
}

func TestSaveFailsDuringDelete(t *testing.T) {
	for i, keyType := range []tea.KeyType{tea.KeyDelete, tea.KeyCtrlG} {
		pw := ProjectWorkspace{
			store:   &failingSaver{},
			project: db.NewProject(db.WithName("my-project")),
		}
		pw.Init()
		pw.Update(tea.KeyMsg{Type: keyType})
		if pw.status.kind != errorStatus {
			t.Errorf("Test #%d: Should have been in error status", i)
		}
	}
}

func initProjectFromWithRecords(projects []db.Project) (ProjectWorkspace, error) {
	database := db.NewInMemoryStore()
	for _, p := range projects {
		if err := database.Save(&p); err != nil {
			return ProjectWorkspace{}, err
		}
	}

	pw := ProjectWorkspace{store: database, project: &projects[0]}
	pw.Init()
	return pw, nil
}

type failingSaver struct{}

func (fs *failingSaver) Save(project *db.Project) error {
	return errors.New("saving failed")
}

func (fs *failingSaver) Load() ([]db.Project, error) {
	return []db.Project{}, nil
}

func (fs *failingSaver) Delete(id uint) error {
	return nil
}

type customFileCreator struct {
	name string
}

func (c *customFileCreator) Create(name string) (musicxml.WriterCloser, error) {
	_ = name
	return os.Create(c.name)
}

func TestGenerateScore(t *testing.T) {
	tmpFile := t.TempDir() + "/test.xml"
	pw := ProjectWorkspace{
		store:   db.NewInMemoryStore(),
		project: db.NewProject(db.WithName("my-project")),
		library: compose.NewStandardLibrary(),
		creator: &customFileCreator{name: tmpFile},
	}
	pw.Init()

	pw.Update(tea.KeyMsg{Type: tea.KeyCtrlG})

	content, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Error(err)
		return
	}
	if !strings.Contains(string(content), "<score-partwise>") {
		t.Errorf("Wanted to find <score-piecewise> in file %s", string(content))
	}
}
