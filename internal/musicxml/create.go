package musicxml

import (
	"encoding/xml"
	"strconv"
)

// Measure constructions
type MeasureOpt func(m *Measure)

func WithRehersalMark(mark string) MeasureOpt {
	return func(m *Measure) {
		dirType := Directiontype{Rehearsal: []Formattedtextid{{Value: mark}}}
		direction := Direction{Directiontype: []Directiontype{dirType}}
		element := MusicDataElement{Direction: &direction, XMLName: xml.Name{Local: "direction"}}
		m.MusicDataElements = append(m.MusicDataElements, element)
	}
}

type EndingType int

const (
	EndingTypeStart = iota
	EndingTypeStop
	EndingTypeDiscontinue
)

func (et EndingType) String() string {
	switch et {
	case EndingTypeStart:
		return "start"
	case EndingTypeStop:
		return "stop"
	case EndingTypeDiscontinue:
		return "discontinue"
	}
	return "start"
}

func WithEndingBarline(number int, endingType EndingType) MeasureOpt {

	return func(m *Measure) {
		m.MusicDataElements = append(m.MusicDataElements, MusicDataElement{
			Barline: &Barline{
				LocationAttr: "left",
				Ending: &Ending{
					TypeAttr:   endingType.String(),
					NumberAttr: strconv.Itoa(number),
				},
			},
			XMLName: xml.Name{Local: "barline"},
		})
	}
}

func NewMeasure(opts ...MeasureOpt) *Measure {
	m := Measure{}
	for _, opt := range opts {
		opt(&m)
	}
	return &m
}

func MustDeepCopyMeasure(m *Measure) *Measure {
	var measure Measure
	marshalled, err := xml.Marshal(m)
	if err != nil {
		// Should not be possible since the datastructure is in the musicxml standard
		panic(err)
	}
	if err := xml.Unmarshal(marshalled, &measure); err != nil {
		panic(err)
	}
	return &measure
}

// Direction constructions
type DirectionOpt func(d *Direction)

func WithTempo(tempo int) DirectionOpt {
	return func(d *Direction) {
		dirType := Directiontype{Metronome: &Metronome{Perminute: &Perminute{Value: tempo}}}
		d.Directiontype = append(d.Directiontype, dirType)
	}
}

func NewDirection(opts ...DirectionOpt) *Direction {
	d := Direction{}
	for _, opt := range opts {
		opt(&d)
	}
	return &d
}

// Scorepartwise constructions
func NewScorePartwise(opts ...func(s *Scorepartwise)) *Scorepartwise {
	score := Scorepartwise{
		XMLName: xml.Name{Local: "score-partwise"},
	}

	for _, opt := range opts {
		opt(&score)
	}
	return &score
}

func WithComposer(composer string) func(s *Scorepartwise) {
	return func(s *Scorepartwise) {
		s.Credit = append(s.Credit, Credit{Credittype: []string{"composer"}, Creditwords: ComposerElement(composer)})
	}
}

// Other constructors
func TitleElement(title string) *Formattedtextid {
	return &Formattedtextid{
		Value: title,
		Textformatting: Textformatting{
			Justify: Justify{JustifyAttr: "center"},
			Printstylealign: Printstylealign{
				Printstyle: Printstyle{
					Font: Font{FontsizeAttr: "22"},
					Position: Position{
						DefaultxAttr: 616.9347,
						DefaultyAttr: 1511.047129,
					},
				},
				ValignAttr: "top",
			},
		},
	}
}

func ComposerElement(title string) *Formattedtextid {
	return &Formattedtextid{
		Value: title,
		Textformatting: Textformatting{
			Justify: Justify{JustifyAttr: "right"},
			Printstylealign: Printstylealign{
				Printstyle: Printstyle{
					Font: Font{FontsizeAttr: "10"},
					Position: Position{
						DefaultxAttr: 1148.144364,
						DefaultyAttr: 1411.047256,
					},
				},
				ValignAttr: "bottom",
			},
		},
	}
}

func DefaultPageMargins(marginType string) *Pagemargins {
	return &Pagemargins{
		TypeAttr: marginType,
		Allmargins: Allmargins{
			Leftrightmargins: Leftrightmargins{
				Leftmargin:  85.725,
				Rightmargin: 85.725,
			},
			Topmargin:    85.725,
			Bottommargin: 85.725,
		},
	}
}
