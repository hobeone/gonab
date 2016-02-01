package db

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/OneOfOne/xxhash"
	"github.com/Sirupsen/logrus"
	"github.com/davecgh/go-spew/spew"
	"github.com/hobeone/gonab/types"
)

// embed regexp.Regexp in a new type so we can extend it
type myRegexp struct {
	*regexp.Regexp
}

// add a new method to our new regular expression type
func (r *myRegexp) FindStringSubmatchMap(s string) map[string]string {
	captures := make(map[string]string)

	match := r.FindStringSubmatch(s)
	if match == nil {
		return captures
	}

	for i, name := range r.SubexpNames() {
		// Ignore the whole regexp match and unnamed groups
		if i == 0 || name == "" {
			continue
		}

		captures[name] = match[i]

	}
	return captures
}

var PartRegex = regexp.MustCompile(`(?i)[\[\( ]((\d{1,3}\/\d{1,3})|(\d{1,3} of \d{1,3})|(\d{1,3}-\d{1,3})|(\d{1,3}~\d{1,3}))[\)\] ]`)

func hasNameAndParts(m map[string]string) bool {
	_, nameok := m["name"]
	_, partok := m["parts"]
	return nameok && partok
}

func makeBinaryHash(name, group, from, totalParts string) string {
	h := xxhash.New64()
	h.Write([]byte(name + group + from + totalParts))
	return fmt.Sprintf("%x", h.Sum64())
}

func TestMakeReleases(t *testing.T) {
	dbh := NewDBHandle("test.db", true, true)
	logrus.SetLevel(logrus.DebugLevel)
	err := dbh.MakeReleases()
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
}

func TestMakeBinaries(t *testing.T) {
	r := `(?i).*?(?P<parts>\d{1,3}\/\d{1,3}).*?\"(?P<name>.*?)\.(sample|mkv|Avi|mp4|vol|ogm|par|rar|sfv|nfo|nzb|srt|ass|mpg|txt|zip|wmv|ssa|r\d{1,3}|7z|tar|mov|divx|m2ts|rmvb|iso|dmg|sub|idx|rm|ac3|t\d{1,2}|u\d{1,3})`
	rc := myRegexp{regexp.MustCompile(r)}
	logrus.SetLevel(logrus.DebugLevel)
	dbh := NewDBHandle("test.db", true, true)
	err := dbh.MakeBinaries()
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	var parts []types.Part
	err = dbh.DB.Where("binary_id is NULL").Find(&parts).Error
	if err != nil {
		t.Fatalf("Error: %v", err)
	}

	binaries := map[string]*types.Binary{}

	for _, p := range parts {

		m := rc.FindStringSubmatchMap(p.Subject)
		if len(m) > 0 {
			for k, v := range m {
				m[k] = strings.TrimSpace(v)
			}
		}
		// fill name if reqid is available
		if reqid, ok := m["reqid"]; ok {
			if _, okname := m["name"]; !okname {
				m["name"] = reqid
			}
		}

		// Generate a name if we don't have one
		if _, ok := m["name"]; !ok {
			var matchvalues []string
			for _, v := range m {
				matchvalues = append(matchvalues, v)
			}
			m["name"] = strings.Join(matchvalues, " ")
		}

		// Look for parts manually if the regex didn't return some
		if _, ok := m["parts"]; !ok {
			partmatch := PartRegex.FindStringSubmatch(p.Subject)
			if partmatch != nil {
				m["parts"] = partmatch[1]
			}
		}
		if !hasNameAndParts(m) {
			fmt.Printf("Couldn't find Name and Parts for %s\n", p.Subject)
			spew.Dump(m)
			continue
		}

		// Clean name of '-', '~', ' of '
		if strings.Index(m["parts"], "/") == -1 {
			m["parts"] = strings.Replace(m["parts"], "-", "/", -1)
			m["parts"] = strings.Replace(m["parts"], "~", "/", -1)
			m["parts"] = strings.Replace(m["parts"], " of ", "/", -1)
			m["parts"] = strings.Replace(m["parts"], "[", "", -1)
			m["parts"] = strings.Replace(m["parts"], "]", "", -1)
			m["parts"] = strings.Replace(m["parts"], "(", "", -1)
			m["parts"] = strings.Replace(m["parts"], ")", "", -1)
		}

		if strings.Index(m["parts"], "/") == -1 {
			fmt.Printf("Couldn't find valid parts information for %s (%s didn't include /)\n", p.Subject, m["parts"])
			continue
		}

		partcounts := strings.SplitN(m["parts"], "/", 2)

		binhash := makeBinaryHash(m["name"], p.Group, p.From, partcounts[1])
		if bin, ok := binaries[binhash]; ok {
			bin.Parts = append(bin.Parts, p)
		} else {
			totalparts, _ := strconv.Atoi(partcounts[1])
			binaries[binhash] = &types.Binary{
				Hash:       binhash,
				Name:       m["name"],
				Posted:     p.Posted,
				From:       p.From,
				Parts:      []types.Part{p},
				Group:      p.Group,
				TotalParts: totalparts,
			}
		}
		err = dbh.DB.Save(binaries[binhash]).Error
		if err != nil {
			t.Fatalf("Error saving binary %v", err)
		}
	}
}
