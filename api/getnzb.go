package api

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/mholt/binding"
)

type nzbReq struct {
	APIKey      string
	ReleaseHash string
}

func (n *nzbReq) FieldMap(req *http.Request) binding.FieldMap {
	return binding.FieldMap{
		&n.APIKey: binding.Field{
			Form:     "apikey",
			Required: true,
		},
		&n.ReleaseHash: binding.Field{
			Form:     "h",
			Required: true,
		},
	}
}

func handleBindingErrors(rw http.ResponseWriter, errs binding.Errors) {
	http.Error(rw, "Error", http.StatusBadRequest)
	for _, err := range errs {
		if err.Kind() == binding.RequiredError {
			fmt.Fprintf(rw, "Missing Required Argument(s): %s\n", strings.Join(err.Fields(), ","))
			continue
		}
		fmt.Fprintf(rw, "%s: %v\n", err.Kind(), err.Fields())
	}
}

func nzbDownloadHandler(rw http.ResponseWriter, r *http.Request) {
	reqargs := new(nzbReq)
	errs := binding.Bind(r, reqargs)
	if errs.Len() > 0 {
		handleBindingErrors(rw, errs)
		return
	}

	dbh := getDB(r)
	rel, err := dbh.FindReleaseByHash(reqargs.ReleaseHash)
	if err != nil {
		http.Error(rw, "No release found", http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s.nzb", rel.Name))
	rw.Header().Set("Content-Type", "application/x-nzb")
	rw.Header().Set("Content-Length", strconv.Itoa(len(rel.NZB)))
	fmt.Fprintf(rw, rel.NZB)
}
