package nntputil

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/OneOfOne/xxhash/native"
	"github.com/Sirupsen/logrus"
	"github.com/hobeone/gonab/db"
	"github.com/hobeone/gonab/types"
	"github.com/hobeone/nntp"
)

const defaultMaxOverview = 100000

//NNTPClient comment
type NNTPClient struct {
	c          NNTPConnection
	MaxScan    int
	SaveMissed bool
}

//NNTPConnection is for creating fakes in testing
type NNTPConnection interface {
	Group(group string) (*nntp.Group, error)
	Overview(begin, end int64) ([]nntp.MessageOverview, error)
	Quit() error
}

// NewClient returns a NNTPClient with the given connection and defaults set.
func NewClient(c NNTPConnection) *NNTPClient {
	return &NNTPClient{
		c:       c,
		MaxScan: defaultMaxOverview,
	}
}

//ConnectAndAuthenticate returns a NNTPClient that is authenticated to the
//server
//TODO: allow for different SSL configs.
func ConnectAndAuthenticate(server, username, password string, useSSL bool) (*NNTPClient, error) {
	var c *nntp.Conn
	var err error

	if useSSL {
		c, err = nntp.NewTLS("tcp", server, nil)
	} else {
		c, err = nntp.New("tcp", server)
	}
	if err != nil {
		return nil, err
	}
	if username != "" {
		err = c.Authenticate(username, password)
		if err != nil {
			return nil, err
		}
	}
	return NewClient(c), nil
}

// Quit closes the underlying connection.
func (n *NNTPClient) Quit() {
	n.c.Quit()
}

// return a hex string rather than the native uint64 as go's sql module doesn't
// deal with those.
func hashOverview(sub, from, groupName string, segmentTotal int) string {
	h := xxhash.New64()
	h.Write([]byte(sub + from + groupName + string(segmentTotal)))
	return fmt.Sprintf("%x", h.Sum64())
}

// Given a start and end message number find the
func findMissingMessages(begin, end int64, overviews []nntp.MessageOverview) types.MessageNumberSet {
	fullset := types.NewMessageNumberSet()
	messages := types.NewMessageNumberSet()
	for _, o := range overviews {
		messages.Add(types.MessageNumber(o.MessageNumber))
	}
	for i := begin; i <= end; i++ {
		fullset.Add(types.MessageNumber(i))
	}
	missed := fullset.Difference(messages)
	return missed
}

// GroupScanForward looks for new messages in a particular Group.
// Returns the number of articles scanned and if an error was encountered
func (n *NNTPClient) GroupScanForward(dbh *db.Handle, group string, limit int) (int, error) {
	ctxLogger := logrus.WithFields(
		logrus.Fields{
			"group": group,
		},
	)
	nntpGroup, err := n.c.Group(group)
	if err != nil {
		return 0, err
	}
	g, err := dbh.FindGroupByName(group)
	if err != nil {
		return 0, err
	}
	if g.First == 0 {
		ctxLogger.Infof("DB Group First not set, setting to lowest message in group: %d", nntpGroup.Low)
		g.First = nntpGroup.Low
	}
	if g.Last == 0 {
		if limit > 0 {
			ctxLogger.Infof("DB Group Last seen not set, setting to most recent message minus max to fetch: %d", nntpGroup.High-int64(limit))
			g.Last = nntpGroup.High - int64(limit)
		} else {
			ctxLogger.Infof("DB Group Last seen not set and no limit given, using default of %d", defaultMaxOverview)
			g.Last = nntpGroup.High - defaultMaxOverview
			if g.Last < nntpGroup.Low {
				g.Last = nntpGroup.Low
			}
		}
	}
	if g.First < nntpGroup.Low {
		ctxLogger.Infof("Group %s first article was older than first on server (%d < %d), resetting to %d", g.Name, g.First, nntpGroup.Low, nntpGroup.Low)
		g.First = nntpGroup.Low
	}
	if g.Last > nntpGroup.High {
		ctxLogger.Errorf("Group %s last article is newer than on server (%d > %d), resetting to %d.", g.Name, g.Last, nntpGroup.High, nntpGroup.High)
		g.Last = nntpGroup.High
	}
	err = dbh.DB.Save(g).Error
	if err != nil {
		return 0, err
	}

	newMessages := nntpGroup.High - g.Last
	var maxToGet int64
	if limit > 0 {
		maxToGet = g.Last + int64(limit)
	} else {
		maxToGet = nntpGroup.High
	}
	if maxToGet > nntpGroup.High {
		maxToGet = nntpGroup.High
	}
	if newMessages < 1 {
		ctxLogger.Info("No new articles")
		return 0, nil
	}
	ctxLogger.Infof("%d new articles limited to getting just %d (%d - %d) in %d article chunks", newMessages, maxToGet-g.Last, g.Last, maxToGet, n.MaxScan)
	begin := g.Last + 1
	totalArticles := 0
	missedMessages := 0
	for begin < maxToGet {
		toGet := begin + int64(n.MaxScan) - 1
		if toGet > maxToGet {
			toGet = maxToGet
		}
		if toGet < begin {
			toGet = begin
		}
		ctxLogger.Debugf("Getting %d-%d", begin, toGet)
		overviews, err := n.c.Overview(begin, toGet)
		if err != nil {
			return len(overviews), err
		}
		var mm types.MessageNumberSet
		if n.SaveMissed {
			mm = findMissingMessages(begin, toGet, overviews)
			missedMessages = missedMessages + mm.Cardinality()
			ctxLogger.Debugf("Got %d messages and %d missed messages", len(overviews), mm.Cardinality())
		} else {
			ctxLogger.Debugf("Got %d messages", len(overviews))
		}
		totalArticles = totalArticles + len(overviews)
		ctxLogger.Debugf("Saving parts and messages to db.")
		err = saveOverviewBatch(dbh, g.Name, overviews, mm)
		if err != nil {
			return len(overviews), err
		}
		begin = toGet + 1
		g.Last = toGet

		err = dbh.DB.Save(g).Error
		if err != nil {
			return 0, err
		}
	}
	if n.SaveMissed {
		ctxLogger.Infof("Got %d messages and %d missed messages", totalArticles, missedMessages)
	} else {
		ctxLogger.Debugf("Got %d messages", totalArticles)
	}
	err = dbh.DB.Save(g).Error
	if err != nil {
		return 0, err
	}

	return totalArticles, nil
}

var segmentRegexp = regexp.MustCompile(`\((\d+)[\/](\d+)\)`)

func saveOverviewBatch(dbh *db.Handle, group string, overviews []nntp.MessageOverview, missed types.MessageNumberSet) error {
	parts := map[string]*types.Part{}

	for _, o := range overviews {
		m := segmentRegexp.FindStringSubmatch(o.Subject)
		if m != nil {
			segNum, _ := strconv.Atoi(m[1])
			segTotal, _ := strconv.Atoi(m[2])
			// Strip the segment information to match the subject to other parts
			newSub := strings.Replace(o.Subject, m[0], "", -1)
			newSub = strings.TrimSpace(newSub)

			hash := hashOverview(newSub, o.From, group, segTotal)
			seg := types.Segment{
				MessageID: o.MessageID,
				Segment:   segNum,
				Size:      int64(o.Bytes),
			}
			if part, ok := parts[hash]; ok {
				part.Segments = append(part.Segments, seg)
			} else {
				parts[hash] = &types.Part{
					Hash:          hash,
					Subject:       newSub,
					Posted:        o.Date,
					From:          o.From,
					GroupName:     group,
					TotalSegments: segTotal,
					Xref:          o.Xref(),
					Segments:      []types.Segment{seg},
				}
			}
		}
	}
	mm := make([]types.MissedMessage, missed.Cardinality())
	i := 0
	for id := range missed.Iter() {
		mm[i] = types.MissedMessage{
			MessageNumber: int64(id),
			GroupName:     group,
			Attempts:      1,
		}
		i++
	}
	logrus.Debugf("Found %d new parts, %d missed messages", len(parts), len(mm))
	return dbh.SavePartsAndMissedMessages(parts, mm)
}
