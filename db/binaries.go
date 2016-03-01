package db

import (
	"crypto/sha1"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/OneOfOne/xxhash/native"
	"github.com/Sirupsen/logrus"
	"github.com/hobeone/gonab/types"
	"github.com/jinzhu/gorm"
)

var (
	musicGroupRegex = regexp.MustCompile(`\.(flac|lossless|mp3|music|sounds)`)

	// File/part count.
	genericSubjectCleaner1 = regexp.MustCompile(`(?i)((( \(\d\d\) -|(\d\d)? - \d\d\.|\d{4} \d\d -) | - \d\d-| \d\d\. [a-z]).+| \d\d of \d\d| \dof\d)\.mp3"?|(\)|\(|\[|\s)\d{1,5}(\/|(\s|_)of(\s|_)|-)\d{1,5}(\)|\]|\s|$|:)|\(\d{1,3}\|\d{1,3}\)|[^\d]{4}-\d{1,3}-\d{1,3}\.|\s\d{1,3}\sof\s\d{1,3}\.|\s\d{1,3}\/\d{1,3}|\d{1,3}of\d{1,3}\.|^\d{1,3}\/\d{1,3}\s|\d{1,3} - of \d{1,3}`)

	// File extensions.
	genericSubjectCleaner2 = regexp.MustCompile(`(?i)([-_](proof|sample|thumbs?))*(\.part\d*(\.rar)?|\.rar|\.7z)?(\d{1,3}\.rev"|\.vol\d+\+\d+.+?"|\.[A-Za-z0-9]{2,4}"|")`)

	// File extensions - If it was not in quotes.
	genericSubjectCleaner3 = regexp.MustCompile(`(?i)(-? [a-z0-9]+-?|\(?\d{4}\)?(_|-)[a-z0-9]+)\.jpg"?| [a-z0-9]+\.mu3"?|((\d{1,3})?\.part(\d{1,5})?|\d{1,5} ?|sample|- Partie \d+)?\.(7z|avi|diz|docx?|epub|idx|iso|jpg|m3u|m4a|mds|mkv|mobi|mp4|nfo|nzb|par(\s?2|")|pdf|rar|rev|rtf|r\d\d|sfv|srs|srr|sub|txt|vol.+(par2)|xls|zip|z{2,3})"?|(\s|(\d{2,3})?-)\d{2,3}\.mp3|\d{2,3}\.pdf|\.part\d{1,4}\.`)

	// File Sizes - Non unique ones.
	genericSubjectCleaner4 = regexp.MustCompile(`(?i)\d{1,3}(,|\.|\/)\d{1,3}\s(k|m|g)b|(\])?\s\d+KB\s(yENC)?|"?\s\d+\sbytes?|[- ]?\d+(\.|,)?\d+\s(g|k|m)?B\s-?(\s?yenc)?|\s\(d{1,3},\d{1,3}\s{K,M,G}B\)\s|yEnc \d+k$|{\d+ yEnc bytes}|yEnc \d+ |\(\d+ ?(k|m|g)?b(ytes)?\) yEnc$`)

	// Random stuff.
	genericSubjectCleaner5 = regexp.MustCompile(`(?i)/AutoRarPar\d{1,5}|\(\d+\)( |  )yEnc|\d+(Amateur|Classic)| \d{4,}[a-z]{4,} |part\d+`)

	// Multi spaces.
	genericSubjectCleaner6 = regexp.MustCompile(`\s\s+`)
)

// NameCleaner rewrites a given string based on it's regexes and a set of
// fallback rules.
type NameCleaner struct {
	Regexes []*types.Regex
}

// NewNameCleaner returns a new NameCleaner.  It will call Compile() on all
// given regexes.
func NewNameCleaner(regexes []*types.Regex) *NameCleaner {
	for _, r := range regexes {
		r.Compile()
	}
	return &NameCleaner{Regexes: regexes}
}

// Clean will return a rewritten name.
func (n *NameCleaner) Clean(name, groupname string) string {
	for _, r := range n.Regexes {
		if !r.CompiledGroupRegex.Regex.MatchString(groupname) {
			continue
		}
		m := r.Compiled.FindStringSubmatchMap(name)
		if len(m) == 0 {
			continue
		}
		// Sort keynames and then concatinate values in that order
		var keys = make([]string, len(m))
		i := 0
		for k := range m {
			keys[i] = k
			i++
		}
		sort.Strings(keys)
		var parts = make([]string, len(m))
		for i, k := range keys {
			parts[i] = m[k]
		}
		return strings.Join(parts, "")
	}
	return n.subjectCleaner(name, groupname)
}

func (n *NameCleaner) subjectCleaner(subject, groupname string) string {
	// try collections regex
	if !musicGroupRegex.MatchString(groupname) {
		subject = genericSubjectCleaner1.ReplaceAllString(subject, " ")
		subject = genericSubjectCleaner2.ReplaceAllString(subject, " ")
		subject = genericSubjectCleaner3.ReplaceAllString(subject, " ")
		subject = genericSubjectCleaner4.ReplaceAllString(subject, " ")
		subject = genericSubjectCleaner5.ReplaceAllString(subject, " ")
		subject = genericSubjectCleaner6.ReplaceAllString(subject, " ")
		return strings.TrimSpace(subject)
	}
	//TODO: handle music subject cleaning
	return subject
}

var partPartsRegex = regexp.MustCompile(`(?i)[[(\s](\d{1,5})(\/|[\s_]of[\s_]|-)(\d{1,5})[])\s$:]`)

func getPartsFromSubject(subject string) (int, int) {
	part, totalparts := 0, 0
	partmatches := partPartsRegex.FindStringSubmatch(subject)
	if len(partmatches) > 0 {
		totalparts, _ = strconv.Atoi(partmatches[3])
		part, _ = strconv.Atoi(partmatches[1])
	}
	return part, totalparts
}

// MakeBinaries makes binaries using the nzedb approach
func (d *Handle) MakeBinaries() error {
	var regex []*types.Regex
	err := d.DB.Table("collection_regex").Find(&regex).Error
	if err != nil {
		return err
	}
	cleaner := NewNameCleaner(regex)

	var parts []types.Part
	var partCount int64
	err = d.DB.Where("binary_id is NULL").Find(&parts).Count(&partCount).Error
	if err != nil {
		return err
	}
	logrus.Infof("Found %d parts to process.", partCount)

	binaries := map[string]*types.Binary{}
	t := time.Now()
	for _, p := range parts {
		cleanedSubject := cleaner.Clean(p.Subject, p.GroupName)
		_, totalparts := getPartsFromSubject(p.Subject)

		binhash := makeHash(cleanedSubject, p.GroupName, p.From, strconv.Itoa(totalparts))
		if bin, ok := binaries[binhash]; ok {
			bin.Parts = append(bin.Parts, p)
		} else {
			b, err := d.FindBinaryByHash(binhash)
			if err != nil {
				logrus.Debugf("New binary found: %s", cleanedSubject)
				binaries[binhash] = &types.Binary{
					Hash:       binhash,
					Name:       p.Subject,
					Posted:     p.Posted,
					From:       p.From,
					Parts:      []types.Part{p},
					GroupName:  p.GroupName,
					TotalParts: totalparts,
				}
			} else {
				b.Parts = append(b.Parts, p)
				binaries[binhash] = b
			}
		}
	}
	logrus.Infof("Found %d new binaries", len(binaries))
	dbt := time.Now()
	tx := d.DB.Begin()
	for _, b := range binaries {
		txerr := saveBinary(tx, b)
		if txerr != nil {
			tx.Rollback()
			return txerr
		}
	}
	tx.Commit()
	logrus.Infof("Saved %d binaries to db in %s", len(binaries), time.Since(dbt))
	logrus.Infof("Processed %d binaries from %d parts in %s", len(binaries), len(parts), time.Since(t))
	return nil

	// Get name cleaner and load regexes
	// Find all parts with no binary
	// Gen hash of that part and see if a Binary already exists
	// If so, add part
	// If not create Binary and add part
}

// SaveBinary will efficently save a binary and it's associated parts.
// Having to do this for efficency is probably a good sign that I should just
// use sqlx or something.
func saveBinary(tx *gorm.DB, b *types.Binary) error {
	var txerr error
	if tx.NewRecord(b) {
		parts := b.Parts
		b.Parts = []types.Part{}
		txerr = tx.Save(b).Error
		if txerr != nil {
			return txerr
		}
		pids := make([]int64, len(parts))
		for i, p := range parts {
			pids[i] = p.ID
		}
		txerr = tx.Model(types.Part{}).Where("id IN (?)", pids).Updates(map[string]interface{}{"binary_id": b.ID}).Error
		if txerr != nil {
			return txerr
		}
	} else {
		pids := make([]int64, len(b.Parts))
		for i, p := range b.Parts {
			pids[i] = p.ID
		}
		txerr = tx.Model(types.Part{}).Where("id IN (?)", pids).Updates(map[string]interface{}{"binary_id": b.ID}).Error
		if txerr != nil {
			return txerr
		}

	}
	return nil
}

func makeShaHash(hashargs ...string) string {
	return fmt.Sprintf("%x", sha1.Sum([]byte(strings.Join(hashargs, ""))))
}

func makeHash(hashargs ...string) string {
	h := xxhash.New64()
	h.Write([]byte(strings.Join(hashargs, ".")))
	return fmt.Sprintf("%x", h.Sum64())
}
