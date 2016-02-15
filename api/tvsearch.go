package api

import (
	"fmt"
	"html/template"
	"net/http"

	"gopkg.in/unrolled/render.v1"

	"github.com/mholt/binding"
)

type tvSearchReq struct {
	searchReq
	TvRageID string
	Season   string
	Episode  string
}

func (s *tvSearchReq) FieldMap(req *http.Request) binding.FieldMap {
	fm := s.searchReq.FieldMap(req)
	fm[&s.TvRageID] = "rid"
	fm[&s.Season] = "season"
	fm[&s.Episode] = "ep"
	return fm
}

func tvSearchHandler(rw http.ResponseWriter, r *http.Request) {
	searchrequest := new(tvSearchReq)
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

	sr := searchResponse{
		URL:          r.RequestURI,
		ContactEmail: "foo@bar.com",
		Offset:       searchrequest.Offset,
		Total:        len(releases),
		Header:       template.HTML(`<?xml version="1.0" encoding="utf-8" ?>`),
		Image: &rssImage{
			URL:         "http://localhost/foo.jpg",
			Title:       "gonab tvsearch",
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
