package db

import (
	"database/sql"
	"fmt"
	"testing"

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
	cats, err := dbh.GetCategories()
	if err != nil {
		t.Fatalf("Error: %s", err)
	}
	for _, c := range cats {
		fmt.Printf("<category id=\"%d\" name=\"%s\">\n", c.ID, c.Name)
		for _, s := range c.SubCategories {
			fmt.Printf("	<subcat id=\"%d\" \"%s\"/>\n", s.ID, s.Name)
		}
		fmt.Println("</category>")
	}
	/*
		var cats []types.DBCategory
		err = dbh.DB.Preload("Parent").Find(&cats).Error
		if err != nil {
			t.Fatal(err)
		}

		for _, c := range cats {
			if c.IsParent() {
				fmt.Println(c.Name)
			} else {
				fmt.Println(c.Name)
			}
		}
	*/
}
