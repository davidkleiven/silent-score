package musicxml

import (
	"strconv"

	"pgregory.net/rapid"
)

func GenerateRandomScorepartwise(t *rapid.T) *Scorepartwise {
	return &Scorepartwise{
		Part: rapid.SliceOfN(generateRandomPart(), 1, 4).Draw(t, "Part"),
	}
}

func generateRandomPart() *rapid.Generator[*Part] {
	return rapid.Custom(func(t *rapid.T) *Part {
		return &Part{
			Partattributes: Partattributes{
				IdAttr: rapid.StringMatching(`[A-Z0-9]+`).Draw(t, "IdAttr"),
			},
			Measure: rapid.SliceOfN(generateRandomMeasure(), 1, 16).Draw(t, "Measure"),
		}
	})
}

func generateRandomMeasure() *rapid.Generator[*Measure] {
	return rapid.Custom(func(t *rapid.T) *Measure {
		return &Measure{
			Measureattributes: Measureattributes{
				NumberAttr: strconv.Itoa(rapid.IntRange(1, 100).Draw(t, "NumberAttr")),
			},
			Musicdata: Musicdata{
				Direction:  rapid.SliceOfN(generateRandomDirection(), 0, 2).Draw(t, "Direction"),
				Attributes: rapid.SliceOfN(generateRandomAttributes(), 0, 2).Draw(t, "Attributes"),
			},
		}
	})
}

func generateRandomDirection() *rapid.Generator[*Direction] {
	return rapid.Custom(func(t *rapid.T) *Direction {
		return &Direction{
			Directiontype: rapid.SliceOfN(generateRandomDirectiontype(), 0, 2).Draw(t, "Directiontype"),
		}
	})
}

func generateRandomDirectiontype() *rapid.Generator[*Directiontype] {
	return rapid.Custom(func(t *rapid.T) *Directiontype {
		return &Directiontype{
			Words:     rapid.SliceOfN(generateRandomFormattedtextid("WordsValue"), 0, 3).Draw(t, "Words"),
			Rehearsal: rapid.SliceOfN(generateRandomFormattedtextid("RehearsalValue"), 0, 3).Draw(t, "Rehearsal"),
			Metronome: generateRandomMetronome().Draw(t, "Metronome"),
		}
	})
}

func generateRandomFormattedtextid(label string) *rapid.Generator[*Formattedtextid] {
	return rapid.Custom(func(t *rapid.T) *Formattedtextid {
		return &Formattedtextid{
			Value: rapid.String().Draw(t, label),
		}
	})
}

func generateRandomAttributes() *rapid.Generator[*Attributes] {
	return rapid.Custom(func(t *rapid.T) *Attributes {
		return &Attributes{
			Time: rapid.SliceOfN(generateRandomTimesignature(), 0, 2).Draw(t, "Time"),
		}
	})
}

func generateRandomTimesignature() *rapid.Generator[*Timesignature] {
	return rapid.Custom(func(t *rapid.T) *Timesignature {
		return &Timesignature{
			Beats:    strconv.Itoa(rapid.IntRange(1, 4).Draw(t, "BeatsAttr")),
			Beattype: strconv.Itoa(rapid.IntRange(2, 8).Draw(t, "Beattype")),
		}
	})
}

func generateRandomMetronome() *rapid.Generator[*Metronome] {
	metronome := rapid.Custom(func(t *rapid.T) *Metronome {
		return &Metronome{
			Perminute: &Perminute{
				Value: strconv.Itoa(rapid.IntRange(40, 200).Draw(t, "PerminuteValue")),
			},
		}
	})

	var nilMetronome *Metronome
	return rapid.OneOf(metronome, rapid.Just(nilMetronome))
}
