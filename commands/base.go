package commands

import "gopkg.in/alecthomas/kingpin.v2"

var (
	// App is the main hook to run
	App        = kingpin.New("gonab", "A usenet indexer")
	debug      = App.Flag("debug", "Enable Debug mode.").Bool()
	configfile = App.Flag("config", "Config file to use").Default("config.json").String()

	scanner = &ScanCommand{}
	s       = App.Command("scan", "Scan usenet groups for new articles").Action(scanner.scan)

	dbcreate = App.Command("createdb", "Create Database and Tables.").Action(createdb)

	bcmd     = &BinariesCommand{}
	binaries = App.Command("makebinaries", "Create binaries from parts").Action(bcmd.run)

	rcmd     = &ReleasesCommand{}
	releases = App.Command("makereleases", "Create releases from binaries").Action(rcmd.run)
)
