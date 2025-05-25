package musicxml

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"errors"
	"io"
)

var ErrNoMusixXMLFileInZip = errors.New("no MusicXML file found in zip archive")
var ErrNoMusicXMLRootFile = errors.New("no MusicXML root file found in container.xml")

func Zip2MusicXMLReader(zipReader io.Reader) (io.Reader, error) {
	buf, err := io.ReadAll(zipReader)
	if err != nil {
		return zipReader, err
	}
	r, err := zip.NewReader(bytes.NewReader(buf), int64(len(buf)))
	if err != nil {
		return zipReader, err
	}

	// Search for a MusicXML file in the zip archive
	var musicXmlFile string
	for _, file := range r.File {
		if file.Name == "META-INF/container.xml" {
			f, err := file.Open()
			if err != nil {
				return zipReader, err
			}
			defer f.Close()
			content, err := io.ReadAll(f)
			if err != nil {
				return zipReader, err
			}

			var container Container
			if err := xml.Unmarshal(content, &container); err != nil {
				return zipReader, err
			}

			if len(container.RootFileList.Files) == 0 {
				return zipReader, ErrNoMusicXMLRootFile
			}

			// According to https://www.w3.org/2021/06/musicxml40/tutorial/compressed-mxl-files/
			// the first root file is the MusicXML file.
			musicXmlFile = container.RootFileList.Files[0].FullPathAttr
			break
		}
	}

	// Open the musicxml file
	for _, file := range r.File {
		if file.Name == musicXmlFile {
			f, err := file.Open()
			if err != nil {
				return zipReader, err
			}
			defer f.Close()
			content, err := io.ReadAll(f)
			if err != nil {
				return zipReader, err
			}
			return bytes.NewReader(content), nil
		}
	}
	return zipReader, ErrNoMusixXMLFileInZip
}
