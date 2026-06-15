package widget

import (
	"fmt"
	"log/slog"
	"sort"
	"unicode"

	"github.com/cruciblehq/crux/crex"
	es "github.com/evanw/esbuild/pkg/api"
)

// Severity levels for esbuild messages from a widget build.
type severity int

const (
	severityWarning severity = iota // Warning severity.
	severityError                   // Error severity.
)

// A normalized esbuild diagnostic from a widget build.
//
// Holds a single error or warning with its location, used to sort and report
// the messages produced by a widget build.
type diagnostic struct {
	severity severity // The severity of the message.
	message  string   // The error or warning message.
	line     int      // The line number for sorting.
	column   int      // The column number for sorting.
}

// Processes the esbuild build result and logs errors and warnings.
//
// It normalizes the messages, sorts them, and logs them. If there are errors,
// it returns an error indicating a general failure of the build.
func processResult(result es.BuildResult) error {
	if len(result.Errors) == 0 && len(result.Warnings) == 0 {
		return nil
	}

	diags := normalizeAndSort(result)

	for _, d := range diags {
		if d.severity == severityWarning {
			slog.Warn(d.message)
		} else {
			slog.Error(d.message)
		}
	}

	if len(result.Errors) == 0 {
		slog.Warn(fmt.Sprintf("build completed with %d warning(s)", len(result.Warnings)))
		return nil
	}

	return crex.Newf(ErrBuild, "%d error(s) encountered during the build process", len(result.Errors))
}

// Normalizes and sorts esbuild results into [diagnostic] structs.
//
// It processes both errors and warnings, normalizing their messages and
// location information. The resulting diagnostics are sorted by line and column
// number to provide a coherent order for reporting.
func normalizeAndSort(result es.BuildResult) []diagnostic {
	var diags []diagnostic

	for _, err := range result.Errors {
		diags = append(diags, normalizeMessage(err, severityError))
	}
	for _, warn := range result.Warnings {
		diags = append(diags, normalizeMessage(warn, severityWarning))
	}

	sort.SliceStable(diags, func(i, j int) bool {
		if diags[i].line == diags[j].line {
			return diags[i].column < diags[j].column
		}
		return diags[i].line < diags[j].line
	})

	return diags
}

// Converts an esbuild message into a [diagnostic].
//
// It uses the provided severity level to create either error or warning
// messages. If location information is available, it includes it in the result
// and keeps track of line and column for sorting purposes.
func normalizeMessage(msg es.Message, sev severity) diagnostic {
	d := diagnostic{
		severity: sev,
		message:  lowerFirst(msg.Text),
	}

	if msg.Location != nil {
		d.message = fmt.Sprintf("%s: %s", formatLocation(*msg.Location), lowerFirst(msg.Text))
		d.line = msg.Location.Line
		d.column = msg.Location.Column
	}

	return d
}

// Formats an esbuild location as "file:line:column".
func formatLocation(loc es.Location) string {
	if loc.File != "" {
		return fmt.Sprintf("%s:%d:%d", loc.File, loc.Line, loc.Column)
	}
	return "(unknown)"
}

// Lowercases the first character of a string.
func lowerFirst(s string) string {
	if s == "" {
		return s
	}
	return string(unicode.ToLower(rune(s[0]))) + s[1:]
}
