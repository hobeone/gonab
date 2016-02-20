# gonab
Golang usenet binary indexer

# Quick Howto:
* Install Mysql
* Create database gonab and give a user permissions on it
* [Install Go](https://golang.org/doc/install)
* Install gonab: `go get -v github.com/hobeone/gonab`
* Go to directory: `cd $GOPATH/src/github.com/hobeone/gonab/`
* Copy config_sample.json to config.json and edit it to your satisfaction
* Compile: `go build`
* Create the database: `./gonab createdb`
* Import regex's (newznab seems to work best): `./gonab importregex`
* Create groups: `./gonab groups add ....`
* Scan groups: `./gonab scan`
* Make Binaries: `./gonab makebinaries`
* Make Releases: `./gonab releases make`

Use the `--debug` flag to see more about what's going on and `--debugdb` to see every SQL command.
