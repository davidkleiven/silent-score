package musicxml

import (
	"encoding/xml"
)

type MeasureOpt func(m *Measure)

func WithRehersalMark(mark string) MeasureOpt {
	return func(m *Measure) {
		dirType := Directiontype{Rehearsal: []*Formattedtextid{{Value: mark}}}
		direction := Direction{Directiontype: []Directiontype{dirType}}
		element := MusicDataElement{Direction: &direction, XMLName: xml.Name{Local: "direction"}}
		m.MusicDataElements = append(m.MusicDataElements, element)
	}

}

func NewMeasure(opts ...MeasureOpt) *Measure {
	m := Measure{}
	for _, opt := range opts {
		opt(&m)
	}
	return &m
}

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
