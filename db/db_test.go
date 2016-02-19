package db

import (
	"database/sql"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/hobeone/gonab/types"
)

func TestDBCategory(t *testing.T) {
	dbh := NewMemoryDBHandle(true)

	rel := types.Release{}
	rel.CategoryID = sql.NullInt64{Int64: int64(types.TV_HD), Valid: true}
	err := dbh.DB.Save(&rel).Error
	if err != nil {
		t.Fatal(err)
	}

	var dbrel types.Release
	err = dbh.DB.Preload("Category").Find(&dbrel, rel.ID).Error
	if err != nil {
		t.Fatal(err)
	}
	spew.Dump(dbrel)
}
