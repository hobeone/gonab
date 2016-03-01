package api

import (
	"fmt"
	"html/template"
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
	Categories []int64
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
	Header       template.HTML
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
	Group       string
}

const maxSearchResults = 100
const defaultLimit = 10

func searchHandler(rw http.ResponseWriter, r *http.Request) {
	rend := render.New()
	searchrequest := new(searchReq)
	errs := binding.Bind(r, searchrequest)
	if errs.Len() > 0 {
		handleBindingErrors(rw, errs)
		return
	}
	if searchrequest.Limit == 0 {
		searchrequest.Limit = defaultLimit
	}

	dbh := getDB(r)
	var cats []types.Category
	for _, c := range searchrequest.Categories {
		cats = append(cats, types.CategoryFromInt(c))
	}

	releases, err := dbh.SearchReleases(searchrequest.Query, searchrequest.Offset, searchrequest.Limit, cats)
	if err != nil {
		rend.Text(rw, http.StatusInternalServerError, fmt.Sprintf("Error: %v", err))
		return
	}

	sr := &searchResponse{
		Header:       template.HTML(`<?xml version="1.0" encoding="UTF-8"?>`),
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
			Link:        makeNZBUrl(rel, r),
			Size:        rel.Size,
			Category:    rel.Category.Name,
			Description: rel.Name,
			GUID:        rel.Hash,
			PermaLink:   true,
			Comments:    fmt.Sprintf("https://%s/nzb/details/%s#comments", r.Host, rel.Hash),
			Date:        rel.Posted,
			Group:       rel.Group.Name,
		}
	}

	searchResponseTemplate.Execute(rw, sr)
}

func makeNZBUrl(rel types.Release, r *http.Request) string {
	return fmt.Sprintf("%s/getnzb?h=%s&apikey=123", getLink(r), rel.Hash)
}
