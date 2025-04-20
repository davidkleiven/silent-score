package musicxml

import "testing"

func TestWithRehersalMark(t *testing.T) {
	measure := NewMeasure(WithRehersalMark("A"))

	if measure.MusicDataElements[0].Direction.Directiontype[0].Rehearsal[0].Value != "A" {
		t.Errorf("Did not find a rehersal mark")
	}
}

func TestDirection(t *testing.T) {
	direction := NewDirection(WithTempo(127))
	if tempo := direction.Directiontype[0].Metronome.Perminute.Value; tempo != 127 {
		t.Errorf("Wanted 127 got %d", tempo)
	}
}

func TestDeepCopyMessage(t *testing.T) {
	m := Measure{}
	if MustDeepCopyMeasure(&m) == &m {
		t.Errorf("Measure was not deepcopied")
	}
}

func TestNewScorePartwise(t *testing.T) {
	score := NewScorePartwise(WithComposer("Beethoven"))
	if score == nil {
		t.Errorf("Scorepartwise was not created")
		return
	}

	if len(score.Credit) != 1 {
		t.Errorf("Composer credit was not created")
		return
	}

	if score.Credit[0].Credittype[0] != "composer" {
		t.Errorf("Composer credit was not created")
	}
	if score.Credit[0].Creditwords.Value != "Beethoven" {
		t.Errorf("Composer credit was not created")
	}
}

func TestTitleElement(t *testing.T) {
	title := TitleElement("Test Title")
	if title.Value != "Test Title" {
		t.Errorf("Title element was not created correctly")
	}
}

func TestDefaultPageMaringsNotNil(t *testing.T) {
	pageMargins := DefaultPageMargins("even")
	if pageMargins == nil {
		t.Errorf("Default page margins should not be nil")
		return
	}

	if pageMargins.TypeAttr != "even" {
		t.Errorf("Default page margins should have type 'even'")
		return
	}
}
