package main

import (
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/davidkleiven/silent-score/internal/db"
	"github.com/davidkleiven/silent-score/internal/ui"
)

func main() {
	config := ui.NewConfig()
	os.Remove(config.LogFile)
	f, err := tea.LogToFile(config.LogFile, "")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	programDb, err := db.GormConnection(config.DbName)
	if err != nil {
		log.Fatal(err)
	}
	if err = db.AutoMigrate(programDb); err != nil {
		log.Fatal(err)
	}

	model := ui.NewAppModel(&db.GormStore{Database: programDb})
	program := tea.NewProgram(model)

	if _, err := program.Run(); err != nil {
		log.Fatal(err)
	}
}
