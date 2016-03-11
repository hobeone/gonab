package processing

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"

	"github.com/hobeone/gonab/types"
)

var (
	tvNameBase = `[^a-z0-9](\d\d-\d\d|\d{1,3}x\d{2,3}|\(?(19|20)\d{2}\)?|(480|720|1080)[ip]|AAC2?|BD-?Rip|Blu-?Ray|D0?\d|DD5|DiVX|DLMux|DTS|DVD(-?Rip)?|E\d{2,3}|[HX][-_. ]?26[45]|ITA(-ENG)?|HEVC|[HPS]DTV|PROPER|REPACK|Season|Episode|S\d+[^a-z0-9]?((E\d+)[abr]?)*|WEB[-_. ]?(DL|Rip)|XViD)[^a-z0-9]?`

	tvNameRegex  = regexp.MustCompile(`(?i)` + tvNameBase)
	tvNameRegex1 = regexp.MustCompile(`(?i)^([^a-z0-9]{2,}|(sample|proof|repost)-)(?P<name>[\w .-]*?)` + tvNameBase)
	tvNameRegex2 = regexp.MustCompile(`(?i)^(?P<name>[a-z0-9][\w\' .-]*?)` + tvNameBase)

	tvDateStartRegex   = regexp.MustCompile(`^\d{6}`)
	tvNameCleanerRegex = regexp.MustCompile(`(?i)\(.*?\)|[._]`)
	tvYearRegex        = types.RegexpUtil{Regex: regexp.MustCompile(`[^a-z0-9](?P<year>(19|20)(\d{2}))[^a-z0-9]`)}
)

func parseNameToTVName(name string) string {
	tvName := ""
	r := types.RegexpUtil{Regex: tvNameRegex1}
	matches := r.FindStringSubmatchMap(name)
	if n, ok := matches["name"]; ok {
		tvName = n
	} else {
		r = types.RegexpUtil{Regex: tvNameRegex2}
		matches = r.FindStringSubmatchMap(name)
		if n, ok = matches["name"]; ok {
			tvName = n
		} else {
			tvName = name
		}
	}

	// Clean any remaining parts
	tvName = tvNameRegex.ReplaceAllString(tvName, " ")
	// Remove date from front
	tvName = tvDateStartRegex.ReplaceAllString(tvName, "")
	// Remove periods, underscored, anything between parenthesis.
	tvName = tvNameCleanerRegex.ReplaceAllString(tvName, " ")
	// Collapse multiple spaces and remove leading and trailing spaces
	tvName = strings.Join(strings.Fields(tvName), " ")
	tvName = strings.Trim(tvName, " ")
	return tvName
}
func isMn(r rune) bool {
	return unicode.Is(unicode.Mn, r) // Mn: nonspacing marks
}

var (
	removeChars  = []string{"'", ":", "!", "\"", "#", "*", "’", ",", "(", ")", "?"}
	spaceChars   = []string{".", "_"}
	badChanRegex = regexp.MustCompile(`^(?i)(history|discovery) channel`)
)

func cleanName(name string) string {
	name = strings.Replace(name, "ß", "ss", -1)
	name = strings.Replace(name, "Σ", "e", -1)
	name = strings.Replace(name, "æ", "a", -1)
	name = strings.Replace(name, "&", "and", -1)
	name = strings.Replace(name, "$", "s", -1)
	for _, c := range removeChars {
		name = strings.Replace(name, c, "", -1)
	}
	for _, c := range spaceChars {
		name = strings.Replace(name, c, " ", -1)
	}

	name = badChanRegex.ReplaceAllString(name, "")
	name = strings.Join(strings.Fields(name), " ")
	t := transform.Chain(norm.NFD, transform.RemoveFunc(isMn), norm.NFC)
	unicodeCleanedName, _, err := transform.String(t, name)

	if err == nil {
		name = unicodeCleanedName
	}

	return strings.Trim(name, ` "`)
}

var (
	// S01E01-E02 and S01E01-02
	seasonEpRegex1 = regexp.MustCompile(`(?i)^(.*?)[^a-z0-9]s(?P<season>\d{1,2})[^a-z0-9]?e(?P<episode>\d{1,3})(?:[e-])(?P<lastepisode>\d{1,3})[^a-z0-9]`)
	//S01E0102 and S01E01E02 - lame no delimit numbering, regex would collide if there was ever 1000 ep season.
	seasonEpRegex2 = regexp.MustCompile(`(?i)^(.*?)[^a-z0-9]s(?P<season>\d{2})[^a-z0-9]?e(?P<episode>\d{2})e?(\d{2})[^a-z0-9]`)
	// S01E01 and S01.E01
	seasonEpRegex3 = regexp.MustCompile(`(?i)^(.*?)[^a-z0-9]s(?P<season>\d{1,2})[^a-z0-9]?e(?P<episode>\d{1,3})[abr]?[^a-z0-9]`)
	// S01
	seasonEpRegex4 = regexp.MustCompile(`(?i)^(.*?)[^a-z0-9]s(?P<season>\d{1,2})[^a-z0-9]`)
	// S01D1 and S1D1
	seasonEpRegex5 = regexp.MustCompile(`(?i)^(.*?)[^a-z0-9]s(?P<season>\d{1,2})[^a-z0-9]?d\d{1}[^a-z0-9]`)
	// 1x01 and 101
	seasonEpRegex6 = regexp.MustCompile(`(?i)^(.*?)[^a-z0-9](?P<season>\d{1,2})x(?P<episode>\d{1,3})[^a-z0-9]`)

	regularShowRegexes = []*regexp.Regexp{
		seasonEpRegex1, seasonEpRegex2, seasonEpRegex3, seasonEpRegex4, seasonEpRegex5, seasonEpRegex6,
	}

	// Date based Shows
	// 2009.01.01 and 2009-01-01
	seasonEpRegex7 = regexp.MustCompile(`(?i)^(.*?)[^a-z0-9](?P<airdate>(19|20)(\d{2})[.\/-](\d{2})[.\/-](\d{2}))[^a-z0-9]`)
	// 01.01.2009
	seasonEpRegex8 = regexp.MustCompile(`(?i)^(.*?)[^a-z0-9](?P<airdate>(\d{2})[^a-z0-9](\d{2})[^a-z0-9](19|20)(\d{2}))[^a-z0-9]`)
	// 01.01.09
	seasonEpRegex9 = regexp.MustCompile(`(?i)^(.*?)[^a-z0-9](\d{2})[^a-z0-9](\d{2})[^a-z0-9](\d{2})[^a-z0-9]`)
	// 2009.E01
	seasonEpRegex10 = regexp.MustCompile(`(?i)^(.*?)[^a-z0-9]20(\d{2})[^a-z0-9](\d{1,3})[^a-z0-9]`)
	// 2009.Part1
	seasonEpRegex11 = regexp.MustCompile(`(?i)^(.*?)[^a-z0-9](19|20)(\d{2})[^a-z0-9]Part(\d{1,2})[^a-z0-9]`)

	// Randos
	// Part1/Pt1
	seasonEpRegex12 = regexp.MustCompile(`(?i)^(.*?)[^a-z0-9](?:Part|Pt)[^a-z0-9](\d{1,2})[^a-z0-9]`)
	//The.Pacific.Pt.VI.HDTV.XviD-XII / Part.IV
	seasonEpRegex13 = regexp.MustCompile(`(?i)^(.*?)[^a-z0-9](?:Part|Pt)[^a-z0-9]([ivx]+)`)
	// Band.Of.Brothers.EP06.Bastogne.DVDRiP.XviD-DEiTY
	seasonEpRegex14 = regexp.MustCompile(`(?i)^(.*?)[^a-z0-9]EP?[^a-z0-9]?(\d{1,3})`)
	// Season.1
	seasonEpRegex15 = regexp.MustCompile(`(?i)^(.*?)[^a-z0-9]Seasons?[^a-z0-9]?(\d{1,2})`)
)

type showSeasonEp struct {
	Season      string
	Episode     string
	LastEpisode string
	Airdate     string
}

func parseSeasonEp(name string) *showSeasonEp {
	for _, r := range regularShowRegexes {
		reu := types.RegexpUtil{Regex: r}
		m := reu.FindStringSubmatchMap(name)
		if len(m) > 0 {
			sse := &showSeasonEp{}
			if s, ok := m["season"]; ok {
				sse.Season = strings.TrimLeft(s, "0")
			}
			if s, ok := m["episode"]; ok {
				sse.Episode = strings.TrimLeft(s, "0")
			}
			if s, ok := m["lastepisode"]; ok {
				sse.LastEpisode = s
			}
			return sse
		}
	}
	m := seasonEpRegex7.FindStringSubmatch(name)
	if m != nil {
		return &showSeasonEp{
			Season:  m[3] + m[4],
			Episode: m[5] + "/" + m[6],
			Airdate: m[2],
		}
	}
	//TODO: add the rest
	return nil
}

// TVNameParseResult represents extracted information from a tv release name
type TVNameParseResult struct {
	Name        string
	CleanedName string
	Country     string
	Episode     string
	Season      string
	Airdate     string
}

// ParseInfo returns information about the given tv show release
func ParseInfo(name string) (*TVNameParseResult, error) {
	res := &TVNameParseResult{}
	res.Name = parseNameToTVName(name)
	// country info
	res.CleanedName = cleanName(name)
	sse := parseSeasonEp(name)
	if sse == nil {
		return nil, fmt.Errorf("error parsing %s", name)
	}
	res.Season = sse.Season
	res.Episode = sse.Episode
	res.Airdate = sse.Airdate

	if res.Airdate == "" {
		// If has year add back to cleaned name for searching
		m := tvYearRegex.FindStringSubmatchMap(name)
		if len(m) > 0 {
			res.CleanedName = fmt.Sprintf("%s (%s)", res.CleanedName, m["year"])
		}
	}
	if (res.Season == "" && res.Episode == "") && res.Airdate == "" {
		return nil, fmt.Errorf("Error parsing %s", name)
	}

	return res, nil
}
