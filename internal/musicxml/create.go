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

func WithBarline(b *Barline) MeasureOpt {

	return func(m *Measure) {
		m.MusicDataElements = append(m.MusicDataElements, MusicDataElement{
			Barline: b,
			XMLName: xml.Name{Local: "barline"},
		})
	}
}

func WithPrint(p *Print) MeasureOpt {
	return func(m *Measure) {
		m.MusicDataElements = append(m.MusicDataElements, MusicDataElement{
			Print:   p,
			XMLName: xml.Name{Local: "print"},
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

func NewPage() MusicDataElement {
	return MusicDataElement{
		XMLName: xml.Name{Local: "print"},
		Print: &Print{
			Printattributes: Printattributes{
				NewpageAttr: "yes",
			},
		},
	}
}

func NewSystem() MusicDataElement {
	return MusicDataElement{
		XMLName: xml.Name{Local: "print"},
		Print: &Print{
			Printattributes: Printattributes{
				NewsystemAttr: "yes",
			},
		},
	}
}

// Barline constructions
type BarlineOpt func(b *Barline)

func WithEnding(ending *Ending) BarlineOpt {
	return func(b *Barline) {
		b.Ending = ending
	}
}

func WithRepeat(repeat *Repeat) BarlineOpt {
	return func(b *Barline) {
		b.Repeat = repeat
	}
}

type BarStyle int

const (
	BarStyleDashed = iota
	BarStyleDotted
	BarStyleHeavy
	BarStyleHeavyHeavy
	BarStyleHeavyLight
	BarStyleLightHeavy
	BarStyleLightLight
	BarStyleRegular
	BarStyleShort
	BarStyleTick
)

func (bs BarStyle) String() string {
	var result string
	switch bs {
	case BarStyleDashed:
		result = "dashed"
	case BarStyleDotted:
		result = "dotted"
	case BarStyleHeavy:
		result = "heavy"
	case BarStyleHeavyHeavy:
		result = "heavy-heavy"
	case BarStyleHeavyLight:
		result = "heavy-light"
	case BarStyleLightHeavy:
		result = "light-heavy"
	case BarStyleLightLight:
		result = "light-light"
	case BarStyleRegular:
		result = "regular"
	case BarStyleShort:
		result = "short"
	case BarStyleTick:
		result = "tick"
	}
	return result
}

func WithBarStyle(style BarStyle) BarlineOpt {
	return func(b *Barline) {
		b.Barstyle = &Barstylecolor{Value: style.String()}
	}
}

func NewBarline(opts ...BarlineOpt) *Barline {
	var b Barline
	for _, opt := range opts {
		opt(&b)
	}
	return &b

}

// Ending constructions
type EndingOpt func(e *Ending)

func WithEndingType(endingType EndingType) EndingOpt {
	return func(e *Ending) {
		e.TypeAttr = endingType.String()
	}
}
func WithEndingNumber(endingNumber int) EndingOpt {
	return func(e *Ending) {
		e.NumberAttr = strconv.Itoa(endingNumber)
	}
}

type EndingType int

const (
	EndingTypeStart = iota
	EndingTypeStop
	EndingTypeDiscontinue
)

func (et EndingType) String() string {
	var result string
	switch et {
	case EndingTypeStart:
		result = "start"
	case EndingTypeStop:
		result = "stop"
	case EndingTypeDiscontinue:
		result = "discontinue"
	}
	return result
}

func NewEnding(opts ...EndingOpt) *Ending {
	var e Ending
	for _, opt := range opts {
		opt(&e)
	}
	return &e
}
