package categorize

import (
	"testing"

	"github.com/hobeone/gonab/types"
)

type categoryMatches struct {
	Name  string
	Group string
	types.Category
}

var miscGroupTuples = []categoryMatches{
	{"aaaaaaaaaaaaaaaaaaaa", "group", types.Other_Misc},
	{"FooTitle-aaaaaaaaaaaaaaaaaaa", "group", types.Unknown},
	{"FooTitle-aaaaaaaaaaaaaaaaaaa", "alt.binaries.audio.warez", types.Unknown},
}

func TestIsMisc(t *testing.T) {
	for _, m := range miscGroupTuples {
		cat := isMisc(m.Name, m.Group)
		if cat != m.Category {
			t.Errorf("Expected name:%s group%s to get category %s. Got: %s", m.Name, m.Group, m.Category, cat)
		}
	}
}

var groupTuples = []categoryMatches{
	{"", "alt.binaries.audio.warez", types.PC_0day},
	{"", "alt.binaries.multimedia.anime", types.TV_Anime},
	{"", "alt.binaries.multimedia.anime.highspeed", types.TV_Anime},
	{"foobar", "", types.Unknown},
	{"foobar-Season 01-1080p", "alt.binaries.moovee", types.TV_HD},
	{"foobar-1080p", "alt.binaries.moovee", types.Movie_HD},
	{"foobar", "alt.binaries.moovee", types.Movie_SD},
}

func TestCategoryFromGroup(t *testing.T) {
	for _, m := range groupTuples {
		cat := categoryFromGroup(m.Name, m.Group)
		if cat != m.Category {
			t.Errorf("Expected name:%s group%s to get category %s. Got: %s", m.Name, m.Group, m.Category, cat)
		}
	}
}

var tvMatches = []categoryMatches{
	{"The.Daily.Show.with.Trevor.Noah.2016.02.08.Gillian.Jacobs.720p.CC.WEBRip.AAC2.0.x264-monkee", "", types.TV_WEBDL},
	{"The.Daily.Show.with.Trevor.Noah.2016.02.08.Gillian.Jacobs.720p.CC.AAC2.0.x264-monkee", "", types.TV_HD},
	{"[HorribleSubs] Haruchika - 07 [1080p]", "", types.TV_Anime},
	{"NBC.Nightly.News.2016.02.17.WEB-DL.x264-2Maverick", "", types.TV_Other},
	{"Lilyhammer.3x08.Un.Nuovo.Inizio.ITA.BDMux.x264-NovaRip", "", types.TV_Foreign},
	{"WWE.NXT.2016.02.17.720p.WEBRip.h264-HatchetGear", "", types.TV_Sport},
	{"Solar.System.The.Secrets.of.the.Universe.E01.2014.DOCU.1080p.BluRay.x264-iFPD", "", types.TV_Documentary},
	{"James.Corden.2016.02.17.Katie.Holmes.HDTV.x264-CROOKS", "", types.TV_SD},
}

func TestIsTV(t *testing.T) {
	for _, m := range tvMatches {
		cat := isTV(m.Name, m.Group)
		if cat != m.Category {
			t.Errorf("Expected name:%s group%s to get category %s. Got: %s", m.Name, m.Group, m.Category, cat)
		}
	}
}

var movieMatches = []categoryMatches{
	{"danish-BluRay", "", types.Movie_Foreign},
	{"Castellano-Bluray", "", types.Movie_Foreign},
	{"Movie.Name.2001-Danish-1080p", "", types.Movie_Foreign},
	{"Movie.Name.2001-AC3-DVD.x264", "", types.Movie_DVD},
	{"Movie.Name.2001-AC3-dvdrip.x264", "", types.Movie_SD},
	{"Movie.Name.2001-3D-1080p.x264", "", types.Movie_3D},
	{"Movie.Name.2001-bluray-1080p.x264", "", types.Movie_BluRay},
	{"Movie.Name.2001-WebRip-1080p.x264", "", types.Movie_WEBDL},
	{"Movie.Name.2001-Web-dl-1080p.x264", "", types.Movie_WEBDL},
	{"SecretUsenet.com-Movie.Name.2001-bluray-1080p.x264", "", types.Movie_HD},
}

func TestIsMovie(t *testing.T) {
	for _, m := range movieMatches {
		cat := isMovie(m.Name, m.Group)
		if cat != m.Category {
			t.Errorf("Expected name:'%s' group:'%s' to get category: '%s'. Got: '%s'", m.Name, m.Group, m.Category, cat)
		}
	}

}
