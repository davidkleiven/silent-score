package musicxml

import (
	"encoding/xml"
	"testing"
)

func TestParseExample(t *testing.T) {
	// Example from https://www.w3.org/2021/06/musicxml40/tutorial/compressed-mxl-files/
	xmlString := `
<container>
  <rootfiles>
    <rootfile full-path="Dichterliebe01.musicxml" media-type="application/vnd.recordare.musicxml+xml"/>
  </rootfiles>
</container>`

	var container Container
	if err := xml.Unmarshal([]byte(xmlString), &container); err != nil {
		t.Error(err)
		return
	}
	if len(container.RootFileList.Files) != 1 {
		t.Errorf("Expected 1 root file, got %d", len(container.RootFileList.Files))
		return
	}
	if container.RootFileList.Files[0].FullPathAttr != "Dichterliebe01.musicxml" {
		t.Errorf("Expected full-path to be 'Dichterliebe01.musicxml', got '%s'", container.RootFileList.Files[0].FullPathAttr)
		return
	}

}
