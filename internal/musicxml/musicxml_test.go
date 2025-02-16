package musicxml

import (
	"encoding/xml"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestRead(t *testing.T) {
	_, currentFile, _, _ := runtime.Caller(0)
	currentDir := filepath.Dir(currentFile)
	testData := filepath.Join(currentDir, "../../test/data/testScore.musicxml")
	data, err := os.ReadFile(testData)
	if err != nil {
		t.Error(err)
		return
	}

	var document Scorepartwise
	err = xml.Unmarshal(data, &document)
	if err != nil {
		t.Error(err)
		return
	}
}
