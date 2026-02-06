// Crex provides structured error management.
//
// Crex errors include a description, reason, fallback suggestion, underlying
// cause, error class, and additional details. They are designed to provide
// user-friendly error messages that explain what happened, why, and what the
// user can do about it.
//
// Note:
//
//	Go's standard error chaining produces output that is often hard to read
//	for end users. Crex provides a more user-friendly representation with
//	structured information.
//
//	Crex isn't meant to replace standard errors. Low-level system interactions
//	should still use standard errors. Crex is designed for higher-level
//	application error management and user-facing context.
//
//	As a rule of thumb, consider what is known about the user's actions when
//	creating errors. If the error can be explained in terms of what the user
//	did or didn't do, it is a good candidate for Crex.
//
// Crex provides factory functions to create errors with different classes:
// [UserError], [SystemError], [ProgrammingError], and [Bug]. Each returns an
// [ErrorBuilder] for further configuration:
//
//	err := crex.UserError("could not retrieve data", "connection timed out").
//		Fallback("Check your network settings.").
//		Detail("server", serverName).
//		Cause(previousError).
//		Err()
//
// The [Error] type implements the standard error interface and can be used
// anywhere a regular error is expected. It also implements [slog.LogValuer]
// for structured logging.
//
// Errors are classified by their source:
//
//   - [ErrorClassUser]: User-caused errors (bad input, invalid options)
//   - [ErrorClassSystem]: External system errors (network, filesystem)
//   - [ErrorClassProgramming]: Application bugs (should not happen)
//   - [ErrorClassUnknown]: Errors from unknown sources (default, avoid in practice)
//
// Crex also provides a custom [slog.Handler] implementation that buffers log
// records until configured. This addresses the common CLI problem where
// logging may occur before command-line arguments are parsed.
//
// The handler supports:
//   - Buffering until a [Formatter] is set
//   - Pretty printing for TTY output
//   - JSON output for non-TTY environments
//   - Log level filtering
//   - Groups and attributes
//
// Example:
//
//	// At startup (before CLI parsing)
//	handler := crex.NewHandler()
//	slog.SetDefault(slog.New(handler))
//
//	// Logs are buffered
//	slog.Info("initializing")
//
//	// After CLI parsing
//	handler.SetLevel(slog.LevelDebug)
//	handler.SetFormatter(crex.NewPrettyFormatter(true))
//	handler.Flush()
//
// Crex errors integrate with slog via [Error.LogValue]:
//
//	slog.Error(crexErr.Description(), "error", crexErr)
//
// Two formatters are provided:
//
//   - [PrettyFormatter]: Human-readable output with optional color
//   - [JSONFormatter]: Machine-readable JSON output
//
// The [PrettyFormatter] automatically detects crex errors and formats them
// as "description: reason. fallback". The [JSONFormatter] outputs all attributes
// as nested JSON objects.
//
// Crex provides [Assert] and [Assertf] for debug-only assertions that are
// stripped in release builds (when built without the "debug" tag).
package crex
