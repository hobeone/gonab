package commands

import (
	"github.com/DavidHuie/gomigrate"
	"github.com/Sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
)

func createdb(c *kingpin.ParseContext) error {
	_, dbh := commonInit()

	d := dbh.DB.DB()
	migrator, err := gomigrate.NewMigrator(d, gomigrate.Mysql{}, "db/migrations/mysql")
	if err != nil {
		logrus.Fatalf("Error starting migration: %v", err)
	}
	err = migrator.Migrate()
	return err
}
