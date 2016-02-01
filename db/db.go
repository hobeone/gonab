package db

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"strings"
	"sync"

	"github.com/Sirupsen/logrus"
	"github.com/davecgh/go-spew/spew"
	"github.com/hobeone/gonab/nzb"
	"github.com/hobeone/gonab/types"
	"github.com/jinzhu/gorm"

	// Import sqlite
	_ "github.com/mattn/go-sqlite3"
)

//Handle Struct
type Handle struct {
	DB           gorm.DB
	writeUpdates bool
	syncMutex    sync.Mutex
}

// debugLogger satisfies Gorm's logger interface
// so that we can log SQL queries at Logrus' debug level
type debugLogger struct{}

func (*debugLogger) Print(msg ...interface{}) {
	logrus.Debug(msg)
}

func openDB(dbType string, dbArgs string, verbose bool) gorm.DB {
	logrus.Infof("Opening database %s:%s", dbType, dbArgs)
	// Error only returns from this if it is an unknown driver.
	d, err := gorm.Open(dbType, dbArgs)
	if err != nil {
		panic(err.Error())
	}
	d.SingularTable(true)
	d.SetLogger(&debugLogger{})
	d.LogMode(verbose)
	// Actually test that we have a working connection
	err = d.DB().Ping()
	if err != nil {
		panic(err.Error())
	}
	return d
}

func setupDB(db gorm.DB) error {
	tx := db.Begin()
	err := tx.AutoMigrate(&types.Group{}, &types.Release{}, &types.Binary{}, &types.Part{}, &types.Segment{}).Error
	if err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	err = db.Exec("PRAGMA journal_mode=WAL;").Error
	if err != nil {
		return err
	}
	err = db.Exec("PRAGMA synchronous = NORMAL;").Error
	if err != nil {
		return err
	}
	err = db.Exec("PRAGMA encoding = \"UTF-8\";").Error
	if err != nil {
		return err
	}

	return nil
}

func createAndOpenDb(dbPath string, verbose bool, memory bool) *Handle {
	mode := "rwc"
	if memory {
		mode = "memory"
	}
	constructedPath := fmt.Sprintf("file:%s?cache=shared&mode=%s", dbPath, mode)
	db := openDB("sqlite3", constructedPath, verbose)
	err := setupDB(db)
	if err != nil {
		panic(err.Error())
	}
	return &Handle{DB: db}
}

// NewDBHandle creates a new DBHandle
//	dbPath: the path to the database to use.
//	verbose: when true database accesses are logged to stdout
//	writeUpdates: when true actually write to the databse (useful for testing)
func NewDBHandle(dbPath string, verbose bool, writeUpdates bool) *Handle {
	d := createAndOpenDb(dbPath, verbose, false)
	d.writeUpdates = writeUpdates
	return d
}

// NewMemoryDBHandle creates a new in memory database.  Only used for testing.
// The name of the database is a random string so multiple tests can run in
// parallel with their own database.
func NewMemoryDBHandle(verbose bool, writeUpdates bool) *Handle {
	d := createAndOpenDb(randString(), verbose, true)
	d.writeUpdates = writeUpdates
	return d
}

func randString() string {
	rb := make([]byte, 32)
	_, err := rand.Read(rb)
	if err != nil {
		fmt.Println(err)
	}
	return base64.URLEncoding.EncodeToString(rb)
}

// CreatePart func
func (d *Handle) CreatePart(p *types.Part) error {
	return d.DB.Save(p).Error
}

// ListParts func
func (d *Handle) ListParts() {
	var parts []types.Part
	err := d.DB.Preload("Segments").Find(&parts).Error
	if err != nil {
		fmt.Printf("Error getting parts: %v\n", err)
	}

	for _, p := range parts {
		fmt.Printf("Part: %s\n", p.Subject)
		fmt.Printf("  Segments: %d\n", p.TotalSegments)
		fmt.Printf("  Available Segments: %d\n", len(p.Segments))
	}
}

// FindGroupByName does what it says
func (d *Handle) FindGroupByName(name string) (*types.Group, error) {
	var g types.Group
	err := d.DB.Where("name = ?", name).First(&g).Error
	if err != nil {
		return nil, err
	}
	return &g, nil
}

// MakeBinaries comment
func (d *Handle) MakeBinaries() error {
	var groupnames []string
	err := d.DB.Model(&types.Part{}).Group(`"group"`).Pluck(`"group"`, &groupnames).Error
	if err != nil {
		return err
	}
	for _, name := range groupnames {
		fmt.Printf("Group: %s\n", name)
	}
	return nil
}

var removeChars = []string{"#", "@", "$", "%", "^", "§", "¨", "©", "Ö"}
var spaceChars = []string{"_", ".", "-"}

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

// MakeReleases comment
func (d *Handle) MakeReleases() error {
	var binaries []types.Binary
	q := `SELECT binary.id, binary.name, binary.posted, binary.total_parts, binary.'group'
	FROM binary
	INNER JOIN (
			SELECT
					part.id, part.binary_id, part.total_segments, count(*) as available_segments
			FROM part
					INNER JOIN segment ON part.id = segment.part_id
			GROUP BY part.id
			) as part
			ON binary.id = part.binary_id
	GROUP BY binary.id
	HAVING count(*) >= binary.total_parts AND (sum(part.available_segments) / sum(part.total_segments)) * 100 >= 100
	ORDER BY binary.posted DESC LIMIT 1`
	err := d.DB.Raw(q, 100).Scan(&binaries).Error
	if err != nil {
		return err
	}
	for _, b := range binaries {
		// See if a Release already exists for this binary name/date
		dbrel := &types.Release{}
		err := d.DB.Where("name = ? and posted = ?", b.Name, b.Posted).First(&dbrel).Error
		if err != nil && err != gorm.RecordNotFound {
			spew.Dump(err)
			return err
		}
		if dbrel.ID != 0 {
			logrus.Infof("Duplicate Binary found, deleting: %s", b.Name)
			//Delete here
			continue
		}

		dbbin := &types.Binary{}
		err = d.DB.Preload("Parts").Preload("Parts.Segments").First(dbbin, b.ID).Error
		if err != nil {
			return err
		}

		// Find size
		// Blacklist
		grp, err := d.FindGroupByName(b.Group)
		if err != nil {
			logrus.Errorf("Unknown group %s for binary %d: %s", b.Group, b.ID, b.Name)
		}

		newrel := &types.Release{
			Name:         b.Name,
			OriginalName: b.Name,
			SearchName:   cleanReleaseName(b.Name),
			Posted:       b.Posted,
			From:         b.From,
			Group:        *grp,
			Size:         dbbin.Size(),
		}

		nzbstr, err := nzb.WriteNZB(dbbin)
		if err != nil {
			return err
		}
		fmt.Println(nzbstr)
		// Categorize

		// Check if size is too small
		// Check if too few files
		spew.Dump(newrel)
		err = ioutil.WriteFile("tesst.nzb", []byte(nzbstr), 0644)
		if err != nil {
			return err
		}
		//spew.Dump(dbbin)

		//Create NZB
	}
	return nil
}
