package db

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/DavidHuie/gomigrate"
	"github.com/Sirupsen/logrus"
	"github.com/hobeone/gonab/types"
	"github.com/jinzhu/gorm"
	// Import mysql
	_ "github.com/go-sql-driver/mysql"

	//Import Sqlite3
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
type debugLogger struct {
	logger *logrus.Logger
}

func newDBLogger() *debugLogger {
	d := &debugLogger{
		logger: logrus.New(),
	}
	d.logger.Level = logrus.DebugLevel
	return d
}

func (d *debugLogger) Print(msg ...interface{}) {
	d.logger.Debug(msg)
}

func openDB(dbType string, dbArgs string, verbose bool) gorm.DB {
	logrus.Infof("Opening database %s:%s", dbType, dbArgs)
	// Error only returns from this if it is an unknown driver.
	d, err := gorm.Open(dbType, dbArgs)
	if err != nil {
		panic(err.Error())
	}
	d.SingularTable(true)
	d.SetLogger(newDBLogger())
	d.LogMode(verbose)
	// Actually test that we have a working connection
	err = d.DB().Ping()
	if err != nil {
		panic(err.Error())
	}
	return d
}

func constructDBPath(dbname, dbuser, dbpass string) string {
	return fmt.Sprintf("%s:%s@/%s?charset=utf8&parseTime=True&loc=Local", dbuser, dbpass, dbname)
}

// NewDBHandle creates a new DBHandle
//	dbpath: the path to the database to use.
//	verbose: when true database accesses are logged to stdout
func NewDBHandle(dbname, dbuser, dbpass string, verbose bool) *Handle {
	constructedPath := constructDBPath(dbname, dbuser, dbpass)
	db := openDB("mysql", constructedPath, verbose)
	return &Handle{DB: db}
}

// NewMemoryDBHandle creates a new in memory database.  Only used for testing.
// The name of the database is a random string so multiple tests can run in
// parallel with their own database.  This will setup the database with the
// all the tables as well.
func NewMemoryDBHandle(verbose bool) *Handle {
	// The DSN is super important here. https://www.sqlite.org/inmemorydb.html
	// We want a named in memory db with a shared cache so that multiple
	// connections from the database layer share the cache but each call to this
	// function will return a different named database so unit tests don't stomp
	// on each other.
	gormdb := openDB("sqlite3", fmt.Sprintf("file:%s?mode=memory&cache=shared", randString()), verbose)

	migrator, err := gomigrate.NewMigrator(gormdb.DB(), gomigrate.Sqlite3{}, "../db/migrations/sqlite3")
	if err != nil {
		panic(err)
	}
	err = migrator.Migrate()
	if err != nil {
		panic(err)
	}

	return &Handle{DB: gormdb}
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

func (d *Handle) SearchReleasesByName(name string) ([]types.Release, error) {
	var releases []types.Release
	err := d.DB.Where("search_name LIKE ?", fmt.Sprintf("%%%s%%", name)).Preload("Group").Find(&releases).Error
	return releases, err
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

// AddGroup adds a new group to the database
func (d *Handle) AddGroup(groupname string) (*types.Group, error) {
	group := types.Group{
		Name:   groupname,
		Active: true,
	}
	err := d.DB.Save(&group).Error
	if err != nil {
		return nil, err
	}
	return &group, nil
}

// DisableGroup sets the Active attribute to false.
func (d *Handle) DisableGroup(groupname string) error {
	g, err := d.FindGroupByName(groupname)
	if err != nil {
		return err
	}
	g.Active = false
	return d.DB.Save(g).Error
}

// GetAllGroups returns all groups
func (d *Handle) GetAllGroups() ([]types.Group, error) {
	var g []types.Group
	err := d.DB.Find(&g).Error
	if err != nil {
		return nil, err
	}
	return g, nil
}

// GetActiveGroups returns all groups marked active in the db
func (d *Handle) GetActiveGroups() ([]types.Group, error) {
	var g []types.Group
	err := d.DB.Where("active = ?", true).Find(&g).Error
	if err != nil {
		return nil, err
	}
	return g, nil
}

// FindPartByHash does what it says.
func (d *Handle) FindPartByHash(hash string) (*types.Part, error) {
	var p types.Part
	err := d.DB.Preload("Segments").Where("hash = ?", hash).Find(&p).Error
	return &p, err
}

// FindBinaryByHash does what it says.
func (d *Handle) FindBinaryByHash(hash string) (*types.Binary, error) {
	var b types.Binary
	err := d.DB.Where("hash = ?", hash).Find(&b).Error
	return &b, err
}

// FindBinaryByName does what it says.
func (d *Handle) FindBinaryByName(name string) (*types.Binary, error) {
	var b types.Binary
	err := d.DB.Where("name = ?", name).Find(&b).Error
	return &b, err
}

func saveSegments(d *gorm.DB, segments []types.Segment, partid int64) error {
	var vals []interface{}
	valstrings := make([]string, len(segments))
	for i, s := range segments {
		valstrings[i] = "(?,?,?,?)"
		vals = append(vals, s.Segment, s.Size, s.MessageID, partid)
	}
	stmtString := fmt.Sprintf("INSERT INTO segment (segment, size, message_id, part_id) VALUES%s;", strings.Join(valstrings, ","))
	return d.Exec(stmtString, vals...).Error
}

// SavePartsAndMissedMessages saves a list of parts and missing message ids
// from an Overview call to the news server.
func (d *Handle) SavePartsAndMissedMessages(parts map[string]*types.Part, missed []types.MissedMessage) error {
	t := time.Now()
	tx := d.DB.Begin()
	newparts, newsegments := 0, 0
	for hash, part := range parts {
		var dbpart types.Part
		err := d.DB.Where("hash = ?", hash).Find(&dbpart).Error
		if err != nil {
			// Save new part
			err = tx.Save(part).Error
			if err != nil {
				tx.Rollback()
				return err
			}
			newparts++
			continue
		}
		err = saveSegments(tx, part.Segments, dbpart.ID)
		if err != nil {
			tx.Rollback()
			return err
		}
		newsegments = newsegments + len(part.Segments)
	}
	logrus.Debugf("Saved %d new parts and %d new messages in %s", newparts, newsegments, time.Since(t))

	t = time.Now()
	for _, mm := range missed {
		var dbMissed types.MissedMessage
		err := tx.Where("group_name = ? and message_number = ?", mm.GroupName, mm.MessageNumber).First(&dbMissed).Error
		if err != nil {
			err = tx.Save(&mm).Error
			if err != nil {
				tx.Rollback()
				return err
			}
			continue
		}
		dbMissed.Attempts++
		err = tx.Save(&dbMissed).Error
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	logrus.Debugf("Saved %d missed messages in %s", len(missed), time.Since(t))
	tx.Commit()

	return nil
}
