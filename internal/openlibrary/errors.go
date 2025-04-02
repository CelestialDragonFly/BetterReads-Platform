package openlibrary

import "errors"

var (
	ErrInternalServer = errors.New("internal server error")
	ErrNotFound       = errors.New("not found")
	ErrBadRequest     = errors.New("bad request")
)
