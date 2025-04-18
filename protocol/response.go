package protocol

import (
	"encoding/json"
	"fmt"
)

// jsonRPCResponse represents a JSON-RPC 2.0 / MCP-compliant response.
type jsonRPCResponse[T IDConstraint] struct {
	// Version of the JSON-RPC protocol.
	JSONRPC string `json:"jsonrpc"`

	// The ID of the request this response corresponds to.
	ID ID[T] `json:"id"`

	// The result of the request. This is optional, but exactly one of Result or Error MUST be set.
	Result interface{} `json:"result,omitempty"`

	// Error object if the request failed.
	Error *RPCError `json:"error,omitempty"`
}

// RPCError represents a structured error object in the MCP/JSON-RPC response.
//
// According to the MCP and JSON-RPC 2.0 specifications, an error response
// must contain at least the "code" and "message" fields.
// Optionally, the "data" field can be provided for additional context or debugging information.
//
// Example:
//
//	{
//	  "code": -32601,
//	  "message": "Method not found",
//	  "data": { "method": "unknownMethod" }
//	}
//
// Fields:
//   - Code:    Error code as an integer. MCP and JSON-RPC use standardized codes for known errors.
//   - Message: Human-readable message describing the error.
//   - Data:    Optional additional structured data about the error.
type RPCError struct {
	// Code is the error code according to MCP / JSON-RPC specification.
	// Codes MUST be integers.
	Code int `json:"code"`

	// Message is a human-readable string describing the error.
	Message string `json:"message"`

	// Data provides optional additional information about the error.
	// This field may contain any structured or unstructured data.
	Data interface{} `json:"data,omitempty"`
}

// Error implements the standard Go error interface for RPCError.
//
// It returns the human-readable error message contained in the RPCError.
// This allows RPCError to be used as a regular Go error.
//
// Example:
//
//	var err error = &RPCError{Code: -32601, Message: "Method not found"}
//	fmt.Println(err.Error()) // Output: Method not found
func (e *RPCError) Error() string {
	return e.Message
}

// NewRPCError creates a new instance of RPCError.
//
// Parameters:
//   - code: integer error code as per MCP / JSON-RPC specification.
//   - message: human-readable description of the error.
//   - data: optional additional structured information about the error.
//
// Example:
//
//	err := NewRPCError(-32601, "Method not found", map[string]string{"method": "unknown"})
//	fmt.Println(err)
//
// Returns:
//   - A pointer to an RPCError instance.
func NewRPCError(code int, message string, data interface{}) *RPCError {
	return &RPCError{
		Code:    code,
		Message: message,
		Data:    data,
	}
}

// validate checks whether the JSON-RPC / MCP response object adheres to protocol rules.
//
// It ensures the following:
//   - The JSONRPC version is correct.
//   - The response ID is present and not empty.
//   - Exactly one of Result or Error is populated (exclusive).
//   - If Error is present, it must include both a non-zero code and a non-empty message.
//
// Returns:
//   - nil if the response is valid.
//   - ValidationError if any rule is violated.
//
// Example:
//
//	response := NewSuccessResponse(NewID("req-123"), &Result{...})
//	if err := response.validate(); err != nil {
//	    log.Fatalf("Invalid response: %v", err)
//	}
func (r jsonRPCResponse[T]) validate() error {
	if r.JSONRPC != JSONRPCVersion {
		return &ValidationError{Reason: fmt.Sprintf("invalid JSON-RPC version: expected %q, got %q", JSONRPCVersion, r.JSONRPC)}
	}

	if r.ID.isEmpty() {
		return &ValidationError{Reason: "response ID must not be empty"}
	}

	// MCP / JSON-RPC rule: must have exactly one of result or error
	if r.Result != nil && r.Error != nil {
		return &ValidationError{Reason: "response MUST NOT contain both result and error"}
	}
	if r.Result == nil && r.Error == nil {
		return &ValidationError{Reason: "response MUST contain either result or error"}
	}

	// Validate error object if present
	if r.Error != nil && r.Error.Message == "" {
		return &ValidationError{Reason: "error must contain non-empty message"}
	}

	return nil
}

// UnmarshalJSON deserializes the JSON data into a JSONRPCResponse object,
// and validates the response according to MCP/JSON-RPC specifications.
//
// This method ensures that even when deserialized, the response adheres to protocol rules.
// It will return an error if the JSON is invalid or if the response structure fails validation.
//
// Parameters:
//   - data: JSON-encoded byte slice to be deserialized.
//
// Returns:
//   - An error if deserialization or validation fails, nil otherwise.
//
// Example:
//
//	var resp JSONRPCResponse[string, MyResult]
//	err := json.Unmarshal([]byte(rawJSON), &resp)
//	if err != nil {
//	    log.Fatalf("Invalid response: %v", err)
//	}
func (r *jsonRPCResponse[T]) UnmarshalJSON(data []byte) error {
	if len(data) == 0 {
		return ErrEmptyJSONData
	}

	type responseNoMethods jsonRPCResponse[T]
	aux := &struct {
		responseNoMethods
	}{}

	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}

	temp := jsonRPCResponse[T](aux.responseNoMethods)
	if err := temp.validate(); err != nil {
		return err
	}
	*r = temp
	return nil

}

func (r *jsonRPCResponse[T]) GetID() any {
	return r.ID.Value
}

func (r *jsonRPCResponse[T]) SetID(v any) error {
	val, ok := v.(T)
	if !ok {
		return fmt.Errorf("invalid ID type: expected %T, got %T", *new(T), v)
	}
	r.ID = ID[T]{Value: val}
	return nil
}
func (r *jsonRPCResponse[T]) GetResult() interface{} {
	return r.Result
}

func (r *jsonRPCResponse[T]) SetResult(v interface{}) error {
	r.Result = v
	return nil
}

func (r *jsonRPCResponse[T]) GetError() *RPCError {
	return r.Error
}

func (r *jsonRPCResponse[T]) SetError(err *RPCError) {
	r.Error = err
}

func (r *jsonRPCResponse[T]) HasResult() bool {
	return r.Result != nil
}

func (r *jsonRPCResponse[T]) HasError() bool {
	return r.Error != nil
}

type Response interface {
	GetID() any
	SetID(any) error
	GetResult() interface{}
	SetResult(interface{}) error
	GetError() *RPCError
	SetError(*RPCError)
	HasResult() bool
	HasError() bool
}

func NewResponse[T IDConstraint](id T, result interface{}) Response {
	return &jsonRPCResponse[T]{
		JSONRPC: JSONRPCVersion,
		ID:      ID[T]{Value: id},
		Result:  result,
	}
}
