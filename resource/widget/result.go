package widget

import (
	"fmt"
	"log/slog"
	"sort"
	"unicode"

	"github.com/cruciblehq/crux/crex"
	es "github.com/evanw/esbuild/pkg/api"
)

// Severity levels for esbuild messages.
type severity int

const (
	severityWarning severity = iota // Warning severity.
	severityError                   // Error severity.
)

// Helper struct for sorting esbuild results.
type resultHelper struct {
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

	helpers := normalizeAndSort(result)

	for _, h := range helpers {
		if h.severity == severityWarning {
			slog.Warn(h.message)
		} else {
			slog.Error(h.message)
		}
	}

	if len(result.Errors) == 0 {
		slog.Warn(fmt.Sprintf("build completed with %d warning(s)", len(result.Warnings)))
		return nil
	}

	return crex.Wrapf(ErrBuild, "%d error(s) encountered during the build process", len(result.Errors))
}

// Normalizes and sorts esbuild results into [resultHelper] structs.
//
// It processes both errors and warnings, normalizing their messages and
// location information. The resulting helpers are sorted by line and column
// number to provide a coherent order for reporting.
func normalizeAndSort(result es.BuildResult) []resultHelper {
	var helpers []resultHelper

	for _, err := range result.Errors {
		helpers = append(helpers, normalizeMessage(err, severityError))
	}
	for _, warn := range result.Warnings {
		helpers = append(helpers, normalizeMessage(warn, severityWarning))
	}

	sort.SliceStable(helpers, func(i, j int) bool {
		if helpers[i].line == helpers[j].line {
			return helpers[i].column < helpers[j].column
		}
		return helpers[i].line < helpers[j].line
	})

	return helpers
}

// Converts an esbuild message into a [resultHelper].
//
// It uses the provided severity level to create either error or warning
// messages. If location information is available, it includes it in the helper
// and keeps track of line and column for sorting purposes.
func normalizeMessage(msg es.Message, sev severity) resultHelper {
	h := resultHelper{
		severity: sev,
		message:  lowerFirst(msg.Text),
	}

	if msg.Location != nil {
		h.message = fmt.Sprintf("%s: %s", formatLocation(*msg.Location), lowerFirst(msg.Text))
		h.line = msg.Location.Line
		h.column = msg.Location.Column
	}

	return h
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
