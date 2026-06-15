package crex

import (
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
)

// Creates an error.
//
// Use for package-level sentinel values that callers match with errors.Is. The
// message is fixed; use Newf when the message needs formatting and a sentinel
// to classify it.
func New(text string) error {
	return errors.New(text)
}

// Builds a tagged error that carries a formatted message but no cause.
//
// Use when a sentinel needs a message attached and there is no underlying error
// to chain. The rendered message is "sentinel: message".
func Newf(sentinel error, format string, args ...any) error {
	return &wrapped{sentinel: sentinel, message: fmt.Sprintf(format, args...)}
}

// Wraps a cause under a sentinel.
//
// The sentinel classifies the error and the cause carries the detail. The
// rendered message is "sentinel: cause". Use for technical errors in library
// code that need classification but no user-facing context.
func Wrap(sentinel, cause error) error {
	return &wrapped{sentinel: sentinel, cause: cause}
}

// Wraps a cause under a sentinel together with a formatted message.
//
// The format states why this layer failed and must not contain a %w verb. The
// rendered message is "sentinel: message: cause", where cause is the foreign
// root of the chain. Intermediate wrapped layers are omitted from the rendered
// string, so an error never accumulates more than one message.
func Wrapf(sentinel, cause error, format string, args ...any) error {
	return &wrapped{sentinel: sentinel, cause: cause, message: fmt.Sprintf(format, args...)}
}

// Attaches additional sentinels to err for errors.Is matching without changing
// the rendered message.
//
// Use when one layer must match more than one sentinel, for example a category
// and a specific cause, while rendering a single woven message. The tags take
// part in errors.Is through Unwrap but never appear in the error string. A nil
// err returns nil; a non-wrapped err is wrapped so the tags can ride along.
func Tag(err error, tags ...error) error {
	if err == nil {
		return nil
	}
	if w, ok := err.(*wrapped); ok {
		nw := *w
		nw.tags = append(append([]error(nil), w.tags...), tags...)
		return &nw
	}
	return &wrapped{sentinel: err, tags: tags}
}

// A single step in an error's structural position path.
type segment struct {
	label string // What kind of location this step names, for example "grant scope".
	value string // Which one, the rendered position token.
	quote bool   // Whether the value renders quoted, for name-like identifiers.
}

// Records a structural position on err using a 1-based ordinal.
//
// Use during propagation to mark where in an ordered collection the failure
// was found, for example a grant inside a scope. Positions accumulate from the
// outermost layer inward, forming a path that renders as a breadcrumb in the
// error string and is kept as separate fields in the structured model. A nil
// err returns nil; a non-wrapped err is wrapped so the position can ride along.
func At(err error, label string, ordinal int) error {
	return at(err, segment{label: label, value: strconv.Itoa(ordinal)})
}

// Records a named structural position on err, rendered quoted.
//
// Use for entities located by identity rather than order, for example a named
// stage. Behaves like At otherwise: the name renders quoted in the breadcrumb
// and is kept as a separate field in the structured model.
func AtName(err error, label, name string) error {
	return at(err, segment{label: label, value: name, quote: true})
}

// Prepends seg to err's position path, wrapping a non-wrapped err so the
// position can ride along. A nil err returns nil.
func at(err error, seg segment) error {
	if err == nil {
		return nil
	}
	if w, ok := err.(*wrapped); ok {
		nw := *w
		nw.path = append([]segment{seg}, w.path...)
		return &nw
	}
	return &wrapped{sentinel: err, path: []segment{seg}}
}

// Sentinel-classified error that renders as a single message.
//
// A wrapped error shows its own sentinel and message followed by the foreign
// root cause of its chain. Intermediate wrapped layers still participate in
// errors.Is matching through Unwrap, but their messages are not rendered, so
// nesting collapses to one message rather than a chain.
type wrapped struct {
	sentinel error     // Classifier matched by errors.Is.
	cause    error     // Underlying cause, or nil.
	message  string    // Why this layer failed, or empty.
	tags     []error   // Extra classifiers matched by errors.Is but not rendered.
	path     []segment // Structural position, outermost step first.
}

// Renders the error as "sentinel: message: cause", omitting empty parts and the
// messages of any intermediate wrapped layers, with the structural position
// appended as a breadcrumb.
func (w *wrapped) Error() string {
	var b strings.Builder
	b.WriteString(w.sentinel.Error())
	if reason := w.reasonString(); reason != "" {
		b.WriteString(": ")
		b.WriteString(reason)
	}
	if path := collectPath(w); len(path) > 0 {
		b.WriteString(" (at ")
		writePath(&b, path)
		b.WriteByte(')')
	}
	return b.String()
}

// Renders the message and foreign root cause without the sentinel or position.
func (w *wrapped) reasonString() string {
	var parts []string
	if w.message != "" {
		parts = append(parts, w.message)
	}
	if w.cause != nil {
		parts = append(parts, rootCause(w.cause))
	}
	return strings.Join(parts, ": ")
}

// Implements slog.LogValuer for the structured model.
//
// Exposes the classification, the rendered reason, and the structural position
// as separate attributes, including the sentinel marker so formatters keep the
// position as its own fields rather than collapsing it into the message.
func (w *wrapped) LogValue() slog.Value {
	attrs := []slog.Attr{
		slog.Bool(crexErrorMarker, true),
		slog.String("description", w.sentinel.Error()),
	}
	if reason := w.reasonString(); reason != "" {
		attrs = append(attrs, slog.String("reason", reason))
	}
	if path := collectPath(w); len(path) > 0 {
		segs := make([]any, 0, len(path)*2)
		for _, seg := range path {
			segs = append(segs, seg.label, seg.value)
		}
		attrs = append(attrs, slog.Group("at", segs...))
	}
	return slog.GroupValue(attrs...)
}

// Exposes the sentinel, cause, and any match-only tags for errors.Is traversal.
func (w *wrapped) Unwrap() []error {
	out := make([]error, 0, 2+len(w.tags))
	out = append(out, w.sentinel)
	if w.cause != nil {
		out = append(out, w.cause)
	}
	out = append(out, w.tags...)
	return out
}

// Renders the foreign root of an error chain.
//
// Wrapped layers are traversed to their cause so intermediate messages are
// skipped. A wrapped leaf with no cause renders its own sentinel and message,
// and any non-wrapped error renders directly.
func rootCause(err error) string {
	w, ok := err.(*wrapped)
	if !ok {
		return err.Error()
	}
	if w.cause != nil {
		return rootCause(w.cause)
	}
	if w.message == "" {
		return w.sentinel.Error()
	}
	return w.sentinel.Error() + ": " + w.message
}

// Gathers the full structural position of an error chain.
//
// Segments are collected from the outermost wrapped layer inward, so the result
// reads from the top of the structure down to where the failure occurred.
func collectPath(w *wrapped) []segment {
	var path []segment
	for cur := w; cur != nil; {
		path = append(path, cur.path...)
		next, ok := cur.cause.(*wrapped)
		if !ok {
			break
		}
		cur = next
	}
	return path
}

// Writes a position path as a breadcrumb like `grant scope 2 > grant 1`.
func writePath(b *strings.Builder, path []segment) {
	for i, seg := range path {
		if i > 0 {
			b.WriteString(" > ")
		}
		b.WriteString(seg.label)
		b.WriteByte(' ')
		if seg.quote {
			fmt.Fprintf(b, "%q", seg.value)
		} else {
			b.WriteString(seg.value)
		}
	}
}
