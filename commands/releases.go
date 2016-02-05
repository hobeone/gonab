package commands

import (
	"github.com/Sirupsen/logrus"
	"github.com/hobeone/gonab/db"
	"gopkg.in/alecthomas/kingpin.v2"
)

// ReleasesCommand comment
type ReleasesCommand struct{}

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

	dbh := db.NewDBHandle(cfg.DB.Name, cfg.DB.Username, cfg.DB.Password, cfg.DB.Verbose)
	dbh.ListParts()
	return nil
}
