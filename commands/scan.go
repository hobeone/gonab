package commands

import (
	"fmt"
	"sync"

	"github.com/Sirupsen/logrus"
	"github.com/hobeone/gonab/config"
	"github.com/hobeone/gonab/db"
	"github.com/hobeone/gonab/nntp"
	"github.com/hobeone/gonab/types"
	"gopkg.in/alecthomas/kingpin.v2"
)

// ScanCommand comment
type ScanCommand struct {
	MaxArticles int
	MaxConns    int
	MaxChunk    int
	Group       string
}

func (s *ScanCommand) configure(app *kingpin.Application) {
	cmd := app.Command("scan", "scan for new messages").Action(s.scan)
	cmd.Flag("limit", "Limit scan to this many messages starting at the oldest.  -1 means get all new messages.").Default("-1").IntVar(&s.MaxArticles)
	cmd.Flag("chunk", "Limit scan to this many messages per overview command to the server").Default("10000").IntVar(&s.MaxChunk)
	cmd.Flag("conn", "Limit to this many simultanious connections.").IntVar(&s.MaxConns)
	cmd.Flag("group", "Only scan this group.").StringVar(&s.Group)
}

// GroupScanner is designed to be run in a goroutine and take requests for
// groups to scan.
type GroupScanner struct {
	conn  *nntputil.NNTPClient
	dbh   *db.Handle
	ident string
}

// NewGroupScanner returns a new GroupScanner
func NewGroupScanner(cfg *config.Config, dbh *db.Handle, ident string) (*GroupScanner, error) {
	n, err := nntputil.ConnectAndAuthenticate(
		fmt.Sprintf("%s:%d", cfg.NewsServer.Host, cfg.NewsServer.Port),
		cfg.NewsServer.Username,
		cfg.NewsServer.Password,
		cfg.NewsServer.UseTLS,
	)
	if err != nil {
		return nil, err
	}
	return &GroupScanner{
		conn:  n,
		dbh:   dbh,
		ident: ident,
	}, nil
}

// Close the connection
func (g *GroupScanner) Close() {
	g.conn.Quit()
}

func (g *GroupScanner) scanGroup(req *scanRequest) *scanResponse {
	g.conn.MaxScan = req.MaxChunk
	articleCount, err := g.conn.GroupScanForward(g.dbh, req.Group, req.Max)
	return &scanResponse{
		Group:    req.Group,
		Articles: articleCount,
		Error:    err,
	}
}

// ScanLoop takes requests for groups to scan, scans them and returns the
// result.
func (g *GroupScanner) ScanLoop(scanRequests chan *scanRequest, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case req, ok := <-scanRequests:
			// channel closed
			if !ok {
				logrus.WithField("worker", g.ident).Debugf("Request channel closed, exiting scanner.")
				return
			}
			logrus.WithField("worker", g.ident).Debugf("Got request for group %s", req.Group)
			req.ResponseChan <- g.scanGroup(req)
		}
	}
}

type scanRequest struct {
	Group        string
	Max          int
	MaxChunk     int
	ResponseChan chan *scanResponse
}

type scanResponse struct {
	Group    string
	Articles int
	Error    error
}

func (s *ScanCommand) scan(c *kingpin.ParseContext) error {
	if *debug {
		logrus.SetLevel(logrus.DebugLevel)
	}
	cfg := loadConfig(*configfile)

	if s.MaxConns < 1 {
		s.MaxConns = cfg.NewsServer.MaxConns
	}
	if s.MaxConns < 1 {
		s.MaxConns = 1
	}

	dbh := db.NewDBHandle(cfg.DB.Name, cfg.DB.Username, cfg.DB.Password, cfg.DB.Verbose)
	var groups []types.Group
	if s.Group != "" {
		g, err := dbh.FindGroupByName(s.Group)
		if err != nil {
			return fmt.Errorf("Error getting group %s, have you added it?", s.Group)
		}
		groups = []types.Group{*g}
	} else {
		var err error
		groups, err = dbh.GetActiveGroups()
		if err != nil {
			return err
		}
		if len(groups) == 0 {
			return fmt.Errorf("No active groups to scan.")
		}
	}
	logrus.Debugf("Got %d groups to scan.", len(groups))

	connsToMake := s.MaxConns
	if len(groups) < s.MaxConns {
		connsToMake = len(groups)
	}

	reqchan := make(chan *scanRequest)
	respchan := make(chan *scanResponse, len(groups))
	var wg sync.WaitGroup

	for i := 0; i < connsToMake; i++ {
		g, err := NewGroupScanner(cfg, dbh, fmt.Sprintf("%d", i))
		if g != nil {
			defer g.Close()
		}
		if err != nil {
			return err
		}
		fmt.Printf("Started scanner %d\n", i)
		wg.Add(1)
		go g.ScanLoop(reqchan, &wg)
	}

	for _, g := range groups {
		logrus.Debugf("Requesting scan of %s", g.Name)
		reqchan <- &scanRequest{
			Group:        g.Name,
			Max:          s.MaxArticles,
			MaxChunk:     s.MaxChunk,
			ResponseChan: respchan,
		}
	}
	close(reqchan) // Causes workers to exit
	wg.Wait()
	close(respchan) // Causes range below to be non-infinite

	for r := range respchan {
		fmt.Printf("Finished scanning %s\n", r.Group)
		fmt.Printf("  %d new Messages\n", r.Articles)
		fmt.Printf("Error: %s\n", r.Error)
	}

	return nil
}
