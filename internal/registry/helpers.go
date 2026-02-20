package registry

// Logs an error with structured context and returns a registry Error.
//
// The error is logged with the provided message and key-value pairs, while the
// returned Error contains only the error code and message for the client. This
// prevents leaking internal implementation details.
func (r *SQLRegistry) logAndReturnError(code ErrorCode, message string, err error, keyvals ...any) *Error {
	args := make([]any, 0, 2+len(keyvals))
	args = append(args, "error", err)
	args = append(args, keyvals...)
	r.logger.Error(message, args...)
	return &Error{
		Code:    code,
		Message: message,
	}
}
