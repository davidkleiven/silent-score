package compose

import (
	"math"
	"testing"
	"time"

	"github.com/davidkleiven/silent-score/internal/musicxml"
)

func eightBarPiece() []musicxml.Measure {
	measures := []musicxml.Measure{
		*musicxml.NewMeasure(),
		*musicxml.NewMeasure(musicxml.WithRehersalMark("A")),
		*musicxml.NewMeasure(),
		*musicxml.NewMeasure(musicxml.WithRehersalMark("B")),
		*musicxml.NewMeasure(),
		*musicxml.NewMeasure(),
		*musicxml.NewMeasure(),
		*musicxml.NewMeasure(),
	}
	enumerateMeasuresInPlace(measures)
	return measures
}

func TestPieceSections(t *testing.T) {
	measures := eightBarPiece()
	sections := pieceSections(measures)
	want := []section{
		{
			start: 0,
			end:   1,
		},
		{
			start: 1,
			end:   3,
		},
		{
			start: 3,
			end:   8,
		},
	}

	if len(sections) != len(want) {
		t.Errorf("Wanted %v got %v\n", sections, want)
	}

	for i := range want {
		if want[i].start != sections[i].start || want[i].end != sections[i].end {
			t.Errorf("Wanted %v got %v\n", want, sections)
			return
		}
	}
}

func TestSectionForScene(t *testing.T) {
	sections := []section{
		{
			start: 0,
			end:   4,
		},
		{
			start: 4,
			end:   16,
		},
	}

	combinedSections := sectionForScene(time.Duration(2*60.0*1e9), 80, 4, sections)
	want := sceneSection{
		sections: []section{{start: 0, end: 4}, {start: 4, end: 16}, {start: 0, end: 4}, {start: 4, end: 16}, {start: 0, end: 4}},
		tempo:    72.0,
	}

	if math.Abs(combinedSections.tempo-want.tempo) > 1e-6 {
		t.Errorf("Wanted %v got %v", want, combinedSections)
		return
	}

	if len(combinedSections.sections) != len(want.sections) {
		t.Errorf("Wanted %v got %v", want, combinedSections)
		return
	}
	for i := range want.sections {
		if want.sections[i].start != combinedSections.sections[i].start || want.sections[i].end != combinedSections.sections[i].end {
			t.Errorf("Wanted %v got %v", want, combinedSections)
			return
		}
	}
}

func TestMeasuresForScene(t *testing.T) {
	bars := eightBarPiece()
	section := sceneSection{
		sections: []section{
			{start: 0, end: 8},
			{start: 0, end: 8},
			{start: 0, end: 8},
			{start: 4, end: 8},
		},
		tempo: 92,
	}
	measures := measuresForScene(bars, section)
	if len(measures) != 28 {
		t.Errorf("Wanted 28 bars got %d", len(measures))
	}
}

func TestMeasuresForSceneEmptyMeasures(t *testing.T) {
	var m []musicxml.Measure
	result := measuresForScene(m, sceneSection{sections: []section{{start: 0, end: 0}}, tempo: 92})
	if len(result) != 0 {
		t.Errorf("Should have empty result")
	}

}
