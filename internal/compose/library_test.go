package compose

import (
	"strconv"
	"testing"

	"github.com/davidkleiven/silent-score/internal/db"
	"github.com/davidkleiven/silent-score/internal/musicxml"
	"github.com/davidkleiven/silent-score/test"
	"pgregory.net/rapid"
)

func measuresAreEnumerated(measures []*musicxml.Measure) bool {
	for i, measure := range measures {
		if measure != nil {
			if measure.NumberAttr != strconv.Itoa(i+1) {
				return false
			}
		}
	}
	return true
}

func atLeastOneSceneDescPartOfComposition(measures []*musicxml.Measure, records []db.ProjectContentRecord) bool {
	if len(measures) == 0 {
		return true
	}
	for _, measure := range measures {
		if measure != nil {
			for _, direction := range measure.Direction {
				if direction != nil {
					for _, dirType := range direction.Directiontype {
						if dirType != nil {
							for _, words := range dirType.Words {
								if words != nil {
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
