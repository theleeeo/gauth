package repo

import "errors"

var (
	// ErrNotFound is returned when the requested resource is not found.
	ErrNotFound = errors.New("not found")
	// ErrAlreadyExists is returned when the requested resource already exists.
	ErrAlreadyExists = errors.New("already exists")
)
