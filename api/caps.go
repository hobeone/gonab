package api

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"

	"gopkg.in/unrolled/render.v1"

	"github.com/hobeone/gonab/types"
	"github.com/mholt/binding"
)

type capsReq struct {
	Output string
}

func (cr *capsReq) FieldMap(req *http.Request) binding.FieldMap {
	return binding.FieldMap{
		&cr.Output: "o",
	}
}

type capServer struct {
	AppVersion string
	Version    string
	Title      string
	Email      string
	URL        string
	Image      string
	Strapline  string
}
type capLimits struct {
	Max     int
	Default int
}

type searchResp struct {
	Name            string
	Available       string
	SupportedParams string
}

type capsResp struct {
	Header       template.HTML       `json:"-"`
	Server       capServer           `json:"server"`
	Limits       capLimits           `json:"limits"`
	Registration map[string]string   `json:"registration"`
	Searching    []searchResp        `json:"searching"`
	Categories   []*types.DBCategory `json:"categories"`
}

func capsHandler(rw http.ResponseWriter, r *http.Request) {
	rend := render.New(render.Options{
		IndentJSON:   true,
		IndentXML:    true,
		UnEscapeHTML: true,
	})

	capsreq := &capsReq{}
	errs := binding.Bind(r, capsreq)
	if errs.Len() > 0 {
		handleBindingErrors(rw, errs)
		return
	}

	cr := &capsResp{
		Header: template.HTML(`<?xml version="1.0" encoding="UTF-8"?>`),
	}
	cr.Server = capServer{
		AppVersion: "0.0.1",
		Version:    "0.1",
		Title:      "gonab",
		Strapline:  "a Go based usenet indexer",
		Email:      "",
		URL:        "http://localhost/",
		Image:      "",
	}
	cr.Limits = capLimits{
		Max:     100,
		Default: 100,
	}

	dbh := getDB(r)
	cats, err := dbh.GetCategories()
	if err != nil {
		rend.Text(rw, http.StatusInternalServerError, fmt.Sprintf("Error: %v", err))
		return
	}
	cr.Categories = cats

	cr.Registration = map[string]string{
		"Available": "yes",
		"Open":      "no",
	}
	cr.Searching = []searchResp{
		{
			Name:            "search",
			Available:       "yes",
			SupportedParams: "q",
		},
		{
			Name:            "tv-search",
			Available:       "yes",
			SupportedParams: "q,rid,tvdbid,vid,traktid,tvmazeid,imdbid,tmdbid,season,ep",
		},
	}
	b := bytes.NewBuffer([]byte{})
	capsResponseTemplate.Execute(b, cr)
	if capsreq.Output == "json" {
		rend.JSON(rw, http.StatusOK, cr)
	} else {
		rend.Text(rw, http.StatusOK, b.String())
	}
}
