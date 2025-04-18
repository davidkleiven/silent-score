package musicxml

import (
	"encoding/xml"
	"strconv"

	"pgregory.net/rapid"
)

func GenerateRandomScorepartwise(t *rapid.T) *Scorepartwise {
	partPointers := rapid.SliceOfN(generateRandomPart(), 1, 4).Draw(t, "Part")
	partValues := make([]Part, len(partPointers))
	for i, part := range partPointers {
		if part != nil {
			partValues[i] = *part
		}
	}
	return &Scorepartwise{
		Part: partValues,
	}
}

func generateRandomPart() *rapid.Generator[*Part] {
	return rapid.Custom(func(t *rapid.T) *Part {
		measurePointers := rapid.SliceOfN(generateRandomMeasure(), 1, 16).Draw(t, "Measure")
		measureValues := make([]Measure, len(measurePointers))
		for i, measure := range measurePointers {
			if measure != nil {
				measureValues[i] = *measure
			}
		}
		return &Part{
			Partattributes: Partattributes{
				IdAttr: rapid.StringMatching(`[A-Z0-9]+`).Draw(t, "IdAttr"),
			},
			Measure: measureValues,
		}
	})
}

func generateRandomMeasure() *rapid.Generator[*Measure] {
	return rapid.Custom(func(t *rapid.T) *Measure {
		attributes := rapid.SliceOfN(generateRandomMusicDataElement(), 0, 4).Draw(t, "MusicDataElements")
		attributeValues := make([]MusicDataElement, len(attributes))
		for i, attr := range attributes {
			if attr != nil {
				attributeValues[i] = *attr
			}
		}

		return &Measure{
			Measureattributes: Measureattributes{
				NumberAttr: strconv.Itoa(rapid.IntRange(1, 100).Draw(t, "NumberAttr")),
			},
			MusicDataElements: attributeValues,
		}
	})
}

func generateRandomDirection() *rapid.Generator[*Direction] {
	return rapid.Custom(func(t *rapid.T) *Direction {
		dirTypePointers := rapid.SliceOfN(generateRandomDirectiontype(), 0, 2).Draw(t, "Directiontype")
		dirTypeValues := make([]Directiontype, len(dirTypePointers))
		for i, dirType := range dirTypePointers {
			if dirType != nil {
				dirTypeValues[i] = *dirType
			}
		}
		return &Direction{
			Directiontype: dirTypeValues,
		}
	})
}

func generateRandomDirectiontype() *rapid.Generator[*Directiontype] {
	return rapid.Custom(func(t *rapid.T) *Directiontype {
		wordsPointers := rapid.SliceOfN(generateRandomFormattedtextid("Words"), 0, 3).Draw(t, "Words")
		rehersalPointers := rapid.SliceOfN(generateRandomFormattedtextid("Rehearsal"), 0, 3).Draw(t, "Rehearsal")
		wordsValues := make([]Formattedtextid, len(wordsPointers))
		rehersalValues := make([]Formattedtextid, len(rehersalPointers))
		for i, words := range wordsPointers {
			if words != nil {
				wordsValues[i] = *words
			}
		}
		for i, rehersal := range rehersalPointers {
			if rehersal != nil {
				rehersalValues[i] = *rehersal
			}
		}
		return &Directiontype{
			Words:     wordsValues,
			Rehearsal: rehersalValues,
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
		timeSigPointers := rapid.SliceOfN(generateRandomTimesignature(), 0, 2).Draw(t, "Time")
		timeSigValues := make([]Timesignature, len(timeSigPointers))
		for i, timeSig := range timeSigPointers {
			if timeSig != nil {
				timeSigValues[i] = *timeSig
			}
		}
		return &Attributes{
			Time: timeSigValues,
		}
	})
}

func generateRandomTimesignature() *rapid.Generator[*Timesignature] {
	return rapid.Custom(func(t *rapid.T) *Timesignature {
		return &Timesignature{
			Beats:    rapid.IntRange(1, 4).Draw(t, "BeatsAttr"),
			Beattype: rapid.IntRange(2, 8).Draw(t, "Beattype"),
		}
	})
}

func generateRandomMetronome() *rapid.Generator[*Metronome] {
	metronome := rapid.Custom(func(t *rapid.T) *Metronome {
		return &Metronome{
			Perminute: &Perminute{
				Value: rapid.IntRange(40, 200).Draw(t, "PerminuteValue"),
			},
		}
	})

	var nilMetronome *Metronome
	return rapid.OneOf(metronome, rapid.Just(nilMetronome))
}

func generateRandomMusicDataElement() *rapid.Generator[*MusicDataElement] {
	names := []string{"Direction", "Attributes"}
	return rapid.Custom(func(t *rapid.T) *MusicDataElement {
		name := rapid.SampledFrom(names).Draw(t, "Name")
		switch name {
		case "Direction":
			return &MusicDataElement{
				XMLName:   xml.Name{Local: "Direction"},
				Direction: generateRandomDirection().Draw(t, "Direction"),
			}
		case "Attributes":
			return &MusicDataElement{
				XMLName:    xml.Name{Local: "Attributes"},
				Attributes: generateRandomAttributes().Draw(t, "Attributes"),
			}

		default:
			return &MusicDataElement{
				XMLName: xml.Name{Local: name},
			}
		}
	})
}
