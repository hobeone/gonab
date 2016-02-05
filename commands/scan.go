package commands

import (
	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/hobeone/gonab/db"
	"github.com/hobeone/gonab/nntp"
	"gopkg.in/alecthomas/kingpin.v2"
)

// ScanCommand comment
type ScanCommand struct {
	MaxArticles int
}

func (s *ScanCommand) scan(c *kingpin.ParseContext) error {
	if *debug {
		logrus.SetLevel(logrus.DebugLevel)
	}
	cfg := loadConfig(*configfile)

	dbh := db.NewDBHandle(cfg.DB.Name, cfg.DB.Username, cfg.DB.Password, cfg.DB.Verbose)
	groups, err := dbh.GetActiveGroups()
	if err != nil {
		return err
	}
	if len(groups) == 0 {
		return fmt.Errorf("No active groups to scan.")
	}
	logrus.Debugf("Got %d groups to scan.", len(groups))

	n, err := nntputil.ConnectAndAuthenticate(
		fmt.Sprintf("%s:%d", cfg.NewsServer.Host, cfg.NewsServer.Port),
		cfg.NewsServer.Username,
		cfg.NewsServer.Password,
		cfg.NewsServer.UseTLS,
	)
	if err != nil {
		return err
	}

	for _, g := range groups {
		_, err := n.GroupScanForward(dbh, g.Name, 100000)
		if err != nil {
			return err
		}
	}
	return nil
}
