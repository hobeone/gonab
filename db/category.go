package db

import (
	"sort"

	"github.com/hobeone/gonab/types"
)

//GetCategories returns a sorted list of parent categories with their cubcategories
//populated.
func (d *Handle) GetCategories() ([]*types.DBCategory, error) {
	var cats []*types.DBCategory
	err := d.DB.Preload("Parent").Find(&cats).Error
	if err != nil {
		return cats, err
	}

	catmap := map[int]*types.DBCategory{}

	for _, c := range cats {
		if c.IsParent() {
			c.SubCategories = []types.DBCategory{}
			catmap[int(c.ID)] = c
		}
	}
	for _, c := range cats {
		if c.Parent != nil {
			c.SubCategories = []types.DBCategory{}
			if _, ok := catmap[int(c.Parent.ID)]; ok {
				p := catmap[int(c.Parent.ID)]
				p.SubCategories = append(p.SubCategories, *c)
			}
		}
	}

	keys := []int{}
	for key := range catmap {
		keys = append(keys, key)
	}
	sort.Ints(keys)
	sortedcats := make([]*types.DBCategory, len(keys))
	for i, key := range keys {
		sortedcats[i] = catmap[key]
	}
	return sortedcats, nil
}
