package musicxml

import (
	"encoding/xml"
	"io"
	"io/fs"
	"log/slog"
	"strconv"
)

// TextFields extracts all fields in the document that should be
// used for matching
func TextFields(document Scorepartwise) []string {
	out := []string{}

	for _, credit := range document.Credit {
		if credit != nil && credit.Creditwords != nil {
			out = append(out, credit.Creditwords.Value)
		}
	}

	for _, part := range document.Part {
		if part != nil {
			for _, measureText := range MeasureText(part.Measure) {
				out = append(out, measureText.Text)
			}
		}
	}
	return out
}

type MeasureTextResult struct {
	Number int
	Text   string
}

func MeasureText(measures []*Measure) []MeasureTextResult {
	result := []MeasureTextResult{}
	for _, measure := range measures {
		if measure != nil {
			result = append(result, DirectionFromMeasure(*measure)...)
		}
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

	for _, direction := range measure.Direction {
		if direction != nil {
			for _, dirType := range direction.Directiontype {
				if dirType != nil {
					dirText := ""
					for _, words := range dirType.Words {
						if words != nil {
							dirText += words.Value
						}
					}

					if dirText != "" {
						result = append(result, MeasureTextResult{Number: num, Text: dirText})
					}

					dirText = ""
					for _, rehersal := range dirType.Rehearsal {
						if rehersal != nil {
							dirText += rehersal.Value
						}
					}

					if dirText != "" {
						result = append(result, MeasureTextResult{Number: num, Text: dirText})
					}
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
