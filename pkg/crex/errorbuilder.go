package crex

import (
	"context"
	"fmt"
	"strings"
)

// Provides a builder pattern for constructing [Error] instances.
//
// [ErrorBuilder] allows constructing [Error] instances using several factory
// and setter methods. The factories are organized according to error class,
// allowing the caller to specify whether the error is a user error with
// [UserError], a system error with [SystemError], or a programming/bug error
// with [Bug] or [ProgrammingError]. Each factory method also includes a
// formatted variant (e.g., [UserErrorf]).
//
// [ErrorBuilder] allows setting various attributes of an error: description,
// reason, fallback, cause, additional details, and context. Description and
// reason are required and must be provided when creating the builder. Fallback,
// cause, details, and context can be set using their respective setter methods:
// [Fallback], [Cause], [Detail], and [Context].
//
// Once all desired attributes are set, [Err] can be called to retrieve the
// constructed [Error] instance.
type ErrorBuilder struct {
	err Error
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

// Creates a programming error with the given description and reason.
func Bug(description, reason string) *ErrorBuilder {
	return ProgrammingError(description, reason)
}

// Creates a programming error with the given description and formatted reason.
func Bugf(description string, reasonFormat string, args ...any) *ErrorBuilder {
	return ProgrammingError(description, fmt.Sprintf(reasonFormat, args...))
}

// Creates a new ErrorBuilder with the specified class, description, and reason.
//
// This is the internal factory method used by the public factory methods.
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

// Sets the fallback suggestion or compromise for the error.
func (b *ErrorBuilder) Fallback(suggestion string) *ErrorBuilder {
	b.err.fallback = strings.TrimSpace(suggestion)
	return b
}

// Sets the formatted fallback suggestion or compromise for the error.
func (b *ErrorBuilder) Fallbackf(format string, args ...any) *ErrorBuilder {
	return b.Fallback(fmt.Sprintf(format, args...))
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

// Validate checks whether the error being built is valid.
//
// Returns an error if description or reason are empty. This is primarily
// useful for validating programmatically constructed errors before calling Err.
func (b *ErrorBuilder) Validate() error {
	if b.err.description == "" {
		return fmt.Errorf("crex: error description is empty")
	}
	if b.err.reason == "" {
		return fmt.Errorf("crex: error reason is empty")
	}
	return nil
}

// Builds and returns the constructed [Error] instance.
func (b *ErrorBuilder) Err() error {
	return &b.err
}
