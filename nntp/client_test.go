package nntputil

import (
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/hobeone/nntp"
)

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
