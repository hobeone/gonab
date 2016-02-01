package main

import (
	"testing"

	"github.com/hobeone/gonab/db"
)

func TestPrintParts(t *testing.T) {
	dbh := db.NewMemoryDBHandle(true, true)
	dbh.ListParts()
}
