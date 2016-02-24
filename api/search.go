package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/hobeone/gonab/types"
	"github.com/mholt/binding"
	"gopkg.in/unrolled/render.v1"
)

type searchReq struct {
	APIKey     string
	Query      string
	Groups     string
	Limit      int
	Categories string
	Output     string
	Attrs      string
	Extended   bool
	Delete     bool
	MaxAge     int
	Offset     int
}

func (s *searchReq) FieldMap(req *http.Request) binding.FieldMap {
	return binding.FieldMap{
		&s.APIKey: binding.Field{
			Form:     "apikey",
			Required: true,
		},
		&s.Query: binding.Field{
			Form:     "q",
			Required: true,
		},
		&s.Groups:     "group",
		&s.Limit:      "limit",
		&s.Categories: "cat",
		&s.Output:     "o",
		&s.Attrs:      "attrs",
		&s.Extended:   "extended",
		&s.Delete:     "del",
		&s.MaxAge:     "maxage",
		&s.Offset:     "offset",
	}
}

type rssImage struct {
	URL         string
	Title       string
	Link        string
	Description string
	Width       int
	Height      int
}

type searchResponse struct {
	URL          string
	ContactEmail string
	Offset       int
	Total        int
	Category     string
	Image        *rssImage
	NZBs         []NZB
}

// NZB represents a single NZB item in search results.
type NZB struct {
	Title       string
	Link        string
	Size        int64
	Category    string
	GUID        string
	PermaLink   bool
	Comments    string
	Description string
	Author      string
	Date        time.Time
}

const maxSearchResults = 100

func searchHandler(rw http.ResponseWriter, r *http.Request) {
	rend := render.New()
	searchrequest := new(searchReq)
	errs := binding.Bind(r, searchrequest)
	if errs.Len() > 0 {
		handleBindingErrors(rw, errs)
		return
	}

	dbh := getDB(r)
	releases, err := dbh.SearchReleasesByName(searchrequest.Query)
	if err != nil {
		rend.Text(rw, http.StatusInternalServerError, fmt.Sprintf("Error: %v", err))
		return
	}

	sr := &searchResponse{
		URL:          r.RequestURI,
		ContactEmail: "foo@bar.com",
		Offset:       searchrequest.Offset,
		Total:        len(releases),
		Image: &rssImage{
			URL:         "http://localhost/foo.jpg",
			Title:       "gonab",
			Link:        "myurl",
			Description: "visit gonab",
		},
	}
	sr.NZBs = make([]NZB, len(releases))
	for i, rel := range releases {
		sr.NZBs[i] = NZB{
			Title:       rel.Name,
			Link:        makeNZBUrl(rel),
			Size:        rel.Size,
			Category:    rel.Category.Name,
			Description: rel.Name,
			GUID:        rel.Hash,
			PermaLink:   true,
			Comments:    fmt.Sprintf("https://localhost/nzb/details/%s#comments", rel.Hash),
			Date:        rel.Posted,
		}
	}

	searchResponseTemplate.Execute(rw, sr)
}

func makeNZBUrl(rel types.Release) string {
	return fmt.Sprintf("https://localhost/getnzb/%s", rel.Hash)
}
