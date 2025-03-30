package compose

import (
	"time"

	"github.com/davidkleiven/silent-score/internal/musicxml"
)

type section struct {
	start int
	end   int
}

func hasRehersalMark(directions []*musicxml.Direction) bool {
	for _, direction := range directions {
		if direction != nil {
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
			if hasRehersalMark(measure.Direction) && i > 0 {
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
	targetNumberOfMeasures := duration.Minutes() * targetTempo / float64(beatsPerMeasure)
	currentNum := 0
	counter := 0
	for currentNum < int(targetNumberOfMeasures) {
		sectionIdx := counter % len(sections)
		numBars := sections[sectionIdx].end - sections[sectionIdx].start
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
	if len(measures) == 0 {
		return result
	}
	chunks := [][]*musicxml.Measure{measures[section.start:], measures[:section.start]}

	counter := 0

	for len(result) < numToAdd {
		for _, content := range chunks[counter%2] {
			result = append(result, musicxml.MustDeepCopyMeasure(content))
		}
		counter += 1
	}

	result[0].Direction = append(result[0].Direction, musicxml.NewDirection(musicxml.WithTempo(int(section.tempo))))
	return result
}
