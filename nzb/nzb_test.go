package nzb

import (
	"bytes"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/hobeone/gonab/types"
)

// Grabbed from golang.org/src/go/printer/printer_test.go
// -- should move this to a library

// lineAt returns the line in text starting at offset offs.
func lineAt(text []byte, offs int) []byte {
	i := offs
	for i < len(text) && text[i] != '\n' {
		i++
	}
	return text[offs:i]
}

// diff compares a and b.
func diff(aname, bname string, a, b []byte) error {
	var buf bytes.Buffer // holding long error message

	// compare lengths
	if len(a) != len(b) {
		fmt.Fprintf(&buf, "\nlength changed: len(%s) = %d, len(%s) = %d", aname, len(a), bname, len(b))
	}

	// compare contents
	line := 1
	offs := 1
	for i := 0; i < len(a) && i < len(b); i++ {
		ch := a[i]
		if ch != b[i] {
			fmt.Fprintf(&buf, "\n%s:%d:%d: %s", aname, line, i-offs+1, lineAt(a, offs))
			fmt.Fprintf(&buf, "\n%s:%d:%d: %s", bname, line, i-offs+1, lineAt(b, offs))
			fmt.Fprintf(&buf, "\n\n")
			break
		}
		if ch == '\n' {
			line++
			offs = i + 1
		}
	}

	if buf.Len() > 0 {
		return errors.New(buf.String())
	}
	return nil
}

var goldenOutput = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE nzb PUBLIC "-//newzBin//DTD NZB 1.1//EN" "http://www.newzbin.com/DTD/nzb/nzb-1.1.dtd">
<nzb xmlns="http://www.newzbin.com/DTD/2003/nzb">
  <head>
    <meta type="category">TV</meta>
    <meta type="name">TestBinary</meta>
  </head>
  <file poster="test@foo.bar" date="-62135596800" subject="TestBinary.r01">
    <groups>
      <group>misc.test</group>
    </groups>
    <segments>
      <segment bytes="12356" number="1">123456@foo.bar</segment>
      <segment bytes="12356" number="2">789@foo.bar</segment>
    </segments>
  </file>
  <file poster="test@foo.bar" date="-62135596800" subject="TestBinary.r02">
    <groups>
      <group>misc.test</group>
    </groups>
    <segments>
      <segment bytes="12356" number="1">456@foo.bar</segment>
      <segment bytes="12356" number="2">123@foo.bar</segment>
    </segments>
  </file>
</nzb>
`

func TestNZBCreate(t *testing.T) {
	// out of order parts and segments.
	parts := []types.Part{
		{
			Subject:   "TestBinary.r02",
			From:      "test@foo.bar",
			Posted:    time.Time{},
			GroupName: "misc.test",
			Segments: []types.Segment{
				{
					MessageID: "123@foo.bar",
					Size:      12356,
					Segment:   2,
				},
				{
					MessageID: "456@foo.bar",
					Size:      12356,
					Segment:   1,
				},
			},
		},
		{
			Subject:   "TestBinary.r01",
			From:      "test@foo.bar",
			Posted:    time.Time{},
			GroupName: "misc.test",
			Segments: []types.Segment{
				{
					MessageID: "789@foo.bar",
					Size:      12356,
					Segment:   2,
				},
				{
					MessageID: "123456@foo.bar",
					Size:      12356,
					Segment:   1,
				},
			},
		},
	}
	b := types.Binary{
		Name:  "TestBinary",
		Parts: parts,
	}
	output, err := WriteNZB(&b)
	if err != nil {
		t.Fatalf("Error creating NZB: %v", err)
	}

	err = diff("test output", "golden", []byte(output), []byte(goldenOutput))
	if err != nil {
		t.Fatal(err)
	}
}
