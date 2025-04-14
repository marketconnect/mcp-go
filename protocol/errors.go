package protocol

import (
	"errors"
	"fmt"
)

// === Error Variables ===

// ErrInvalidID is returned when the ID in the JSON is invalid.
//
// Example:
//
//	if errors.Is(err, protocol.ErrInvalidID) {
//		log.Println("Invalid ID")
//	}
var ErrInvalidID = errors.New("invalid ID")

// ErrEmptyRequestID is returned when the request ID is empty.
//
// Example:
//
//	if errors.Is(err, protocol.ErrEmptyRequestID) {
//		log.Println("Request ID is empty")
//	}
var ErrEmptyRequestID = errors.New("request ID cannot be empty")

// ErrSoftTimeoutNotPositive is returned when the soft timeout is not positive.
var ErrSoftTimeoutNotPositive = errors.New("soft timeout must be greater than zero")

// ErrMaximumTimeoutNotPositive is returned when the maximum timeout is not positive.
var ErrMaximumTimeoutNotPositive = errors.New("maximum timeout must be greater than zero")

// ErrSoftTimeoutExceedsMaximum is returned when the soft timeout exceeds or equals the maximum timeout.
var ErrSoftTimeoutExceedsMaximum = errors.New("soft timeout exceeds or equals maximum timeout")

// ErrDuplicateRequestID is returned when a request with the same ID has already been started in this session.
//
// Example:
//
//	if errors.Is(err, protocol.ErrDuplicateRequestID) {
//		log.Println("Duplicate request ID")
//	}
var ErrDuplicateRequestID = errors.New("request ID already used in this session")

// ErrRequestNotFound is returned when attempting to operate on a request that does not exist.
var ErrRequestNotFound = errors.New("request not found")

// ErrTimeoutCallbackNil is returned when a nil callback function is provided.
var ErrCallbackNil = errors.New("callback must not be nil")

// === JSON-RPC Error Codes ===

const (
	ParseError     = -32700
	InvalidRequest = -32600
	MethodNotFound = -32601
	InvalidParams  = -32602
	InternalError  = -32603

	// Server-defined errors
	CapabilityDisabled = -32001
	ResourceNotFound   = -32002
)

// === Custom Error Types ===

// ValidationError represents an error when a request or response fails validation.
//
// Example:
//
//	err := &protocol.ValidationError{Reason: "response must contain result or error"}
//	fmt.Println(err)
type ValidationError struct {
	Reason string
}

// Error implements the error interface.
func (e *ValidationError) Error() string {
	return e.Reason
}

// InvalidIDError wraps ErrInvalidID with additional context.
//
// Example:
//
//	err := &protocol.InvalidIDError{Err: errors.New("unexpected format")}
//	fmt.Println(err)
type InvalidIDError struct {
	Err error
}

// Error implements the error interface.
func (e *InvalidIDError) Error() string {
	return fmt.Sprintf("%v: %v", ErrInvalidID, e.Err)
}

// Unwrap allows errors.Is and errors.As to work with InvalidIDError.
func (e *InvalidIDError) Unwrap() error {
	return ErrInvalidID
}

// Is allows errors.Is to work directly with InvalidIDError.
func (e *InvalidIDError) Is(target error) bool {
	return target == ErrInvalidID
}

// === Error Factory ===

// NewValidationError creates a new ValidationError with a formatted reason.
func NewValidationError(format string, args ...interface{}) *ValidationError {
	return &ValidationError{Reason: fmt.Sprintf(format, args...)}
}

// NewInvalidIDError creates a new InvalidIDError with formatted context.
func NewInvalidIDError(format string, args ...interface{}) *InvalidIDError {
	return &InvalidIDError{Err: fmt.Errorf(format, args...)}
}
