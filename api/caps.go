package api

import (
	"net/http"

	"gopkg.in/unrolled/render.v1"

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

type category struct {
	ID             int
	Title          string
	ParentID       int
	Status         int
	Description    string
	DisablePreview int
	MinSize        int
	SubCategories  []category
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
	Available       string
	SupportedParams string
}

type capsResp struct {
	Server capServer `json:"server"`
	Limits capLimits `json:"limits"`
	//Registration map[string]string     `json:"registration"`
	//Searching    map[string]searchResp `json:"searching"`
	Categories []category `json:"categories"`
}

func capsHandler(rw http.ResponseWriter, r *http.Request) {
	rend := render.New(render.Options{
		IndentJSON:   true,
		UnEscapeHTML: true,
	})
	cr := &capsResp{}
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
	/*
		cr.Registration = map[string]string{
			"available": "yes",
			"open":      "no",
		}
		cr.Searching = map[string]searchResp{
			"search": {
				Available:       "yes",
				SupportedParams: "q",
			},
		}

	*/
	cr.Categories = []category{
		{
			ID:             0,
			Title:          "Other",
			Status:         1,
			DisablePreview: 0,
			MinSize:        0,
			SubCategories: []category{
				{
					ID:             20,
					Title:          "Hashed",
					ParentID:       0,
					Status:         1,
					Description:    "Hashed",
					DisablePreview: 0,
					MinSize:        0,
				},
			},
		},
	}
	rend.XML(rw, http.StatusOK, cr)
}
