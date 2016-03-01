package db

import (
	"fmt"
	"testing"
	"time"

	"github.com/hobeone/gonab/types"
	. "github.com/onsi/gomega"
)

func TestNameCleaner(t *testing.T) {
	RegisterTestingT(t)
	r := &types.Regex{
		Regex:      `(?i)^(?P<name>.*?\]) \[(?P<parts>\d{1,3}\/\d{1,3})`,
		GroupRegex: `.*`,
	}
	nc := NewNameCleaner([]*types.Regex{r})

	subj := `Long Show Name - 01 [1080p] [01/20] - "Long Show Name - 01.mkv.rar" yEnc`
	match := nc.Clean(subj, "misc.test")
	Expect(match).Should(Equal("Long Show Name - 01 [1080p]01/20"))

	nc = NewNameCleaner([]*types.Regex{})
	match = nc.Clean(subj, "misc.test")
	Expect(match).Should(Equal("Long Show Name - 01 [1080p] - Long Show Name - yEnc"))
}

// nzedb uses match0, match1... instead of parts and name for their capture
// groups.
func TestMatchPartWithNzedbRegex(t *testing.T) {
	RegisterTestingT(t)
	r := &types.Regex{
		Regex:      `^\[\d+\/\d+ (?P<match1>.+?)(\.(part\d*|rar|avi|iso|mp4|mkv|mpg))?(\d{1,3}\.rev"|\.vol.+?"|\.[A-Za-z0-9]{2,4}"|") yEnc$`,
		GroupRegex: `.*`,
	}
	nc := NewNameCleaner([]*types.Regex{r})

	subj := `[04/20] Geroellheimer - S03E19 - Freudige Überraschung Geroellheimer - S03E19 - Freudige Überraschung.mp4.004" yEnc`

	match := nc.Clean(subj, "misc.text")
	Expect(match).Should(Equal("Geroellheimer - S03E19 - Freudige Überraschung Geroellheimer - S03E19 - Freudige Überraschung yEnc"))
}

/*
*
* Need to setup fixtures for this to really work
* Regexes
* Segments
* Parts
 */
func TestMakeBinaries(t *testing.T) {
	RegisterTestingT(t)
	dbh := NewMemoryDBHandle(false, false)
	err := loadFixtures(dbh)
	if err != nil {
		t.Fatalf("Error creating fixtures: %v", err)
	}

	err = loadRegexFixtures(dbh)
	if err != nil {
		t.Fatalf("Error saving regex: %v", err)
	}

	err = dbh.MakeBinaries()
	if err != nil {
		t.Fatalf("Error creating binaries: %v", err)
	}

	bin, err := dbh.FindBinaryByName(`Long Show Name - 01 [1080p] [01/01] - "Long Show Name - 01.mkv.rar" yEnc`)

	if err != nil {
		t.Fatalf("Error finding expected binary: %v", err)
	}
	Expect(bin.TotalParts).To(Equal(1))

	//TODO: make binaries where binary exists and new parts are added.
}

func loadRegexFixtures(dbh *Handle) error {
	r := `(?i)^(?P<name>.*?\]) \[(?P<parts>\d{1,3}\/\d{1,3})`

	regex := &types.Regex{
		GroupRegex: ".*",
		Regex:      r,
		Ordinal:    1,
	}
	return dbh.DB.Save(regex).Error
}

func loadFixtures(dbh *Handle) error {
	segments := []types.Segment{}
	seg1 := types.Segment{
		Segment:   1,
		Size:      1024,
		MessageID: fmt.Sprintf("%s@news.astraweb.com", randString()),
	}
	seg2 := types.Segment{
		Segment:   2,
		Size:      1024,
		MessageID: fmt.Sprintf("%s@news.astraweb.com", randString()),
	}

	segments = append(segments, seg1, seg2)

	part := types.Part{
		Hash:          "abcdefg1234",
		Subject:       `Long Show Name - 01 [1080p] [01/01] - "Long Show Name - 01.mkv.rar" yEnc`,
		TotalSegments: 2,
		Posted:        time.Now(),
		From:          "foo@bar.com",
		Xref:          "12345 news.giganews.com",
		GroupName:     "misc.test",
		Segments:      segments,
	}

	err := dbh.DB.Save(&part).Error
	if err != nil {
		return err
	}
	g := types.Group{
		Name: "misc.test",
	}
	err = dbh.DB.Save(&g).Error
	return err
}
