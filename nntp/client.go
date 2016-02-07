package nntputil

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/OneOfOne/xxhash/native"
	log "github.com/Sirupsen/logrus"
	"github.com/hobeone/gonab/db"
	"github.com/hobeone/gonab/types"
	"github.com/hobeone/nntp"
	"github.com/jinzhu/gorm"
)

const defaultMaxOverview = 10000

//NNTPClient comment
type NNTPClient struct {
	c       NNTPConnection
	MaxScan int
}

//NNTPConnection is for creating fakes in testing
type NNTPConnection interface {
	Group(group string) (*nntp.Group, error)
	Overview(begin, end int64) ([]nntp.MessageOverview, error)
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

// GroupScanForward looks for new messages in a particular Group
func (n *NNTPClient) GroupScanForward(dbh *db.Handle, group string, limit int) error {
	nntpGroup, err := n.c.Group(group)
	if err != nil {
		return err
	}
	g, err := dbh.FindGroupByName(group)
	if err != nil {
		return err
	}
	if g.First == 0 {
		log.Infof("DB Group First not set, setting to lowest message in group: %d", nntpGroup.Low)
		g.First = nntpGroup.Low
	}
	if g.Last == 0 {
		log.Infof("DB Group Last seen not set, setting to most recent message minus max to fetch: %d", nntpGroup.High-int64(limit))
		g.Last = nntpGroup.High - int64(limit)
	}
	if g.First < nntpGroup.Low {
		log.Infof("Group %s first article was older than first on server (%d < %d), resetting to %d", g.Name, g.First, nntpGroup.Low, nntpGroup.Low)
		g.First = nntpGroup.Low
	}
	if g.Last > nntpGroup.High {
		log.Errorf("Group %s last article is newer than on server (%d > %d), resetting to %d.", g.Name, g.Last, nntpGroup.High, nntpGroup.High)
	}
	err = dbh.DB.Save(g).Error
	if err != nil {
		return err
	}

	newMessages := nntpGroup.High - g.Last
	maxToGet := g.Last + int64(limit)
	if maxToGet > nntpGroup.High {
		maxToGet = nntpGroup.High
	}
	log.Infof("%d new articles limited to getting just %d. (%d - %d)", newMessages, limit, g.Last, maxToGet)
	log.Debugf("Max messages per overview = %d", n.MaxScan)
	begin := g.Last + 1
	o := []nntp.MessageOverview{}
	missedMessages := types.NewMessageNumberSet()
	for begin < maxToGet {
		toGet := begin + int64(n.MaxScan) - 1
		if toGet > maxToGet {
			toGet = maxToGet
		}
		if toGet < begin {
			toGet = begin
		}
		log.Debugf("Getting %d-%d", begin, toGet)
		overviews, err := n.c.Overview(begin, toGet)
		if err != nil {
			return err
		}
		mm := findMissingMessages(begin, toGet, overviews)
		missedMessages = missedMessages.Union(mm)
		log.Debugf("Got %d messages and %d missed messages", len(overviews), mm.Cardinality())
		o = append(o, overviews...)
		begin = toGet + 1
		g.Last = toGet
	}
	log.Infof("Got %d messages and %d missed messages", len(o), missedMessages.Cardinality())

	parts := overviewToParts(dbh, g.Name, o)
	tx := dbh.DB.Begin()
	var txErr error
	for _, p := range parts {
		txErr = tx.Save(p).Error
		if txErr != nil {
			tx.Rollback()
			return txErr
		}
	}
	txErr = saveMissedMessages(tx, g.Name, missedMessages)
	if txErr != nil {
		tx.Rollback()
		return txErr
	}
	txErr = tx.Save(&g).Error
	if txErr != nil {
		tx.Rollback()
		return txErr
	}
	tx.Commit()

	return nil
}

func saveMissedMessages(tx *gorm.DB, groupName string, ms types.MessageNumberSet) error {
	// Get existing misses in the range for the group
	// Find previously missed and increment their attempt
	// Save those
	// Create new ones0

	for id := range ms.Iter() {
		var dbMissed types.MissedMessage
		err := tx.Where("group_name = ? and message_number = ?", groupName, id).First(&dbMissed).Error
		if err != nil {
			dbMissed = types.MissedMessage{
				MessageNumber: int64(id),
				GroupName:     groupName,
				Attempts:      1,
			}
		} else {
			dbMissed.Attempts++
		}
		err = tx.Save(&dbMissed).Error
		if err != nil {
			return err
		}
	}
	return nil
}

var segmentRegexp = regexp.MustCompile(`\((\d+)[\/](\d+)\)`)

func overviewToParts(dbh *db.Handle, group string, overviews []nntp.MessageOverview) map[string]*types.Part {
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
				part, err := dbh.FindPartByHash(hash)
				if err != nil {
					log.Debugf("New part found: %s", newSub)
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
				} else {
					dbh.DB.Model(part).Association("Segments").Append(seg)
					parts[hash] = part
				}
			}
		}
	}
	return parts
}
