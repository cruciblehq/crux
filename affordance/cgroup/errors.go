package cgroup

import "errors"

var (
	ErrInvalidGrant = errors.New("invalid cgroup grant")
	ErrConflict     = errors.New("cgroup conflict")
	ErrUnknownKnob  = errors.New("unknown cgroup knob")
)
