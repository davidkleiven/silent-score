package musicxml

import (
	"archive/zip"
	"bytes"
	"errors"
	"io/fs"
	"testing"
)

type FailingReader struct{}

func (f *FailingReader) Read(p []byte) (n int, err error) {
	return 0, fs.ErrNotExist
}

func TestErrorPropagatedWhenReaderFails(t *testing.T) {
	_, err := Zip2MusicXMLReader(&FailingReader{})

	if !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("Expected fs.ErrNotExist, got %v", err)
	}
}

func TestInvalidZip(t *testing.T) {
	buf := bytes.NewBuffer([]byte("This is not a valid MusicXML file."))
	_, err := Zip2MusicXMLReader(buf)
	if !errors.Is(err, zip.ErrFormat) {
		t.Errorf("Expected fs.ErrNotExist, got %v", err)
	}
}

func TestErrorWhenNoMusicXml(t *testing.T) {
	writer := bytes.NewBuffer([]byte{})
	zipWriter := zip.NewWriter(writer)
	_, err := zipWriter.Create("test.txt")
	if err != nil {
		t.Error(err)
		return
	}
	if err := zipWriter.Close(); err != nil {
		t.Error(err)
		return
	}

	_, err = Zip2MusicXMLReader(writer)
	if !errors.Is(err, ErrNoMusixXMLFileInZip) {
		t.Errorf("Expected ErrNoMusixXMLFileInZip, got %v", err)
	}
}

func TestUnmarshalErrorWhenWrongXML(t *testing.T) {
	xmlString := `This is not a valid XML string`

	byteBuffer := bytes.NewBuffer([]byte{})
	zipWriter := zip.NewWriter(byteBuffer)
	fileWriter, err := zipWriter.Create("META-INF/container.xml")
	if err != nil {
		t.Error(err)
		return
	}

	fileWriter.Write([]byte(xmlString))
	zipWriter.Close()

	if byteBuffer.Len() == 0 {
		t.Error("Expected non-empty buffer, got empty")
		return
	}

	_, err = Zip2MusicXMLReader(byteBuffer)
	if err == nil {
		t.Errorf("Expected fs.ErrNotExist, got %v", err)
	}
}

func TestNoMusicXmlFileSpecifiedInContainer(t *testing.T) {
	xmlString := "<container><rootfiles></rootfiles></container>"
	byteBuffer := bytes.NewBuffer([]byte{})
	zipWriter := zip.NewWriter(byteBuffer)
	fileWriter, err := zipWriter.Create("META-INF/container.xml")
	if err != nil {
		t.Error(err)
		return
	}
	fileWriter.Write([]byte(xmlString))
	zipWriter.Close()
	if byteBuffer.Len() == 0 {
		t.Error("Expected non-empty buffer, got empty")
		return
	}
	_, err = Zip2MusicXMLReader(byteBuffer)
	if !errors.Is(err, ErrNoMusicXMLRootFile) {
		t.Errorf("Expected ErrNoMusicXMLRootFile, got %v", err)
	}
}
