package net

import "errors"

var (
	ErrInvalidGrant       = errors.New("invalid net grant")
	ErrInvalidSpec        = errors.New("invalid network spec")
	ErrInvalidIngressRule = errors.New("invalid network ingress rule")
	ErrInvalidEgressRule  = errors.New("invalid network egress rule")
)
