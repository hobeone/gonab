package main

import (
	"log"
	"os"

	"gopkg.in/alecthomas/kingpin.v2"

	"net/http"
	_ "net/http/pprof"

	"github.com/hobeone/gonab/commands"
)

func main() {
	go func() {
		log.Println(http.ListenAndServe("0.0.0.0:6060", nil))
	}()

	commands.SetupCommands()
	kingpin.MustParse(commands.App.Parse(os.Args[1:]))
}
