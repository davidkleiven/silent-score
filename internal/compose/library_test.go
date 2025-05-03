package compose

import (
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

	measures := []musicxml.Measure{
		*musicxml.NewMeasure(),
		*musicxml.NewMeasure(musicxml.WithEndingBarline(1, musicxml.EndingTypeStart)),
		*musicxml.NewMeasure(),
		*musicxml.NewMeasure(musicxml.WithEndingBarline(1, musicxml.EndingTypeStop)),
		*musicxml.NewMeasure(musicxml.WithEndingBarline(2, musicxml.EndingTypeStart)),
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
