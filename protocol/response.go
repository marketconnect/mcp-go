package protocol

import (
	"encoding/json"
	"errors"
	"fmt"
)

// JSONRPCResponse represents a JSON-RPC 2.0 / MCP-compliant response.
type JSONRPCResponse[T IDConstraint, R any] struct {
	// Version of the JSON-RPC protocol.
	JSONRPC string `json:"jsonrpc"`

	// The ID of the request this response corresponds to.
	ID IDType[T] `json:"id"`

	// The result of the request. This is optional, but exactly one of Result or Error MUST be set.
	Result *R `json:"result,omitempty"`

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

// NewSuccessResponse creates a new MCP / JSON-RPC-compliant success response.
//
// Parameters:
//   - id: the unique identifier of the request that this response corresponds to.
//   - result: the result of the operation, can be any type or nil.
//
// Example:
//
//	type SumResult struct {
//	    Total int `json:"total"`
//	}
//
//	result := &SumResult{Total: 42}
//	response := NewSuccessResponse(NewID("req-123"), result)
//	fmt.Printf("%+v\n", response)
//
// Returns:
//   - JSONRPCResponse[T, R]: a populated successful response object.
func NewSuccessResponse[T IDConstraint, R any](id IDType[T], result *R) JSONRPCResponse[T, R] {
	return JSONRPCResponse[T, R]{
		JSONRPC: JSONRPCVersion,
		ID:      id,
		Result:  result,
	}
}

// NewErrorResponse creates a new MCP / JSON-RPC-compliant error response.
//
// Use this function to construct a response for a failed request with a specific
// error code, message, and optional data.
//
// Parameters:
//   - id: the identifier of the original request this response corresponds to.
//   - code: the error code as per MCP / JSON-RPC specification.
//   - message: a human-readable message describing the error.
//   - data: optional additional structured data providing error context.
//
// Example:
//
//	response := NewErrorResponse(NewID("req-456"), -32601, "Method not found", map[string]string{"method": "unknown"})
//	fmt.Printf("%+v\n", response)
//
// Returns:
//   - JSONRPCResponse[T, R]: a populated error response object.
func NewErrorResponse[T IDConstraint, R any](id IDType[T], code int, message string, data interface{}) JSONRPCResponse[T, R] {
	return JSONRPCResponse[T, R]{
		JSONRPC: JSONRPCVersion,
		ID:      id,
		Error: &RPCError{
			Code:    code,
			Message: message,
			Data:    data,
		},
	}
}

// NewErrorResponseFromError creates an MCP / JSON-RPC error response
// using an existing *RPCError instance.
//
// Use this function when you already have a pre-defined RPCError and want to
// wrap it into a JSONRPCResponse for transport.
//
// Parameters:
//   - id: the identifier of the original request.
//   - rpcErr: the RPCError instance to use in the response.
//
// Example:
//
//	err := NewRPCError(-32602, "Invalid params", map[string]string{"param": "age"})
//	response := NewErrorResponseFromError(NewID("req-789"), err)
//	fmt.Printf("%+v\n", response)
//
// Returns:
//   - JSONRPCResponse[T, R]: a populated error response object.
func NewErrorResponseFromError[T IDConstraint, R any](id IDType[T], rpcErr *RPCError) JSONRPCResponse[T, R] {
	return JSONRPCResponse[T, R]{
		JSONRPC: JSONRPCVersion,
		ID:      id,
		Error:   rpcErr,
	}
}

// Validate checks whether the JSON-RPC / MCP response object adheres to protocol rules.
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
//	if err := response.Validate(); err != nil {
//	    log.Fatalf("Invalid response: %v", err)
//	}
func (r JSONRPCResponse[T, R]) Validate() error {
	if r.JSONRPC != JSONRPCVersion {
		return ValidationError{Reason: fmt.Sprintf("invalid JSON-RPC version: expected %q, got %q", JSONRPCVersion, r.JSONRPC)}
	}

	if r.ID.IsEmpty() {
		return ValidationError{Reason: "response ID must not be empty"}
	}

	// MCP / JSON-RPC rule: must have exactly one of result or error
	if r.Result != nil && r.Error != nil {
		return ValidationError{Reason: "response MUST NOT contain both result and error"}
	}
	if r.Result == nil && r.Error == nil {
		return ValidationError{Reason: "response MUST contain either result or error"}
	}

	// Validate error object if present
	if r.Error != nil {
		if r.Error.Code == 0 {
			return ValidationError{Reason: "error code must be non-zero integer"}
		}
		if r.Error.Message == "" {
			return ValidationError{Reason: "error message must not be empty"}
		}
	}

	return nil
}

// MarshalJSON serializes the JSONRPCResponse into JSON format according to MCP/JSON-RPC 2.0 specifications.
//
// This method allows for custom serialization behavior if needed in the future.
// Currently, it behaves identically to the standard JSON marshaling.
//
// Note:
//   - If no special behavior is required, the default encoding/json marshaler is sufficient.
//   - This method is kept for consistency and potential extensibility.
//
// Returns:
//   - A JSON-encoded byte slice representing the response.
//   - An error, if serialization fails.
// func (r JSONRPCResponse[T, R]) MarshalJSON() ([]byte, error) {
// 	type Alias JSONRPCResponse[T, R]
// 	return json.Marshal(&struct {
// 		Alias
// 	}{
// 		Alias: (Alias)(r),
// 	})
// }

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
func (r *JSONRPCResponse[T, R]) UnmarshalJSON(data []byte) error {
	if len(data) == 0 {
		return errors.New("empty JSON data")
	}

	type Alias JSONRPCResponse[T, R]
	aux := &struct {
		Alias
	}{}

	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}

	*r = JSONRPCResponse[T, R](aux.Alias)
	return r.Validate()
}

// GetID returns the underlying raw value of the response ID.
//
// Useful for logging, correlating requests and responses, or internal tracking systems.
// The returned ID will be of the underlying type specified by the generic parameter T.
//
// Example:
//
//	response := NewSuccessResponse(NewID("req-123"), &MyResult{})
//	fmt.Println(response.GetID()) // Output: req-123
func (r JSONRPCResponse[T, R]) GetID() T {
	return r.ID.Value
}

// IsError reports whether the response represents an error.
//
// This is a convenience method to quickly check if the response
// contains an error object.
//
// Returns:
//   - true if the response is an error response.
//   - false if the response is a success response.
//
// Example:
//
//	if response.IsError() {
//	    log.Printf("Request failed: %v", response.Error)
//	}
func (r JSONRPCResponse[T, R]) IsError() bool {
	return r.Error != nil
}

// IsSuccess reports whether the response represents a success.
//
// This is a convenience method to quickly check if the response
// contains a result object.
//
// Returns:
//   - true if the response is a success response.
//   - false if the response is an error response.
//
// Example:
//
//	if response.IsSuccess() {
//	    log.Printf("Request succeeded: %v", response.Result)
//	}
func (r JSONRPCResponse[T, R]) IsSuccess() bool {
	return r.Error == nil && r.Result != nil
}
