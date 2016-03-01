package db

import (
	"database/sql"
	"testing"

	"github.com/hobeone/gonab/types"
)

func TestDBCategory(t *testing.T) {
	dbh := NewMemoryDBHandle(false, true)

	rel := types.Release{}
	rel.CategoryID = sql.NullInt64{Int64: int64(types.TV_HD), Valid: true}
	err := dbh.DB.Save(&rel).Error
	if err != nil {
		t.Fatal(err)
	}

	var dbrel types.Release
	err = dbh.DB.Preload("Category").Preload("Category.Parent").Find(&dbrel, rel.ID).Error
	if err != nil {
		t.Fatal(err)
	}
	if dbrel.Category.ID != 5040 {
		t.Fatalf("Unexpected category id: %s", dbrel.Category.ID)
	}
	if dbrel.Category.Parent.ID != 5000 {
		t.Fatalf("Unexpected category id: %s", dbrel.Category.Parent.ID)
	}
}
