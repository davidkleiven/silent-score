package db

import "errors"

var (
	ErrProjectIdsNotUnique = errors.New("records can only be inserted for one project at the time")
)
