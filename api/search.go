package api

import (
	"fmt"
	"html/template"
	"net/http"
	"time"

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
	URL         string `xml:"url"`
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description,omitempty"`
	Width       int    `xml:"width,omitempty"`
	Height      int    `xml:"height,omitempty"`
}

type searchResponse struct {
	URL          string
	ContactEmail string
	Offset       int
	Total        int
	Header       template.HTML
	Category     string
	Image        *rssImage
	NZBs         []rawNZB
}

// rawNZB represents a single NZB item in search results.
type rawNZB struct {
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

// https://newznab.readthedocs.org/en/latest/misc/api/#search
func searchHandler(rw http.ResponseWriter, r *http.Request) {
	searchrequest := new(searchReq)
	errs := binding.Bind(r, searchrequest)
	// Better error handling
	if errs.Handle(rw) {
		return
	}
	rend := render.New()

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
		Header:       template.HTML(`<?xml version="1.0" encoding="utf-8" ?>`),
		Image: &rssImage{
			URL:         "http://localhost/foo.jpg",
			Title:       "gonab",
			Link:        "myurl",
			Description: "visit gonab",
		},
	}
	sr.NZBs = make([]rawNZB, len(releases))
	for i, rel := range releases {
		sr.NZBs[i] = rawNZB{
			Title:       rel.Name,
			Link:        "https://www.packetspike.net:81/getnzb/1487ae6e44de505ffe43bae63269d2c693142ce0.nzb&i=1&r=d46718373be42282260a342878962f35",
			Size:        rel.Size,
			Category:    "TV > HD",
			Description: rel.Name,
			GUID:        rel.Hash,
			PermaLink:   true,
			Comments:    "http://localhost/details/nzbhash22222222222222222222222222222222222222222222222222222#comments",
			Date:        rel.Posted,
		}
	}

	searchResponseTemplate.Execute(rw, sr)
}
