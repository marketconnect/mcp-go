package protocol

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

const JSONRPCVersion = "2.0"

type JSONRPCRequest[T IDConstraint] struct {
	// A String specifying the version of the JSON-RPC protocol. MUST be exactly "2.0".
	JSONRPC string `json:"jsonrpc"`
	// A String containing the name of the method to be invoked.
	// Method names that begin with the word rpc followed by a period character (U+002E or ASCII 46) are reserved for rpc-internal methods
	// and extensions and MUST NOT be used for anything else.
	Method string `json:"method"`
	// A Structured value that holds the parameter values to be used during the invocation of the method.
	// This member MAY be omitted.
	Params interface{} `json:"params,omitempty"` // optional parameters

	ID IDType[T] `json:"id"`
}

// NewRequest creates a new JSON-RPC request with the given method name, raw ID, and parameters.
//
// This is the primary constructor for creating requests.
// It accepts a raw ID value (string or integer) and automatically wraps it in an IDType.
//
// Example:
//
//	req := protocol.NewRequest("sum", 123, map[string]interface{}{"a": 1, "b": 2})
//
// For guaranteed unique IDs, you can use NextIntID or NextStringID:
//
//	req := protocol.NewRequest("sum", protocol.NextIntID().Value, params)
//
// If you already have an IDType, consider using NewRequestWithID.
//
// See also: NewRequestWithID for advanced use cases.
func NewRequest[T IDConstraint](method string, id T, params interface{}) JSONRPCRequest[T] {
	return JSONRPCRequest[T]{
		JSONRPC: JSONRPCVersion,
		ID:      NewID(id),
		Method:  method,
		Params:  params,
	}
}

// NewRequestWithID creates a new JSON-RPC request with a pre-wrapped IDType.
//
// This constructor is useful when you already have an IDType[T],
// for example, when using custom ID generators or session managers.
//
// For most cases, prefer NewRequest, which accepts a raw ID value (string or int).
//
// Example:
//
//	id := protocol.NextIntID()
//	req := protocol.NewRequestWithID("sum", id, map[string]interface{}{"a": 1, "b": 2})
func NewRequestWithID[T IDConstraint](method string, id IDType[T], params interface{}) JSONRPCRequest[T] {
	return JSONRPCRequest[T]{
		JSONRPC: JSONRPCVersion,
		ID:      id,
		Method:  method,
		Params:  params,
	}
}

// Validate checks if the JSON-RPC request is valid.
func (r JSONRPCRequest[T]) Validate() error {
	if r.JSONRPC != JSONRPCVersion {
		return &ValidationError{Reason: fmt.Sprintf("invalid JSON-RPC version: expected %q, got %q", JSONRPCVersion, r.JSONRPC)}
	}

	if strings.TrimSpace(r.Method) == "" {
		return &ValidationError{Reason: "method name cannot be empty or whitespace"}
	}

	if r.ID.IsEmpty() {
		return &ValidationError{Reason: "id must not be empty"}
	}

	// https://www.jsonrpc.org/specification#request_object
	// Method names that begin with the word rpc followed by
	// a period character (U+002E or ASCII 46) are reserved
	// for rpc-internal methods and extensions and MUST NOT be used for anything else.
	if strings.HasPrefix(r.Method, "rpc.") {
		return &ValidationError{Reason: fmt.Sprintf("method names starting with 'rpc.' are reserved, got: %q", r.Method)}
	}

	return nil
}

// UnmarshalJSON implements the json.Unmarshaler interface.
//
// It parses the JSON-RPC request from JSON data and validates its correctness.
// If the input is invalid or violates the JSON-RPC specification, an error is returned.
//
// This method automatically calls Validate() after unmarshaling.
//
// Example:
//
//	var req protocol.JSONRPCRequest[int]
//	err := json.Unmarshal(data, &req)
//	if err != nil {
//	    log.Fatal(err)
//	}
func (r *JSONRPCRequest[T]) UnmarshalJSON(data []byte) error {
	if len(data) == 0 {
		return errors.New("empty JSON data")
	}

	type Alias JSONRPCRequest[T]
	aux := &struct {
		Alias
	}{}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	temp := JSONRPCRequest[T](aux.Alias)
	if err := temp.Validate(); err != nil {
		return err
	}
	*r = temp
	return nil

}

// GetID returns the raw ID value of the request.
//
// Useful when you need to access the request ID for logging, tracking, or matching responses.
//
// Example:
//
//	req := protocol.NewRequest("sum", 123, params)
//	id := req.GetID()
//	fmt.Println(id) // Output: 123
func (r JSONRPCRequest[T]) GetID() T {
	return r.ID.Value
}
