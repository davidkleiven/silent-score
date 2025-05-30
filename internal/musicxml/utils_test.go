package musicxml

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

func newMeasure(measureNo string, dirType Directiontype) Measure {
	direction := Direction{Directiontype: []Directiontype{dirType}}
	return Measure{
		Measureattributes: Measureattributes{NumberAttr: measureNo},
		MusicDataElements: []MusicDataElement{{Direction: &direction, XMLName: xml.Name{Local: "Direction"}}},
	}
}

func TestDirectionFromMeasure(t *testing.T) {
	formattedText := Formattedtextid{Value: "value"}

	dirTypeRehersalFilled := Directiontype{Rehearsal: []Formattedtextid{formattedText}}
	dirTypeWordFilled := Directiontype{Words: []Formattedtextid{formattedText}}

	for _, test := range []struct {
		measure Measure
		want    []MeasureTextResult
		desc    string
	}{
		{
			newMeasure("1", dirTypeRehersalFilled),
			[]MeasureTextResult{{Number: 1, Text: "value"}},
			"Rehersal non nil",
		},
		{
			newMeasure("1", dirTypeWordFilled),
			[]MeasureTextResult{{Number: 1, Text: "value"}},
			"Word filled non nil",
		},
		{
			newMeasure("non-int", dirTypeWordFilled),
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
	dirType := Directiontype{Rehearsal: []Formattedtextid{value}}

	m1 := newMeasure("1", dirType)
	m2 := newMeasure("2", dirType)
	measures := []Measure{m1, m2}

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

func TestReadFromXmlNonMuusicXmlContent(t *testing.T) {
	data := []byte("<not-musicxml>content</not-musicxml>")
	reader := bytes.NewReader(data)
	score, err := ReadFromFile(reader)
	if err == nil {
		t.Errorf("Expected error reading non-MusicXML content, got nil")
	}
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

type failingZipReader struct{}

func (f *failingZipReader) OpenFile(name string) (fs.File, error) {
	return nil, os.ErrNotExist
}
func (f *failingZipReader) OpenZipFile(name string) (*zip.ReadCloser, error) {
	return nil, os.ErrNotExist
}

func TestWriteScore(t *testing.T) {
	var buf bytes.Buffer
	score := Scorepartwise{
		Scoreheader: Scoreheader{
			Work: &Work{
				Worktitle: "Test",
			},
		},
	}
	err := WriteScore(&buf, &score)
	if err != nil {
		t.Errorf("Error writing score: %v", err)
		return
	}

	expected := `<score-partwise>
  <work>
    <work-title>Test</work-title>
  </work>
</score-partwise>`

	if buf.String() != expected {
		t.Errorf("Expected\n%s\ngot\n%s", expected, buf.String())
	}
}

type failingCreator struct{}

func (f *failingCreator) Create(name string) (WriterCloser, error) {
	return nil, os.ErrNotExist
}

func TestWriteScoreToFile(t *testing.T) {
	score := Scorepartwise{}
	err := WriteScoreToFile(&failingCreator{}, "score.xml", &score)
	if !errors.Is(err, os.ErrNotExist) {
		t.Errorf("Expected error writing score, got nil")
	}
}

func TestWriteScoreToFileWithFileCreator(t *testing.T) {
	score := Scorepartwise{
		Scoreheader: Scoreheader{
			Work: &Work{
				Worktitle: "Test",
			},
		},
	}

	filename := t.TempDir() + "/score.xml"
	err := WriteScoreToFile(&FileCreator{}, filename, &score)
	if err != nil {
		t.Errorf("Error writing score: %v", err)
		return
	}

	expected := `<score-partwise>
  <work>
    <work-title>Test</work-title>
  </work>
</score-partwise>`
	fileContent, err := os.ReadFile(filename)
	if err != nil {
		t.Error(err)
	}
	if string(fileContent) != expected {
		t.Errorf("Expected\n%s\ngot\n%s", expected, string(fileContent))
	}
}

func TestFileNameFromScore(t *testing.T) {
	for _, test := range []struct {
		score    Scorepartwise
		expected string
	}{
		{
			Scorepartwise{
				Scoreheader: Scoreheader{
					Work: &Work{
						Worktitle: "Test title",
					},
				},
			},
			"Test_title.musicxml",
		},
		{
			Scorepartwise{},
			"silent-score.musicxml",
		},
		{
			Scorepartwise{
				Scoreheader: Scoreheader{
					Work: &Work{
						Worktitle: "",
					},
				},
			},
			"silent-score.musicxml",
		},
	} {
		result := FileNameFromScore(&test.score)
		if result != test.expected {
			t.Errorf("Expected %s, got %s", test.expected, result)
		}
	}
}

func TestTempoAtBeginning(t *testing.T) {
	for _, test := range []struct {
		tempo   int
		measure *Measure
		desc    string
	}{
		{
			tempo:   92,
			measure: &Measure{},
			desc:    "Bar without elements",
		},
		{
			tempo: 92,
			measure: &Measure{
				MusicDataElements: []MusicDataElement{
					{XMLName: xml.Name{Local: "direction"}},
				},
			},
			desc: "Has one element with direction field",
		},
		{
			tempo: 92,
			measure: &Measure{
				MusicDataElements: []MusicDataElement{
					{XMLName: xml.Name{Local: "note"}, Note: &Note{}},
					{XMLName: xml.Name{Local: "direction"}},
				},
			},
			desc: "Has one element with direction after a note",
		},
		{
			tempo: 92,
			measure: &Measure{
				MusicDataElements: []MusicDataElement{
					{XMLName: xml.Name{Local: "direction"}},
					{XMLName: xml.Name{Local: "note"}, Note: &Note{}},
				},
			},
			desc: "Has one element with direction before a note",
		},
	} {
		t.Run(test.desc, func(t *testing.T) {
			metronome := Metronome{Perminute: &Perminute{Value: test.tempo}}
			SetTempoAtBeginning(test.measure, &metronome)
			hasTempo := false
			for _, element := range test.measure.MusicDataElements {
				if element.Note != nil {
					break
				}
				if element.Direction != nil {
					for _, dirType := range element.Direction.Directiontype {
						if dirType.Metronome != nil && dirType.Metronome.Perminute != nil {
							if dirType.Metronome.Perminute.Value == test.tempo {
								hasTempo = true
							}
						}
					}
				}
			}

			if !hasTempo {
				t.Errorf("No musical elements had a tempot")
			}
		})
	}
}

func TestSetSystemTextAtBeginning(t *testing.T) {
	for _, test := range []struct {
		text    string
		measure *Measure
		desc    string
	}{
		{
			text:    "Test",
			measure: &Measure{},
			desc:    "Bar without elements",
		},
		{
			text: "Test",
			measure: &Measure{
				MusicDataElements: []MusicDataElement{
					{XMLName: xml.Name{Local: "direction"}},
				},
			},
			desc: "Has one element with direction field",
		},
	} {
		t.Run(test.desc, func(t *testing.T) {
			SetSystemTextAtBeginning(test.measure, test.text)
			hasText := false
			for _, element := range test.measure.MusicDataElements {
				if element.Note != nil {
					break
				}
				if element.Direction != nil {
					for _, dirType := range element.Direction.Directiontype {
						if len(dirType.Words) > 0 && dirType.Words[0].Value == test.text {
							hasText = true
						}
					}
				}
			}

			if !hasText {
				t.Errorf("No musical elements had a system text")
			}
		})
	}
}

func TestSetTimeSignatureAtBeginning(t *testing.T) {
	for _, test := range []struct {
		timeSignature Timesignature
		measure       *Measure
		desc          string
	}{
		{
			timeSignature: Timesignature{Beats: 4, Beattype: 4},
			measure:       &Measure{},
			desc:          "Bar without elements",
		},
		{
			timeSignature: Timesignature{Beats: 4, Beattype: 4},
			measure: &Measure{
				MusicDataElements: []MusicDataElement{
					{XMLName: xml.Name{Local: "attributes"}},
				},
			},
			desc: "Has one element with attributes field",
		},
	} {
		t.Run(test.desc, func(t *testing.T) {
			SetTimeSignatureAtBeginning(test.measure, test.timeSignature)
			hasTimeSig := false
			for _, element := range test.measure.MusicDataElements {
				if element.Attributes != nil && len(element.Attributes.Time) > 0 {
					if element.Attributes.Time[0].Beats == test.timeSignature.Beats &&
						element.Attributes.Time[0].Beattype == test.timeSignature.Beattype {
						hasTimeSig = true
					}
				}
			}

			if !hasTimeSig {
				t.Errorf("No musical elements had a time signature")
			}
		})
	}
}

func TestApplyBeforeFirstNote(t *testing.T) {
	newName := "element-was-modified"
	fn := func(m *MusicDataElement) { m.XMLName.Local = newName }

	for _, test := range []struct {
		name    string
		measure *Measure
		desc    string
	}{
		{
			name: "direction",
			measure: &Measure{
				MusicDataElements: []MusicDataElement{
					{XMLName: xml.Name{Local: "note"}, Note: &Note{}},
					{XMLName: xml.Name{Local: "direction"}},
				},
			},
			desc: "Has one element with direction after a note",
		},
		{
			name:    "direction",
			measure: &Measure{},
			desc:    "Bar without elements",
		},
		{
			name: "direction",
			desc: "One note at end",
			measure: &Measure{
				MusicDataElements: []MusicDataElement{
					{XMLName: xml.Name{Local: "direction"}},
					{XMLName: xml.Name{Local: "note"}, Note: &Note{}},
				},
			},
		},
	} {
		t.Run(test.desc, func(t *testing.T) {
			applyBeforeFirstNote(test.measure, test.name, true, fn)
			hasApplied := false
			for _, element := range test.measure.MusicDataElements {
				if element.Note != nil {
					break
				}
				if element.XMLName.Local == newName {
					hasApplied = true
					break
				}
			}

			if !hasApplied {
				t.Errorf("No musical elements had the applied function")
			}
		})
	}
}

func TestApplyBeforeFirstNoteNoteNotRemoved(t *testing.T) {
	measure := &Measure{
		MusicDataElements: []MusicDataElement{
			{XMLName: xml.Name{Local: "direction"}},
			{XMLName: xml.Name{Local: "note"}, Note: &Note{}},
			{XMLName: xml.Name{Local: "direction"}},
		},
	}

	fn := func(m *MusicDataElement) { m.XMLName.Local = "modified" }
	applyBeforeFirstNote(measure, "attributes", false, fn)
	if len(measure.MusicDataElements) != 4 {
		t.Errorf("Element not added")
	}

	if measure.MusicDataElements[2].Note == nil {
		t.Errorf("Note does not appear in the right place")
	}

}

func TestSetTempoWithExistingMetronome(t *testing.T) {
	tempo := 92
	element := &MusicDataElement{
		XMLName: xml.Name{Local: "direction"},
		Direction: &Direction{
			Directiontype: []Directiontype{
				{
					Metronome: &Metronome{
						Perminute: &Perminute{Value: 100},
					},
				},
			},
		},
	}

	setTempo(element, &Metronome{Perminute: &Perminute{Value: tempo}})
	if element.Direction.Directiontype[0].Metronome.Perminute.Value != tempo {
		t.Errorf("Tempo not set correctly, expected %d, got %d", tempo, element.Direction.Directiontype[0].Metronome.Perminute.Value)
	}

}

func TestClefEqual(t *testing.T) {
	clef1 := Clef{ClefDesc: ClefDesc{Sign: "G", Line: 2}}
	clef2 := Clef{ClefDesc: ClefDesc{Sign: "G", Line: 2}}
	clef3 := Clef{ClefDesc: ClefDesc{Sign: "F", Line: 2}}
	clef4 := Clef{ClefDesc: ClefDesc{Sign: "G", Line: 4}}

	for i, test := range []struct {
		clef1, clef2 *Clef
		expected     bool
	}{
		{&clef1, &clef2, true},
		{&clef1, &clef3, false},
		{&clef1, &clef4, false},
		{&clef1, nil, false},
		{nil, &clef2, false},
		{nil, nil, false},
	} {
		t.Run(fmt.Sprintf("Test %d", i), func(t *testing.T) {
			result := ClefEquals(test.clef1, test.clef2)
			if result != test.expected {
				t.Errorf("Expected %v, got %v", test.expected, result)
			}
		})
	}
}

func TestReadCompressedFile(t *testing.T) {
	data := testFile("Saltarello.mxl")
	zipReader, err := zip.OpenReader(data)
	if err != nil {
		t.Error(err)
		return
	}
	defer zipReader.Close()
	score, err := ReadCompressedFile(&zipReader.Reader)
	if err != nil {
		t.Errorf("Error reading compressed file: %v", err)
		return
	}

	if score.Scoreheader.Movementtitle != "Saltarello" {
		t.Errorf("Expected score with title 'Saltarello', got %v", score.Work)
		return
	}

	if len(score.Part) != 1 {
		t.Errorf("Expected one part, got %d", len(score.Part))
		return
	}

	if len(score.Part[0].Measure) != 22 {
		t.Errorf("Expected 22 measures, got %d", len(score.Part[0].Measure))
		return
	}
}

func TestErrorWhenNoMusicXmlFileExists(t *testing.T) {
	reader := &zip.Reader{}
	_, err := ReadCompressedFile(reader)
	if !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("Expected fs.ErrNotExist error, got %v", err)
		return
	}
}

func TestReadCompressedFileFromFileName(t *testing.T) {
	data := testFile("Saltarello.mxl")
	reader := os.DirFS(filepath.Dir(data))
	score := ReadFromFileName(reader, filepath.Base(data))
	if score.Scoreheader.Movementtitle != "Saltarello" {
		t.Errorf("Expected score with title 'Saltarello', got %v", score.Work)
		return
	}
}

func TestIsRepeatJump(t *testing.T) {
	for _, test := range []struct {
		element *Direction
		want    bool
		desc    string
	}{
		{
			element: NewDirection(WithSegno(Segno{})),
			want:    true,
			desc:    "Dal segno",
		},
		{
			element: NewDirection(WithCoda(Coda{})),
			want:    true,
			desc:    "Contains coda",
		},
		{
			element: NewDirection(WithWords("D. S.")),
			want:    true,
			desc:    "D. S. is a repeat jump",
		},
		{
			element: NewDirection(WithWords("Fine")),
			want:    true,
			desc:    "Fine is a repeat jump",
		},
		{
			element: NewDirection(WithWords("To Coda")),
			want:    true,
			desc:    "To Coda is a repeat jump",
		},
		{
			element: NewDirection(WithWords("<sym>Coda</sym>")),
			want:    true,
			desc:    "Coda with sym tag is a repeat jump",
		},
		{
			element: NewDirection(WithWords("This is not a repeat")),
			want:    false,
			desc:    "This is not a repeat jump",
		},
	} {
		t.Run(test.desc, func(t *testing.T) {
			result := IsRepeatJump(test.element)
			if result != test.want {
				t.Errorf("Expected %v, got %v", test.want, result)
			}
		})
	}
}
