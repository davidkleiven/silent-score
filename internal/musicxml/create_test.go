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

func TestWithEnding(t *testing.T) {
	for i, endNum := range []EndingType{EndingTypeStart, EndingTypeStop, EndingTypeDiscontinue} {
		measure := NewMeasure(WithBarline(NewBarline(WithEnding(NewEnding(WithEndingNumber(i), WithEndingType(endNum))))))
		if measure.MusicDataElements[0].Barline == nil {
			t.Errorf("Test #%d: Should have a barline", i)
			return
		}

		if measure.MusicDataElements[0].Barline.Ending == nil {
			t.Errorf("Test #%d: Should have a ending", i)
			return
		}
	}
}

func TestNewPage(t *testing.T) {
	n := NewPage()
	if n.Print == nil {
		t.Errorf("Page should have a print element")
		return
	}
}

func TestNewSystem(t *testing.T) {
	n := NewSystem()
	if n.Print == nil {
		t.Errorf("System should have a print element")
		return
	}
}

func TestWithRepeat(t *testing.T) {
	b := NewBarline(WithRepeat(&Repeat{}))
	if b.Repeat == nil {
		t.Errorf("Barline should have a repeat element")
		return
	}
}

func TestWithPrint(t *testing.T) {
	p := NewMeasure(WithPrint(&Print{}))

	hasPrint := false
	for _, element := range p.MusicDataElements {
		if element.Print != nil {
			hasPrint = true
			break
		}
	}
	if !hasPrint {
		t.Errorf("Measure should have a print element")
	}
}

func TestWithBarStyle(t *testing.T) {
	for i, style := range []BarStyle{BarStyleDashed, BarStyleDotted, BarStyleHeavy, BarStyleHeavyHeavy, BarStyleShort, BarStyleHeavyLight, BarStyleLightLight, BarStyleTick, BarStyleLightHeavy, BarStyleRegular} {
		barline := NewBarline(WithBarStyle(style))
		if barline.Barstyle.Value != style.String() {
			t.Errorf("Test #%d: Barline should have a barstyle %s", i, style.String())
			return
		}
	}
}

func TestWithPart(t *testing.T) {
	score := NewScorePartwise(WithPart(Part{}))
	if len(score.Part) != 1 {
		t.Errorf("Scorepartwise should have a part")
		return
	}
}
