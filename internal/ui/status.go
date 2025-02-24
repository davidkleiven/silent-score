package ui

import "fmt"

type StatusKind int

const (
	okStatus StatusKind = iota
	errorStatus
)

type Status struct {
	kind StatusKind
	msg  string
}

func (s *Status) Render(prefix string) string {
	switch s.kind {
	case errorStatus:
		return pad2.Foreground(errorColor).Render(fmt.Sprintf("[%s] %s\n", prefix, s.msg))
	default:
		return pad2.Foreground(okColor).Render(fmt.Sprintf("[%s] %s\n", prefix, s.msg))
	}
}

func (s *Status) Set(msg string, err error) {
	if err != nil {
		s.kind = errorStatus
		s.msg = err.Error()
	} else {
		s.kind = okStatus
		s.msg = msg
	}
}

func NewStatus() *Status {
	return &Status{
		kind: okStatus,
		msg:  "OK",
	}
}
