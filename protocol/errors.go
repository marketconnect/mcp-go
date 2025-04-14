package protocol

import (
	"errors"
	"fmt"
)

var (
	// ErrInvalidID is returned when the ID in the JSON is invalid.
	ErrInvalidID = errors.New("invalid ID")

	// ErrEmptyRequestID is returned when the request ID is empty.
	ErrEmptyRequestID = errors.New("lifecycle: request ID cannot be empty")
	// ErrSoftTimeoutMustBePositive is returned when the soft timeout is not a positive duration.
	ErrSoftTimeoutMustBePositive = errors.New("soft timeout must be greater than zero")

	// ErrMaximumTimeoutMustBePositive is returned when the maximum timeout is not a positive duration.
	ErrMaximumTimeoutMustBePositive = errors.New("maximum timeout must be greater than zero")

	// ErrSoftTimeoutExceedsMaximum is returned when the soft timeout exceeds or equals the maximum timeout.
	ErrSoftTimeoutExceedsMaximum = errors.New("soft timeout exceeds or equals maximum timeout")

	// ErrDuplicateRequestID is returned when a request with the same ID has already been started in this session.
	ErrDuplicateRequestID = errors.New("request ID already used in this session")

	// ErrRequestNotFound is returned when attempting to operate on a request that does not exist.
	ErrRequestNotFound = errors.New("request not found")

	// ErrNilTimeoutCallback is returned when a nil callback function is provided.
	ErrNilTimeoutCallback = errors.New("timeout callback must not be nil")
)

const (
	// Parse error	Invalid JSON was received by the server.
	// An error occurred on the server while parsing the JSON text.
	ParseError = -32700
	// Invalid Request	The JSON sent is not a valid Request object.
	InvalidRequest = -32600
	// Method not found	The method does not exist / is not available.
	MethodNotFound = -32601
	// Invalid params	Invalid method parameter(s).
	InvalidParams = -32602
	// Internal error	Internal JSON-RPC error.
	InternalError = -32603

	// -32000 to -32099	Server error	Reserved for implementation-defined server-errors.

)

const (
	CapabilityDisabled = -32001 // a requested capability or resource is not available
	ResourceNotFound   = -32002 // a requested resource was not found
)

// ValidationError represents an error that occurs when a response or request fails validation
// against MCP or JSON-RPC specifications.
//
// ValidationError is used internally and can also be returned to users
// when decoding and validating incoming responses.
//
// Example:
//
//	err := ValidationError{Reason: "response MUST contain either result or error"}
//	fmt.Println(err) // Output: response MUST contain either result or error
type ValidationError struct {
	// Reason provides a human-readable explanation of the validation failure.
	Reason string
}

// Error implements the standard Go error interface for ValidationError.
//
// It returns the reason for the validation failure.
func (e ValidationError) Error() string {
	return e.Reason
}

type InvalidIDError struct {
	Err error
}

func (e *InvalidIDError) Error() string {
	return fmt.Sprintf("%v: %v", ErrInvalidID, e.Err)
}

func (e *InvalidIDError) Unwrap() error {
	return e.Err
}
