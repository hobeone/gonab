package nzb

import (
	"bytes"
	"encoding/xml"
	"strings"

	"github.com/hobeone/gonab/types"
)

// NZB is the top level structure
type NZB struct {
	XMLName  xml.Name `xml:"nzb"`
	Xmlns    string   `xml:"xmlns,attr,omitempty"`
	Metadata []Meta   `xml:"head>meta"`
	Files    []File   `xml:"file"` // xml:tag name doesn't work?
}

// Meta information for the header
type Meta struct {
	Type  string `xml:"type,attr"`
	Value string `xml:",innerxml"`
}

// File wraps the file structure in the nzb xml
type File struct {
	Groups   []string  `xml:"groups>group"`
	Segments []Segment `xml:"segments>segment"`
	Poster   string    `xml:"poster,attr"`
	Date     int64     `xml:"date,attr"`
	Subject  string    `xml:"subject,attr"`
	Part     int       `xml:"-"`
}

// Segment represents each segment of a file
type Segment struct {
	XMLName xml.Name `xml:"segment"`
	Bytes   int64    `xml:"bytes,attr"`
	Number  int      `xml:"number,attr"`
	ID      string   `xml:",innerxml"`
}

var nzbHeader = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE nzb PUBLIC "-//newzBin//DTD NZB 1.1//EN" "http://www.newzbin.com/DTD/nzb/nzb-1.1.dtd">
`

// WriteNZB takes a Binary and returns a NZB document as a string.
func WriteNZB(b *types.Binary) (string, error) {
	nz := NZB{
		Xmlns: "http://www.newzbin.com/DTD/2003/nzb",
	}
	nz.Metadata = []Meta{
		{
			Type:  "category",
			Value: "TV",
		},
		{
			Type:  "name",
			Value: b.Name,
		},
	}

	nz.Files = make([]File, len(b.Parts))
	for i, part := range b.Parts {
		segs := make([]Segment, len(part.Segments))
		for i, seg := range part.Segments {
			msgid := strings.TrimRight(seg.MessageID, ">")
			msgid = strings.TrimLeft(msgid, "<")
			segs[i] = Segment{
				Bytes:  seg.Size,
				Number: seg.Segment,
				ID:     msgid,
			}
		}
		nz.Files[i] = File{
			Subject:  part.Subject,
			Segments: segs,
			Poster:   part.From,
			Date:     part.Posted.Unix(),
			Groups:   []string{part.Group},
		}
	}

	xmlWriter := bytes.NewBufferString("")
	xmlWriter.WriteString(nzbHeader)

	enc := xml.NewEncoder(xmlWriter)
	enc.Indent("", "  ")
	if err := enc.Encode(nz); err != nil {
		return "", err
	}
	return xmlWriter.String() + "\n", nil
}
