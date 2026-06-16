package crex

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

// Provides a builder pattern for constructing [Error] instances.
//
// [ErrorBuilder] allows constructing [Error] instances using several factory
// and setter methods. The factories are organized according to error class,
// allowing the caller to specify whether the error is a user error with
// [UserError], a system error with [SystemError], or a programming/bug error
// with [ProgrammingError]. Each factory method also includes a formatted
// variant (e.g., [UserErrorf]).
//
// [ErrorBuilder] allows setting various attributes of an error: description,
// reason, recovery, cause, additional details, and context. Description and
// reason are required and must be provided when creating the builder. Recovery,
// cause, details, and context can be set using their respective setter methods:
// [Recovery], [Cause], [Detail], and [Context].
//
// Once all desired attributes are set, [Err] can be called to retrieve the
// constructed [Error] instance.
type ErrorBuilder struct {
	err Error // The error being built.
}

// Creates a user error with the given description and reason.
func UserError(description, reason string) *ErrorBuilder {
	return newError(ErrorClassUser, description, reason)
}

// Creates a user error with the given description and formatted reason.
func UserErrorf(description string, reasonFormat string, args ...any) *ErrorBuilder {
	return UserError(description, fmt.Sprintf(reasonFormat, args...))
}

// Creates a system error with the given description and reason.
func SystemError(description, reason string) *ErrorBuilder {
	return newError(ErrorClassSystem, description, reason)
}

// Creates a system error with the given description and formatted reason.
func SystemErrorf(description string, reasonFormat string, args ...any) *ErrorBuilder {
	return SystemError(description, fmt.Sprintf(reasonFormat, args...))
}

// Creates a programming error with the given description and reason.
func ProgrammingError(description, reason string) *ErrorBuilder {
	return newError(ErrorClassProgramming, description, reason)
}

// Creates a programming error with the given description and formatted reason.
func ProgrammingErrorf(description string, reasonFormat string, args ...any) *ErrorBuilder {
	return ProgrammingError(description, fmt.Sprintf(reasonFormat, args...))
}

// Creates a new ErrorBuilder with the specified class, description, and reason.
//
// Panics if description or reason are empty after trimming whitespace in order
// to enforce error construction conventions.
func newError(class ErrorClass, description, reason string) *ErrorBuilder {
	description = strings.TrimSpace(description)
	reason = strings.TrimSpace(reason)

	if description == "" {
		panic("crex: error description cannot be empty")
	}
	if reason == "" {
		panic("crex: error reason cannot be empty")
	}

	return &ErrorBuilder{
		err: Error{
			class:       class,
			description: description,
			reason:      reason,
		},
	}
}

// Sets the recovery suggestion or compromise for the error.
func (b *ErrorBuilder) Recovery(suggestion string) *ErrorBuilder {
	b.err.recovery = strings.TrimSpace(suggestion)
	return b
}

// Sets the formatted recovery suggestion or compromise for the error.
func (b *ErrorBuilder) Recoveryf(format string, args ...any) *ErrorBuilder {
	return b.Recovery(fmt.Sprintf(format, args...))
}

// Sets the underlying cause error for the error.
func (b *ErrorBuilder) Cause(cause error) *ErrorBuilder {
	b.err.cause = cause
	return b
}

// Adds an additional detail to the error with the given key and value.
func (b *ErrorBuilder) Detail(key string, value any) *ErrorBuilder {
	if b.err.details == nil {
		b.err.details = make(map[string]any)
	}
	b.err.details[key] = value
	return b
}

// Sets the context associated with the error.
func (b *ErrorBuilder) Context(ctx context.Context) *ErrorBuilder {
	b.err.context = ctx
	return b
}

// Builds and returns the constructed [Error] instance.
func (b *ErrorBuilder) Err() error {
	return &b.err
}

// Builds the error with err as its cause unless err is already classful.
//
// When err already carries an [Error] anywhere in its chain, the existing
// classification wins and err is returned unchanged. Otherwise err becomes
// the cause of the error being built. A nil err returns nil. Use to attach
// a class without overriding one that a deeper layer already set.
func (b *ErrorBuilder) Ensure(err error) error {
	if err == nil {
		return nil
	}
	var existing *Error
	if errors.As(err, &existing) {
		return err
	}
	b.err.cause = err
	return &b.err
}

// Builds the error with err as its cause, overriding any existing classification.
//
// The error being built becomes the outermost classification regardless of
// whether err already carries one, so the new class takes precedence while the
// prior classification stays in the chain for errors.Is. A nil err returns nil.
// Use when an upper layer has better context than a class set below it.
func (b *ErrorBuilder) Reclassify(err error) error {
	if err == nil {
		return nil
	}
	b.err.cause = err
	return &b.err
}
