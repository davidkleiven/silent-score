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
				if dirType != nil {
					if len(dirType.Rehearsal) > 0 {
						return true
					}
				}
			}
		}
	}
	return false
}

func pieceSections(measures []*musicxml.Measure) []section {
	var sections []section
	start := 0
	for i, measure := range measures {
		if measure != nil {
			if hasRehersalMark(measure.MusicDataElements) && i > 0 {
				sections = append(sections, section{start: start, end: i})
				start = i
			}
		}
	}

	if len(measures) > 0 {
		sections = append(sections, section{start: start, end: len(measures)})
	}
	return sections
}

type sceneSection struct {
	start int
	end   int
	tempo float64
}

func sectionForScene(duration time.Duration, targetTempo float64, beatsPerMeasure int, sections []section) sceneSection {
	targetNumberOfMeasures := int(duration.Minutes() * targetTempo / float64(beatsPerMeasure))
	currentNum := 0
	counter := 0
	for range 1000 {
		sectionIdx := counter % len(sections)
		numBars := sections[sectionIdx].end - sections[sectionIdx].start

		nextNum := currentNum + numBars
		remaining := targetNumberOfMeasures - currentNum
		overshooting := nextNum - targetNumberOfMeasures
		if nextNum > targetNumberOfMeasures {
			if remaining < overshooting {
				break
			} else {
				currentNum = nextNum
				break
			}
		}

		currentNum += numBars
		counter += 1
	}
	return sceneSection{
		start: sections[0].start,
		end:   currentNum,
		tempo: float64(beatsPerMeasure) * float64(currentNum) / duration.Minutes(),
	}
}

func measuresForScene(measures []*musicxml.Measure, section sceneSection) []*musicxml.Measure {
	numToAdd := section.end - section.start
	result := make([]*musicxml.Measure, 0, numToAdd)
	if len(measures) == 0 || numToAdd == 0 {
		return result
	}
	chunks := [][]*musicxml.Measure{measures[section.start:], measures[:section.start]}

	counter := 0

	for len(result) < numToAdd {
		currentChunk := chunks[counter%2]
		end := len(currentChunk)
		if len(result)+end > numToAdd {
			end = numToAdd - len(result)
		}
		for i := range end {
			result = append(result, musicxml.MustDeepCopyMeasure(currentChunk[i]))
		}
		counter += 1
	}
	return result
}
