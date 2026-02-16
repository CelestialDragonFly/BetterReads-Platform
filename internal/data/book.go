package data

import "time"

type LibraryBook struct {
	UserID     string
	BookID     string
	Title      string
	AuthorName string
	BookImage  string
	Rating     int32
	Source     int32
	ShelfIDs   []string
	AddedAt    time.Time
	UpdatedAt  time.Time
}
