package db

import (
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/hobeone/gonab/types"
)

func TestSearchReleases(t *testing.T) {
	dbh := NewMemoryDBHandle(true)

	r := types.Release{
		Name:       "foo",
		SearchName: "foo.bar",
		Category:   types.DBCategory{ID: int64(types.TV_HD)},
	}

	err := dbh.DB.Create(&r).Error
	if err != nil {
		t.Fatalf("Error creating release: %s", err)
	}

	dbrel, err := dbh.SearchReleases("foo", 0, 10, []types.Category{types.TV_HD, types.TV_SD})
	if err != nil {
		t.Fatalf("Error searching for release: %s", err)
	}
	spew.Dump(dbrel)
}
