package musicxml

import (
	"encoding/xml"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func testScore() (Scorepartwise, error) {
	_, currentFile, _, _ := runtime.Caller(0)
	currentDir := filepath.Dir(currentFile)
	testData := filepath.Join(currentDir, "../../test/data/testScore.musicxml")
	data, err := os.ReadFile(testData)
	var document Scorepartwise
	if err != nil {
		return document, err
	}

	err = xml.Unmarshal(data, &document)
	return document, err
}

func TestRead(t *testing.T) {
	document, err := testScore()
	if err != nil {
		t.Error(err)
		return
	}

	fields := TextFields(document)
	fieldMap := make(map[string]bool)
	for _, v := range fields {
		fieldMap[v] = true
	}

	expect := []string{"Unknown composer", "Test score", "Score subtitle", "Calmly", "Agitato", "A"}
	for _, v := range expect {
		if _, ok := fieldMap[v]; !ok {
			t.Errorf("%s not in %v\n", v, fieldMap)
			return
		}
	}
}
