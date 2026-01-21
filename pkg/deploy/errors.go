package deploy

import "errors"

var (
	ErrInvalidPlan      = errors.New("invalid plan file")
	ErrInvalidState     = errors.New("invalid state file")
	ErrProviderNotFound = errors.New("provider not found")

	ErrAWSOperation     = errors.New("aws operation failed")
	ErrNotImplemented   = errors.New("not implemented")
	ErrServiceOperation = errors.New("service operation failed")
	ErrGatewayOperation = errors.New("gateway operation failed")
)
