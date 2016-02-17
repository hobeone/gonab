package db

import "database/sql"

// Category is the DB representation of a category
type Category struct {
	ID             int64
	Name           string
	Active         bool
	Description    string
	DisablePreview string
	MinSize        int64
	ParentID       sql.NullInt64
}
