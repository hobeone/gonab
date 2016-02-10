package commands

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/hobeone/gonab/db"
	"github.com/hobeone/gonab/types"
	"github.com/hobeone/rss2go/httpclient"
	"gopkg.in/alecthomas/kingpin.v2"
)

var nnpplusRegexURL = `http://www.newznab.com/getregex.php?newznabID=%s`

//RegexImporter comment
type RegexImporter struct{}

var regexWithOption = regexp.MustCompile(`\/(\w+)$`)

func cleanRegex(re string) string {
	regex := strings.TrimSpace(re)
	regex = strings.Replace(regex, `\\`, `\`, -1)
	regex = strings.TrimLeft(regex, `/`)
	regexOpts := regexWithOption.FindStringSubmatch(regex)
	if regexOpts != nil {
		if strings.Contains(regexOpts[1], "i") {
			regex = fmt.Sprintf("(?i)") + regex // add re2 ignore case option
		}
		regex = strings.TrimRight(regex, regexOpts[0])
	} else {
		regex = strings.TrimRight(regex, "/")
	}
	return regex
}

// parsed order:
// 0 full string
// 1 id
// 2 group
// 3 regex
// 4 ordinal
// 5 status
// 6 description
// 7 categoryID
func newzNabRegexToRegex(parsed []string) (*types.Regex, error) {
	if len(parsed) != 8 {
		return nil, fmt.Errorf("parsed newznab regex should be 8 items")
	}
	status, err := strconv.ParseBool(parsed[5])
	if err != nil {
		logrus.Errorf("Couldn't parse %s to bool, assuming true", parsed[5])
		status = true
	}
	id, err := strconv.Atoi(parsed[1])
	if err != nil {
		logrus.Errorf("Couldn't parse id %s, skipping", parsed[1])
	}
	ord, err := strconv.Atoi(parsed[4])
	if err != nil {
		logrus.Errorf("Couldn't parse ordinal %s, skipping", parsed[4])
	}
	regex := cleanRegex(parsed[3])
	dbregex := types.Regex{
		ID:          id,
		GroupName:   parsed[2],
		Regex:       regex,
		Status:      status,
		Description: parsed[6],
		Ordinal:     ord,
	}
	_, err = regexp.Compile(dbregex.Regex)
	if err != nil {
		return nil, fmt.Errorf("Error compiling regex, skipping: %v", err)
	}
	return &dbregex, nil
}

var splitRegex = regexp.MustCompile(`\((\d+), \'(.*)\', \'(.*)\', (\d+), (\d+), (.*), (.*)\);$`)

func parseNewzNabRegexes(b []byte) ([]*types.Regex, error) {
	r := bufio.NewReader(bytes.NewReader(b))
	newregexes := []*types.Regex{}
	lines := 0
	for {
		lines++
		record, err := r.ReadString('\n')
		if err == io.EOF {
			break
		}

		if lines <= 2 {
			// Skip the revision and delete lines
			continue
		}
		record = strings.TrimSpace(record)
		matches := splitRegex.FindStringSubmatch(record)

		if len(matches) != 8 {
			logrus.Errorf("Invalid line in regex file: %s", record)
			continue
		} else {
			newregex, err := newzNabRegexToRegex(matches)
			if err != nil {
				logrus.Errorf("Couldn't create Regex from %v: %v", record, err)
				continue
			}
			newregexes = append(newregexes, newregex)
		}
	}
	return newregexes, nil
}

// Field Order
// 0 id
// 1 group_regex
// 2 regex
// 3 status
// 4 description
// 5 ordinal
func nzedbRegexToRegex(parsed []string) (*types.Regex, error) {
	if len(parsed) != 6 {
		return nil, fmt.Errorf("nzedb regexes must be 6 items")
	}
	status, err := strconv.ParseBool(parsed[3])
	if err != nil {
		logrus.Errorf("Couldn't parse %s to bool, assuming true", parsed[3])
		status = true
	}
	id, err := strconv.Atoi(parsed[0])
	if err != nil {
		return nil, fmt.Errorf("Couldn't parse id %s, skipping", parsed[0])
	}
	ord, err := strconv.Atoi(parsed[5])
	if err != nil {
		return nil, fmt.Errorf("Couldn't parse ordinal %s, skipping", parsed[5])
	}
	regex := cleanRegex(parsed[2])
	// nzedb has actual regexes for group names but they are mostly full group
	// names.
	groupname := strings.Replace(parsed[1], `\\`, "", -1)
	groupname = strings.TrimRight(groupname, "$")
	groupname = strings.TrimLeft(groupname, "^")
	dbregex := types.Regex{
		ID:          id,
		GroupName:   groupname,
		Regex:       regex,
		Status:      status,
		Description: parsed[4],
		Ordinal:     ord,
	}
	_, err = regexp.Compile(dbregex.Regex)
	if err != nil {
		return nil, fmt.Errorf("Error compiling regex, skipping: %v", err)
	}
	return &dbregex, nil
}

func parseNzedbRegexes(b []byte) ([]*types.Regex, error) {
	r := bufio.NewReader(bytes.NewReader(b))
	newregexes := []*types.Regex{}
	lines := 0

	for {
		lines++
		record, err := r.ReadString('\n')
		if err == io.EOF {
			break
		}

		if lines <= 1 {
			continue
		}
		record = strings.TrimSpace(record)
		fields := strings.Split(record, "\t")
		regex, err := nzedbRegexToRegex(fields)
		if err != nil {
			logrus.Errorf("Error parsing nZEDb regex (line %d): %v", lines, err)
			continue
		}
		newregexes = append(newregexes, regex)
	}
	return newregexes, nil
}

func (regeximporter *RegexImporter) run(c *kingpin.ParseContext) error {
	if *debug {
		logrus.SetLevel(logrus.DebugLevel)
	}
	logrus.Infof("Reading config %s\n", *configfile)

	cfg := loadConfig(*configfile)
	url := cfg.Regex.URL
	logrus.Infof("Crawling %v", url)

	// Defaults to 1 second for connect and read
	connectTimeout := (5 * time.Second)
	readWriteTimeout := (15 * time.Second)

	client := httpclient.NewTimeoutClient(connectTimeout, readWriteTimeout)

	resp, err := client.Get(url)

	if err != nil {
		logrus.Infof("Error getting %s: %s", url, err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("feed %s returned a non 200 status code: %s", url, resp.Status)
		logrus.Error(err)
		return err
	}
	var b []byte
	if resp.ContentLength > 0 {
		b = make([]byte, resp.ContentLength)
		_, err := io.ReadFull(resp.Body, b)
		if err != nil {
			return fmt.Errorf("error reading response for %s: %s", url, err)
		}
	} else {
		b, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("error reading response for %s: %s", url, err)
		}
	}

	var regexes []*types.Regex
	switch cfg.Regex.Type {
	case "nnplus":
		regexes, err = parseNewzNabRegexes(b)
	case "nzedb":
		regexes, err = parseNzedbRegexes(b)
	default:
		return fmt.Errorf("Unknown Regex type: %s", cfg.Regex.Type)
	}
	if err != nil {
		return err
	}

	dbh := db.NewDBHandle(cfg.DB.Name, cfg.DB.Username, cfg.DB.Password, cfg.DB.Verbose)
	tx := dbh.DB.Begin()
	tx.Where("id < ?", 100000).Delete(&types.Regex{})
	for _, dbregex := range regexes {
		err = tx.Create(dbregex).Error
		if err != nil {
			logrus.Errorf("Error saving regex: %v", err)
			tx.Rollback()
			return err
		}
	}
	tx.Commit()
	return nil
}
