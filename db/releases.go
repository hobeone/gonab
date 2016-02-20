package db

import (
	"database/sql"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/hobeone/gonab/categorize"
	"github.com/hobeone/gonab/nzb"
	"github.com/hobeone/gonab/types"
	"github.com/jinzhu/gorm"
)

// MakeReleases comment
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
	for _, b := range binaries {
		// See if a Release already exists for this binary name/date
		dbrel := &types.Release{}
		err := d.DB.Where("name = ? and posted = ?", b.Name, b.Posted).First(&dbrel).Error
		if err != nil && err != gorm.RecordNotFound {
			return err
		}
		if dbrel.ID != 0 {
			logrus.Infof("Duplicate Binary found, deleting: %s", b.Name)
			d.DB.Delete(&b)
			continue
		}

		// Get Binary and all it's parts and segments.
		dbbin := &types.Binary{}
		err = d.DB.Preload("Parts").Preload("Parts.Segments").First(dbbin, b.ID).Error
		if err != nil {
			return err
		}

		// Find size
		// Blacklist
		grp, err := d.FindGroupByName(b.GroupName)
		if err != nil {
			logrus.Errorf("Unknown group %s for binary %d: %s. Skipping...", b.GroupName, b.ID, b.Name)
			continue
		}
		nzbstr, err := nzb.WriteNZB(dbbin)
		if err != nil {
			return err
		}
		logrus.Infof("New release found: %s", b.Name)
		newrel := &types.Release{
			Name:         b.Name,
			OriginalName: b.Name,
			SearchName:   cleanReleaseName(b.Name),
			Posted:       b.Posted,
			From:         b.From,
			Group:        *grp,
			Size:         dbbin.Size(),
			NZB:          nzbstr,
		}

		// Categorize
		cat := categorize.Categorize(newrel.Name, newrel.Group.Name)
		newrel.CategoryID = sql.NullInt64{Int64: int64(cat), Valid: true}

		// Check if size is too small
		// Check if too few files
		tx := d.DB.Begin()
		err = tx.Save(newrel).Error
		if err != nil {
			tx.Rollback()
			return err
		}

		// Delete Parts
		err = tx.Where("binary_id = ?", dbbin.ID).Delete(types.Part{}).Error
		if err != nil {
			tx.Rollback()
			return err
		}

		// Delete segments.
		partids := make([]int64, len(dbbin.Parts))
		for i, p := range dbbin.Parts {
			partids[i] = p.ID
		}
		err = tx.Where("part_id in (?)", partids).Delete(types.Segment{}).Error
		if err != nil {
			tx.Rollback()
			return err
		}

		// Delete binary
		err = tx.Delete(dbbin).Error
		if err != nil {
			tx.Rollback()
			return err
		}
		tx.Commit()
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
