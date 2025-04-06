package compose

import (
	"embed"
	"iter"
	"slices"
	"strconv"
	"strings"
	"time"

	_ "embed"
	"log/slog"

	"github.com/davidkleiven/silent-score/internal/db"
	"github.com/davidkleiven/silent-score/internal/musicxml"
)

const (
	defaultTempo = 82
	defaultBpm   = 4
)

type Library interface {
	BestMatch(desc string) *musicxml.Scorepartwise
}

//go:embed assets/*.musicxml
var standardLib embed.FS

type StandardLibrary struct {
	directory string
}

func NewStandardLibrary() *StandardLibrary {
	return &StandardLibrary{directory: "assets"}
}

func (sl *StandardLibrary) readNames() []string {
	entries, err := standardLib.ReadDir(sl.directory)
	if err != nil {
		slog.Error("Failed to read standard library directory", "error", err)
		return []string{}
	}
	names := make([]string, len(entries))
	for i, entry := range entries {
		names[i] = entry.Name()
	}
	slog.Info("Standard library loaded", "count", len(names))
	return names
}

func (sl *StandardLibrary) scores() iter.Seq[*musicxml.Scorepartwise] {
	names := sl.readNames()
	return func(yield func(item *musicxml.Scorepartwise) bool) {
		for _, name := range names {
			score := musicxml.ReadFromFileName(standardLib, sl.directory+"/"+name)
			if !yield(&score) {
				break
			}
		}
	}
}

func (sl *StandardLibrary) BestMatch(desc string) *musicxml.Scorepartwise {
	texts := collectTextFields(sl.scores())
	bestMatch := bestMatchForDesc(desc, texts)
	name := sl.readNames()[bestMatch]
	score := musicxml.ReadFromFileName(standardLib, sl.directory+"/"+name)
	return &score
}

type InMemoryLibrary struct {
	scores []*musicxml.Scorepartwise
}

func (l *InMemoryLibrary) BestMatch(desc string) *musicxml.Scorepartwise {
	texts := collectTextFields(slices.Values(l.scores))
	return l.scores[bestMatchForDesc(desc, texts)]
}

func collectTextFields(scoreIter iter.Seq[*musicxml.Scorepartwise]) []string {
	var texts []string
	for score := range scoreIter {
		texts = append(texts, strings.Join(musicxml.TextFields(*score), " "))
	}
	return texts
}

func bestMatchForDesc(desc string, texts []string) int {
	normalizedDesc := normalize(desc)
	normalizedTexts := make([]string, len(texts))
	for i, text := range texts {
		normalizedTexts[i] = normalize(text)
	}
	bestMatch := orderPieces(normalizedDesc, normalizedTexts)
	return bestMatch[0].Index
}

func tempoIfGiven(tempo int, measures []*musicxml.Measure) int {
	if tempo > 0 {
		return tempo
	}
	for _, measure := range measures {
		if measure != nil {
			for _, direction := range measure.Direction {
				if direction != nil {
					for _, dirType := range direction.Directiontype {
						if dirType != nil && dirType.Metronome != nil {
							if perminute, err := strconv.Atoi(dirType.Metronome.Perminute.Value); err != nil {
								return perminute
							}
						}
					}
				}
			}
		}
	}
	return defaultTempo
}

func beatsPerMeasure(measures []*musicxml.Measure) *musicxml.Timesignature {
	for _, measure := range measures {
		if measure != nil {
			for _, attr := range measure.Attributes {
				if attr != nil && attr.Time != nil {
					for _, timesig := range attr.Time {
						if timesig != nil {
							return timesig
						}
					}
				}
			}
		}
	}
	return &musicxml.Timesignature{
		Beats:    "4",
		Beattype: "4",
	}
}

func addSystemText(measure *musicxml.Measure, text string) {
	if measure != nil {
		measure.Direction = append(measure.Direction, &musicxml.Direction{
			Directiontype: []*musicxml.Directiontype{
				{Words: []*musicxml.Formattedtextid{{Value: text}}},
			},
		})
	}
}

func addTimeSignature(measure *musicxml.Measure, timeSignature *musicxml.Timesignature) {
	if measure != nil {
		measure.Attributes = append(measure.Attributes, &musicxml.Attributes{
			Time: []*musicxml.Timesignature{timeSignature},
		})
	}
}

func clearTempoMarkings(measures []*musicxml.Measure) {
	for _, measure := range measures {
		if measure != nil {
			for i, direction := range measure.Direction {
				if direction != nil {
					for j, dirType := range direction.Directiontype {
						if dirType != nil && dirType.Metronome != nil {
							measure.Direction[i].Directiontype[j].Metronome = nil
						}
					}
				}
			}
		}
	}
}

func addTempo(measure *musicxml.Measure, tempo int) {
	if measure != nil {
		measure.Direction = append(measure.Direction, &musicxml.Direction{
			Directiontype: []*musicxml.Directiontype{
				{Metronome: &musicxml.Metronome{
					Perminute: &musicxml.Perminute{
						Value: strconv.Itoa(tempo),
					},
				}},
			},
		})
	}
}

func enumerateMeasuresInPlace(measures []*musicxml.Measure) {
	for i, measure := range measures {
		if measure != nil {
			measure.NumberAttr = strconv.Itoa(i + 1)
		}
	}
}

type selection struct {
	measures []*musicxml.Measure
}

func title(score *musicxml.Scorepartwise) string {
	if score != nil && score.Scoreheader.Work != nil {
		return score.Scoreheader.Work.Worktitle
	}
	return ""
}

func pickMeasures(library Library, records []db.ProjectContentRecord) selection {
	var measures []*musicxml.Measure
	for _, record := range records {
		piece := library.BestMatch(record.Keywords)
		slog.Info("Picking piece", "keywords", record.Keywords, "sceneDesc", record.SceneDesc, "title", title(piece))
		if piece != nil {
			if len(piece.Part) > 0 {
				sections := pieceSections(piece.Part[0].Measure)
				timeSignature := beatsPerMeasure(piece.Part[0].Measure)
				beatsInTimeSig := beatsPerMinutes(timeSignature.Beats)
				tempo := tempoIfGiven(int(record.Tempo), piece.Part[0].Measure)
				sceneSection := sectionForScene(time.Duration(record.DurationSec)*time.Second, float64(tempo), beatsInTimeSig, sections)
				measuresForScene := measuresForScene(piece.Part[0].Measure, sceneSection)
				clearTempoMarkings(measuresForScene)

				if len(measuresForScene) > 0 {
					addSystemText(measuresForScene[0], record.SceneDesc)
					addTimeSignature(measuresForScene[0], timeSignature)
					addTempo(measuresForScene[0], int(sceneSection.tempo))
				}
				measures = append(measures, measuresForScene...)
			}
		}
	}
	enumerateMeasuresInPlace(measures)
	return selection{measures: measures}
}

func CreateComposition(library Library, project *db.Project) *musicxml.Scorepartwise {
	result := pickMeasures(library, project.Records)
	slog.Info("Creating composition", "projectName", project.Name, "measuresCount", len(result.measures))
	composition := musicxml.Scorepartwise{
		Part: []*musicxml.Part{
			{
				Partattributes: musicxml.Partattributes{
					IdAttr: "P1",
				},
				Measure: result.measures,
			},
		},
		Scoreheader: musicxml.Scoreheader{
			Work: &musicxml.Work{
				Worktitle: project.Name,
			},
		},
	}
	return &composition
}

func beatsPerMinutes(beats string) int {
	beatsInTimeSig, err := strconv.Atoi(beats)
	if err != nil {
		slog.Warn("Using default of 4 when picking tempo.", "error", err)
		beatsInTimeSig = defaultBpm
	}
	return beatsInTimeSig
}
