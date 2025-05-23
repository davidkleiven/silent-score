package ui

import "github.com/davidkleiven/silent-score/internal/db"

type toProjectOverview struct{}
type toProjectWorkspace struct {
	project *db.Project
}

type toLibraryList struct{}
type toLibraryContent struct{}
