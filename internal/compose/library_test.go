package compose

import (
	"encoding/xml"
	"strconv"
	"testing"

	"github.com/davidkleiven/silent-score/internal/db"
	"github.com/davidkleiven/silent-score/internal/musicxml"
	"github.com/davidkleiven/silent-score/test"
	"pgregory.net/rapid"
)

func measuresAreEnumerated(measures []musicxml.Measure) bool {
	for i, measure := range measures {
		if measure.NumberAttr != strconv.Itoa(i+1) {
			return false
		}
	}
	return true
}

func atLeastOneSceneDescPartOfComposition(measures []musicxml.Measure, records []db.ProjectContentRecord) bool {
	if len(measures) == 0 {
		return true
	}
	for _, measure := range measures {
		for _, element := range measure.MusicDataElements {
			if direction := element.Direction; direction != nil {
				for _, dirType := range direction.Directiontype {
					for _, words := range dirType.Words {
						for _, record := range records {
							if words.Value == record.SceneDesc {
								return true
							}
						}
					}
				}
			}
		}
	}
	return false
}

func TestCompose(t *testing.T) {

	scoreGen := rapid.Custom(func(t *rapid.T) *musicxml.Scorepartwise {
		return musicxml.GenerateRandomScorepartwise(t)
	})

	rapid.Check(t, func(t *rapid.T) {
		scores := rapid.SliceOfN(scoreGen, 1, 3).Draw(t, "scores")
		library := &InMemoryLibrary{scores: scores}
		project := test.GenerateProjects(t)[0]
		composition := CreateComposition(library, &project)

		if !measuresAreEnumerated(composition.Part[0].Measure) {
			t.Errorf("Measures are not enumerated correctly")
		}

		if !atLeastOneSceneDescPartOfComposition(composition.Part[0].Measure, project.Records) {
			t.Errorf("No scene description found in the composition")
		}

		if composition.Work.Worktitle != project.Name {
			t.Errorf("Expected work title %s, got %s", project.Name, composition.Work.Worktitle)
		}
	})
}

func TestStandardLibrary(t *testing.T) {
	sl := NewStandardLibrary()
	names := sl.readNames()
	if len(names) == 0 {
		t.Errorf("Expected non-empty names, got %v", names)
	}
}

func TestStandardLibraryBestMatch(t *testing.T) {
	sl := NewStandardLibrary()
	desc := "Andante Doloroso, No. 70"
	score := sl.BestMatch(desc)
	if score.Work.Worktitle != desc {
		t.Errorf("Expected work title %s, got %s", desc, score.Work.Worktitle)
	}
}

func TestStandardLibraryErrorOnOpen(t *testing.T) {
	sl := StandardLibrary{directory: "invalid/path"}
	names := sl.readNames()
	if len(names) != 0 {
		t.Errorf("Expected empty names, got %v", names)
	}
}

func TestUnitPerBeat(t *testing.T) {
	for _, test := range []struct {
		metronome     *musicxml.Metronome
		timesignature *musicxml.Timesignature
		expectedUnit  int
		expectedNum   int
		desc          string
	}{
		{
			metronome:     &musicxml.Metronome{Beatunit: musicxml.Beatunit{Beatunit: "quarter"}},
			timesignature: &musicxml.Timesignature{Beats: 4, Beattype: 4},
			expectedUnit:  4,
			expectedNum:   1,
			desc:          "4/4 time with quarter note",
		},
		{
			metronome:     &musicxml.Metronome{Beatunit: musicxml.Beatunit{Beatunit: "eighth"}},
			timesignature: &musicxml.Timesignature{Beats: 12, Beattype: 8},
			expectedUnit:  8,
			expectedNum:   1,
			desc:          "12/8 one beat per 8th note",
		},
		{
			metronome:     &musicxml.Metronome{Beatunit: musicxml.Beatunit{Beatunit: "quarter", Beatunitdot: []musicxml.Empty{{}}}},
			timesignature: &musicxml.Timesignature{Beats: 12, Beattype: 8},
			expectedUnit:  8,
			expectedNum:   3,
			desc:          "12/8 one beat per 8th note",
		},
	} {
		t.Run(test.desc, func(t *testing.T) {
			unit, num := unitsPerBeat(test.metronome)
			if unit != test.expectedUnit {
				t.Errorf("Expected unit %d, got %d", test.expectedUnit, unit)
			}
			if num != test.expectedNum {
				t.Errorf("Expected num %d, got %d", test.expectedNum, num)
			}
		})
	}
}

func TestComposer(t *testing.T) {
	score := musicxml.NewScorePartwise(musicxml.WithComposer("Beethoven"))
	if c := composer(score); c != "Beethoven" {
		t.Errorf("Expected composer Beethoven, got %s", c)
	}
}

func TestComposerEmpty(t *testing.T) {
	score := musicxml.NewScorePartwise()
	if c := composer(score); c != "" {
		t.Errorf("Expected empty composer, got %s", c)
	}
}

func TestRemoveRepititions(t *testing.T) {
	ending1 := musicxml.NewEnding(musicxml.WithEndingNumber(1), musicxml.WithEndingType(musicxml.EndingTypeStart))
	ending2 := musicxml.NewEnding(musicxml.WithEndingNumber(1), musicxml.WithEndingType(musicxml.EndingTypeStop))
	ending3 := musicxml.NewEnding(musicxml.WithEndingNumber(2), musicxml.WithEndingType(musicxml.EndingTypeStart))

	barline1 := musicxml.NewBarline(musicxml.WithEnding(ending1))
	barline2 := musicxml.NewBarline(musicxml.WithEnding(ending2))
	barline3 := musicxml.NewBarline(musicxml.WithEnding(ending3))
	measures := []musicxml.Measure{
		*musicxml.NewMeasure(),
		*musicxml.NewMeasure(musicxml.WithBarline(barline1)),
		*musicxml.NewMeasure(),
		*musicxml.NewMeasure(musicxml.WithBarline(barline2)),
		*musicxml.NewMeasure(musicxml.WithBarline(barline3)),
	}

	result := removeRepetitions(measures)

	if len(measures) != 5 {
		t.Errorf("Original measures should not be modified")
	}

	if len(result) != 2 {
		t.Errorf("Should be 2 measures left got %d", len(result))
	}

	for _, measure := range result {
		for _, element := range measure.MusicDataElements {
			if element.Barline != nil && element.Barline.Ending != nil {
				t.Errorf("There should be no endings left in the measures")
			}
		}
	}
}

func TestClearRepeatSigns(t *testing.T) {
	repeat := musicxml.NewBarline(musicxml.WithRepeat(&musicxml.Repeat{}))
	measures := []musicxml.Measure{
		*musicxml.NewMeasure(musicxml.WithBarline(repeat)),
		*musicxml.NewMeasure(),
	}
	result := clearRepeatSigns(measures)
	for _, measure := range result {
		for _, element := range measure.MusicDataElements {
			if element.Barline != nil && element.Barline.Repeat != nil {
				t.Errorf("There should be no repeat signs left in the measures")
			}
		}
	}
}

func TestEnsurePageBreak(t *testing.T) {
	for i, test := range []struct {
		elements []musicxml.MusicDataElement
		f        func([]musicxml.MusicDataElement) []musicxml.MusicDataElement
	}{
		{
			elements: []musicxml.MusicDataElement{{}, {Print: &musicxml.Print{}}},
			f:        ensurePageBreak,
		},
		{
			elements: []musicxml.MusicDataElement{{}},
			f:        ensurePageBreak,
		},
		{
			elements: []musicxml.MusicDataElement{{}, {Print: &musicxml.Print{}}},
			f:        ensureLineBreak,
		},
		{
			elements: []musicxml.MusicDataElement{{}},
			f:        ensureLineBreak,
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			result := test.f(test.elements)
			hasPrint := false
			for _, element := range result {
				if element.Print != nil {
					hasPrint = true
					break
				}
			}
			if !hasPrint {
				t.Errorf("Expected print element, got %v", result)
			}
		})
	}
}

func TestRemoveRedundantClefs(t *testing.T) {
	clef1 := musicxml.Clef{ClefDesc: musicxml.ClefDesc{Sign: "G"}}
	clef2 := musicxml.Clef{}

	for _, test := range []struct {
		clefs        map[int][]musicxml.Clef
		wantNumClefs int
		desc         string
	}{
		{
			clefs:        map[int][]musicxml.Clef{},
			wantNumClefs: 0,
			desc:         "No clefs",
		},
		{
			clefs:        map[int][]musicxml.Clef{0: {clef1}},
			wantNumClefs: 1,
			desc:         "One clef at beginning",
		},
		{
			clefs:        map[int][]musicxml.Clef{0: {clef1}, 4: {clef1}},
			wantNumClefs: 1,
			desc:         "One redundant clef at bar 4",
		},
		{
			clefs:        map[int][]musicxml.Clef{0: {clef1}, 4: {clef2}},
			wantNumClefs: 2,
			desc:         "New clef in bar 4",
		},
		{
			clefs:        map[int][]musicxml.Clef{0: {clef1, clef2}, 4: {clef1}},
			wantNumClefs: 3,
			desc:         "Clef split in bar 0 and back to old clef in bar 4",
		},
		{
			clefs:        map[int][]musicxml.Clef{0: {clef1, clef2}, 4: {clef2}},
			wantNumClefs: 2,
			desc:         "Clef split in bar 0 and redundant clef in bar 4",
		},
	} {
		t.Run(test.desc, func(t *testing.T) {
			measures := eightBarPiece()

			// Add clefs to the measures
			for i, clefs := range test.clefs {
				measures[i].MusicDataElements = append(measures[i].MusicDataElements, musicxml.MusicDataElement{
					Attributes: &musicxml.Attributes{
						Clef: clefs,
					},
					XMLName: xml.Name{Local: "attributes"},
				})
			}
			removeRedundantClefs(measures)
			numClefs := 0
			for _, measure := range measures {
				for _, elem := range measure.MusicDataElements {
					if elem.Attributes != nil {
						numClefs += len(elem.Attributes.Clef)
					}
				}
			}

			if numClefs != test.wantNumClefs {
				t.Errorf("Expected %d clefs, got %d", test.wantNumClefs, numClefs)
			}
		})
	}

}
