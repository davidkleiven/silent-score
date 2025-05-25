package musicxml

type Container struct {
	RootFileList RootFiles `xml:"rootfiles"`
}

type RootFiles struct {
	Files []RootFile `xml:"rootfile"`
}

type RootFile struct {
	FullPathAttr  string `xml:"full-path,attr"`
	MediaTypeAttr string `xml:"media-type,attr"`
}
