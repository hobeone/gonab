package commands

import (
	"github.com/Sirupsen/logrus"
	"github.com/hobeone/gonab/api"
	"gopkg.in/alecthomas/kingpin.v2"
)

type ServerCommand struct{}

func (b *ServerCommand) configure(app *kingpin.Application) {
	app.Command("server", "start a server for the API").Action(b.run)
}

func (b *ServerCommand) run(c *kingpin.ParseContext) error {
	if *debug {
		logrus.SetLevel(logrus.DebugLevel)
	}
	cfg := loadConfig(*configfile)
	api.RunAPIServer(cfg)
	return nil
}
