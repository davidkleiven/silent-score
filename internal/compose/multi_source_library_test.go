package compose

import (
	"testing"

	"github.com/davidkleiven/silent-score/internal/musicxml"
)

func TestMultiLibraryBestMatch(t *testing.T) {
	score1 := musicxml.NewScorePartwise(musicxml.WithComposer("Beethoven"))
	score2 := musicxml.NewScorePartwise(musicxml.WithComposer("Bach"))

	library := MultiSourceLibrary{
		libraries: []Library{
			&InMemoryLibrary{
				scores: []*musicxml.Scorepartwise{score1},
			},
			&InMemoryLibrary{
				scores: []*musicxml.Scorepartwise{score2},
			},
		},
	}

	bm := library.BestMatch("Bach")
	if bm.Scoreheader.Credit[0].Creditwords.Value != "Bach" {
		t.Errorf("Expected 'Bach', got '%s'", bm.Scoreheader.Credit[0].Creditwords.Value)
	}
}
