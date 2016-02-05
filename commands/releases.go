package commands

import (
	"fmt"
	"io/ioutil"
	"path"

	"github.com/Sirupsen/logrus"
	"github.com/hobeone/gonab/db"
	"github.com/hobeone/gonab/types"
	"gopkg.in/alecthomas/kingpin.v2"
)

// ReleasesCommand should probably be split up for the sub commands.
type ReleasesCommand struct {
	Limit     int
	ReleaseID int64
	FilePath  string
	DirPath   string
}

func (r *ReleasesCommand) configure(app *kingpin.Application) {
	rgrp := App.Command("releases", "manipulate releases")
	rgrp.Command("make", "Create releases from binaries").Action(r.run)

	rgrpList := rgrp.Command("list", "List releases").Action(r.list)
	rgrpList.Flag("limit", "Number of releases to list").Short('l').Default("10").IntVar(&r.Limit)

	rgrpExportNZB := rgrp.Command("exportnzb", "Write NZB for release to file").Action(r.exportNZB)
	rgrpExportNZB.Flag("id", "ID of release to export").Required().Int64Var(&r.ReleaseID)
	rgrpExportNZB.Flag("file", "Filename to write to.  If not given use the name of the release +'.nzb'").StringVar(&r.FilePath)
	rgrpExportNZB.Flag("dir", "Directory to write to.").Default(".").StringVar(&r.DirPath)
}

func (r *ReleasesCommand) run(c *kingpin.ParseContext) error {
	if *debug {
		logrus.SetLevel(logrus.DebugLevel)
	}
	cfg := loadConfig(*configfile)

	dbh := db.NewDBHandle(cfg.DB.Name, cfg.DB.Username, cfg.DB.Password, cfg.DB.Verbose)
	err := dbh.MakeReleases()
	return err
}

func (r *ReleasesCommand) list(c *kingpin.ParseContext) error {
	cfg := loadConfig(*configfile)
	if *debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	limit := r.Limit
	if limit == 0 {
		limit = 10
	}

	dbh := db.NewDBHandle(cfg.DB.Name, cfg.DB.Username, cfg.DB.Password, cfg.DB.Verbose)
	dbh.ListReleases(r.Limit)
	return nil
}

func (r *ReleasesCommand) exportNZB(c *kingpin.ParseContext) error {
	cfg := loadConfig(*configfile)
	if *debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	dbh := db.NewDBHandle(cfg.DB.Name, cfg.DB.Username, cfg.DB.Password, cfg.DB.Verbose)

	var rel types.Release
	err := dbh.DB.First(&rel, r.ReleaseID).Error
	if err != nil {
		return err
	}

	filename := r.FilePath
	if filename == "" {
		filename = rel.Name + ".nzb"
	}
	fullpath := path.Join(r.DirPath, filename)

	fmt.Printf("Writing NZB for %s to %s\n", rel.Name, fullpath)
	err = ioutil.WriteFile(fullpath, []byte(rel.NZB), 0644)

	if err != nil {
		return err
	}

	return nil
}
