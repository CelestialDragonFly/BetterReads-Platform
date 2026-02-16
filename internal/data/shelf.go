package data

import "time"

type Shelf struct {
	ID        string
	Name      string
	UserID    string
	CreatedAt time.Time
	UpdatedAt time.Time
}
