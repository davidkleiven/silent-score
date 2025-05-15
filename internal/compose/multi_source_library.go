package compose

import "github.com/davidkleiven/silent-score/internal/musicxml"

type MultiSourceLibrary struct {
	libraries []Library
}

func (m *MultiSourceLibrary) BestMatch(desc string) matchResult {
	var best *musicxml.Scorepartwise
	bestScore := 0
	for _, lib := range m.libraries {
		result := lib.BestMatch(desc)
		if best == nil || result.similarity > bestScore {
			best = result.score
			bestScore = result.similarity
		}
	}
	return matchResult{score: best, similarity: bestScore}
}

func NewMultiSourceLibrary(libraries ...Library) *MultiSourceLibrary {
	return &MultiSourceLibrary{
		libraries: libraries,
	}
}
