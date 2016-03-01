package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hobeone/gonab/db"
)

func TestCaps(t *testing.T) {
	dbh := db.NewMemoryDBHandle(false, true)
	n := configRoutes(dbh)

	req, err := http.NewRequest("GET", "/gonab/api?t=caps&o=json", nil)
	if err != nil {
		t.Fatalf("Error setting up request: %s", err)
	}
	respRec := httptest.NewRecorder()
	n.ServeHTTP(respRec, req)

	if respRec.Code != http.StatusOK {
		t.Fatalf("Error running caps api: %d", respRec.Code)
	}
}
