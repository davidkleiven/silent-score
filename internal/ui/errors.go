package ui

import "errors"

var (
	ErrWrongTimeFormat    = errors.New("time must be given in the format HH:MM:SS (e.g. 00:01:30)")
	ErrTempoMustBeInteger = errors.New("tempo must be an integer")
)
