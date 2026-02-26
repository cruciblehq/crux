package internal

import (
	"strconv"
	"sync/atomic"
)

var (
	quietMode atomic.Bool // Indicates whether quiet mode is enabled.
	debugMode atomic.Bool // Indicates whether debug logging is enabled.
)

// Parses the linker flags into usable runtime variables.
//
// The rawQuiet and rawDebug variables should be set via ldflags during the
// build process. If not set, they default to "false".
func init() {
	if v, err := strconv.ParseBool(rawQuiet); err == nil {
		quietMode.Store(v)
	}
	if v, err := strconv.ParseBool(rawDebug); err == nil {
		debugMode.Store(v)
	}
}

// Returns true if quiet mode is enabled.
func IsQuiet() bool {
	return quietMode.Load()
}

// Returns true if debug mode is enabled.
func IsDebug() bool {
	return debugMode.Load()
}
