package crex

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sort"
	"strings"
)

// Sentinel key used to identify crex errors in slog output.
const crexErrorMarker = "!github.com/cruciblehq/crex.Error"

// Represents an error with rich context.
//
// crex errors are composed of a description (what failed), a reason (why it
// failed), and an optional recovery (how to fix it or system compromise). The
// recovery either provides a suggestion to the user to recover from the error
// or indicates how the system has compromised itself to continue operating.
//
// Errors also have a class (user, system, programming/bug) that indicates the
// nature of the error. User errors are caused by incorrect user input or actions,
// and can typically be resolved by the user. System errors are caused by
// external system failures (e.g., network issues, file system errors) and may
// require user intervention or retries. Programming/bug errors indicate flaws
// in the code itself and should be reported to the developers. The CLI tools
// may suggest reporting programming/bug errors to the development team.
//
// Errors can carry additional details as key-value pairs for more context,
// wrap an underlying cause error, and include a context.
type Error struct {
	description string          // What failed.
	reason      string          // Why it failed.
	recovery    string          // Recovery suggestion or compromise, if any.
	cause       error           // Underlying cause error, if any.
	class       ErrorClass      // Classification of the error.
	details     map[string]any  // Additional key-value details about the error.
	context     context.Context // Context associated with the error, if any.
}

// Returns the error description.
func (r *Error) Description() string {
	return r.description
}

// Returns the error reason.
func (r *Error) Reason() string {
	return r.reason
}

// Returns the error recovery suggestion or compromise.
func (r *Error) Recovery() string {
	return r.recovery
}

// Returns the underlying cause error, if any.
func (r *Error) Cause() error {
	return r.cause
}

// Returns the error classification.
func (r *Error) Class() ErrorClass {
	return r.class
}

// Returns the error class carried by err, walking the chain to the outermost
// [Error].
//
// The boolean is false when no layer of err carries a class, in which case the
// returned class is [ErrorClassUnknown].
func ClassOf(err error) (ErrorClass, bool) {
	var e *Error
	if errors.As(err, &e) {
		return e.class, true
	}
	return ErrorClassUnknown, false
}

// Returns the context associated with the error, or nil if none was set.
func (r *Error) Context() context.Context {
	return r.context
}

// Returns the value of a specific detail by key.
//
// The boolean return value indicates whether the detail was found.
func (r *Error) Detail(key string) (any, bool) {
	if r.details == nil {
		return nil, false
	}
	val, ok := r.details[key]
	return val, ok
}

// Returns a copy of all error details as a map.
func (r *Error) Details() map[string]any {
	if r.details == nil {
		return map[string]any{}
	}
	copy := make(map[string]any, len(r.details))
	for k, v := range r.details {
		copy[k] = v
	}
	return copy
}

// Returns a sorted list of all detail keys.
func (r *Error) DetailKeys() []string {
	if len(r.details) == 0 {
		return []string{}
	}
	keys := make([]string, 0, len(r.details))
	for k := range r.details {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// Implements the error interface.
func (r *Error) Error() string {
	return r.String()
}

// Implements error unwrapping.
func (r *Error) Unwrap() error {
	return r.cause
}

// Returns a string representation of the error, including description, reason,
// and recovery.
//
// The format is: "description: reason. recovery", omitting any empty parts.
func (r *Error) String() string {
	var b strings.Builder
	b.WriteString(r.description)
	if r.reason != "" {
		b.WriteString(": ")
		b.WriteString(r.reason)
	}
	if r.recovery != "" {
		b.WriteString(". ")
		b.WriteString(r.recovery)
	}
	return b.String()
}

// Implements custom formatting for the error.
//
// The 'v' and 's' verbs produce the same output as [Error.String]. The 'q' verb
// produces a quoted string representation.
func (r *Error) Format(f fmt.State, verb rune) {
	switch verb {
	case 'v', 's':
		fmt.Fprint(f, r.String())
	case 'q':
		fmt.Fprintf(f, "%q", r.String())
	default:
		fmt.Fprint(f, r.String()) // Unknown verbs are treated as %v
	}
}

// Implements [slog.LogValuer].
//
// Returns a grouped value containing the error's class, description, reason,
// recovery, cause, and details (if present). Includes a sentinel marker for
// reliable identification by formatters.
func (r *Error) LogValue() slog.Value {
	attrs := []slog.Attr{
		slog.Bool(crexErrorMarker, true),
		slog.String("class", string(r.class)),
		slog.String("description", r.description),
	}

	if r.reason != "" {
		attrs = append(attrs, slog.String("reason", r.reason))
	}

	if r.recovery != "" {
		attrs = append(attrs, slog.String("recovery", r.recovery))
	}

	if r.cause != nil {
		attrs = append(attrs, slog.String("cause", r.cause.Error()))
	}

	if len(r.details) > 0 {
		detailAttrs := make([]any, 0, len(r.details)*2)
		for _, k := range r.DetailKeys() {
			detailAttrs = append(detailAttrs, k, r.details[k])
		}
		attrs = append(attrs, slog.Group("details", detailAttrs...))
	}

	return slog.GroupValue(attrs...)
}
