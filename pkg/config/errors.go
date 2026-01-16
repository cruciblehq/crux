package config

import "errors"

var (
	ErrProviderNotFound    = errors.New("provider not found")
	ErrInvalidProvider     = errors.New("invalid provider configuration")
	ErrInvalidProviderType = errors.New("invalid provider type")
	ErrEmptyCredentials    = errors.New("empty credentials")
	ErrNoProvider          = errors.New("no provider configured")
)
