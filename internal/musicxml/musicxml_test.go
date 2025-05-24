package musicxml

import (
	"bytes"
	"encoding/xml"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/ucarion/c14n"
)

func testFile(name string) string {
	_, currentFile, _, _ := runtime.Caller(0)
	currentDir := filepath.Dir(currentFile)
	return filepath.Join(currentDir, "../../test/data/", name)
}

func testScoreBytes() ([]byte, error) {
	testData := testFile("testScore.musicxml")
	return os.ReadFile(testData)
}

func testScore() (Scorepartwise, error) {
	data, err := testScoreBytes()
	if err != nil {
		return Scorepartwise{}, err
	}
	return ReadFromFile(bytes.NewReader(data))
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

func TestRoundTrip(t *testing.T) {
	data, err := testScoreBytes()
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

	serializedBytes, err := xml.MarshalIndent(document, "", "  ")
	if err != nil {
		t.Error(err)
		return
	}

	// Canonicalize the XML
	decoder := xml.NewDecoder(bytes.NewBuffer(serializedBytes))
	canonicalBytes, err := c14n.Canonicalize(decoder)
	if err != nil {
		t.Error(err)
		return
	}

	// Add newline character at end
	canonicalBytes = append(canonicalBytes, '\n')

	if !bytes.Equal(canonicalBytes, data) {
		first := 0
		for i := range canonicalBytes {
			if canonicalBytes[i] != data[i] {
				first = i
				break
			}
		}
		t.Errorf("Serialized bytes differ from original bytes. First differeing byte %d", first)
	}
}
