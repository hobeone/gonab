package types

import (
	"database/sql"
	"time"
)

//Group struct
type Group struct {
	ID     int64
	Active bool `sql:"index"`
	First  int64
	Last   int64
	Name   string
}

//Release struct
type Release struct {
	ID           int64
	Hash         string
	CreatedAt    time.Time
	Posted       time.Time
	Name         string
	SearchName   string
	OriginalName string
	From         string
	Status       int
	Grabs        int
	Size         int64
	Group        Group
	GroupID      sql.NullInt64
	NZB          string `sql:"size:0" gorm:"column:nzb"`
	// Category
	// Regex
}

// Binary struct
type Binary struct {
	ID         int64
	Hash       string `sql:"size:16"`
	Name       string `sql:"size:512"`
	TotalParts int
	Posted     time.Time
	From       string
	Xref       string `sql:"size:1024"`
	Group      string
	Parts      []Part
	//Regex
	//RegexID
	//Parts
}

// Part struct
type Part struct {
	ID            int64
	Hash          string `sql:"index,size:16"`
	Subject       string `sql:"size:512"`
	TotalSegments int    `sql:"index"`
	Posted        time.Time
	From          string
	Xref          string `sql:"size:1024"`
	Group         string `sql:"index"`
	Binary        Binary
	BinaryID      sql.NullInt64
	Segments      []Segment
}

//Segment struct
type Segment struct {
	ID        int64
	Segment   int
	Size      int64
	MessageID string
	Part      Part
	PartID    sql.NullInt64
}

// Regex Comment
type Regex struct {
	ID          int
	Regex       string `sql:"size:2048"`
	Description string
	Status      bool
	Ordinal     int
	GroupName   string
}

// Size computes size of Binary
func (b *Binary) Size() int64 {
	size := int64(0)
	for _, p := range b.Parts {
		for _, s := range p.Segments {
			size = size + s.Size
		}
	}
	return size
}
