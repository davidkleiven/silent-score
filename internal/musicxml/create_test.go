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
