package config

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

// Config is the base config struct for gonab
type Config struct {
	NewsServer newsServer
	DB         dbConfig
	Regex      regexSource
}

type newsServer struct {
	Host     string
	Port     int
	Username string
	Password string
	UseTLS   bool
}
type dbConfig struct {
	Path     string
	Verbose  bool   // turn on verbose db logging
	UpdateDB bool   // if we should update db items during crawl
	Type     string // file or memory (for testing)
}

type regexSource struct {
	URL string //
}

// NewConfig returns a new Config struct with defaults filled out
func NewConfig() *Config {
	return &Config{
		NewsServer: newsServer{},
		Regex:      regexSource{},
		DB: dbConfig{
			Verbose:  true,
			UpdateDB: true,
			Type:     "file",
		},
	}
}

func replaceTildeInPath(path string) string {
	usr, _ := user.Current()
	dir := usr.HomeDir
	return strings.Replace(path, "~", dir, 1)
}

// ReadConfig Decodes and evaluates a json config file, watching for include cycles.
func (c *Config) ReadConfig(configPath string) error {
	absConfigPath, err := filepath.Abs(replaceTildeInPath(configPath))
	if err != nil {
		return fmt.Errorf("failed to expand absolute path for %s", configPath)
	}

	var f *os.File
	if f, err = os.Open(absConfigPath); err != nil {
		return fmt.Errorf("failed to open config: %v", err)
	}
	defer f.Close()

	filecont, err := ioutil.ReadAll(f)

	if err = json.Unmarshal(filecont, c); err != nil {
		extra := ""
		if serr, ok := err.(*json.SyntaxError); ok {
			if _, serr := f.Seek(0, os.SEEK_SET); serr != nil {
				fmt.Printf("seek error: %v\n", serr)
			}
			line, col, highlight := highlightBytePosition(f, serr.Offset)
			extra = fmt.Sprintf(":\nError at line %d, column %d (file offset %d):\n%s",
				line, col, serr.Offset, highlight)
		}
		return fmt.Errorf("error parsing JSON object in config file %s%s\n%v",
			f.Name(), extra, err)
	}
	return nil
}

// HighlightBytePosition takes a reader and the location in bytes of a parse
// error (for instance, from json.SyntaxError.Offset) and returns the line, column,
// and pretty-printed context around the error with an arrow indicating the exact
// position of the syntax error.
//
// Lifted from camlistore
func highlightBytePosition(f io.Reader, pos int64) (line, col int, highlight string) {
	line = 1
	br := bufio.NewReader(f)
	lastLine := ""
	thisLine := new(bytes.Buffer)
	for n := int64(0); n < pos; n++ {
		b, err := br.ReadByte()
		if err != nil {
			break
		}
		if b == '\n' {
			lastLine = thisLine.String()
			thisLine.Reset()
			line++
			col = 1
		} else {
			col++
			thisLine.WriteByte(b)
		}
	}
	if line > 1 {
		highlight += fmt.Sprintf("%5d: %s\n", line-1, lastLine)
	}
	highlight += fmt.Sprintf("%5d: %s\n", line, thisLine.String())
	highlight += fmt.Sprintf("%s^\n", strings.Repeat(" ", col+5))
	return
}
