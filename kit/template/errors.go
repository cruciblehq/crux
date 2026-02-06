package template

import (
	"errors"
	"io/fs"
	"os"
	"text/template"
)

var (
	ErrTemplate = errors.New("template error") // Error related to template input
	ErrOutput   = errors.New("output error")   // Error related to output writing
)

// Wraps an error that occurred while reading/processing a template.
type TemplateError struct {
	Path string
	Err  error
}

// Implements the error interface.
func (e *TemplateError) Error() string {
	return "template " + e.Path + ": " + innerMessage(e.Err)
}

// Unwraps to the underlying error.
func (e *TemplateError) Unwrap() []error {
	return []error{ErrTemplate, e.Err}
}

// Wraps an error that occurred while writing output.
type OutputError struct {
	Path string
	Err  error
}

// Implements the error interface.
func (e *OutputError) Error() string {
	return "output " + e.Path + ": " + innerMessage(e.Err)
}

// Unwraps to the underlying error.
func (e *OutputError) Unwrap() []error {
	return []error{ErrOutput, e.Err}
}

// Extracts the inner error message, stripping redundant path info.
func innerMessage(err error) string {
	var pathErr *fs.PathError
	if errors.As(err, &pathErr) {
		return pathErr.Err.Error()
	}

	var linkErr *os.LinkError
	if errors.As(err, &linkErr) {
		return linkErr.Err.Error()
	}

	return err.Error()
}

// Converts a template execution error into a TemplateError, preserving the
// original error message and context.
func errorFromExecuteError(err error, path string) error {
	if err == nil {
		return nil
	}

	// Template error (e.g., "is not a method but has arguments")
	if _, ok := err.(*template.ExecError); ok {
		return &TemplateError{
			Path: path,
			Err:  err,
		}
	}

	// This should only happen on I/O errors, but [strings.Builder.Write]
	// never returns an error. If we're here, we don't know why.
	return &TemplateError{
		Path: path,
		Err:  err,
	}
}
