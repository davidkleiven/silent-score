package compose

import (
	"time"

	"github.com/davidkleiven/silent-score/internal/musicxml"
)

type section struct {
	start int
	end   int
}

func hasRehersalMark(elements []musicxml.MusicDataElement) bool {
	for _, element := range elements {
		if direction := element.Direction; direction != nil {
			for _, dirType := range direction.Directiontype {
				if len(dirType.Rehearsal) > 0 {
					return true
				}
			}
		}
	}
	return false
}

func pieceSections(measures []musicxml.Measure) []section {
	var sections []section
	start := 0
	for i, measure := range measures {
		if hasRehersalMark(measure.MusicDataElements) && i > 0 {
			sections = append(sections, section{start: start, end: i})
			start = i
		}
	}

	if len(measures) > 0 {
		sections = append(sections, section{start: start, end: len(measures)})
	}
	return sections
}

type sceneSection struct {
	sections []section
	tempo    float64
}

func sectionForScene(duration time.Duration, targetTempo float64, beatsPerMeasure int, sections []section) sceneSection {
	targetNumberOfMeasures := int(duration.Minutes() * targetTempo / float64(beatsPerMeasure))
	currentNum := 0
	var chosenSections []section
	for i := range 1000 {
		sectionIdx := i % len(sections)
		numBars := sections[sectionIdx].end - sections[sectionIdx].start

		nextNum := currentNum + numBars
		remaining := targetNumberOfMeasures - currentNum
		overshooting := nextNum - targetNumberOfMeasures

		if overshooting < remaining {
			chosenSections = append(chosenSections, sections[sectionIdx])
			currentNum += numBars
		}

		if nextNum > targetNumberOfMeasures {
			break
		}
	}
	return sceneSection{
		sections: chosenSections,
		tempo:    float64(beatsPerMeasure) * float64(currentNum) / duration.Minutes(),
	}
}

func measuresForScene(measures []musicxml.Measure, section sceneSection) []musicxml.Measure {
	var result []musicxml.Measure
	if len(measures) == 0 {
		return result
	}

	for _, section := range section.sections {
		for i := section.start; i < section.end; i++ {
			result = append(result, *musicxml.MustDeepCopyMeasure(&measures[i]))
		}
	}
	return result
}
