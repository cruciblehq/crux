package crex

// Categorizes the nature of an Error.
//
// It helps in distinguishing between different types of errors. This can be
// useful for error handling, logging, and providing feedback to users or
// developers. For example, programming errors might indicate bugs in the code
// that need to be fixed, while user errors suggest incorrect usage or input.
// The classification can guide the response to the error, such as whether to
// log it, display it to the user, or take corrective action.
type ErrorClass string

const (

	// Indicates no specific error class. This is the default value, but
	// should be avoided in practice.
	ErrorClassUnknown ErrorClass = "unknown"

	// Indicates an error caused by user action or input. The user can
	// potentially resolve the issue by changing their input or workflow.
	ErrorClassUser ErrorClass = "user"

	// Indicates an error caused by system/environment issues. These issues are
	// typically outside the user's or the process's control and may require
	// external intervention to resolve.
	ErrorClassSystem ErrorClass = "system"

	// Indicates an error likely caused by a bug in the code. These issues
	// should be addressed by Crucible developers and are not expected to be
	// resolved by end-users. They may warrant reporting to a remote error
	// tracking service for further investigation.
	ErrorClassProgramming ErrorClass = "programming"
)
