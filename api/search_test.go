package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/hobeone/gonab/db"
)

func TestSearch(t *testing.T) {
	dbh := db.NewMemoryDBHandle(false)
	n := configRoutes(dbh)

	req, err := http.NewRequest("GET", "/api?t=search&q=foo&apikey=123", nil)
	if err != nil {
		t.Fatalf("Error setting up request: %s", err)
	}
	respRec := httptest.NewRecorder()
	n.ServeHTTP(respRec, req)

	if respRec.Code != http.StatusOK {
		spew.Dump(respRec)
		t.Fatalf("Error running caps api: %d", respRec.Code)
	}
}
