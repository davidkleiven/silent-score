package ui

import "errors"

var (
	ErrTempoMustBeInteger    = errors.New("tempo must be an integer")
	ErrDurationMustBeInteger = errors.New("duration must be an integer")
)
