package commands

import (
	"github.com/hobeone/gonab/config"
	"github.com/hobeone/gonab/db"
	"gopkg.in/alecthomas/kingpin.v2"
)

func createdb(c *kingpin.ParseContext) error {
	cfg := config.NewConfig()
	err := cfg.ReadConfig(*configfile)
	if err != nil {
		return err
	}

	_, err = db.CreateAndMigrateDB(cfg.DB.Path, cfg.DB.Verbose)
	if err != nil {
		return err
	}
	return nil
}
