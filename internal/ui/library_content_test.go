package ui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/davidkleiven/silent-score/internal/compose"
	"github.com/davidkleiven/silent-score/internal/musicxml"
)

func TestToProjectOverview(t *testing.T) {
	content := LibraryContentView{
		lib: compose.NewStandardLibrary(),
	}

	content.Init()
	_, cmd := content.Update(tea.KeyMsg{Type: tea.KeyEsc})

	if cmd == nil {
		t.Errorf("Expected a command, got nil")
		return
	}

	result, ok := cmd().(tea.BatchMsg)
	if !ok {
		t.Errorf("Expected tea.Cmd, got %T", cmd())
		return
	}

	hasToProjectOverview := false
	for _, msg := range result {
		if _, ok := msg().(toProjectOverview); ok {
			hasToProjectOverview = true
			break
		}
	}
	if !hasToProjectOverview {
		t.Errorf("Expected toProjectOverview message, got %v", result)
		return
	}
}

func TestDoNotGoToProjectOverviewWhenFiltering(t *testing.T) {
	view := LibraryContentView{
		lib: compose.NewStandardLibrary(),
	}

	view.Init()
	view.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")})

	if view.content.FilterState() != list.Filtering {
		t.Errorf("Expected filtering state, got %v", view.content.FilterState())
		return
	}

	_, cmd := view.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if cmd != nil {
		t.Errorf("Expected no command, got %v", cmd)
		return
	}
}

func TestView(t *testing.T) {
	content := LibraryContentView{
		lib: &compose.InMemoryLibrary{
			Scores: []*musicxml.Scorepartwise{
				musicxml.NewScorePartwise(musicxml.WithComposer("Chopin")),
				musicxml.NewScorePartwise(musicxml.WithComposer("Bach")),
				musicxml.NewScorePartwise(musicxml.WithComposer("Beethoven")),
				musicxml.NewScorePartwise(musicxml.WithComposer("Beethoven")),
			},
		},
		width:  80,
		height: 80,
	}

	content.Init()
	result := content.View()

	for _, substr := range []string{"Bach", "Beethoven", "Chopin"} {
		if !strings.Contains(result, substr) {
			t.Errorf("Expected %s in view, got %s", substr, result)
			return
		}
	}
}

func TestWindowResizePassedToList(t *testing.T) {
	content := LibraryContentView{
		lib:    &compose.InMemoryLibrary{},
		width:  80,
		height: 80,
	}

	content.Init()
	content.Update(tea.WindowSizeMsg{Width: 100, Height: 100})

	if content.content.Width() != 100 {
		t.Errorf("Expected width 100, got %d", content.content.Width())
		return
	}
	if content.content.Height() != 99 {
		t.Errorf("Expected height %d, got %d", 99, content.content.Height())
		return
	}
}
