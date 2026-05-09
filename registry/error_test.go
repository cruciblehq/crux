package registry

import (
	"errors"
	"testing"
)

func TestError_Error(t *testing.T) {
	e := &Error{Code: ErrorCodeNotFound, Message: "namespace not found"}
	got := e.Error()
	expected := "not_found: namespace not found"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestError_Validate_Valid(t *testing.T) {
	e := &Error{Code: ErrorCodeNotFound, Message: "not found"}
	if err := e.Validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestError_Validate_Invalid_Code(t *testing.T) {
	e := &Error{Code: "unknown_code", Message: "some message"}
	err := e.Validate()
	if !errors.Is(err, ErrErrorCodeInvalid) {
		t.Errorf("expected ErrErrorCodeInvalid, got %v", err)
	}
}

func TestError_Validate_Invalid_EmptyMessage(t *testing.T) {
	e := &Error{Code: ErrorCodeNotFound, Message: ""}
	err := e.Validate()
	if !errors.Is(err, ErrErrorMessageEmpty) {
		t.Errorf("expected ErrErrorMessageEmpty, got %v", err)
	}
}

func TestError_Validate_AllCodes(t *testing.T) {
	codes := []ErrorCode{
		ErrorCodeBadRequest,
		ErrorCodeNotFound,
		ErrorCodeNamespaceExists,
		ErrorCodeNamespaceNotEmpty,
		ErrorCodeResourceExists,
		ErrorCodeResourceHasPublished,
		ErrorCodeVersionExists,
		ErrorCodeVersionPublished,
		ErrorCodeChannelExists,
		ErrorCodePreconditionFailed,
		ErrorCodeInternalError,
	}
	for _, code := range codes {
		e := &Error{Code: code, Message: "test message"}
		if err := e.Validate(); err != nil {
			t.Errorf("unexpected error for code %q: %v", code, err)
		}
	}
}
