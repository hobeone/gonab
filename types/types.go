package types

import (
	"database/sql"
	"regexp"
	"time"
)

//Group struct
type Group struct {
	ID     int64
	Active bool `sql:"index"`
	First  int64
	Last   int64
	Name   string `sql:"unique"`
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
	GroupName  string
	Parts      []Part
	//Regex
	//RegexID
}

// Part struct
type Part struct {
	ID            int64
	Hash          string `sql:"index;size:16"`
	Subject       string `sql:"size:512"`
	TotalSegments int    `sql:"index"`
	Posted        time.Time
	From          string
	Xref          string `sql:"size:1024"`
	GroupName     string `sql:"index"`
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

// MissedMessage represents a message we expected to get from the server in an
// OVERVIEW range but wasn't returned.  Save the message id for retry later.
type MissedMessage struct {
	ID            int64
	MessageNumber int64
	GroupName     string
	Attempts      int
}

// Regex Comment
type Regex struct {
	ID          int
	Regex       string `sql:"size:2048"`
	Description string
	Status      bool
	Ordinal     int
	GroupName   string
	Compiled    *RegexpUtil `sql:"-"` // Ignore for DB
}

// Compile the Regex and stores it in the Compiled attribute.
func (r *Regex) Compile() error {
	c, err := regexp.Compile(r.Regex)
	if err != nil {
		return err
	}
	r.Compiled = &RegexpUtil{Regex: c}
	return nil
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
