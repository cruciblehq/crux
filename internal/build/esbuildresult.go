package build

import (
	"fmt"
	"log/slog"
	"sort"
	"unicode"

	"github.com/cruciblehq/crux/pkg/crex"
	es "github.com/evanw/esbuild/pkg/api"
)

// Severity levels for esbuild messages.
type esbuildSeverity int

const (
	esbuildSeverityWarning esbuildSeverity = iota // Warning severity
	esbuildSeverityError                          // Error severity
)

// Helper struct for sorting esbuild results.
type esbuildResultSortHelper struct {
	severity esbuildSeverity // The severity of the message
	message  string          // The error or warning message
	line     int             // The line number for sorting
	column   int             // The column number for sorting
}

// Processes the esbuild build result and logs errors and warnings.
//
// It normalizes the messages, sorts them, and logs them. If there are errors,
// it returns an indicating a general failure of the build.
func processBuildResult(result es.BuildResult) error {

	// Clean build
	if len(result.Errors) == 0 && len(result.Warnings) == 0 {
		slog.Debug("build completed successfully")
		return nil
	}

	// Normalize and sort
	helpers := normalizeAndSortEsbuildResult(result)

	// Log everything
	for _, h := range helpers {
		if h.severity == esbuildSeverityWarning {
			slog.Warn(h.message)
		} else {
			slog.Error(h.message)
		}
	}

	// Warnings only
	if len(result.Errors) == 0 {
		slog.Warn(fmt.Sprintf("build completed with %d warning(s)", len(result.Warnings)))
		return nil
	}

	// Errors present
	return crex.UserErrorf("build failed", "%d error(s) encountered during the build process", len(result.Errors)).
		Fallback("Fix the error(s) and try again.").
		Err()
}

// Normalizes and sorts esbuild results into [esbuildResultSortHelper] structs.
//
// It processes both errors and warnings, normalizing their messages and
// location information. The resulting helpers are sorted by line and column
// number to provide a coherent order for reporting.
func normalizeAndSortEsbuildResult(result es.BuildResult) []esbuildResultSortHelper {
	var helpers []esbuildResultSortHelper

	// Process errors
	for _, err := range result.Errors {
		helpers = append(helpers, normalizeEsbuildMessage(err, esbuildSeverityError))
	}

	// Process warnings
	for _, warn := range result.Warnings {
		helpers = append(helpers, normalizeEsbuildMessage(warn, esbuildSeverityWarning))
	}

	// Sort reports by line and column if location info is available
	sort.SliceStable(helpers, func(i, j int) bool {
		if helpers[i].line == helpers[j].line {
			return helpers[i].column < helpers[j].column
		}
		return helpers[i].line < helpers[j].line
	})

	return helpers
}

// Converts esbuild errors into [esbuildResultSortHelper] structs.
//
// It uses the provided severity level to create either error or warning
// messages. If location information is available, it includes it in the helper
// and keeps track of line and column for sorting purposes.
func normalizeEsbuildMessage(msg es.Message, severity esbuildSeverity) esbuildResultSortHelper {

	helper := esbuildResultSortHelper{
		severity: severity,
		message:  lowerFirst(msg.Text),
	}

	if msg.Location != nil {
		helper.message = fmt.Sprintf("%s: %s",
			normalizeEsbuildLocation(*msg.Location),
			lowerFirst(msg.Text))
		helper.line = msg.Location.Line
		helper.column = msg.Location.Column
	}

	return helper
}

// Formats an esbuild location as "file:line:column".
func normalizeEsbuildLocation(loc es.Location) string {
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
