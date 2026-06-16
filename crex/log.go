package crex

import (
	"context"
	"log/slog"
)

// Logs err to logger.
//
// If err is nil, this is a no-op. Otherwise, the error is adapted to a
// headline description and structured value and logged at error level.
func LogError(logger *slog.Logger, err error) {
	if err == nil {
		return
	}
	description, value := surface(err)
	logger.LogAttrs(context.Background(), slog.LevelError, description, slog.Any("error", value))
}

// Returns a headline description and a structured value for err so it renders
// through the crex formatter no matter its type.
func surface(err error) (string, slog.LogValuer) {
	switch e := err.(type) {
	case *Error:
		return e.description, e
	case *wrapped:
		return e.sentinel.Error(), e
	default:
		adapted := &Error{
			class:       ErrorClassUnknown,
			description: err.Error(),
			cause:       err,
		}
		return adapted.description, adapted
	}
}
