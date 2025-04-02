package main

import (
	"cmp"
	"log"
	"os"
	"slices"

	"github.com/beevik/etree"
)

type SortTask struct {
	Name      string
	NodeOrder []string
}

// The purpose of this file is to fix things where test data from for example MuseScore
// does not comply with the MusicXML schema. It is not desired to use the XML-deserialization
// routines within silent-score directly as that may skip information because of mis-configuratoin
// Therefore, an element tree is used as part of the linting process
func main() {
	if len(os.Args) != 3 {
		log.Fatalf("Expect two argument got %d", len(os.Args))
	}

	fname := os.Args[1]
	outfile := os.Args[2]
	doc := etree.NewDocument()
	if err := doc.ReadFromFile(fname); err != nil {
		log.Fatal(err)
	}

	sortTasks := []SortTask{
		{
			Name:      "measure",
			NodeOrder: []string{"note", "backup", "forward", "direction", "attributes", "harmony", "figured-bass", "print", "sound", "listening", "barline", "grouping", "link", "bookmark"},
		},
		{
			Name:      "note",
			NodeOrder: []string{"chord", "pitch", "unpitched", "rest", "duration", "grace", "tie", "cue", "instrument", "type", "dot", "accidental", "time-modification", "stem", "notehead", "notehead-text", "beam", "notations", "lyric", "play", "listen"},
		},
	}

	for _, sortTask := range sortTasks {
		performSortTask(doc, sortTask)
	}
	if err := doc.WriteToFile(outfile); err != nil {
		log.Fatal(err)
	}
}

func performSortTask(doc *etree.Document, task SortTask) {
	order := make(map[string]int)
	for i, name := range task.NodeOrder {
		order[name] = i
	}
	for _, element := range doc.FindElements("//measure") {
		slices.SortStableFunc(element.Child, func(e1, e2 etree.Token) int {
			elem1, ok1 := e1.(*etree.Element)
			elem2, ok2 := e2.(*etree.Element)
			if !ok1 || !ok2 {
				return 0
			}
			return cmp.Compare(order[elem1.Tag], order[elem2.Tag])
		})
		element.ReindexChildren()
	}
}
