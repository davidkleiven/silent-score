package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/davidkleiven/silent-score/internal/db"
	"gorm.io/gorm"
)

func mustConnectInMemDb() *gorm.DB {
	database, err := db.InMemoryGormConnection()
	if err != nil {
		panic(err)
	}
	return database
}

func TestInteravtoveTableEvent(t *testing.T) {
	for _, test := range []struct {
		msg   tea.Msg
		check func(*testing.T, *InteractiveTable)
		desc  string
	}{
		{
			msg: tea.KeyDown,
			check: func(t *testing.T, it *InteractiveTable) {
				if it.table.Cursor() != 0 {
					t.Errorf("Cursor should be 0")
				}

				if len(it.table.Rows()) != len(it.iRows) {
					t.Errorf("Inconsistent rows %d and %d", len(it.table.Rows()), len(it.iRows))
				}
			},
		},
	} {
		it := NewInteractiveTable()
		it.Init()
		it.Update(test.msg)
		t.Run(test.desc, func(t *testing.T) { test.check(t, it) })
	}
}
