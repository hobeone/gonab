package db

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"sync"

	"github.com/Sirupsen/logrus"
	"github.com/hobeone/gonab/types"
	"github.com/jinzhu/gorm"

	// Import mysql
	_ "github.com/go-sql-driver/mysql"
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

func setupDB(db gorm.DB) error {
	tx := db.Begin()
	err := tx.AutoMigrate(
		&types.Group{},
		&types.Release{},
		&types.Binary{},
		&types.Part{},
		&types.Segment{},
		&types.Regex{},
		&types.MissedMessage{},
	).Error
	if err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()

	return nil
}

func constructDBPath(dbname, dbuser, dbpass string) string {
	return fmt.Sprintf("%s:%s@/%s?charset=utf8&parseTime=True&loc=Local", dbuser, dbpass, dbname)
}

// CreateAndMigrateDB will create a new database on disk and create all tables.
func CreateAndMigrateDB(dbname, dbuser, dbpass string, verbose bool) (*Handle, error) {
	constructedPath := constructDBPath(dbname, dbuser, dbpass)
	db := openDB("mysql", constructedPath, verbose)
	err := setupDB(db)
	if err != nil {
		return nil, err
	}
	return &Handle{DB: db}, nil
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
	dbpath := randString()
	dbpath = fmt.Sprintf("file:%s?mode=memory", dbpath)
	db := openDB("sqlite3", dbpath, verbose)
	err := setupDB(db)
	if err != nil {
		panic(err.Error())
	}
	return &Handle{DB: db}
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

// SearchReleasesByName searches releases for those matching the given string
func (d *Handle) SearchReleasesByName(name string) ([]types.Release, error) {
	var releases []types.Release
	err := d.DB.Where("search_name LIKE ?", fmt.Sprintf("%%%s%%", name)).Find(&releases).Error
	return releases, err
}

// ListReleases func
func (d *Handle) ListReleases(limit int) error {
	var rels []types.Release
	err := d.DB.Limit(limit).Find(&rels).Error
	if err != nil {
		return err
	}

	for _, rel := range rels {
		fmt.Printf("Release (%d):\n", rel.ID)
		fmt.Printf("  Name: %s\n", rel.Name)
	}
	return nil
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
	err := d.DB.Where("hash = ?", hash).Find(&p).Error
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
