package db

import (
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/hobeone/gonab/types"
	. "github.com/onsi/gomega"
)

func TestMatchPart(t *testing.T) {
	RegisterTestingT(t)
	r := regexp.MustCompile(`(?i)^(?P<name>.*?\]) \[(?P<parts>\d{1,3}\/\d{1,3})`)

	regex := &types.RegexpUtil{
		Regex: r,
	}

	part := &types.Part{
		Subject: `Long Show Name - 01 [1080p] [01/20] - "Long Show Name - 01.mkv.rar" yEnc`,
	}

	match, err := matchPart(regex, part)
	if err != nil {
		t.Fatalf("Error matching: %v", err)
	}
	Expect(match).Should(HaveKeyWithValue("name", "Long Show Name - 01 [1080p]"))
	Expect(match).Should(HaveKeyWithValue("parts", "01/20"))

	// Default part matcher
	r = regexp.MustCompile(`(?i)^(?P<name>.*?\]) `)
	match, err = matchPart(regex, part)
	if err != nil {
		t.Fatalf("Error matching: %v", err)
	}
	Expect(match).Should(HaveKeyWithValue("parts", "01/20"))
}

// nzedb uses match0, match1... instead of parts and name for their capture
// groups.
func TestMatchPartWithNzedbRegex(t *testing.T) {
	RegisterTestingT(t)
	r := regexp.MustCompile(`^\[\d+\/\d+ (?P<match1>.+?)(\.(part\d*|rar|avi|iso|mp4|mkv|mpg))?(\d{1,3}\.rev"|\.vol.+?"|\.[A-Za-z0-9]{2,4}"|") yEnc$`)
	regex := &types.RegexpUtil{
		Regex: r,
	}

	part := &types.Part{
		Subject: `[04/20 Geroellheimer - S03E19 - Freudige ?berraschung Geroellheimer - S03E19 - Freudige ?berraschung.mp4.004" yEnc`,
	}

	match, err := matchPart(regex, part)
	if err != nil {
		t.Fatalf("Error matching: %v", err)
	}
	Expect(match).Should(HaveKeyWithValue("parts", "04/20"))
	Expect(match).Should(HaveKeyWithValue("name", "Geroellheimer - S03E19 - Freudige ?berraschung Geroellheimer - S03E19 - Freudige ?berraschung"))
}

func TestGetRegexesForGroups(t *testing.T) {
	RegisterTestingT(t)
	groups := []string{"misc.test"}
	dbh := NewMemoryDBHandle(true)
	r := types.Regex{
		GroupName: `.*`,
	}
	dbh.DB.Save(&r)

	regs, err := dbh.getRegexesForGroups(groups, true)
	if err != nil {
		t.Fatalf("Error getting regexes: %v", err)
	}
	Expect(regs).To(HaveLen(1))

	regs, err = dbh.getRegexesForGroups(groups, false)
	if err != nil {
		t.Fatalf("Error getting regexes: %v", err)
	}
	Expect(regs).To(BeEmpty())

	r = types.Regex{
		GroupName: "misc.test",
	}
	dbh.DB.Save(&r)
	regs, err = dbh.getRegexesForGroups(groups, false)
	if err != nil {
		t.Fatalf("Error getting regexes: %v", err)
	}
	Expect(regs).To(HaveLen(1))
	regs, err = dbh.getRegexesForGroups(groups, true)
	if err != nil {
		t.Fatalf("Error getting regexes: %v", err)
	}
	Expect(regs).To(HaveLen(2))
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
	dbh := NewMemoryDBHandle(false)
	err := loadFixtures(dbh)
	if err != nil {
		t.Fatalf("Error creating fixtures: %v", err)
	}

	err = dbh.MakeBinaries()
	if err == nil {
		t.Fatalf("Expected MakeBinaries to error with no regexes.")
	}

	err = loadRegexFixtures(dbh)
	if err != nil {
		t.Fatalf("Error saving regex: %v", err)
	}

	err = dbh.MakeBinaries()
	if err != nil {
		t.Fatalf("Error creating binaries: %v", err)
	}

	bin, err := dbh.FindBinaryByName("Long Show Name - 01 [1080p]")
	if err != nil {
		t.Fatalf("Error finding expected binary: %v", err)
	}
	Expect(bin.TotalParts).To(Equal(2))

	//TODO: make binaries were binary exists and new parts are added.
}

func loadRegexFixtures(dbh *Handle) error {
	r := `(?i)^(?P<name>.*?\]) \[(?P<parts>\d{1,3}\/\d{1,3})`

	regex := &types.Regex{
		GroupName: "misc.test",
		Regex:     r,
		Ordinal:   1,
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
		Subject:       `Long Show Name - 01 [1080p] [01/02] - "Long Show Name - 01.mkv.rar" yEnc`,
		TotalSegments: 2,
		Posted:        time.Now(),
		From:          "foo@bar.com",
		Xref:          "12345 news.giganews.com",
		GroupName:     "misc.test",
		Segments:      segments,
	}

	err := dbh.DB.Save(&part).Error
	return err
}
