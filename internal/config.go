package internal

import (
	"strconv"
	"sync/atomic"
)

var (
	debugMode   atomic.Bool // Indicates whether debug logging is enabled.
	traceMode   atomic.Bool // Indicates whether trace logging is enabled.
	verboseMode atomic.Bool // Indicates whether verbose logging is enabled.
)

// Parses the linker flags into usable runtime variables.
//
// The rawDebug, rawTrace, and rawVerbose variables should be set via ldflags
// during the build process. If not set, they default to "false".
func init() {
	if v, err := strconv.ParseBool(rawDebug); err == nil {
		debugMode.Store(v)
	}
	if v, err := strconv.ParseBool(rawTrace); err == nil {
		traceMode.Store(v)
	}
	if v, err := strconv.ParseBool(rawVerbose); err == nil {
		verboseMode.Store(v)
	}
}

// Enables or disables debug mode.
func SetDebug(enabled bool) {
	debugMode.Store(enabled)
}

// Returns true if debug mode is enabled.
func IsDebug() bool {
	return debugMode.Load()
}

// Enables or disables trace mode.
func SetTrace(enabled bool) {
	traceMode.Store(enabled)
}

// Returns true if trace mode is enabled.
func IsTrace() bool {
	return traceMode.Load()
}

// Enables or disables verbose logging.
func SetVerbose(enabled bool) {
	verboseMode.Store(enabled)
}

// Returns true if verbose logging is enabled.
func IsVerbose() bool {
	return verboseMode.Load()
}
