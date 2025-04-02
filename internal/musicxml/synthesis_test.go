package musicxml

import (
	"testing"

	"pgregory.net/rapid"
)

func TestSynthesisWork(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		score := GenerateRandomScorepartwise(t)
		if score == nil {
			t.Error("No score generated")
		}
	})
}
