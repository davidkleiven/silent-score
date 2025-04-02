package musicxml

import (
	"io/fs"
	"os"
	"testing"
)

func newMeasure(measureNo string, dirType *Directiontype) Measure {
	return Measure{
		Measureattributes: Measureattributes{NumberAttr: measureNo},
		Musicdata:         Musicdata{Direction: []*Direction{{Directiontype: []*Directiontype{dirType}}}},
	}
}

func TestDirectionFromMeasure(t *testing.T) {
	formattedText := Formattedtextid{Value: "value"}
	dirTypeRehersalNil := Directiontype{Rehearsal: []*Formattedtextid{nil}}
	dirTypeWOrdNil := Directiontype{Words: []*Formattedtextid{nil}}

	dirTypeRehersalFilled := Directiontype{Rehearsal: []*Formattedtextid{&formattedText}}
	dirTypeWordFilled := Directiontype{Words: []*Formattedtextid{&formattedText}}

	for _, test := range []struct {
		measure Measure
		want    []MeasureTextResult
		desc    string
	}{
		{
			Measure{Musicdata: Musicdata{Direction: []*Direction{nil}}},
			[]MeasureTextResult{},
			"Direction type nil",
		},
		{
			newMeasure("1", &dirTypeRehersalNil),
			[]MeasureTextResult{},
			"Rehersal content nil",
		},
		{
			newMeasure("1", &dirTypeWOrdNil),
			[]MeasureTextResult{},
			"Word content nil",
		},
		{
			newMeasure("1", &dirTypeRehersalFilled),
			[]MeasureTextResult{{Number: 1, Text: "value"}},
			"Rehersal non nil",
		},
		{
			newMeasure("1", &dirTypeWordFilled),
			[]MeasureTextResult{{Number: 1, Text: "value"}},
			"Word filled non nil",
		},
		{
			newMeasure("non-int", &dirTypeWordFilled),
			[]MeasureTextResult{},
			"Word filled wrong measure number",
		},
	} {
		t.Run(test.desc, func(t *testing.T) {

			result := DirectionFromMeasure(test.measure)

			if len(result) != len(test.want) {
				t.Errorf("Wanted %v got %v", test.want, result)
			}

			for i := range result {
				if result[i].Number != test.want[i].Number || result[i].Text != test.want[i].Text {
					t.Errorf("Wanted %v got %v", test.want, result)
				}
			}
		})
	}
}

func TestMeasureText(t *testing.T) {
	value := Formattedtextid{Value: "value"}
	dirType := Directiontype{Rehearsal: []*Formattedtextid{&value}}

	m1 := newMeasure("1", &dirType)
	m2 := newMeasure("2", &dirType)
	measures := []*Measure{&m1, &m2}

	want := []MeasureTextResult{
		{
			Number: 1,
			Text:   "value",
		},
		{
			Number: 2,
			Text:   "value",
		},
	}

	result := MeasureText(measures)

	if len(result) != len(want) {
		t.Errorf("Wanted %v got %v\n", result, want)
		return
	}

	for i := range result {
		if result[i].Number != want[i].Number || result[i].Text != want[i].Text {
			t.Errorf("Wanted %v got %v\n", result[i], want[i])
		}
	}

}

type openFailFs struct{}

func (o *openFailFs) Open(name string) (fs.File, error) {
	return nil, os.ErrNotExist
}
func TestReadFromFileNameOpenFails(t *testing.T) {
	score := ReadFromFileName(&openFailFs{}, "file.xml")
	if score.Work != nil {
		t.Errorf("Expected empty score, got %v", score)
	}
}

type nonXmlFile struct{}

func (n *nonXmlFile) Stat() (fs.FileInfo, error) {
	return nil, os.ErrNotExist
}
func (n *nonXmlFile) Read(b []byte) (int, error) {
	return 0, os.ErrNotExist
}
func (n *nonXmlFile) Close() error {
	return nil
}

type nonXmlFileFs struct{}

func (n *nonXmlFileFs) Open(name string) (fs.File, error) {
	return &nonXmlFile{}, nil
}

func TestReadFromFileNameNonXmlFile(t *testing.T) {
	score := ReadFromFileName(&nonXmlFileFs{}, "file.xml")
	if score.Work != nil {
		t.Errorf("Expected empty score, got %v", score)
	}
}
