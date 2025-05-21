package compose

import (
	"embed"
	"fmt"
	"io/fs"
	"iter"
	"os"
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
	defaultTempo      = 82
	defaultBpm        = 4
	numBarsInCueSheet = 4
)

var beatUnitMap = map[string]int{
	"quarter": 4,
	"eighth":  8,
	"half":    2,
	"whole":   1,
	"16th":    16,
	"32nd":    32,
	"64th":    64,
}

//go:embed assets/*.musicxml
var standardLib embed.FS

type FileNameProvider interface {
	Names() []string
	Fs() fs.FS
}

type matchResult struct {
	score      *musicxml.Scorepartwise
	similarity int
}

type StandardLibraryFileNameProvider struct {
	directory string
}

func (s *StandardLibraryFileNameProvider) Fs() fs.FS {
	return standardLib
}

func (s *StandardLibraryFileNameProvider) Names() []string {
	entries, err := standardLib.ReadDir(s.directory)
	if err != nil {
		slog.Error("Failed to read standard library directory", "error", err)
		return []string{}
	}
	names := make([]string, len(entries))
	for i, entry := range entries {
		names[i] = s.directory + "/" + entry.Name()
	}
	slog.Info("Standard library loaded", "count", len(names))
	return names
}

func NewStandardLibraryFileNameProvider() *StandardLibraryFileNameProvider {
	return &StandardLibraryFileNameProvider{directory: "assets"}
}

type LocalFileNameProvider struct {
	fs fs.FS
}

func (s *LocalFileNameProvider) Names() []string {
	entries, err := fs.Glob(s.fs, "*.musicxml")
	if err != nil {
		slog.Error("Failed to read local library directory", "error", err)
		return []string{}
	}
	slog.Info("Local library loaded", "count", len(entries))
	return entries
}

func (s *LocalFileNameProvider) Fs() fs.FS {
	return s.fs
}

func NewLocalLibraryFileNameProvider(directory string) *LocalFileNameProvider {
	return &LocalFileNameProvider{fs: os.DirFS(directory)}
}

type Library interface {
	BestMatch(desc string) matchResult
}

type FsLibrary struct {
	nameProvider FileNameProvider
}

func NewStandardLibrary() *FsLibrary {
	return &FsLibrary{nameProvider: NewStandardLibraryFileNameProvider()}
}

func NewLocalLibrary(directory string) *FsLibrary {
	return &FsLibrary{nameProvider: NewLocalLibraryFileNameProvider(directory)}
}

func (sl *FsLibrary) scores() iter.Seq[*musicxml.Scorepartwise] {
	names := sl.nameProvider.Names()
	return func(yield func(item *musicxml.Scorepartwise) bool) {
		for _, name := range names {
			score := musicxml.ReadFromFileName(sl.nameProvider.Fs(), name)
			if !yield(&score) {
				break
			}
		}
	}
}

func (sl *FsLibrary) BestMatch(desc string) matchResult {
	texts := collectTextFields(sl.scores())
	bestMatch := bestMatchForDesc(desc, texts)
	name := sl.nameProvider.Names()[bestMatch.Index]
	score := musicxml.ReadFromFileName(sl.nameProvider.Fs(), name)
	return matchResult{
		score:      &score,
		similarity: bestMatch.Similarity}
}

type InMemoryLibrary struct {
	scores []*musicxml.Scorepartwise
}

func (l *InMemoryLibrary) BestMatch(desc string) matchResult {
	texts := collectTextFields(slices.Values(l.scores))
	result := bestMatchForDesc(desc, texts)
	return matchResult{
		score:      l.scores[result.Index],
		similarity: result.Similarity,
	}
}

func collectTextFields(scoreIter iter.Seq[*musicxml.Scorepartwise]) []string {
	var texts []string
	for score := range scoreIter {
		texts = append(texts, strings.Join(musicxml.TextFields(*score), " "))
	}
	return texts
}

func bestMatchForDesc(desc string, texts []string) Score {
	normalizedDesc := normalize(desc)
	normalizedTexts := make([]string, len(texts))
	for i, text := range texts {
		normalizedTexts[i] = normalize(text)
	}
	return orderPieces(normalizedDesc, normalizedTexts)[0]
}

func tempoIfGiven(tempo int, measures []musicxml.Measure) *musicxml.Metronome {
	var metronome *musicxml.Metronome
	for _, measure := range measures {
		for _, element := range measure.MusicDataElements {
			if direction := element.Direction; direction != nil {
				for _, dirType := range direction.Directiontype {
					if dirType.Metronome != nil {
						metronome = dirType.Metronome
						break
					}
				}
			}

			if metronome != nil {
				break
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
		metronome.Perminute.Value = tempo
	}
	return metronome
}

func defaultMetronome() *musicxml.Metronome {
	return &musicxml.Metronome{
		Perminute: &musicxml.Perminute{
			Value: 82,
		},
		Beatunit: musicxml.Beatunit{
			Beatunit: "quarter",
		},
	}
}

func timesignature(measures []musicxml.Measure) *musicxml.Timesignature {
	for _, measure := range measures {
		for _, element := range measure.MusicDataElements {
			if attr := element.Attributes; attr != nil {
				for _, timesig := range attr.Time {
					return &timesig
				}
			}
		}
	}
	return &musicxml.Timesignature{
		Beats:    4,
		Beattype: 4,
	}
}

func clearTempoMarkings(measures []musicxml.Measure) {
	for measNo, measure := range measures {
		for i, element := range measure.MusicDataElements {
			if direction := element.Direction; direction != nil {
				for j, dirType := range direction.Directiontype {
					if dirType.Metronome != nil {
						measures[measNo].MusicDataElements[i].Direction.Directiontype[j].Metronome = nil
					}
				}
			}
		}
	}
}

func enumerateMeasuresInPlace(measures []musicxml.Measure) {
	for i := range measures {
		measures[i].NumberAttr = strconv.Itoa(i + 1)
	}
}

type selection struct {
	measures []musicxml.Measure
	pieces   []pieceInfo
}

func (s *selection) composer() string {
	c := ""
	for _, piece := range s.pieces {
		if c != "" {
			c += ", "
		}
		c += piece.composer
	}
	return c
}

type pieceInfo struct {
	title    string
	composer string
	cue      []musicxml.Measure
}

func firstN(measures []musicxml.Measure, n int) []musicxml.Measure {
	if n > len(measures) {
		n = len(measures)
	}

	result := make([]musicxml.Measure, n)
	for i, measure := range measures[:n] {
		result[i] = *musicxml.MustDeepCopyMeasure(&measure)
	}
	return result
}

func title(score *musicxml.Scorepartwise) string {
	if score != nil && score.Scoreheader.Work != nil {
		return score.Scoreheader.Work.Worktitle
	}
	return ""
}

func composer(score *musicxml.Scorepartwise) string {
	if score != nil {
		for _, credit := range score.Scoreheader.Credit {
			if slices.Contains(credit.Credittype, "composer") {
				return credit.Creditwords.Value
			}
		}
	}
	return ""
}

func pickMeasures(library Library, records []db.ProjectContentRecord) selection {
	var measures []musicxml.Measure
	var pieces []pieceInfo
	scoresByTheme := make(map[uint]matchResult)
	for _, record := range records {
		bm, ok := scoresByTheme[record.Theme]
		if !ok {
			bm = library.BestMatch(record.Keywords)
		}

		if record.Theme > 0 {
			scoresByTheme[record.Theme] = bm
		}
		piece := bm.score

		if piece != nil {
			if len(piece.Part) > 0 {
				measuresWithNoRepeats := removeRepetitions(piece.Part[0].Measure)
				sections := pieceSections(measuresWithNoRepeats)
				slog.Info("Extracted sections", "title", title(piece), "num-sections", len(sections))
				timeSignature := timesignature(measuresWithNoRepeats)
				metronome := tempoIfGiven(int(record.Tempo), measuresWithNoRepeats)
				beatsInTimeSig := beatsPerMeasure(timeSignature, metronome)
				sceneSection := sectionForScene(time.Duration(record.DurationSec)*time.Second, float64(metronome.Perminute.Value), beatsInTimeSig, sections)

				// Update tempo with result from scence selection
				metronome.Perminute.Value = int(sceneSection.tempo)
				measuresForScene := measuresForScene(measuresWithNoRepeats, sceneSection)
				clearTempoMarkings(measuresForScene)

				if len(measuresForScene) > 0 {
					musicxml.SetSystemTextAtBeginning(&measuresForScene[0], record.SceneDesc)
					musicxml.SetTimeSignatureAtBeginning(&measuresForScene[0], *timeSignature)
					musicxml.SetTempoAtBeginning(&measuresForScene[0], metronome)
					barline := musicxml.NewBarline(musicxml.WithBarStyle(musicxml.BarStyleLightLight))
					musicxml.SetBarlineAtEnd(&measuresForScene[len(measuresForScene)-1], barline)
				}
				measures = append(measures, measuresForScene...)
				slog.Info("Picking piece",
					"keywords", record.Keywords,
					"sceneDesc", record.SceneDesc,
					"similarity-score", bm.similarity,
					"title", title(piece),
					"timeSignature", fmt.Sprintf("%d/%d", timeSignature.Beats, timeSignature.Beattype),
					"tempo", metronome.Perminute.Value,
				)

				pieces = append(pieces, pieceInfo{
					title:    title(piece),
					composer: composer(piece),
					cue:      firstN(measuresForScene, numBarsInCueSheet),
				},
				)
			}
		}
	}
	enumerateMeasuresInPlace(measures)

	// Add barline at very end
	barline := musicxml.NewBarline(musicxml.WithBarStyle(musicxml.BarStyleLightHeavy))
	if len(measures) > 0 {
		musicxml.SetBarlineAtEnd(&measures[len(measures)-1], barline)
	}

	removeRedundantClefs(measures)
	return selection{measures: measures, pieces: pieces}
}

func CreateComposition(library Library, project *db.Project) *musicxml.Scorepartwise {
	result := pickMeasures(library, project.Records)
	slog.Info("Creating composition", "projectName", project.Name, "measuresCount", len(result.measures))

	// Insert page breaks and line breaks
	var allMeasures []musicxml.Measure

	if len(result.measures) > 0 {
		result.measures[0].MusicDataElements = ensurePageBreak(result.measures[0].MusicDataElements)
	}
	for i := 0; i < len(result.pieces); i++ {
		if len(result.pieces[i].cue) == 0 {
			continue
		}
		for j := 0; j < len(result.pieces[i].cue); j++ {
			result.pieces[i].cue[j].MusicDataElements = clearPrint(result.pieces[i].cue[j].MusicDataElements)
		}
		var elements []musicxml.MusicDataElement
		if i == 0 {
			elements = ensurePageBreak(result.pieces[i].cue[0].MusicDataElements)
		} else {
			elements = ensureLineBreak(result.pieces[i].cue[0].MusicDataElements)
		}

		result.pieces[i].cue[0].MusicDataElements = elements
	}

	// Merge all measures
	for _, piece := range result.pieces {
		allMeasures = append(allMeasures, piece.cue...)
	}
	allMeasures = append(allMeasures, result.measures...)
	enumerateMeasuresInPlace(allMeasures)

	composition := musicxml.Scorepartwise{
		Documentattributes: musicxml.Documentattributes{
			VersionAttr: "4.0",
		},
		Part: []musicxml.Part{
			{
				Partattributes: musicxml.Partattributes{
					IdAttr: "P1",
				},
				Measure: allMeasures,
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
			Defaults: &musicxml.Defaults{
				Scaling: &musicxml.Scaling{
					Millimeters: 6.99912,
					Tenths:      40,
				},
				Pagelayout: &musicxml.Pagelayout{
					Pageheight:  1596.77,
					Pagewidth:   1233.87,
					Pagemargins: []musicxml.Pagemargins{*musicxml.DefaultPageMargins("even"), *musicxml.DefaultPageMargins("odd")},
				},
			},
			Credit: []musicxml.Credit{
				{
					PageAttr:    1,
					Credittype:  []string{"title"},
					Creditwords: musicxml.TitleElement(project.Name),
				},
				{
					PageAttr:    1,
					Credittype:  []string{"composer"},
					Creditwords: musicxml.ComposerElement(result.composer()),
				},
			},
		},
	}
	return &composition
}

func beatsPerMeasure(timesignature *musicxml.Timesignature, metronome *musicxml.Metronome) int {
	unit, num := unitsPerBeat(metronome)
	return timesignature.Beats * unit / (timesignature.Beattype * num)
}

func unitsPerBeat(metronome *musicxml.Metronome) (int, int) {
	unit, ok := beatUnitMap[metronome.Beatunit.Beatunit]
	if !ok {
		slog.Warn("Unknown beat unit. Ensure tempo mark is given in the score", "unit", metronome.Beatunit.Beatunit)
		unit = 4
	}
	num := 1
	for _ = range metronome.Beatunit.Beatunitdot {
		unit *= 2
		num *= 2
		num++
	}
	return unit, num
}

func removeRepetitions(measures []musicxml.Measure) []musicxml.Measure {
	result := make([]musicxml.Measure, 0, len(measures))

	takeMeasures := true
	for _, measure := range measures {
		if firstEndingStarts(&measure) {
			takeMeasures = false
		}
		if firstEndingEnds(&measure) {
			takeMeasures = true
			continue
		}

		if takeMeasures {
			result = append(result, *musicxml.MustDeepCopyMeasure(&measure))
		}
	}
	result = clearEndings(result)
	return clearRepeatSigns(result)
}

func firstEndingStarts(measure *musicxml.Measure) bool {
	for _, element := range measure.MusicDataElements {
		if element.Barline != nil && element.Barline.Ending != nil && element.Barline.Ending.NumberAttr == "1" && element.Barline.Ending.TypeAttr == "start" {
			return true
		}
	}
	return false
}

func firstEndingEnds(measure *musicxml.Measure) bool {
	for _, element := range measure.MusicDataElements {
		if element.Barline != nil && element.Barline.Ending != nil && element.Barline.Ending.NumberAttr == "1" && element.Barline.Ending.TypeAttr == "stop" {
			return true
		}
	}
	return false
}

func clearEndings(measures []musicxml.Measure) []musicxml.Measure {
	for i := range measures {
		for j := range measures[i].MusicDataElements {
			if measures[i].MusicDataElements[j].Barline != nil && measures[i].MusicDataElements[j].Barline.Ending != nil {
				measures[i].MusicDataElements[j].Barline = nil
			}
		}
	}
	return measures
}

func clearRepeatSigns(measures []musicxml.Measure) []musicxml.Measure {
	for i := range measures {
		for j := range measures[i].MusicDataElements {
			if measures[i].MusicDataElements[j].Barline != nil && measures[i].MusicDataElements[j].Barline.Repeat != nil {
				measures[i].MusicDataElements[j].Barline.Repeat = nil
			}
		}
	}
	return measures
}

func ensurePageBreak(elements []musicxml.MusicDataElement) []musicxml.MusicDataElement {
	for i := range elements {
		if elements[i].Print != nil {
			elements[i].Print.NewpageAttr = "yes"
			return elements
		}
	}
	return append([]musicxml.MusicDataElement{musicxml.NewPage()}, elements...)
}

func ensureLineBreak(elements []musicxml.MusicDataElement) []musicxml.MusicDataElement {
	for i := range elements {
		if elements[i].Print != nil {
			elements[i].Print.NewsystemAttr = "yes"
			return elements
		}
	}
	return append([]musicxml.MusicDataElement{musicxml.NewSystem()}, elements...)
}

func clearPrint(elements []musicxml.MusicDataElement) []musicxml.MusicDataElement {
	return slices.DeleteFunc(elements, func(element musicxml.MusicDataElement) bool {
		return element.Print != nil
	})
}

func removeRedundantClefs(measures []musicxml.Measure) {
	var currentClef *musicxml.Clef
	for i := range measures {
		for j, elem := range measures[i].MusicDataElements {
			if elem.Attributes != nil {
				for k := 0; k < len(elem.Attributes.Clef); k++ {
					if musicxml.ClefEquals(&elem.Attributes.Clef[k], currentClef) {
						measures[i].MusicDataElements[j].Attributes.Clef = slices.Delete(measures[i].MusicDataElements[j].Attributes.Clef, k, k+1)
						k--
					} else {
						currentClef = &elem.Attributes.Clef[k]
					}
				}
			}
		}
	}
}
