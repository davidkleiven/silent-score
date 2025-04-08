package compose

import (
	"embed"
	"fmt"
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

func tempoIfGiven(tempo int, measures []*musicxml.Measure) *musicxml.Metronome {
	var metronome *musicxml.Metronome
	for _, measure := range measures {
		if measure != nil {
			for _, element := range measure.MusicDataElements {
				if direction := element.Direction; direction != nil {
					for _, dirType := range direction.Directiontype {
						if dirType != nil && dirType.Metronome != nil {
							metronome = dirType.Metronome
							break
						}
					}
				}

				if metronome != nil {
					break
				}
			}
		}

		if metronome != nil {
			break
		}
	}

	if metronome == nil {
		metronome = defaultMetronome()
	}

	if tempo > 0 {
		metronome.Perminute.Value = strconv.Itoa(tempo)
	}
	return metronome
}

func defaultMetronome() *musicxml.Metronome {
	return &musicxml.Metronome{
		Perminute: &musicxml.Perminute{
			Value: "82",
		},
		Beatunit: &musicxml.Beatunit{
			Beatunit: "quarter",
		},
	}
}

func beatsPerMeasure(measures []*musicxml.Measure) *musicxml.Timesignature {
	for _, measure := range measures {
		if measure != nil {
			for _, element := range measure.MusicDataElements {
				if attr := element.Attributes; attr != nil && attr.Time != nil {
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

func clearTempoMarkings(measures []*musicxml.Measure) {
	for _, measure := range measures {
		if measure != nil {
			for i, element := range measure.MusicDataElements {
				if direction := element.Direction; direction != nil {
					for j, dirType := range direction.Directiontype {
						if dirType != nil && dirType.Metronome != nil {
							measure.MusicDataElements[i].Direction.Directiontype[j].Metronome = nil
						}
					}
				}
			}
		}
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
		if piece != nil {
			if len(piece.Part) > 0 {
				sections := pieceSections(piece.Part[0].Measure)
				timeSignature := beatsPerMeasure(piece.Part[0].Measure)
				beatsInTimeSig := beatsPerMinutes(timeSignature.Beats)
				metronome := tempoIfGiven(int(record.Tempo), piece.Part[0].Measure)
				tempo, err := strconv.Atoi(metronome.Perminute.Value)
				if err != nil {
					slog.Warn("Using default of 82 when picking tempo.", "error", err)
					tempo = defaultTempo
				}

				sceneSection := sectionForScene(time.Duration(record.DurationSec)*time.Second, float64(tempo), beatsInTimeSig, sections)

				// Update tempo with result from scence selection
				metronome.Perminute.Value = strconv.Itoa(int(sceneSection.tempo))
				measuresForScene := measuresForScene(piece.Part[0].Measure, sceneSection)
				clearTempoMarkings(measuresForScene)

				if len(measuresForScene) > 0 {
					musicxml.SetSystemTextAtBeginning(measuresForScene[0], record.SceneDesc)
					musicxml.SetTimeSignatureAtBeginning(measuresForScene[0], timeSignature)
					musicxml.SetTempoAtBeginning(measuresForScene[0], metronome)
				}
				measures = append(measures, measuresForScene...)
				slog.Info("Picking piece",
					"keywords", record.Keywords,
					"sceneDesc", record.SceneDesc,
					"title", title(piece),
					"timeSignature", fmt.Sprintf("%s/%s", timeSignature.Beats, timeSignature.Beattype),
					"tempo", tempo,
				)
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
		Documentattributes: musicxml.Documentattributes{
			VersionAttr: "4.0",
		},
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
			Partlist: &musicxml.Partlist{
				Scorepart: &musicxml.Scorepart{
					IdAttr:           "P1",
					Partname:         &musicxml.Partname{Value: "Piano"},
					Partabbreviation: &musicxml.Partname{Value: "Pno."},
				},
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
