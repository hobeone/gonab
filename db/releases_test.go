package db

import (
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/hobeone/gonab/types"
)

func TestSearchReleases(t *testing.T) {
	dbh := NewMemoryDBHandle(false, true)

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
	if len(dbrel) != 1 {
		t.Fatalf("Expected length 1 for search result, got %d", len(dbrel))
	}
}

func TestMakeReleases(t *testing.T) {
	logrus.SetLevel(logrus.ErrorLevel)
	dbh := NewMemoryDBHandle(false, true)
	loadFixtures(dbh)
	err := dbh.MakeBinaries()
	if err != nil {
		t.Fatalf("Error making binaries: %s", err)
	}

	err = dbh.MakeReleases()
	if err != nil {
		t.Fatalf("Error making releases: %s", err)
	}
	var rels []types.Release

	err = dbh.DB.Find(&rels).Error
	if err != nil {
		t.Fatalf("Error getting releases: %v", err)
	}
	if len(rels) != 47 {
		t.Fatalf("Unexpected number of releases: %d != 47", len(rels))
	}
}
