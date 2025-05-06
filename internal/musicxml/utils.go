package musicxml

import (
	"encoding/xml"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"slices"
	"strconv"
	"strings"
)

// TextFields extracts all fields in the document that should be
// used for matching
func TextFields(document Scorepartwise) []string {
	out := []string{}

	for _, credit := range document.Credit {
		if credit.Creditwords != nil {
			out = append(out, credit.Creditwords.Value)
		}
	}

	for _, part := range document.Part {
		for _, measureText := range MeasureText(part.Measure) {
			out = append(out, measureText.Text)
		}
	}
	return out
}

type MeasureTextResult struct {
	Number int
	Text   string
}

func MeasureText(measures []Measure) []MeasureTextResult {
	result := []MeasureTextResult{}
	for _, measure := range measures {
		result = append(result, DirectionFromMeasure(measure)...)
	}
	return result
}

func DirectionFromMeasure(measure Measure) []MeasureTextResult {
	result := []MeasureTextResult{}
	num, err := strconv.Atoi(measure.NumberAttr)
	if err != nil {
		slog.Warn(err.Error())
		return result
	}

	for _, element := range measure.MusicDataElements {
		if direction := element.Direction; direction != nil {
			for _, dirType := range direction.Directiontype {

				dirText := ""
				for _, words := range dirType.Words {
					dirText += words.Value
				}

				if dirText != "" {
					result = append(result, MeasureTextResult{Number: num, Text: dirText})
				}

				dirText = ""
				for _, rehersal := range dirType.Rehearsal {
					dirText += rehersal.Value
				}

				if dirText != "" {
					result = append(result, MeasureTextResult{Number: num, Text: dirText})
				}
			}
		}
	}
	return result
}

func ReadFromFile(reader io.Reader) (Scorepartwise, error) {
	var score Scorepartwise
	content, err := io.ReadAll(reader)
	if err != nil {
		return score, err
	}

	if err := xml.Unmarshal(content, &score); err != nil {
		return score, err
	}
	return score, nil
}

func ReadFromFileName(fs fs.FS, name string) Scorepartwise {
	file, err := fs.Open(name)
	if err != nil {
		slog.Error("Failed to open file", "file", name, "error", err)
		return Scorepartwise{}
	}
	defer file.Close()
	score, err := ReadFromFile(file)
	if err != nil {
		slog.Error("Failed to read score", "file", name, "error", err)
	}
	return score
}

func WriteScore(writer io.Writer, score *Scorepartwise) error {
	encoder := xml.NewEncoder(writer)
	encoder.Indent("", "  ")
	if err := encoder.Encode(score); err != nil {
		return err
	}
	return encoder.Flush()
}

type Creator interface {
	Create(name string) (WriterCloser, error)
}

type WriterCloser interface {
	io.Writer
	io.Closer
}

type FileCreator struct{}

func (fc *FileCreator) Create(name string) (WriterCloser, error) {
	return os.Create(name)
}

func WriteScoreToFile(creator Creator, name string, score *Scorepartwise) error {
	file, err := creator.Create(name)
	if err != nil {
		return err
	}
	defer file.Close()

	return WriteScore(file, score)
}

func FileNameFromScore(score *Scorepartwise) string {
	if score.Scoreheader.Work != nil && score.Scoreheader.Work.Worktitle != "" {
		return strings.ReplaceAll(score.Scoreheader.Work.Worktitle, " ", "_") + ".musicxml"
	}
	return "silent-score.musicxml"
}

func SetTempoAtBeginning(measure *Measure, metronome *Metronome) {
	applyBeforeFirstNote(measure, "direction", false, func(m *MusicDataElement) { setTempo(m, metronome) })
}

func SetSystemTextAtBeginning(measure *Measure, text string) {
	applyBeforeFirstNote(measure, "direction", true, func(m *MusicDataElement) { setSystemText(m, text) })
}

func setTimeSignature(element *MusicDataElement, timeSignature Timesignature) {
	ensureAttributes(element)
	element.Attributes.Time = []Timesignature{timeSignature}
}

func SetTimeSignatureAtBeginning(measure *Measure, timeSignature Timesignature) {
	applyBeforeFirstNote(measure, "attributes", true, func(m *MusicDataElement) { setTimeSignature(m, timeSignature) })
}

func setTempo(element *MusicDataElement, metronome *Metronome) {
	ensureDirection(element)
	metronomeIsSet := false
	for i, dirType := range element.Direction.Directiontype {
		if dirType.Metronome != nil {
			element.Direction.Directiontype[i].Metronome = metronome
			metronomeIsSet = true
		}
	}

	if !metronomeIsSet {
		element.Direction.Directiontype = append(element.Direction.Directiontype, Directiontype{Metronome: metronome})
	}
}

func ensureDirection(m *MusicDataElement) {
	if m != nil && m.Direction == nil {
		m.Direction = &Direction{}
	}
}

func ensureAttributes(m *MusicDataElement) {
	if m != nil && m.Attributes == nil {
		m.Attributes = &Attributes{}
	}
}

func setSystemText(element *MusicDataElement, text string) {
	ensureDirection(element)
	element.Direction.Directiontype = append(element.Direction.Directiontype, Directiontype{Words: []Formattedtextid{{Value: text}}})
}

func applyBeforeFirstNote(measure *Measure, name string, reuseExisting bool, fn func(m *MusicDataElement)) {
	for i, element := range measure.MusicDataElements {
		if element.Note != nil {
			newElement := MusicDataElement{XMLName: xml.Name{Local: name}}
			fn(&newElement)
			measure.MusicDataElements = slices.Insert(measure.MusicDataElements, i, newElement)
			return
		}
		if element.XMLName.Local == name && reuseExisting {
			fn(&measure.MusicDataElements[i])
			return
		}
	}

	newElement := MusicDataElement{XMLName: xml.Name{Local: name}}
	fn(&newElement)
	measure.MusicDataElements = append(measure.MusicDataElements, newElement)
}

func SetBarlineAtEnd(measure *Measure, barline *Barline) {
	measure.MusicDataElements = slices.DeleteFunc(measure.MusicDataElements, func(m MusicDataElement) bool {
		return m.Barline != nil
	})
	measure.MusicDataElements = append(measure.MusicDataElements, MusicDataElement{
		Barline: barline,
		XMLName: xml.Name{Local: "barline"},
	})
}

func ClefEquals(clef1, clef2 *Clef) bool {
	if clef1 == nil || clef2 == nil {
		return false
	}
	return clef1.Sign == clef2.Sign && clef1.Line == clef2.Line && clef1.OctaveChange == clef2.OctaveChange
}
