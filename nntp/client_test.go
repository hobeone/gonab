package nntputil

import (
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/davecgh/go-spew/spew"
	"github.com/hobeone/gonab/db"
	"github.com/hobeone/gonab/types"
	"github.com/hobeone/nntp"
)

func TestGroupScan(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)
	n, err := ConnectAndAuthenticate("fake.localhost:119", "username", "password", true)
	if err != nil {
		t.Fatalf("Error %v", err)
	}
	d := db.NewDBHandle("test.db", true, true)

	var g types.Group

	err = d.DB.FirstOrCreate(
		&g,
		types.Group{
			Name: "alt.binaries.multimedia.anime.highspeed"},
	).Error
	if err != nil {
		t.Fatalf("Error %v", err)
	}

	_, err = n.GroupScanForward(d, "alt.binaries.multimedia.anime.highspeed", 100000)
	if err != nil {

		t.Fatalf("Error %v", err)
	}
}

func TestRegexp(t *testing.T) {

	s := `[AnimeRG-FTS] Ajin (2016) - 02 [720p] [31FBC4AE] [16/16] - "[AnimeRG-FTS] Ajin (2016) - 02 [720p] [31FBC4AE].mkv.vol63+29.par2" yEnc (27/30)`
	overview := nntp.MessageOverview{
		Subject: s,
		Bytes:   1024,
		From:    "<foobaz@bar.com>",
	}

	groupName := "misc.test"
	parts := Scan(groupName, []nntp.MessageOverview{overview})
	spew.Dump(parts)
}
