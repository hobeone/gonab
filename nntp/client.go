package nntputil

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/OneOfOne/xxhash"
	log "github.com/Sirupsen/logrus"
	"github.com/hobeone/gonab/db"
	"github.com/hobeone/gonab/types"
	"github.com/hobeone/nntp"
)

//NNTPClient comment
type NNTPClient struct {
	c *nntp.Conn
}

//ConnectAndAuthenticate comment
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
	n := &NNTPClient{
		c: c,
	}
	return n, nil
}

var segmentRegexp = regexp.MustCompile(`(?i)\((\d+)[\/](\d+)\)`)

// return a hex string rather than the native uint64 as go's sql module doesn't
// deal with those.
func hashOverview(sub, from, groupName string, segmentTotal int) string {
	h := xxhash.New64()
	h.Write([]byte(sub + from + groupName + string(segmentTotal)))
	return fmt.Sprintf("%x", h.Sum64())
}

var maxScan = 10000

// GroupScanForward comment
func (n *NNTPClient) GroupScanForward(dbh *db.Handle, group string, limit int) ([]nntp.MessageOverview, error) {

	nntpGroup, err := n.c.Group(group)
	if err != nil {
		return nil, err
	}
	g, err := dbh.FindGroupByName(group)
	if err != nil {
		return nil, err
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
		log.Errorf("Group %s first article was older than first on server (%d < %d), resetting to %d", g.Name, g.First, nntpGroup.Low, nntpGroup.Low)
		g.First = nntpGroup.Low
	}
	if g.Last > nntpGroup.High {
		log.Errorf("Group %s last article is newer than on server (%d > %d), resetting to %d.", g.Name, g.Last, nntpGroup.High, nntpGroup.High)
	}
	err = dbh.DB.Save(g).Error
	if err != nil {
		return nil, err
	}

	numToGet := nntpGroup.High - g.Last
	log.Debugf("Need to get %d articles", numToGet)
	log.Debugf("Max overview = %d", maxScan)
	begin := g.Last
	o := []nntp.MessageOverview{}
	for begin < nntpGroup.High {
		toGet := begin + int64(maxScan)
		if toGet > nntpGroup.High {
			toGet = nntpGroup.High
		}
		log.Debugf("Getting %d-%d", begin, toGet)
		overviews, err := n.c.Overview(begin, toGet)
		if err != nil {
			return nil, err
		}
		o = append(o, overviews...)
		begin = toGet + 1
		g.Last = toGet
	}
	parts := Scan(g.Name, o)
	tx := dbh.DB.Begin()
	var txErr error
	for _, p := range parts {
		txErr = tx.Create(p).Error
		if txErr != nil {
			tx.Rollback()
			break
		}
	}
	txErr = tx.Save(&g).Error
	if txErr != nil {
		tx.Rollback()
	} else {
		tx.Commit()
	}
	return o, nil
}

// Scan comment
func Scan(group string, overviews []nntp.MessageOverview) map[string]*types.Part {
	parts := map[string]*types.Part{}

	for _, o := range overviews {
		m := segmentRegexp.FindAllStringSubmatch(o.Subject, -1)
		if len(m) > 0 {
			segNum, _ := strconv.Atoi(m[len(m)-1][1])
			segTotal, _ := strconv.Atoi(m[len(m)-1][2])
			newSub := strings.Replace(o.Subject, m[0][0], "", -1)

			hash := hashOverview(newSub, o.From, group, segTotal)
			seg := types.Segment{
				MessageID: o.MessageID,
				Segment:   segNum,
				Size:      int64(o.Bytes),
			}
			if part, ok := parts[hash]; ok {
				log.Debugf("Adding segment %d to part %s", seg.Segment, newSub)
				part.Segments = append(part.Segments, seg)
			} else {
				log.Debugf("New part found: %s", newSub)
				parts[hash] = &types.Part{
					Hash:          hash,
					Subject:       newSub,
					Posted:        o.Date,
					From:          o.From,
					GroupName:     group,
					TotalSegments: segTotal,
					Segments:      []types.Segment{seg},
				}
			}
		}
	}
	return parts
}
