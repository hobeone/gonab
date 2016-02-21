package categorize

import (
	"regexp"

	"github.com/hobeone/gonab/types"
)

type testFunc func(string, string) types.Category

// Categorize takes a release name and usenet group name and tries to
// categorize it.
func Categorize(name, group string) types.Category {
	tests := []testFunc{
		isMisc,
		categoryFromGroup,
		isTV,
		isMovie,
	}
	for _, f := range tests {
		if cat := f(name, group); cat != types.Unknown {
			return cat
		}
	}
	return types.Unknown
}

// Translated from nZEDb to start with

var (
	miscMiscRegex1   = regexp.MustCompile(`(?i)[a-z0-9]{20,}`)
	miscMiscRegex2   = regexp.MustCompile(`(?i)^[A-Z0-9]{1,}$`)
	miscHashRegex    = regexp.MustCompile(`(?i)[a-f0-9]{32,64}`)
	miscNotMiscRegex = regexp.MustCompile(`(?i)[^a-z0-9]((480|720|1080)[ip]|s\d{1,3}[-._ ]?[ed]\d{1,3}([ex]\d{1,3}|[-.\w ]))[^a-z0-9]`)
)

func isMisc(name, group string) types.Category {
	switch {
	case miscNotMiscRegex.MatchString(name):
		return types.Unknown
	case miscHashRegex.MatchString(name):
		return types.Other_Hashed
	case miscMiscRegex1.MatchString(name) || miscMiscRegex2.MatchString(name):
		return types.Other_Misc
	}

	return types.Unknown
}

var (
	animeGroupRegex = regexp.MustCompile(`alt\.binaries\.(multimedia\.erotica\.|cartoons\.french\.|dvd\.|multimedia\.)?anime(\.highspeed|\.repost|s-fansub|\.german)?`)
)

func categoryFromGroup(name, group string) types.Category {
	switch {
	case group == "alt.binaries.audio.warez":
		return types.PC_0day
	case animeGroupRegex.MatchString(group):
		return types.TV_Anime
	case group == "alt.binaries.moovee":
		if cat := isTV(name, group); cat != types.Unknown {
			return cat
		}
		if cat := isMovieHD(name, group); cat != types.Unknown {
			return cat
		}
		return types.Movie_SD
	}
	return types.Unknown
}

var (
	tvRegex         = regexp.MustCompile(`(?i)Daily[-_\.]Show|Nightly News|^\[[a-zA-Z\.\-]+\].*[-_].*\d{1,3}[-_. ]((\[|\()(h264-)?\d{3,4}(p|i)(\]|\))\s?(\[AAC\])?|\[[a-fA-F0-9]{8}\]|(8|10)BIT|hi10p)(\[[a-fA-F0-9]{8}\])?|(\d\d-){2}[12]\d{3}|[12]\d{3}(\.\d\d){2}|\d+x\d+|\.e\d{1,3}\.|s\d{1,3}[-._ ]?[ed]\d{1,3}([ex]\d{1,3}|[-.\w ])|[-._ ](\dx\d\d|C4TV|Complete[-._ ]Season|DSR|(D|H|P|S)DTV|EP[-._ ]?\d{1,3}|S\d{1,3}.+Extras|SUBPACK|Season[-._ ]\d{1,2})([-._ ]|$)|TVRIP|TV[-._ ](19|20)\d\d|TrollHD`)
	tvRegexNegative = regexp.MustCompile(`(?i)[-._ ](flac|imageset|mp3|xxx)[-._ ]|[ .]exe$`)
	tvSportsRegex   = regexp.MustCompile(`(?i)[-._ ]((19|20)\d\d[-._ ]\d{1,2}[-._ ]\d{1,2}[-._ ]VHSRip|Indy[-._ ]?Car|(iMPACT|Smoky[-._ ]Mountain|Texas)[-._ ]Wrestling|Moto[-._ ]?GP|NSCS[-._ ]ROUND|NECW[-._ ]TV|(Per|Post)\-Show|PPV|WrestleMania|WCW|WEB[-._ ]HD|WWE[-._ ](Monday|NXT|RAW|Smackdown|Superstars|WrestleMania))[-._ ]`)

	tvMatchFuncs = []testFunc{
		isOtherTV,
		isForeignTV,
		isSportTV,
		isDocumentaryTV,
		isWebDL,
		isAnimeTV,
		isHDTV,
		isSDTV,
		isOtherTV2,
	}
)

func isTV(name, group string) types.Category {
	if tvRegex.MatchString(name) && !tvRegexNegative.MatchString(name) {
		for _, f := range tvMatchFuncs {
			if cat := f(name, group); cat != types.Unknown {
				return cat
			}
		}
		return types.TV_Other
	}

	if tvSportsRegex.MatchString(name) {
		if cat := isSportTV(name, group); cat != types.Unknown {
			return types.TV_Sport
		}
		return types.TV_Other
	}
	return types.Unknown
}

var (
	hdtvRegex  = regexp.MustCompile(`(?i)1080(i|p)|720p|bluray`)
	webdlRegex = regexp.MustCompile(`(?i)web[-._ ]dl|web-?rip`)
)

func isHDTV(name, group string) types.Category {
	if hdtvRegex.MatchString(name) {
		return types.TV_HD
	}
	return types.Unknown
}

var (
	sdtvRegex1 = regexp.MustCompile(`(?i)(360|480|576)p|Complete[-._ ]Season|dvdr(ip)?|dvd5|dvd9|\.pdtv|SD[-._ ]TV|TVRip|NTSC|BDRip|hdtv|xvid`)
	sdtvRegex2 = regexp.MustCompile(`(?i)((H|P)D[-._ ]?TV|DSR|WebRip)[-._ ]x264`)
	sdtvRegex3 = regexp.MustCompile(`(?i)s\d{1,3}[-._ ]?[ed]\d{1,3}([ex]\d{1,3}|[-.\w ])|\s\d{3,4}\s`)
	sdtvRegex4 = regexp.MustCompile(`(?i)(H|P)D[-._ ]?TV|BDRip[-._ ]x264`)
)

func isSDTV(name, group string) types.Category {
	if sdtvRegex1.MatchString(name) || sdtvRegex2.MatchString(name) {
		return types.TV_SD
	}
	if sdtvRegex3.MatchString(name) && sdtvRegex4.MatchString(name) {
		return types.TV_SD
	}
	return types.Unknown
}

func isWebDL(name, group string) types.Category {
	if webdlRegex.MatchString(name) {
		return types.TV_WEBDL
	}
	return types.Unknown
}

var (
	otherTVRegex  = regexp.MustCompile(`(?i)[-._ ]S\d{1,3}.+(EP\d{1,3}|Extras|SUBPACK)[-._ ]|News`)
	otherTVRegex2 = regexp.MustCompile(`(?i)[-._ ]s\d{1,3}[-._ ]?(e|d(isc)?)\d{1,3}([-._ ]|$)`)
)

func isOtherTV(name, group string) types.Category {
	if otherTVRegex.MatchString(name) {
		return types.TV_Other
	}
	return types.Unknown
}

func isOtherTV2(name, group string) types.Category {
	if otherTVRegex2.MatchString(name) {
		return types.TV_Other
	}
	return types.Unknown
}

var (
	foreignTV1     = regexp.MustCompile(`(?i)[-._ ](chinese|dk|fin|french|ger?|heb|ita|jap|kor|nor|nordic|nl|pl|swe)[-._ ]?(sub|dub)(ed|bed|s)?|<German>`)
	foreignTV2     = regexp.MustCompile(`(?i)[-._ ](brazilian|chinese|croatian|danish|deutsch|dutch|estonian|flemish|finnish|french|german|greek|hebrew|icelandic|italian|ita|latin|mandarin|nordic|norwegian|polish|portuguese|japenese|japanese|russian|serbian|slovenian|spanish|spanisch|swedish|thai|turkish).+(720p|1080p|Divx|DOKU|DUB(BED)?|DLMUX|NOVARIP|RealCo|Sub(bed|s)?|Web[-._ ]?Rip|WS|Xvid|x264)[-._ ]`)
	foreignTV3     = regexp.MustCompile(`(?i)[-._ ](720p|1080p|Divx|DOKU|DUB(BED)?|DLMUX|NOVARIP|RealCo|Sub(bed|s)?|Web[-._ ]?Rip|WS|Xvid).+(brazilian|chinese|croatian|danish|deutsch|dutch|estonian|flemish|finnish|french|german|greek|hebrew|icelandic|italian|ita|latin|mandarin|nordic|norwegian|polish|portuguese|japenese|japanese|russian|serbian|slovenian|spanish|spanisch|swedish|thai|turkish)[-._ ]`)
	foreignTV4     = regexp.MustCompile(`(?i)(S\d\d[EX]\d\d|DOCU(MENTAIRE)?|TV)?[-._ ](FRENCH|German|Dutch)[-._ ](720p|1080p|dv(b|d)r(ip)?|LD|HD\-?TV|TV[-._ ]?RIP|x264)[-._ ]`)
	foreignTV5     = regexp.MustCompile(`(?i)[-._ ]FastSUB|NL|nlvlaams|patrfa|RealCO|Seizoen|slosinh|Videomann|Vostfr|xslidian[-._ ]|x264\-iZU`)
	foreignRegexes = []*regexp.Regexp{foreignTV1, foreignTV2, foreignTV3, foreignTV4, foreignTV5}
)

func isForeignTV(name, group string) types.Category {
	for _, i := range foreignRegexes {
		if i.MatchString(name) {
			return types.TV_Foreign
		}
	}
	return types.Unknown
}

var animeTVRegex = regexp.MustCompile(`(?i)[-._ ]Anime[-._ ]|^\[[a-zA-Z\.\-]+\].*[-_].*\d{1,3}[-_. ]((\[|\()((\d{1,4}x\d{1,4})|(h264-)?\d{3,4}(p|i))(\]|\))\s?(\[AAC\])?|\[[a-fA-F0-9]{8}\]|(8|10)BIT|hi10p)(\[[a-fA-F0-9]{8}\])?`)

func isAnimeTV(name, group string) types.Category {
	if animeTVRegex.MatchString(name) {
		return types.TV_Anime
	}
	return types.Unknown
}

var (
	sportTVNegRegex = regexp.MustCompile(`(?i)s\d{1,3}[-._ ]?[ed]\d{1,3}([ex]\d{1,3}|[-.\w ])`)
	sportTVRegex1   = regexp.MustCompile(`(?i)[-._ ]?(Bellator|bundesliga|EPL|ESPN|FIA|la[-._ ]liga|MMA|motogp|NFL|NHL|NCAA|PGA|red[-._ ]bull.+race|Sengoku|Strikeforce|supercup|uefa|UFC|wtcc|WWE)[-._ ]`)
	sportTVRegex2   = regexp.MustCompile(`(?i)[-._ ]?(AFL|Grand Prix|Indy[-._ ]Car|(iMPACT|Smoky[-._ ]Mountain|Texas)[-._ ]Wrestling|Moto[-._ ]?GP|NSCS[-._ ]ROUND|NECW|Poker|PWX|Rugby|WCW)[-._ ]`)
	sportTVRegex3   = regexp.MustCompile(`(?i)[-._ ]?(Horse)[-._ ]Racing[-._ ]`)

	sportRegexes = []*regexp.Regexp{sportTVRegex1, sportTVRegex2, sportTVRegex3}
)

func isSportTV(name, group string) types.Category {
	if sportTVNegRegex.MatchString(name) {
		return types.Unknown
	}
	for _, i := range sportRegexes {
		if i.MatchString(name) {
			return types.TV_Sport
		}
	}
	return types.Unknown
}

var (
	documentaryRegex = regexp.MustCompile(`(?i)[-._ ](Docu|Documentary)[-._ ]`)
)

func isDocumentaryTV(name, group string) types.Category {
	if documentaryRegex.MatchString(name) {
		return types.TV_Documentary
	}
	return types.Unknown
}

var (
	movieRegex       = regexp.MustCompile(`(?i)[-._ ]AVC[-._ ]|[BH][DR]RIP|Bluray|BD[-._ ]?(25|50)?|\bBR\b|Camrip|[-._ ]\d{4}[-._ ].+(720p|1080p|Cam|HDTS)|DIVX|[-._ ]DVD[-._ ]|DVD-?(5|9|R|Rip)|Untouched|VHSRip|XVID|[-._ ](DTS|TVrip)[-._ ]`)
	movieNegRegex    = regexp.MustCompile(`(?i)auto(cad|desk)|divx[-._ ]plus|[-._ ]exe$|[-._ ](jav|XXX)[-._ ]|SWE6RUS|\wXXX(1080p|720p|DVD)|Xilisoft`)
	movieClassifiers = []testFunc{
		isMovieForeign,
		isMovieDVD,
		isMovieWebDL,
		isMovieSD,
		isMovie3D,
		isMovieBluRay,
		isMovieHD,
		isMovieOther,
	}
)

func isMovie(name, group string) types.Category {
	if movieRegex.MatchString(name) && !movieNegRegex.MatchString(name) {
		for _, f := range movieClassifiers {
			if cat := f(name, group); cat != types.Unknown {
				return cat
			}
		}
	}
	return types.Unknown
}

var (
	movieForeignRegex1  = regexp.MustCompile(`(?i)(danish|flemish|Deutsch|dutch|french|german|heb|hebrew|Castellano|nl[-._ ]?sub|dub(bed|s)?|\.NL|norwegian|swedish|swesub|spanish|Staffel)[-._ ]|\(german\)|Multisub`)
	movieForeignRegex3  = regexp.MustCompile(`(?i)(720p|1080p|AC3|AVC|DIVX|DVD(5|9|RIP|R)|XVID)[-._ ](Dutch|French|German|ITA)|\(?(Dutch|French|German|ITA)\)?[-._ ](720P|1080p|AC3|AVC|DIVX|DVD(5|9|RIP|R)|HD[-._ ]|XVID)`)
	movieForeignRegexes = []*regexp.Regexp{
		movieForeignRegex1, movieForeignRegex3,
	}
)

func isMovieForeign(name, group string) types.Category {
	for _, i := range movieForeignRegexes {
		if i.MatchString(name) {
			return types.Movie_Foreign
		}
	}
	return types.Unknown
}

var (
	movieDVDRegex = regexp.MustCompile(`(?i)(dvd\-?r|[-._ ]dvd|dvd9|dvd5|[-._ ]r5)[-._ ]`)
)

func isMovieDVD(name, group string) types.Category {
	if movieDVDRegex.MatchString(name) {
		return types.Movie_DVD
	}
	return types.Unknown
}

var (
	movieSDRegex = regexp.MustCompile(`(?i)(divx|dvdscr|extrascene|dvdrip|\.CAM|HDTS(-LINE)?|vhsrip|xvid(vd)?)[-._ ]`)
)

func isMovieSD(name, group string) types.Category {
	if movieSDRegex.MatchString(name) {
		return types.Movie_SD
	}
	return types.Unknown
}

var (
	movie3DRegex = regexp.MustCompile(`(?i)[-._ ]3D\s?[\.\-_\[ ](1080p|(19|20)\d\d|AVC|BD(25|50)|Blu[-._ ]?ray|CEE|Complete|GER|MVC|MULTi|SBS|H(-)?SBS)[-._ ]`)
)

func isMovie3D(name, group string) types.Category {
	if movie3DRegex.MatchString(name) {
		return types.Movie_3D
	}
	return types.Unknown
}

var (
	movieBluRayRegex    = regexp.MustCompile(`(?i)bluray\-|[-._ ]bd?[-._ ]?(25|50)|blu-ray|Bluray\s\-\sUntouched|[-._ ]untouched[-._ ]`)
	movieBluRayNegRegex = regexp.MustCompile(`(?i)SecretUsenet\.com`)
)

func isMovieBluRay(name, group string) types.Category {
	if movieBluRayRegex.MatchString(name) && !movieBluRayNegRegex.MatchString(name) {
		return types.Movie_BluRay
	}
	return types.Unknown
}

var (
	movieHDRegex = regexp.MustCompile(`(?i)720p|1080p|AVC|VC1|VC\-1|web\-dl|wmvhd|x264|XvidHD|bdrip`)
)

func isMovieHD(name, group string) types.Category {
	if movieHDRegex.MatchString(name) {
		return types.Movie_HD
	}
	return types.Unknown
}

var (
	movieOtherRegex = regexp.MustCompile(`(?i)[-._ ]cam[-._ ]`)
)

func isMovieOther(name, group string) types.Category {
	if movieOtherRegex.MatchString(name) {
		return types.Movie_Other
	}
	return types.Unknown
}

var (
	movieWebDLRegex = regexp.MustCompile(`(?i)web[-._ ]dl|web-?rip`)
)

func isMovieWebDL(name, group string) types.Category {
	if movieWebDLRegex.MatchString(name) {
		return types.Movie_WEBDL
	}
	return types.Unknown
}
