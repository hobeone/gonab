package commands

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"text/tabwriter"

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

	Categories []int64
	SearchTerm string
}

func (r *ReleasesCommand) configure(app *kingpin.Application) {
	rgrp := app.Command("releases", "manipulate releases")
	rgrp.Command("make", "Create releases from binaries").Action(r.run)

	rgrpList := rgrp.Command("search", "Search releases").Action(r.list)
	rgrpList.Flag("limit", "Number of releases to list").Short('l').Default("10").IntVar(&r.Limit)
	rgrpList.Flag("categories", "Only show releases from this category").Short('c').Int64ListVar(&r.Categories)
	rgrpList.Flag("search", "Only show releases that match this search term").Short('s').StringVar(&r.SearchTerm)

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
	_, dbh := commonInit()

	limit := r.Limit
	if limit == 0 {
		limit = 10
	}

	var cats []types.Category
	for _, c := range r.Categories {
		cats = append(cats, types.CategoryFromInt(c))
	}

	releases, err := dbh.SearchReleases(r.SearchTerm, 0, r.Limit, cats)
	if err != nil {
		return err
	}
	fmt.Printf("Found %d releases matching your search criteria\n", len(releases))
	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 5, 0, 1, ' ', 0)
	fmt.Fprintln(w, "Name\tCategory\tDate\tGroup\tHash")
	for _, r := range releases {
		fmt.Fprintln(w, fmt.Sprintf("%s\t%s\t%s\t%s\t%s", r.Name, r.CategoryName(), r.Posted, r.Group.Name, r.Hash))
	}
	w.Flush()
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
