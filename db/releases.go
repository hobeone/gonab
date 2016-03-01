package db

import (
	"database/sql"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/hobeone/gonab/categorize"
	"github.com/hobeone/gonab/nzb"
	"github.com/hobeone/gonab/types"
	"github.com/jinzhu/gorm"
)

// SearchReleases does what it says.
// query is matched against the name of the Releases
// limit limits the number of returned releases to no more than that
// categories restricts the searched releases to be in those categories
func (d *Handle) SearchReleases(query string, offset, limit int, categories []types.Category) ([]types.Release, error) {
	qParts := []string{}
	var vals []interface{}
	if query != "" {
		qParts = append(qParts, "search_name LIKE ?")
		vals = append(vals, fmt.Sprintf("%%%s%%", query))
	}
	if len(categories) > 0 {
		qParts = append(qParts, "category_id IN (?)")
		for _, cat := range categories {
			vals = append(vals, int64(cat))
		}
	}
	q := strings.Join(qParts, " AND ")
	var releases []types.Release
	err := d.DB.Where(q, vals...).Preload("Category").Preload("Group").Offset(offset).Limit(limit).Order("posted desc").Find(&releases).Error
	return releases, err
}

//FindReleaseByHash returns
func (d *Handle) FindReleaseByHash(h string) (*types.Release, error) {
	var rel types.Release
	err := d.DB.Where("hash = ?", h).First(&rel).Error
	return &rel, err
}

/*
*					"cleansubject"  => $title['title'],
					"properlynamed" => true,
					"predb"         => $title['id'],
					"requestid"     => true
*/

// RegexCleaner cleans strings based on it's regexes
type RegexCleaner struct {
	Regexes []*types.Regex
}

//NewRegexCleaner returns a new RegexCleaner and compiles all the given Regexes
func NewRegexCleaner(regexes []*types.Regex) (*RegexCleaner, error) {
	for _, r := range regexes {
		err := r.Compile()
		if err != nil {
			return nil, fmt.Errorf("error compiling regex for %s Regex %d: %v", r.Kind, r.ID, err)
		}
	}
	return &RegexCleaner{Regexes: regexes}, nil
}

// Clean tries to rewrite the given name using it's regexes that match the
// given group.
func (r *RegexCleaner) Clean(name, groupname string) string {
	for _, r := range r.Regexes {
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
	return name
}

// ReleaseCleaner rewrites release names based on regexes and hard coded rules.
type ReleaseCleaner struct {
	RegexCleaner *RegexCleaner
}

//NewReleaseCleaner returns a new ReleaseCleaner
func NewReleaseCleaner(regexes []*types.Regex) (*ReleaseCleaner, error) {
	regclean, err := NewRegexCleaner(regexes)
	if err != nil {
		return nil, err
	}
	return &ReleaseCleaner{RegexCleaner: regclean}, nil
}

// Clean rewrites the release name based on regexes and hard coded rules.
func (r *ReleaseCleaner) Clean(name, poster, groupname string, size int64) string {
	// TODO: predb check
	// TODO: deal with reqid

	cleanedName := r.RegexCleaner.Clean(name, groupname)
	if cleanedName != name {
		return cleanedName
	}

	// Try release_naming_regexes
	// Try to clean with hardcoded www.town.ag regexes
	// switch on groupname
	// teevee -> clean with teevee hardcode
	// default -> clean with generic
	return name
}

// MakeReleases searchs for complete binaries and create releases from them deleting the
// binaries in the process.
func (d *Handle) MakeReleases() error {
	var binaries []types.Binary
	q := `SELECT binary.id, binary.name, binary.posted, binary.total_parts, binary.group_name
	FROM ` + "`binary`" + `
	INNER JOIN (
		SELECT
			part.id, part.binary_id, part.total_segments, count(*) as available_segments
		FROM part
			INNER JOIN segment ON part.id = segment.part_id
		GROUP BY part.id
	) as part
	ON binary.id = part.binary_id
	GROUP BY binary.id
	HAVING count(*) >= binary.total_parts AND (sum(part.available_segments) / sum(part.total_segments)) * 100 >= ?
	ORDER BY binary.posted DESC`
	err := d.DB.Raw(q, 100).Scan(&binaries).Error
	if err != nil {
		return err
	}
	logrus.Infof("Got %d binaries to scan", len(binaries))
	var regex []*types.Regex
	err = d.DB.Find(&regex).Error
	if err != nil {
		return err
	}
	cleaner, err := NewReleaseCleaner(regex)
	if err != nil {
		return err
	}

	for _, b := range binaries {
		// Get Binary and all it's parts and segments.
		dbbin := &types.Binary{}
		err = d.DB.Preload("Parts").Preload("Parts.Segments").First(dbbin, b.ID).Error
		if err != nil {
			return err
		}

		cleanName := cleaner.Clean(b.Name, b.From, b.GroupName, dbbin.Size())

		hash := makeShaHash(cleanName, b.GroupName, strconv.FormatInt(b.Posted.Unix(), 10), strconv.FormatInt(dbbin.Size(), 10))
		rel, err := d.FindReleaseByHash(hash)
		if err != nil && err != gorm.RecordNotFound {
			return err
		}
		if err == nil {
			logrus.Infof("Found duplicate release hash: %s for binary %s", rel.Hash, b.Name)
			err = deleteBinary(&d.DB, dbbin)
			if err != nil {
				return err
			}
			continue
		}

		grp, err := d.FindGroupByName(b.GroupName)
		if err != nil {
			logrus.Errorf("Unknown group %s for binary %d: %s. Skipping...", b.GroupName, b.ID, b.Name)
			continue
		}

		// Check if too few files
		if len(dbbin.Parts) < grp.MinFiles {
			logrus.Infof("Too few files for %s in group %s (%d < %d)", b.Name, grp.Name, len(dbbin.Parts), grp.MinFiles)
			err = deleteBinary(&d.DB, dbbin)
			if err != nil {
				return err
			}
			continue
		}

		nzbstr, err := nzb.WriteNZB(dbbin)
		if err != nil {
			return err
		}
		newrel := &types.Release{
			Name:         cleanName,
			OriginalName: b.Name,
			SearchName:   cleanReleaseName(cleanName),
			Posted:       b.Posted,
			From:         b.From,
			Group:        *grp,
			Size:         dbbin.Size(),
			NZB:          nzbstr,
			Hash:         hash,
		}

		// Categorize
		cat := categorize.Categorize(newrel.Name, newrel.Group.Name)
		newrel.CategoryID = sql.NullInt64{Int64: int64(cat), Valid: true}
		logrus.Infof("New %s release found: %s", cat, b.Name)
		// Check if size is too small
		tx := d.DB.Begin()
		logrus.Infof("Saving new release: %s", newrel.Name)
		err = tx.Save(newrel).Error
		if err != nil {
			tx.Rollback()
			return err
		}
		err = deleteBinary(tx, dbbin)
		if err != nil {
			tx.Rollback()
			return err
		}
		tx.Commit()
	}
	return nil
}

func deleteBinary(tx *gorm.DB, dbbin *types.Binary) error {
	// Delete Parts
	err := tx.Where("binary_id = ?", dbbin.ID).Delete(types.Part{}).Error
	if err != nil {
		return err
	}

	// Delete segments.
	partids := make([]int64, len(dbbin.Parts))
	for i, p := range dbbin.Parts {
		partids[i] = p.ID
	}
	err = tx.Where("part_id in (?)", partids).Delete(types.Segment{}).Error
	if err != nil {
		return err
	}

	// Delete binary
	err = tx.Delete(dbbin).Error
	if err != nil {
		return err
	}
	return nil
}

// from SE: [...]T syntax is sugar for [123]T. It creates a fixed size array,
// but lets the compiler figure out how many elements are in it.
var removeChars = [...]string{"#", "@", "$", "%", "^", "§", "¨", "©", "Ö"}
var spaceChars = [...]string{"_", ".", "-"}

// Strip bad chars out of name for API to match on
func cleanReleaseName(name string) string {
	for _, c := range removeChars {
		name = strings.Replace(name, c, "", -1)
	}
	for _, c := range spaceChars {
		name = strings.Replace(name, c, " ", -1)
	}
	return name
}
