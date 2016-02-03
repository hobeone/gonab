package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/davecgh/go-spew/spew"
	"github.com/hobeone/gonab/commands"
	"github.com/hobeone/gonab/db"
	"github.com/hobeone/gonab/types"
)

func convertRegexes() error {
	dbh := db.NewDBHandle("gonab.db", true)
	var regexes []types.CollectionRegex
	err := dbh.DB.Find(&regexes).Error
	if err != nil {
		return err
	}
	for _, r := range regexes {
		r.Regex = strings.TrimRight(r.Regex, "/")
		r.Regex = strings.TrimLeft(r.Regex, "/")
		//fmt.Println(r.Regex)
		rc, err := regexp.Compile(r.Regex)
		if err != nil {
			fmt.Printf("Error compiling %v: %v", r.Regex, err)
			continue
		}
		matches := rc.FindStringSubmatch(r.Description)
		spew.Dump(matches)
	}
	return nil
}

func main() {
	kingpin.MustParse(commands.App.Parse(os.Args[1:]))
}
