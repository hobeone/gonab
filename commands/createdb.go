package commands

import (
	"github.com/hobeone/gonab/db"
	"gopkg.in/alecthomas/kingpin.v2"
)

func createdb(c *kingpin.ParseContext) error {
	cfg := loadConfig(*configfile)
	_, err := db.CreateAndMigrateDB(cfg.DB.Name, cfg.DB.Username, cfg.DB.Password, cfg.DB.Verbose)
	return err
}
