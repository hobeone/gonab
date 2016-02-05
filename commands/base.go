package commands

import (
	"github.com/Sirupsen/logrus"
	"github.com/hobeone/gonab/config"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	// App is the main hook to run
	App        = kingpin.New("gonab", "A usenet indexer")
	debug      = App.Flag("debug", "Enable Debug mode.").Bool()
	debugdb    = App.Flag("debugdb", "Log Database queries (noisy).").Default("false").Bool()
	configfile = App.Flag("config", "Config file to use").Default("config.json").ExistingFile()

	scanner = &ScanCommand{}
	s       = App.Command("scan", "Scan usenet groups for new articles").Action(scanner.scan)

	dbcreate = App.Command("createdb", "Create Database and Tables.").Action(createdb)

	bcmd     = &BinariesCommand{}
	binaries = App.Command("makebinaries", "Create binaries from parts").Action(bcmd.run)

	rcmd = &ReleasesCommand{}

	rgrp     = App.Command("releases", "manipulate releases")
	releases = rgrp.Command("make", "Create releases from binaries").Action(rcmd.run)
	rgrpList = rgrp.Command("list", "List releases").Action(rcmd.list)

	regexcmd    = &RegexImporter{}
	regexcmdrun = App.Command("importregex", "Import regexes from nzedb").Action(regexcmd.run)
)

func loadConfig(cfile string) *config.Config {
	if len(cfile) == 0 {
		logrus.Infof("No --config_file given.  Using default: %s", *configfile)
		cfile = *configfile
	}

	logrus.Infof("Got config file: %s\n", cfile)
	cfg := config.NewConfig()
	err := cfg.ReadConfig(cfile)
	if err != nil {
		logrus.Fatal(err)
	}

	// Override cfg from flags
	cfg.DB.Verbose = *debugdb
	return cfg
}
