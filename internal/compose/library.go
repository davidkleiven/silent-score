package compose

import (
	"strings"

	"github.com/davidkleiven/silent-score/internal/musicxml"
)

type Library interface {
	BestMatch(desc string) *musicxml.Scorepartwise
}

type InMemoryLibrary struct {
	scores []*musicxml.Scorepartwise
}

func (l *InMemoryLibrary) BestMatch(desc string) *musicxml.Scorepartwise {
	texts := make([]string, 0, len(l.scores))
	for _, score := range l.scores {
		texts = append(texts, strings.Join(musicxml.TextFields(*score), " "))
	}
	normalizedDesc := normalize(desc)
	normalizedTexts := make([]string, len(texts))
	for i, text := range texts {
		normalizedTexts[i] = normalize(text)
	}
	bestMatch := orderPieces(normalizedDesc, normalizedTexts)
	return l.scores[bestMatch[0].Index]
}
