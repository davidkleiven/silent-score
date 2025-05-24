package musicxml

import (
	"archive/zip"
	"bytes"
	"errors"
	"io"
	"strings"
)

var ErrNoMusixXMLFileInZip = errors.New("no MusicXML file found in zip archive")

func Zip2MusicXMLReader(zipReader io.Reader) (io.Reader, error) {
	buf, err := io.ReadAll(zipReader)
	if err != nil {
		return zipReader, err
	}
	r, err := zip.NewReader(bytes.NewReader(buf), int64(len(buf)))
	if err != nil {
		return zipReader, err
	}

	for _, file := range r.File {
		if strings.HasSuffix(file.Name, ".musicxml") {
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
